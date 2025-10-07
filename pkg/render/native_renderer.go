package render

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yourusername/sheduled-reports-app/pkg/grafana"
	"github.com/yourusername/sheduled-reports-app/pkg/htmlgen"
	"github.com/yourusername/sheduled-reports-app/pkg/model"
)

// NativeRenderer renders dashboards without grafana-image-renderer
type NativeRenderer struct {
	grafanaClient *grafana.Client
	htmlGenerator *htmlgen.Generator
	config        model.RendererConfig
}

// NewNativeRenderer creates a new native renderer
func NewNativeRenderer(grafanaURL string, config model.RendererConfig) *NativeRenderer {
	timeout := time.Duration(config.TimeoutMS) * time.Millisecond
	return &NativeRenderer{
		grafanaClient: grafana.NewClient(grafanaURL, timeout),
		htmlGenerator: htmlgen.NewGenerator(),
		config:        config,
	}
}

// RenderDashboardHTML renders a dashboard to HTML
func (r *NativeRenderer) RenderDashboardHTML(ctx context.Context, schedule *model.Schedule) ([]byte, error) {
	// Add delay before rendering to let queries finish (if configured)
	if r.config.DelayMS > 0 {
		time.Sleep(time.Duration(r.config.DelayMS) * time.Millisecond)
	}

	// Fetch dashboard definition
	dashboard, err := r.grafanaClient.GetDashboard(ctx, schedule.DashboardUID, schedule.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard: %w", err)
	}

	log.Printf("Fetched dashboard: %s (UID: %s) with %d panels", dashboard.Title, dashboard.UID, len(dashboard.Panels))

	// Update schedule with dashboard title if not set
	if schedule.DashboardTitle == "" {
		schedule.DashboardTitle = dashboard.Title
	}

	// Fetch data for each panel
	var panelsData []htmlgen.PanelData

	for _, panel := range dashboard.Panels {
		// Skip panels without targets or row panels
		if len(panel.Targets) == 0 || panel.Type == "row" {
			continue
		}

		log.Printf("Querying panel: %s (ID: %d, Type: %s)", panel.Title, panel.ID, panel.Type)

		// Query panel data
		results, err := r.grafanaClient.QueryPanelData(ctx, schedule, panel)
		if err != nil {
			log.Printf("Warning: Failed to query panel %d: %v", panel.ID, err)
			// Continue with empty results instead of failing the entire report
			results = []grafana.QueryResult{}
		}

		panelsData = append(panelsData, htmlgen.PanelData{
			Panel:   panel,
			Results: results,
		})
	}

	// Generate HTML
	htmlBytes, err := r.htmlGenerator.Generate(dashboard, panelsData, schedule)
	if err != nil {
		return nil, fmt.Errorf("failed to generate HTML: %w", err)
	}

	log.Printf("Generated HTML report: %d bytes", len(htmlBytes))

	return htmlBytes, nil
}
