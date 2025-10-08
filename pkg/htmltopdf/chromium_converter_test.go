package htmltopdf

import (
	"os"
	"testing"
	"time"
)

func TestChromiumConverter_Convert(t *testing.T) {
	// Check if Chromium is available
	if !IsChromiumAvailable() {
		t.Skip("Chromium not available, skipping test")
	}

	converter := NewChromiumConverter(30 * time.Second)

	htmlContent := []byte(`
<!DOCTYPE html>
<html>
<head>
	<title>Test Report</title>
	<style>
		body { font-family: Arial, sans-serif; padding: 20px; }
		h1 { color: #333; }
		p { color: #666; }
	</style>
</head>
<body>
	<h1>Test Report</h1>
	<p>This is a test PDF generated using Chromium.</p>
</body>
</html>
	`)

	pdfBytes, err := converter.Convert(htmlContent)
	if err != nil {
		t.Fatalf("Failed to convert HTML to PDF: %v", err)
	}

	if len(pdfBytes) == 0 {
		t.Fatal("PDF bytes are empty")
	}

	// Verify PDF signature
	if len(pdfBytes) < 4 || string(pdfBytes[:4]) != "%PDF" {
		t.Fatal("Output does not appear to be a valid PDF")
	}

	t.Logf("Successfully converted HTML to PDF: %d bytes", len(pdfBytes))
}

func TestChromiumConverter_ConvertWithOptions_Landscape(t *testing.T) {
	// Check if Chromium is available
	if !IsChromiumAvailable() {
		t.Skip("Chromium not available, skipping test")
	}

	converter := NewChromiumConverter(30 * time.Second)

	htmlContent := []byte(`
<!DOCTYPE html>
<html>
<head>
	<title>Landscape Report</title>
</head>
<body>
	<h1>Landscape Report</h1>
	<p>This page should be in landscape orientation.</p>
</body>
</html>
	`)

	pdfBytes, err := converter.ConvertWithOptions(htmlContent, true, "A4")
	if err != nil {
		t.Fatalf("Failed to convert HTML to PDF with landscape: %v", err)
	}

	if len(pdfBytes) == 0 {
		t.Fatal("PDF bytes are empty")
	}

	// Verify PDF signature
	if len(pdfBytes) < 4 || string(pdfBytes[:4]) != "%PDF" {
		t.Fatal("Output does not appear to be a valid PDF")
	}

	t.Logf("Successfully converted HTML to PDF (landscape): %d bytes", len(pdfBytes))
}

func TestChromiumConverter_ConvertWithOptions_DifferentSizes(t *testing.T) {
	// Check if Chromium is available
	if !IsChromiumAvailable() {
		t.Skip("Chromium not available, skipping test")
	}

	converter := NewChromiumConverter(30 * time.Second)

	htmlContent := []byte(`
<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body><h1>Paper Size Test</h1></body>
</html>
	`)

	testCases := []struct {
		name      string
		landscape bool
		paperSize string
	}{
		{"A4 Portrait", false, "A4"},
		{"A4 Landscape", true, "A4"},
		{"Letter Portrait", false, "Letter"},
		{"Letter Landscape", true, "Letter"},
		{"A3 Portrait", false, "A3"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pdfBytes, err := converter.ConvertWithOptions(htmlContent, tc.landscape, tc.paperSize)
			if err != nil {
				t.Fatalf("Failed to convert with options %s: %v", tc.name, err)
			}

			if len(pdfBytes) == 0 {
				t.Fatalf("PDF bytes are empty for %s", tc.name)
			}

			if len(pdfBytes) < 4 || string(pdfBytes[:4]) != "%PDF" {
				t.Fatalf("Output does not appear to be a valid PDF for %s", tc.name)
			}

			t.Logf("Successfully converted %s: %d bytes", tc.name, len(pdfBytes))
		})
	}
}

func TestChromiumConverter_ConvertWithChartJS(t *testing.T) {
	// Check if Chromium is available
	if !IsChromiumAvailable() {
		t.Skip("Chromium not available, skipping test")
	}

	converter := NewChromiumConverter(30 * time.Second)

	htmlContent := []byte(`
<!DOCTYPE html>
<html>
<head>
	<title>Chart Test</title>
	<script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.0/dist/chart.umd.min.js"></script>
	<style>
		body { font-family: Arial, sans-serif; padding: 20px; }
		.chart-container { width: 600px; height: 400px; }
	</style>
</head>
<body>
	<h1>Chart.js Test</h1>
	<div class="chart-container">
		<canvas id="myChart"></canvas>
	</div>
	<script>
		const ctx = document.getElementById('myChart');
		new Chart(ctx, {
			type: 'bar',
			data: {
				labels: ['Red', 'Blue', 'Yellow', 'Green', 'Purple', 'Orange'],
				datasets: [{
					label: 'Test Data',
					data: [12, 19, 3, 5, 2, 3],
					borderWidth: 1
				}]
			},
			options: {
				scales: {
					y: {
						beginAtZero: true
					}
				}
			}
		});
	</script>
</body>
</html>
	`)

	pdfBytes, err := converter.Convert(htmlContent)
	if err != nil {
		t.Fatalf("Failed to convert HTML with Chart.js to PDF: %v", err)
	}

	if len(pdfBytes) == 0 {
		t.Fatal("PDF bytes are empty")
	}

	// Verify PDF signature
	if len(pdfBytes) < 4 || string(pdfBytes[:4]) != "%PDF" {
		t.Fatal("Output does not appear to be a valid PDF")
	}

	t.Logf("Successfully converted HTML with Chart.js to PDF: %d bytes", len(pdfBytes))
}

func TestChromiumConverter_ConvertWithCustomDelay(t *testing.T) {
	// Check if Chromium is available
	if !IsChromiumAvailable() {
		t.Skip("Chromium not available, skipping test")
	}

	converter := NewChromiumConverter(30 * time.Second)

	htmlContent := []byte(`
<!DOCTYPE html>
<html>
<head><title>Delay Test</title></head>
<body>
	<h1>Content loaded after delay</h1>
	<div id="delayed-content"></div>
	<script>
		setTimeout(function() {
			document.getElementById('delayed-content').innerHTML = '<p>This content appeared after 2 seconds</p>';
		}, 2000);
	</script>
</body>
</html>
	`)

	// Use custom delay of 3 seconds to ensure delayed content is captured
	pdfBytes, err := converter.ConvertWithCustomDelay(htmlContent, 3000, true, "A4")
	if err != nil {
		t.Fatalf("Failed to convert HTML with custom delay: %v", err)
	}

	if len(pdfBytes) == 0 {
		t.Fatal("PDF bytes are empty")
	}

	if len(pdfBytes) < 4 || string(pdfBytes[:4]) != "%PDF" {
		t.Fatal("Output does not appear to be a valid PDF")
	}

	t.Logf("Successfully converted HTML with custom delay: %d bytes", len(pdfBytes))
}

func TestIsChromiumAvailable(t *testing.T) {
	available := IsChromiumAvailable()
	t.Logf("Chromium available: %v", available)

	if !available {
		t.Log("Warning: Chromium is not available. Some tests will be skipped.")
	}
}

// Benchmark tests
func BenchmarkChromiumConverter_Convert(b *testing.B) {
	if !IsChromiumAvailable() {
		b.Skip("Chromium not available, skipping benchmark")
	}

	converter := NewChromiumConverter(30 * time.Second)

	htmlContent := []byte(`
<!DOCTYPE html>
<html>
<head><title>Benchmark</title></head>
<body><h1>Benchmark Test</h1><p>Simple HTML content</p></body>
</html>
	`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := converter.Convert(htmlContent)
		if err != nil {
			b.Fatalf("Conversion failed: %v", err)
		}
	}
}

// Example test that saves output to file (for manual inspection)
func TestChromiumConverter_SaveToFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping file output test in short mode")
	}

	if !IsChromiumAvailable() {
		t.Skip("Chromium not available, skipping test")
	}

	converter := NewChromiumConverter(30 * time.Second)

	htmlContent := []byte(`
<!DOCTYPE html>
<html>
<head>
	<title>Sample Report</title>
	<style>
		body { font-family: Arial, sans-serif; padding: 40px; }
		h1 { color: #2c3e50; border-bottom: 3px solid #3498db; padding-bottom: 10px; }
		.section { margin: 20px 0; }
		.highlight { background-color: #f1c40f; padding: 2px 5px; }
	</style>
</head>
<body>
	<h1>Sample Grafana Report</h1>
	<div class="section">
		<h2>Dashboard: System Metrics</h2>
		<p>Generated: 2025-10-07</p>
		<p>Time Range: Last 24 hours</p>
	</div>
	<div class="section">
		<h3>Key Findings</h3>
		<ul>
			<li>CPU usage peaked at <span class="highlight">85%</span></li>
			<li>Memory consumption stable at 60%</li>
			<li>Network throughput averaged 1.2 Gbps</li>
		</ul>
	</div>
</body>
</html>
	`)

	pdfBytes, err := converter.Convert(htmlContent)
	if err != nil {
		t.Fatalf("Failed to convert HTML to PDF: %v", err)
	}

	// Save to temporary file
	tmpFile := "/tmp/chromium_test_output.pdf"
	err = os.WriteFile(tmpFile, pdfBytes, 0644)
	if err != nil {
		t.Fatalf("Failed to write PDF to file: %v", err)
	}

	t.Logf("PDF saved to %s (%d bytes)", tmpFile, len(pdfBytes))
}
