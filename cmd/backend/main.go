package main

import (
	"fmt"
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
		// Fallback: use executable directory + data
		execPath, err := os.Executable()
		if err == nil {
			dataPath = filepath.Join(filepath.Dir(execPath), "data")
		} else {
			dataPath = "./data"
		}
	}

	log.Printf("Using data path: %s", dataPath)

	// Ensure data directory exists
	if err := os.MkdirAll(dataPath, 0755); err != nil {
		log.Printf("Failed to create data directory %s: %v", dataPath, err)
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Initialize database
	dbPath := filepath.Join(dataPath, "reporting.db")
	log.Printf("Initializing database at: %s", dbPath)
	st, err := store.NewStore(dbPath)
	if err != nil {
		log.Printf("Failed to initialize database: %v", err)
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer st.Close()

	// Get configuration from environment
	grafanaURL := os.Getenv("GF_GRAFANA_URL")
	if grafanaURL == "" {
		grafanaURL = "http://localhost:3000"
	}

	log.Println("Using Grafana-managed service account for authentication")
	log.Println("Token will be retrieved automatically from plugin context")

	// Create artifacts directory
	artifactsPath := filepath.Join(dataPath, "artifacts")
	log.Printf("Creating artifacts directory: %s", artifactsPath)
	if err := os.MkdirAll(artifactsPath, 0755); err != nil {
		log.Printf("Failed to create artifacts directory: %v", err)
		return fmt.Errorf("failed to create artifacts directory: %w", err)
	}

	// Initialize scheduler (token will be retrieved from context on first API call)
	maxConcurrent := 5 // Default max concurrent renders
	log.Printf("Initializing scheduler (max concurrent: %d)", maxConcurrent)
	scheduler := cron.NewScheduler(st, grafanaURL, artifactsPath, maxConcurrent)

	// Start scheduler
	log.Println("Starting scheduler...")
	if err := scheduler.Start(); err != nil {
		log.Printf("Failed to start scheduler: %v", err)
		return fmt.Errorf("failed to start scheduler: %w", err)
	}
	defer scheduler.Stop()
	log.Println("Scheduler started")

	// Create API handler
	log.Println("Creating API handler...")
	handler := api.NewHandler(st, scheduler)

	// Serve plugin
	log.Println("Starting plugin server...")
	return backend.Serve(backend.ServeOpts{
		CallResourceHandler: handler,
	})
}
