package htmltopdf

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// ChromiumConverter converts HTML to PDF using headless Chromium
type ChromiumConverter struct {
	timeout time.Duration
}

// NewChromiumConverter creates a new Chromium-based HTML-to-PDF converter
func NewChromiumConverter(timeout time.Duration) *ChromiumConverter {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &ChromiumConverter{
		timeout: timeout,
	}
}

// createChromiumContext creates a chromedp context with appropriate flags for Docker/containerized environments
func (c *ChromiumConverter) createChromiumContext(ctx context.Context) (context.Context, context.CancelFunc) {
	// Default Chrome flags
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-breakpad", true),
		chromedp.Flag("disable-client-side-phishing-detection", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-features", "site-per-process,Translate,BlinkGenPropertyTrees"),
		chromedp.Flag("disable-hang-monitor", true),
		chromedp.Flag("disable-ipc-flooding-protection", true),
		chromedp.Flag("disable-popup-blocking", true),
		chromedp.Flag("disable-prompt-on-repost", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("force-color-profile", "srgb"),
		chromedp.Flag("metrics-recording-only", true),
		chromedp.Flag("safebrowsing-disable-auto-update", true),
		chromedp.Flag("enable-automation", true),
		chromedp.Flag("password-store", "basic"),
		chromedp.Flag("use-mock-keychain", true),
	)

	// Add no-sandbox flag if running in Docker or if explicitly requested
	if os.Getenv("CHROMEDP_NO_SANDBOX") == "1" {
		opts = append(opts,
			chromedp.Flag("no-sandbox", true),
			chromedp.Flag("disable-setuid-sandbox", true),
		)
		log.Println("Chromium running with --no-sandbox flag (Docker mode)")
	}

	// Set Chrome binary path if specified
	if chromeBin := os.Getenv("CHROME_BIN"); chromeBin != "" {
		opts = append(opts, chromedp.ExecPath(chromeBin))
		log.Printf("Using Chrome binary: %s", chromeBin)
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)

	// Create Chrome instance context
	chromiumCtx, chromiumCancel := chromedp.NewContext(allocCtx)

	// Return a cancel function that cancels both contexts
	cancel := func() {
		chromiumCancel()
		allocCancel()
	}

	return chromiumCtx, cancel
}

// Convert converts HTML to PDF using Chromium
func (c *ChromiumConverter) Convert(htmlContent []byte) ([]byte, error) {
	return c.ConvertWithOptions(htmlContent, true, "A4")
}

// ConvertWithOptions converts HTML to PDF with custom page options
func (c *ChromiumConverter) ConvertWithOptions(htmlContent []byte, landscape bool, paperSize string) ([]byte, error) {
	// Create chromium context with proper flags
	ctx, cancel := c.createChromiumContext(context.Background())
	defer cancel()

	// Set overall timeout
	ctx, timeoutCancel := context.WithTimeout(ctx, c.timeout)
	defer timeoutCancel()

	var pdfBuffer []byte

	// Convert paper size to dimensions (in inches)
	width, height := c.getPaperDimensions(paperSize, landscape)

	// Set print parameters
	printParams := page.PrintToPDF().
		WithPrintBackground(true).
		WithPaperWidth(width).
		WithPaperHeight(height).
		WithMarginTop(0.4).      // ~10mm
		WithMarginBottom(0.4).   // ~10mm
		WithMarginLeft(0.4).     // ~10mm
		WithMarginRight(0.4).    // ~10mm
		WithScale(1.0).
		WithPreferCSSPageSize(false)

	if landscape {
		printParams = printParams.WithLandscape(true)
	}

	// Create data URL from HTML content
	dataURL := "data:text/html;charset=utf-8," + string(htmlContent)

	// Execute chromedp tasks
	err := chromedp.Run(ctx,
		// Navigate to the data URL
		chromedp.Navigate(dataURL),

		// Wait for page to be fully loaded
		chromedp.WaitReady("body"),

		// Wait for Chart.js to render all charts (if any)
		chromedp.Sleep(3*time.Second),

		// Wait for all canvas elements to be ready
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Additional wait to ensure all dynamic content is rendered
			time.Sleep(1 * time.Second)
			return nil
		}),

		// Print to PDF
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuffer, _, err = printParams.Do(ctx)
			if err != nil {
				return fmt.Errorf("failed to print PDF: %w", err)
			}
			return nil
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("chromium conversion failed: %w", err)
	}

	log.Printf("Converted HTML to PDF using Chromium: %d bytes", len(pdfBuffer))

	return pdfBuffer, nil
}

// getPaperDimensions returns paper dimensions in inches
func (c *ChromiumConverter) getPaperDimensions(paperSize string, landscape bool) (width, height float64) {
	// Dimensions in inches
	var w, h float64

	switch paperSize {
	case "Letter":
		w, h = 8.5, 11
	case "A4":
		w, h = 8.27, 11.69 // 210mm x 297mm
	case "A3":
		w, h = 11.69, 16.54 // 297mm x 420mm
	default:
		w, h = 8.27, 11.69 // Default to A4
	}

	if landscape {
		// Swap width and height for landscape
		return h, w
	}

	return w, h
}

// IsChromiumAvailable checks if Chromium/Chrome is available
// This creates a temporary context to verify Chrome can be launched
func IsChromiumAvailable() bool {
	// Create a context with a short timeout
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Try to perform a simple operation
	err := chromedp.Run(ctx,
		chromedp.Navigate("about:blank"),
	)

	return err == nil
}

// ConvertWithWaitForSelector converts HTML to PDF and waits for a specific selector
// This is useful when you need to ensure specific elements are rendered before capturing
func (c *ChromiumConverter) ConvertWithWaitForSelector(htmlContent []byte, selector string, landscape bool, paperSize string) ([]byte, error) {
	// Create chromium context with proper flags
	ctx, cancel := c.createChromiumContext(context.Background())
	defer cancel()

	// Set overall timeout
	ctx, timeoutCancel := context.WithTimeout(ctx, c.timeout)
	defer timeoutCancel()

	var pdfBuffer []byte

	// Convert paper size to dimensions
	width, height := c.getPaperDimensions(paperSize, landscape)

	// Set print parameters
	printParams := page.PrintToPDF().
		WithPrintBackground(true).
		WithPaperWidth(width).
		WithPaperHeight(height).
		WithMarginTop(0.4).
		WithMarginBottom(0.4).
		WithMarginLeft(0.4).
		WithMarginRight(0.4).
		WithScale(1.0).
		WithPreferCSSPageSize(false)

	if landscape {
		printParams = printParams.WithLandscape(true)
	}

	// Create data URL from HTML content
	dataURL := "data:text/html;charset=utf-8," + string(htmlContent)

	// Execute chromedp tasks
	err := chromedp.Run(ctx,
		// Navigate to the data URL
		chromedp.Navigate(dataURL),

		// Wait for specific selector
		chromedp.WaitVisible(selector, chromedp.ByQuery),

		// Additional wait for JavaScript execution
		chromedp.Sleep(2*time.Second),

		// Print to PDF
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuffer, _, err = printParams.Do(ctx)
			if err != nil {
				return fmt.Errorf("failed to print PDF: %w", err)
			}
			return nil
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("chromium conversion with selector failed: %w", err)
	}

	log.Printf("Converted HTML to PDF using Chromium with selector wait: %d bytes", len(pdfBuffer))

	return pdfBuffer, nil
}

// ConvertWithCustomDelay converts HTML to PDF with a custom JavaScript execution delay
func (c *ChromiumConverter) ConvertWithCustomDelay(htmlContent []byte, delayMS int, landscape bool, paperSize string) ([]byte, error) {
	// Create chromium context with proper flags
	ctx, cancel := c.createChromiumContext(context.Background())
	defer cancel()

	// Set overall timeout (add delay to base timeout)
	totalTimeout := c.timeout + time.Duration(delayMS)*time.Millisecond
	ctx, timeoutCancel := context.WithTimeout(ctx, totalTimeout)
	defer timeoutCancel()

	var pdfBuffer []byte

	// Convert paper size to dimensions
	width, height := c.getPaperDimensions(paperSize, landscape)

	// Set print parameters
	printParams := page.PrintToPDF().
		WithPrintBackground(true).
		WithPaperWidth(width).
		WithPaperHeight(height).
		WithMarginTop(0.4).
		WithMarginBottom(0.4).
		WithMarginLeft(0.4).
		WithMarginRight(0.4).
		WithScale(1.0).
		WithPreferCSSPageSize(false)

	if landscape {
		printParams = printParams.WithLandscape(true)
	}

	// Create data URL from HTML content
	dataURL := "data:text/html;charset=utf-8," + string(htmlContent)

	// Execute chromedp tasks
	err := chromedp.Run(ctx,
		// Navigate to the data URL
		chromedp.Navigate(dataURL),

		// Wait for page to be ready
		chromedp.WaitReady("body"),

		// Custom delay for JavaScript execution
		chromedp.Sleep(time.Duration(delayMS)*time.Millisecond),

		// Print to PDF
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuffer, _, err = printParams.Do(ctx)
			if err != nil {
				return fmt.Errorf("failed to print PDF: %w", err)
			}
			return nil
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("chromium conversion with custom delay failed: %w", err)
	}

	log.Printf("Converted HTML to PDF using Chromium with %dms delay: %d bytes", delayMS, len(pdfBuffer))

	return pdfBuffer, nil
}

// Helper function to convert mm to inches
func mmToInches(mm float64) float64 {
	return mm / 25.4
}

// Helper function to round to 2 decimal places
func round(val float64) float64 {
	return math.Round(val*100) / 100
}
