package render

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/yourusername/sheduled-reports-app/pkg/model"
)

// Renderer handles dashboard rendering using Chromium
type Renderer struct {
	grafanaURL string
	config     model.RendererConfig
	browser    *rod.Browser
}

// NewRenderer creates a new renderer instance with embedded Chromium
func NewRenderer(grafanaURL string, config model.RendererConfig) *Renderer {
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
		config.DeviceScaleFactor = 2.0
	}
	// Enable headless by default
	if !config.Headless {
		config.Headless = true
	}

	return &Renderer{
		grafanaURL: grafanaURL,
		config:     config,
		browser:    nil, // Lazy initialization
	}
}

// getBrowser initializes or returns existing browser instance
func (r *Renderer) getBrowser() (*rod.Browser, error) {
	if r.browser != nil {
		return r.browser, nil
	}

	// Configure launcher
	l := launcher.New()

	// Set custom Chromium path if provided
	if r.config.ChromiumPath != "" {
		l = l.Bin(r.config.ChromiumPath)
	}

	// Configure browser flags
	if r.config.Headless {
		l = l.Headless(true)
	}
	if r.config.DisableGPU {
		l = l.Set("disable-gpu")
	}
	if r.config.NoSandbox {
		l = l.Set("no-sandbox", "disable-setuid-sandbox")
	}

	// Skip TLS verification if configured
	if r.config.SkipTLSVerify {
		l = l.Set("ignore-certificate-errors")
		log.Printf("WARNING: TLS certificate verification disabled for renderer")
	}

	// Launch browser
	launchURL, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}

	browser := rod.New().ControlURL(launchURL)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to browser: %w", err)
	}

	r.browser = browser
	log.Printf("Chromium browser initialized successfully")
	return browser, nil
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

// RenderDashboard renders a dashboard to PNG using Chromium
func (r *Renderer) RenderDashboard(ctx context.Context, schedule *model.Schedule) ([]byte, error) {
	// Get service account token
	saToken, err := getServiceAccountToken(ctx)
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

	// Get or initialize browser
	browser, err := r.getBrowser()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize browser: %w", err)
	}

	// Create a new page
	page, err := browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}
	defer page.Close()

	// Set viewport size
	if err := page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:             r.config.ViewportWidth,
		Height:            r.config.ViewportHeight,
		DeviceScaleFactor: r.config.DeviceScaleFactor,
		Mobile:            false,
	}); err != nil {
		return nil, fmt.Errorf("failed to set viewport: %w", err)
	}

	// Set authentication header
	router := page.HijackRequests()
	router.MustAdd("*", func(ctx *rod.Hijack) {
		ctx.Request.Req().Header.Set("Authorization", "Bearer "+saToken)
		ctx.ContinueRequest(&proto.FetchContinueRequest{})
	})
	go router.Run()

	// Set timeout
	page = page.Timeout(time.Duration(r.config.TimeoutMS) * time.Millisecond)

	// Navigate to dashboard
	if err := page.Navigate(dashboardURL); err != nil {
		return nil, fmt.Errorf("failed to navigate to dashboard: %w", err)
	}

	// Wait for page to load
	if err := page.WaitLoad(); err != nil {
		return nil, fmt.Errorf("failed to wait for page load: %w", err)
	}

	// Additional delay for queries to finish (if configured)
	if r.config.DelayMS > 0 {
		time.Sleep(time.Duration(r.config.DelayMS) * time.Millisecond)
		log.Printf("DEBUG: Waited %dms for dashboard queries to complete", r.config.DelayMS)
	}

	// Take screenshot
	imageData, err := page.Screenshot(true, &proto.PageCaptureScreenshot{
		Format:  proto.PageCaptureScreenshotFormatPng,
		Quality: nil, // PNG doesn't use quality parameter
	})
	if err != nil {
		return nil, fmt.Errorf("failed to capture screenshot: %w", err)
	}

	// Verify it's a PNG by checking magic bytes
	if len(imageData) < 8 || string(imageData[1:4]) != "PNG" {
		return nil, fmt.Errorf("screenshot is not a PNG image (got %d bytes)", len(imageData))
	}

	log.Printf("DEBUG: Screenshot captured successfully (%d bytes)", len(imageData))
	return imageData, nil
}

// Close closes the browser instance
func (r *Renderer) Close() error {
	if r.browser != nil {
		log.Printf("Closing Chromium browser")
		return r.browser.Close()
	}
	return nil
}

// buildDashboardURL constructs the Grafana dashboard URL
func (r *Renderer) buildDashboardURL(schedule *model.Schedule) (string, error) {
	// Use configured grafanaURL
	baseURL := r.grafanaURL

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	// Ensure we're using the container hostname, not localhost
	if u.Host == "localhost:3000" || u.Host == "127.0.0.1:3000" {
		u.Host = "grafana:3000"
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
