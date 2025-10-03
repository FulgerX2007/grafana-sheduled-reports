package render

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/yourusername/grafana-app-reporting/pkg/model"
)

// Renderer handles dashboard rendering
type Renderer struct {
	grafanaURL  string
	rendererURL string
	saToken     string
	config      model.RendererConfig
	client      *http.Client
}

// NewRenderer creates a new renderer instance
func NewRenderer(grafanaURL, saToken string, config model.RendererConfig) *Renderer {
	return &Renderer{
		grafanaURL:  grafanaURL,
		rendererURL: config.URL,
		saToken:     saToken,
		config:      config,
		client: &http.Client{
			Timeout: time.Duration(config.TimeoutMS) * time.Millisecond,
		},
	}
}

// RenderDashboard renders a dashboard to PNG
func (r *Renderer) RenderDashboard(ctx context.Context, schedule *model.Schedule) ([]byte, error) {
	// Use Grafana's render API instead of calling renderer directly
	renderURL, err := r.buildGrafanaRenderURL(schedule)
	if err != nil {
		return nil, fmt.Errorf("failed to build render URL: %w", err)
	}

	fmt.Printf("DEBUG: Grafana Render URL: %s\n", renderURL)
	fmt.Printf("DEBUG: Has token: %v (length: %d)\n", r.saToken != "", len(r.saToken))

	// Add delay before rendering to let queries finish
	if r.config.DelayMS > 0 {
		time.Sleep(time.Duration(r.config.DelayMS) * time.Millisecond)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", renderURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add auth token for Grafana
	if r.saToken != "" {
		req.Header.Set("Authorization", "Bearer "+r.saToken)
		fmt.Printf("DEBUG: Added Authorization header\n")
	} else {
		fmt.Printf("DEBUG: No token available\n")
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("render failed with status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Verify it's a PNG by checking magic bytes
	if len(data) < 8 || string(data[1:4]) != "PNG" {
		return nil, fmt.Errorf("response is not a PNG image (got %d bytes, content type: %s)", len(data), resp.Header.Get("Content-Type"))
	}

	return data, nil
}

// buildGrafanaRenderURL constructs the Grafana render API URL
func (r *Renderer) buildGrafanaRenderURL(schedule *model.Schedule) (string, error) {
	u, err := url.Parse(r.grafanaURL)
	if err != nil {
		return "", err
	}

	// Ensure we're using the container hostname, not localhost
	if u.Host == "localhost:3000" || u.Host == "127.0.0.1:3000" {
		u.Host = "grafana:3000"
	}

	// Use Grafana's render API endpoint for full dashboard
	u.Path = fmt.Sprintf("/render/d/%s", schedule.DashboardUID)

	q := u.Query()
	q.Set("from", schedule.RangeFrom)
	q.Set("to", schedule.RangeTo)
	q.Set("orgId", strconv.FormatInt(schedule.OrgID, 10))
	q.Set("tz", schedule.Timezone)
	q.Set("width", strconv.Itoa(r.config.ViewportWidth))
	q.Set("height", strconv.Itoa(r.config.ViewportHeight))
	q.Set("theme", "light")
	q.Set("kiosk", "true") // Hide menu, header, and time picker

	// Add dashboard variables
	for k, v := range schedule.Variables {
		q.Set("var-"+k, v)
	}

	u.RawQuery = q.Encode()

	return u.String(), nil
}

// buildDashboardURL constructs the Grafana dashboard URL
func (r *Renderer) buildDashboardURL(schedule *model.Schedule) (string, error) {
	u, err := url.Parse(r.grafanaURL)
	if err != nil {
		return "", err
	}

	// Ensure we're using the container hostname, not localhost
	// Replace localhost with grafana for Docker network
	if u.Host == "localhost:3000" || u.Host == "127.0.0.1:3000" {
		u.Host = "grafana:3000"
	}

	u.Path = fmt.Sprintf("/d/%s", schedule.DashboardUID)

	q := u.Query()
	q.Set("from", schedule.RangeFrom)
	q.Set("to", schedule.RangeTo)
	q.Set("kiosk", "tv")
	q.Set("orgId", strconv.FormatInt(schedule.OrgID, 10))
	q.Set("tz", schedule.Timezone)

	// Add dashboard variables
	for k, v := range schedule.Variables {
		q.Set("var-"+k, v)
	}

	u.RawQuery = q.Encode()

	return u.String(), nil
}

// buildRendererURL constructs the renderer service URL
func (r *Renderer) buildRendererURL(dashboardURL string) (string, error) {
	u, err := url.Parse(r.rendererURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("url", dashboardURL)
	q.Set("width", strconv.Itoa(r.config.ViewportWidth))
	q.Set("height", strconv.Itoa(r.config.ViewportHeight))
	q.Set("deviceScaleFactor", "1")
	q.Set("timeout", strconv.Itoa(r.config.TimeoutMS/1000)) // Convert to seconds

	// Pass authorization token for renderer to use when accessing Grafana
	if r.saToken != "" {
		q.Set("token", r.saToken)
	}

	u.RawQuery = q.Encode()

	return u.String(), nil
}
