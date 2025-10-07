package grafana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/yourusername/sheduled-reports-app/pkg/model"
)

// Client is a Grafana API client
type Client struct {
	baseURL string
	client  *http.Client
}

// NewClient creates a new Grafana API client
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// getServiceAccountToken retrieves the service account token from context or environment
func getServiceAccountToken(ctx context.Context) (string, error) {
	// Try to get token from Grafana config (for managed service accounts in Grafana 10.3+)
	cfg := backend.GrafanaConfigFromContext(ctx)
	if cfg != nil {
		token, err := cfg.PluginAppClientSecret()
		if err == nil && token != "" {
			return token, nil
		}
		if err != nil {
			log.Printf("Warning: Failed to get managed service account token: %v", err)
		}
	}

	// Fallback to environment variable for backwards compatibility
	token := os.Getenv("GF_PLUGIN_SA_TOKEN")
	if token == "" {
		return "", fmt.Errorf("no service account token available: neither managed service account nor GF_PLUGIN_SA_TOKEN environment variable is set")
	}

	return token, nil
}

// Dashboard represents a Grafana dashboard
type Dashboard struct {
	UID       string          `json:"uid"`
	Title     string          `json:"title"`
	Panels    []Panel         `json:"panels"`
	Time      TimeSettings    `json:"time"`
	Timezone  string          `json:"timezone"`
	Templating TemplateSettings `json:"templating"`
}

// Panel represents a dashboard panel
type Panel struct {
	ID          int64            `json:"id"`
	Type        string           `json:"type"`
	Title       string           `json:"title"`
	Datasource  *Datasource      `json:"datasource"`
	Targets     []Target         `json:"targets"`
	Options     json.RawMessage  `json:"options"`
	FieldConfig json.RawMessage  `json:"fieldConfig"`
	GridPos     GridPos          `json:"gridPos"`
}

// Datasource represents a panel datasource
type Datasource struct {
	Type string `json:"type"`
	UID  string `json:"uid"`
}

// Target represents a panel query target
type Target struct {
	RefID      string          `json:"refId"`
	Datasource *Datasource     `json:"datasource"`
	Expr       string          `json:"expr"`
	RawSQL     string          `json:"rawSql"`
	Format     string          `json:"format"`
	Hide       bool            `json:"hide"`
}

// GridPos represents panel grid position
type GridPos struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

// TimeSettings represents dashboard time settings
type TimeSettings struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// TemplateSettings represents dashboard template variables
type TemplateSettings struct {
	List []TemplateVar `json:"list"`
}

// TemplateVar represents a template variable
type TemplateVar struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Current struct {
		Value interface{} `json:"value"`
		Text  string      `json:"text"`
	} `json:"current"`
}

// QueryResult represents the result of a panel query
type QueryResult struct {
	Frames []Frame `json:"frames"`
}

// Frame represents a data frame
type Frame struct {
	Name   string   `json:"name"`
	Fields []Field  `json:"fields"`
}

// Field represents a data field
type Field struct {
	Name   string        `json:"name"`
	Type   string        `json:"type"`
	Values []interface{} `json:"values"`
	Labels map[string]string `json:"labels"`
}

// GetDashboard fetches a dashboard by UID
func (c *Client) GetDashboard(ctx context.Context, uid string, orgID int64) (*Dashboard, error) {
	token, err := getServiceAccountToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}

	// Replace localhost with grafana for Docker network
	if u.Host == "localhost:3000" || u.Host == "127.0.0.1:3000" {
		u.Host = "grafana:3000"
	}

	u.Path = fmt.Sprintf("/api/dashboards/uid/%s", uid)

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Grafana-Org-Id", strconv.FormatInt(orgID, 10))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get dashboard: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Dashboard Dashboard `json:"dashboard"`
		Meta      struct {
			IsStarred bool   `json:"isStarred"`
			FolderID  int64  `json:"folderId"`
		} `json:"meta"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Dashboard, nil
}

// QueryPanelData executes a panel's queries and returns the results
func (c *Client) QueryPanelData(ctx context.Context, schedule *model.Schedule, panel Panel) ([]QueryResult, error) {
	token, err := getServiceAccountToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}

	// Replace localhost with grafana for Docker network
	if u.Host == "localhost:3000" || u.Host == "127.0.0.1:3000" {
		u.Host = "grafana:3000"
	}

	u.Path = "/api/ds/query"

	var results []QueryResult

	for _, target := range panel.Targets {
		if target.Hide {
			continue
		}

		// Build query request
		queryReq := map[string]interface{}{
			"from": schedule.RangeFrom,
			"to":   schedule.RangeTo,
			"queries": []map[string]interface{}{
				{
					"refId":      target.RefID,
					"datasource": target.Datasource,
					"expr":       target.Expr,
					"rawSql":     target.RawSQL,
					"format":     target.Format,
				},
			},
		}

		// Add dashboard variables
		if len(schedule.Variables) > 0 {
			scopedVars := make(map[string]interface{})
			for k, v := range schedule.Variables {
				scopedVars[k] = map[string]interface{}{
					"text":  v,
					"value": v,
				}
			}
			queryReq["scopedVars"] = scopedVars
		}

		reqBody, err := json.Marshal(queryReq)
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequestWithContext(ctx, "POST", u.String(), bytes.NewReader(reqBody))
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Grafana-Org-Id", strconv.FormatInt(schedule.OrgID, 10))

		resp, err := c.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("query failed: status %d, body: %s", resp.StatusCode, string(body))
		}

		var result struct {
			Results map[string]QueryResult `json:"results"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}

		if qr, ok := result.Results[target.RefID]; ok {
			results = append(results, qr)
		}
	}

	return results, nil
}
