package htmltopdf

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	wkhtmltopdf "github.com/SebastiaanKlippert/go-wkhtmltopdf"
)

// Converter converts HTML to PDF using wkhtmltopdf
type Converter struct {
	timeout time.Duration
}

// NewConverter creates a new HTML-to-PDF converter
func NewConverter(timeout time.Duration) *Converter {
	return &Converter{
		timeout: timeout,
	}
}

// Convert converts HTML to PDF
func (c *Converter) Convert(htmlContent []byte) ([]byte, error) {
	// Create a new PDF generator
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF generator: %w", err)
	}

	// Set global options
	pdfg.Dpi.Set(300)
	pdfg.NoCollate.Set(false)
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	pdfg.Orientation.Set(wkhtmltopdf.OrientationLandscape)
	pdfg.MarginTop.Set(10)
	pdfg.MarginBottom.Set(10)
	pdfg.MarginLeft.Set(10)
	pdfg.MarginRight.Set(10)

	// Create a temporary file for HTML content
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("report-%d.html", time.Now().UnixNano()))

	if err := os.WriteFile(tmpFile, htmlContent, 0644); err != nil {
		return nil, fmt.Errorf("failed to write temp HTML file: %w", err)
	}
	defer os.Remove(tmpFile)

	// Add the HTML file as a page
	page := wkhtmltopdf.NewPage(tmpFile)
	page.EnableLocalFileAccess.Set(true)
	page.PrintMediaType.Set(true)
	page.NoBackground.Set(false) // Include background colors/images
	page.LoadErrorHandling.Set("ignore")
	page.LoadMediaErrorHandling.Set("ignore")

	// Wait for JavaScript to execute (for Chart.js rendering)
	page.JavascriptDelay.Set(3000) // 3 seconds

	pdfg.AddPage(page)

	// Create PDF document
	err = pdfg.Create()
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF: %w", err)
	}

	// Get PDF bytes from buffer
	pdfBytes := pdfg.Bytes()
	log.Printf("Converted HTML to PDF: %d bytes", len(pdfBytes))

	return pdfBytes, nil
}

// ConvertWithOptions converts HTML to PDF with custom page options
func (c *Converter) ConvertWithOptions(htmlContent []byte, landscape bool, paperSize string) ([]byte, error) {
	// Create a new PDF generator
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF generator: %w", err)
	}

	// Set global options
	pdfg.Dpi.Set(300)
	pdfg.NoCollate.Set(false)

	// Set page size
	switch paperSize {
	case "Letter":
		pdfg.PageSize.Set(wkhtmltopdf.PageSizeLetter)
	case "A4":
		pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	case "A3":
		pdfg.PageSize.Set(wkhtmltopdf.PageSizeA3)
	default:
		pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	}

	// Set orientation
	if landscape {
		pdfg.Orientation.Set(wkhtmltopdf.OrientationLandscape)
	} else {
		pdfg.Orientation.Set(wkhtmltopdf.OrientationPortrait)
	}

	pdfg.MarginTop.Set(10)
	pdfg.MarginBottom.Set(10)
	pdfg.MarginLeft.Set(10)
	pdfg.MarginRight.Set(10)

	// Create a temporary file for HTML content
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("report-%d.html", time.Now().UnixNano()))

	if err := os.WriteFile(tmpFile, htmlContent, 0644); err != nil {
		return nil, fmt.Errorf("failed to write temp HTML file: %w", err)
	}
	defer os.Remove(tmpFile)

	// Add the HTML file as a page
	page := wkhtmltopdf.NewPage(tmpFile)
	page.EnableLocalFileAccess.Set(true)
	page.PrintMediaType.Set(true)
	page.NoBackground.Set(false)
	page.LoadErrorHandling.Set("ignore")
	page.LoadMediaErrorHandling.Set("ignore")

	// Wait for JavaScript to execute (for Chart.js rendering)
	page.JavascriptDelay.Set(3000) // 3 seconds

	pdfg.AddPage(page)

	// Create PDF document
	err = pdfg.Create()
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF: %w", err)
	}

	// Get PDF bytes from buffer
	pdfBytes := pdfg.Bytes()
	log.Printf("Converted HTML to PDF: %d bytes", len(pdfBytes))

	return pdfBytes, nil
}

// IsAvailable checks if wkhtmltopdf is installed on the system
func IsAvailable() bool {
	_, err := wkhtmltopdf.NewPDFGenerator()
	return err == nil
}
