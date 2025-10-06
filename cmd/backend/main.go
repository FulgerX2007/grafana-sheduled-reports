package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/yourusername/sheduled-reports-app/pkg/api"
	"github.com/yourusername/sheduled-reports-app/pkg/cron"
	"github.com/yourusername/sheduled-reports-app/pkg/store"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Get plugin data path
	dataPath := os.Getenv("GF_PLUGIN_APP_DATA_PATH")
	if dataPath == "" {
		dataPath = "./data"
	}

	// Ensure data directory exists
	if err := os.MkdirAll(dataPath, 0755); err != nil {
		return err
	}

	// Initialize database
	dbPath := filepath.Join(dataPath, "reporting.db")
	st, err := store.NewStore(dbPath)
	if err != nil {
		return err
	}
	defer st.Close()

	// Get configuration from environment
	grafanaURL := os.Getenv("GF_GRAFANA_URL")
	if grafanaURL == "" {
		grafanaURL = "http://localhost:3000"
	}

	saToken := os.Getenv("GF_PLUGIN_SA_TOKEN")
	if saToken == "" {
		log.Println("Warning: GF_PLUGIN_SA_TOKEN not set. Rendering may fail.")
	}

	// Create artifacts directory
	artifactsPath := filepath.Join(dataPath, "artifacts")
	if err := os.MkdirAll(artifactsPath, 0755); err != nil {
		return err
	}

	// Initialize scheduler
	maxConcurrent := 5 // Default max concurrent renders
	scheduler := cron.NewScheduler(st, grafanaURL, saToken, artifactsPath, maxConcurrent)

	// Start scheduler
	if err := scheduler.Start(); err != nil {
		return err
	}
	defer scheduler.Stop()

	// Create API handler
	handler := api.NewHandler(st, scheduler)

	// Serve plugin
	return backend.Serve(backend.ServeOpts{
		CallResourceHandler: handler,
	})
}
