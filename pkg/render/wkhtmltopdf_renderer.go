package render

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"

	wkhtmltopdf "github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/yourusername/sheduled-reports-app/pkg/model"
)

// WkhtmltopdfRenderer handles dashboard rendering using wkhtmltopdf
type WkhtmltopdfRenderer struct {
	grafanaURL string
	config     model.RendererConfig
}

// NewWkhtmltopdfRenderer creates a new wkhtmltopdf renderer instance
func NewWkhtmltopdfRenderer(grafanaURL string, config model.RendererConfig) *WkhtmltopdfRenderer {
	// Set defaults
	if config.ViewportWidth == 0 {
		config.ViewportWidth = 1920
	}
	if config.ViewportHeight == 0 {
		config.ViewportHeight = 1080
	}
	if config.TimeoutMS == 0 {
		config.TimeoutMS = 30000
	}
	if config.DeviceScaleFactor == 0 {
		config.DeviceScaleFactor = 1.0 // wkhtmltopdf uses zoom instead
	}

	return &WkhtmltopdfRenderer{
		grafanaURL: grafanaURL,
		config:     config,
	}
}

// getServiceAccountToken retrieves the service account token from context or environment
func (r *WkhtmltopdfRenderer) getServiceAccountToken(ctx context.Context) (string, error) {
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

// RenderDashboard renders a dashboard to PNG using wkhtmltopdf
func (r *WkhtmltopdfRenderer) RenderDashboard(ctx context.Context, schedule *model.Schedule) ([]byte, error) {
	// Get service account token
	saToken, err := r.getServiceAccountToken(ctx)
	if err != nil {
		log.Printf("Warning: Failed to get service account token: %v", err)
		return nil, fmt.Errorf("no service account token available: %w", err)
	}

	// Build dashboard URL
	dashboardURL, err := r.buildDashboardURL(schedule)
	if err != nil {
		return nil, fmt.Errorf("failed to build dashboard URL: %w", err)
	}

	log.Printf("DEBUG: Dashboard URL: %s", dashboardURL)
	log.Printf("DEBUG: Using service account token (length: %d)", len(saToken))

	// Set binary path if configured
	if r.config.WkhtmltopdfPath != "" {
		wkhtmltopdf.SetPath(r.config.WkhtmltopdfPath)
	}

	// Create new PDF generator
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF generator: %w", err)
	}

	// Set global options
	pdfg.Dpi.Set(300) // High quality
	pdfg.NoCollate.Set(false)
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	pdfg.Orientation.Set(wkhtmltopdf.OrientationLandscape)
	pdfg.MarginTop.Set(10)
	pdfg.MarginBottom.Set(10)
	pdfg.MarginLeft.Set(10)
	pdfg.MarginRight.Set(10)

	// Create page from URL
	page := wkhtmltopdf.NewPage(dashboardURL)

	// Set page-specific options
	// Note: JavaScript is enabled by default in wkhtmltopdf
	page.NoStopSlowScripts.Set(true)
	page.LoadErrorHandling.Set("ignore")
	page.LoadMediaErrorHandling.Set("ignore")

	// Set zoom based on device scale factor
	if r.config.DeviceScaleFactor > 0 {
		page.Zoom.Set(r.config.DeviceScaleFactor)
	}

	// Add custom header with auth token (CustomHeader is a mapOption)
	page.CustomHeader.Set("Authorization", "Bearer "+saToken)

	// JavaScript delay to let queries finish
	if r.config.DelayMS > 0 {
		page.JavascriptDelay.Set(uint(r.config.DelayMS))
		log.Printf("DEBUG: Waiting %dms for dashboard queries to complete", r.config.DelayMS)
	} else {
		page.JavascriptDelay.Set(2000) // Default 2 second delay
	}

	// Add page to document
	pdfg.AddPage(page)

	// Note: Timeout and SSL verification are handled at command level in wkhtmltopdf
	// The library doesn't expose these as options in the current API

	// Generate PDF
	err = pdfg.Create()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	// Get PDF bytes
	pdfBytes := pdfg.Bytes()

	// wkhtmltopdf generates PDF directly, but we need PNG for consistency
	// For now, return the PDF bytes (the caller will handle conversion if needed)
	log.Printf("DEBUG: PDF generated successfully (%d bytes)", len(pdfBytes))

	// Note: wkhtmltopdf generates PDF, not PNG
	// We return PDF bytes here and let the caller decide if conversion is needed
	return pdfBytes, nil
}

// Close cleans up resources (wkhtmltopdf doesn't need cleanup)
func (r *WkhtmltopdfRenderer) Close() error {
	// No resources to clean up for wkhtmltopdf
	return nil
}

// Name returns the backend name
func (r *WkhtmltopdfRenderer) Name() string {
	return "wkhtmltopdf"
}

// buildDashboardURL constructs the Grafana dashboard URL
func (r *WkhtmltopdfRenderer) buildDashboardURL(schedule *model.Schedule) (string, error) {
	// Use configured grafanaURL
	baseURL := r.grafanaURL

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	// Only convert localhost to grafana hostname if explicitly configured to do so
	// This is needed for Docker deployments where the plugin runs in a separate container
	// For non-Docker deployments, use the actual configured hostname
	// Note: This conversion should only happen if GRAFANA_HOSTNAME env var is set
	if targetHost := os.Getenv("GRAFANA_HOSTNAME"); targetHost != "" {
		if u.Host == "localhost:3000" || u.Host == "127.0.0.1:3000" || u.Host == "localhost" || u.Host == "127.0.0.1" {
			// Parse target to preserve protocol
			if u.Port() != "" {
				u.Host = fmt.Sprintf("%s:%s", targetHost, u.Port())
			} else {
				u.Host = targetHost
			}
			log.Printf("DEBUG: Converted localhost to %s for Docker deployment", u.Host)
		}
	}

	// Preserve any subpath from base URL (e.g., /dna from root_url)
	basePath := u.Path
	if basePath == "" || basePath == "/" {
		basePath = ""
	}

	u.Path = fmt.Sprintf("%s/d/%s", basePath, schedule.DashboardUID)

	q := u.Query()
	q.Set("from", schedule.RangeFrom)
	q.Set("to", schedule.RangeTo)
	q.Set("kiosk", "tv") // Hide menu, header, and time picker
	q.Set("orgId", strconv.FormatInt(schedule.OrgID, 10))
	q.Set("tz", schedule.Timezone)

	// Add dashboard variables
	for k, v := range schedule.Variables {
		q.Set("var-"+k, v)
	}

	u.RawQuery = q.Encode()

	return u.String(), nil
}
