package render

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/sheduled-reports-app/pkg/htmltopdf"
	"github.com/yourusername/sheduled-reports-app/pkg/model"
	"github.com/yourusername/sheduled-reports-app/pkg/pdf"
)

// DashboardRenderer is an interface for rendering dashboards
type DashboardRenderer interface {
	// RenderToHTML renders a dashboard to HTML
	RenderToHTML(ctx context.Context, schedule *model.Schedule) ([]byte, error)

	// RenderToPDF renders a dashboard to PDF
	RenderToPDF(ctx context.Context, schedule *model.Schedule) ([]byte, error)
}

// ImageRenderer uses grafana-image-renderer (legacy mode)
type ImageRenderer struct {
	renderer *Renderer
}

// NewImageRenderer creates a new image-based renderer
func NewImageRenderer(grafanaURL string, config model.RendererConfig) *ImageRenderer {
	return &ImageRenderer{
		renderer: NewRenderer(grafanaURL, config),
	}
}

// RenderToHTML renders dashboard to HTML (wraps PNG in HTML)
func (r *ImageRenderer) RenderToHTML(ctx context.Context, schedule *model.Schedule) ([]byte, error) {
	// For image renderer, we render to PNG and return it as-is
	// The scheduler can decide how to handle it
	return r.renderer.RenderDashboard(ctx, schedule)
}

// RenderToPDF renders dashboard to PDF using gofpdf
func (r *ImageRenderer) RenderToPDF(ctx context.Context, schedule *model.Schedule) ([]byte, error) {
	imageData, err := r.renderer.RenderDashboard(ctx, schedule)
	if err != nil {
		return nil, err
	}

	// Use the existing PDF generator
	pdfGen := pdf.NewGenerator()
	return pdfGen.GenerateFromImages([][]byte{imageData}, schedule)
}

// NativeRendererAdapter adapts the native renderer to the interface
type NativeRendererAdapter struct {
	nativeRenderer *NativeRenderer
	pdfConverter   *htmltopdf.Converter
}

// NewNativeRendererAdapter creates a new native renderer adapter
func NewNativeRendererAdapter(grafanaURL string, config model.RendererConfig) *NativeRendererAdapter {
	timeout := time.Duration(config.TimeoutMS) * time.Millisecond
	return &NativeRendererAdapter{
		nativeRenderer: NewNativeRenderer(grafanaURL, config),
		pdfConverter:   htmltopdf.NewConverter(timeout),
	}
}

// RenderToHTML renders dashboard to HTML
func (r *NativeRendererAdapter) RenderToHTML(ctx context.Context, schedule *model.Schedule) ([]byte, error) {
	return r.nativeRenderer.RenderDashboardHTML(ctx, schedule)
}

// RenderToPDF renders dashboard to PDF
func (r *NativeRendererAdapter) RenderToPDF(ctx context.Context, schedule *model.Schedule) ([]byte, error) {
	htmlBytes, err := r.nativeRenderer.RenderDashboardHTML(ctx, schedule)
	if err != nil {
		return nil, err
	}

	return r.pdfConverter.Convert(htmlBytes)
}

// NewDashboardRenderer creates a dashboard renderer based on config
func NewDashboardRenderer(grafanaURL string, config model.RendererConfig) (DashboardRenderer, error) {
	mode := config.Mode
	if mode == "" {
		mode = "image-renderer" // Default to legacy mode for backward compatibility
	}

	switch mode {
	case "image-renderer":
		return NewImageRenderer(grafanaURL, config), nil
	case "native":
		return NewNativeRendererAdapter(grafanaURL, config), nil
	default:
		return nil, fmt.Errorf("unknown renderer mode: %s", mode)
	}
}
