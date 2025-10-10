package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/yourusername/sheduled-reports-app/pkg/api"
	"github.com/yourusername/sheduled-reports-app/pkg/cron"
	"github.com/yourusername/sheduled-reports-app/pkg/store"
)

func main() {
	if err := backend.Serve(backend.ServeOpts{
		CallResourceHandler: newHandler(),
	}); err != nil {
		log.DefaultLogger.Error("Plugin serve error", "error", err)
		os.Exit(1)
	}
}

func newHandler() backend.CallResourceHandler {
	handler, err := createHandler()
	if err != nil {
		log.DefaultLogger.Error("Failed to create handler", "error", err)
		// Return a dummy handler that will report the error
		return &errorHandler{err: err}
	}
	return handler
}

type errorHandler struct {
	err error
}

func (h *errorHandler) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	return h.err
}

func createHandler() (backend.CallResourceHandler, error) {
	// Get plugin data path from environment or use temp directory
	dataPath := os.Getenv("GF_PLUGIN_APP_DATA_PATH")
	if dataPath == "" {
		// Use system temp directory as fallback
		tmpDir := os.TempDir()
		dataPath = filepath.Join(tmpDir, "grafana-reporting-plugin")
	}

	// Ensure data directory exists
	if err := os.MkdirAll(dataPath, 0755); err != nil {
		return nil, err
	}

	// Initialize database
	dbPath := filepath.Join(dataPath, "reporting.db")
	st, err := store.NewStore(dbPath)
	if err != nil {
		return nil, err
	}

	// Build Grafana URL from Grafana's configuration environment variables
	grafanaURL := buildGrafanaURL()

	// Create artifacts directory
	artifactsPath := filepath.Join(dataPath, "artifacts")
	if err := os.MkdirAll(artifactsPath, 0755); err != nil {
		return nil, err
	}

	// Initialize scheduler (token will be retrieved from context on first API call)
	maxConcurrent := 5 // Default max concurrent renders
	scheduler := cron.NewScheduler(st, grafanaURL, artifactsPath, maxConcurrent)

	// Start scheduler
	if err := scheduler.Start(); err != nil {
		return nil, err
	}

	// Create API handler
	handler := api.NewHandler(st, scheduler)

	return handler, nil
}

// buildGrafanaURL constructs the Grafana URL from environment variables
func buildGrafanaURL() string {
	// Try to get from explicit override first
	if url := os.Getenv("GF_GRAFANA_URL"); url != "" {
		return strings.TrimSuffix(url, "/")
	}

	// Build from Grafana server configuration
	protocol := os.Getenv("GF_SERVER_PROTOCOL")
	if protocol == "" {
		protocol = "http"
	}

	domain := os.Getenv("GF_SERVER_DOMAIN")
	httpAddr := os.Getenv("GF_SERVER_HTTP_ADDR")
	if domain != "" && domain != "localhost" {
		// Use domain if explicitly set
		httpAddr = domain
	} else if httpAddr == "" || httpAddr == "0.0.0.0" || httpAddr == "::" {
		// Use localhost if addr is not set or is wildcard
		httpAddr = "127.0.0.1"
	}

	httpPort := os.Getenv("GF_SERVER_HTTP_PORT")
	if httpPort == "" {
		httpPort = "3000"
	}

	// Build base URL (without root_url path)
	url := fmt.Sprintf("%s://%s:%s", protocol, httpAddr, httpPort)

	log.DefaultLogger.Info("Built Grafana URL", "url", url, "protocol", protocol, "addr", httpAddr, "port", httpPort)

	return url
}
