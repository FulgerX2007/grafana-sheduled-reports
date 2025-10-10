package render

import (
	"context"

	"github.com/yourusername/sheduled-reports-app/pkg/model"
)

// Backend defines the interface for different rendering backends
type Backend interface {
	// RenderDashboard renders a Grafana dashboard to an image
	RenderDashboard(ctx context.Context, schedule *model.Schedule) ([]byte, error)

	// Close cleans up resources used by the backend
	Close() error

	// Name returns the name of the backend (e.g., "chromium", "wkhtmltopdf")
	Name() string
}

// BackendType represents the type of rendering backend
type BackendType string

const (
	BackendChromium    BackendType = "chromium"
	BackendWkhtmltopdf BackendType = "wkhtmltopdf"
)

// NewBackend creates a new rendering backend based on the specified type
func NewBackend(backendType BackendType, grafanaURL string, config model.RendererConfig) (Backend, error) {
	switch backendType {
	case BackendChromium:
		return NewChromiumRenderer(grafanaURL, config), nil
	case BackendWkhtmltopdf:
		return NewWkhtmltopdfRenderer(grafanaURL, config), nil
	default:
		return NewChromiumRenderer(grafanaURL, config), nil // Default to Chromium
	}
}
