# Installation Guide

## Native Renderer Installation

The plugin now includes a native renderer that generates HTML and PDF reports without requiring the `grafana-image-renderer` service.

### Rendering Modes

The plugin supports two rendering modes:

1. **Native Mode** (Default) - No external dependencies for HTML, lightweight for PDF
2. **Image Renderer Mode** (Legacy) - Requires `grafana-image-renderer` service

## Dependencies

### For HTML Export Only

**No additional dependencies required!**

The native renderer generates self-contained HTML with:
- Embedded Chart.js from CDN
- Direct data fetching from Grafana API
- All visualization types supported

### For PDF Export (Optional)

If you want to export PDF files, you need to install `wkhtmltopdf`:

#### Ubuntu/Debian
```bash
sudo apt-get update
sudo apt-get install -y wkhtmltopdf
```

#### Alpine Linux (Docker)
```bash
apk add --no-cache wkhtmltopdf
```

#### RHEL/CentOS/Fedora
```bash
sudo yum install -y wkhtmltopdf
# or
sudo dnf install -y wkhtmltopdf
```

#### macOS
```bash
brew install wkhtmltopdf
```

#### Windows
Download from: https://wkhtmltopdf.org/downloads.html

### Verify Installation

```bash
wkhtmltopdf --version
```

You should see output like:
```
wkhtmltopdf 0.12.6
```

## Docker Installation

### HTML Only (Minimal)

```dockerfile
FROM grafana/grafana:latest

# Copy plugin files
COPY dist/ /var/lib/grafana/plugins/grafana-app-reporting/

# No additional dependencies needed!
```

### HTML + PDF Support

```dockerfile
FROM grafana/grafana:latest

# Install wkhtmltopdf
USER root
RUN apt-get update && \
    apt-get install -y wkhtmltopdf && \
    rm -rf /var/lib/apt/lists/*

# Copy plugin files
COPY dist/ /var/lib/grafana/plugins/grafana-app-reporting/

USER grafana
```

### Full Docker Compose Example

```yaml
version: '3.8'

services:
  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    volumes:
      - ./dist:/var/lib/grafana/plugins/grafana-app-reporting
      - grafana-data:/var/lib/grafana
    environment:
      - GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=grafana-app-reporting
      - GF_INSTALL_PLUGINS=grafana-app-reporting
    # Install wkhtmltopdf for PDF support
    entrypoint: >
      sh -c "
      apt-get update &&
      apt-get install -y wkhtmltopdf &&
      /run.sh
      "

volumes:
  grafana-data:
```

## Configuration

### Settings in Plugin UI

Navigate to **Plugin Settings** in Grafana:

```json
{
  "renderer_config": {
    "mode": "native",
    "timeout_ms": 60000,
    "delay_ms": 1000,
    "viewport_width": 1920,
    "viewport_height": 1080
  }
}
```

### Renderer Modes

#### Native Mode (Default)
```json
{
  "renderer_config": {
    "mode": "native"
  }
}
```

**Pros:**
- No external service dependency
- Direct access to Grafana data
- Lighter infrastructure
- Supports HTML export without any dependencies

**Cons:**
- Requires wkhtmltopdf for PDF export

#### Image Renderer Mode (Legacy)
```json
{
  "renderer_config": {
    "mode": "image-renderer",
    "url": "http://renderer:8081/render"
  }
}
```

**Pros:**
- Pixel-perfect rendering
- No local dependencies

**Cons:**
- Requires grafana-image-renderer service
- Additional infrastructure complexity

## Deployment Scenarios

### Scenario 1: HTML-Only Deployment (Simplest)

✅ **No dependencies**
- Deploy plugin as-is
- Only HTML format available in schedules
- Perfect for lightweight deployments

### Scenario 2: HTML + PDF (Recommended)

✅ **Install wkhtmltopdf**
- Full functionality
- Both HTML and PDF formats available
- Minimal overhead (~50MB for wkhtmltopdf)

### Scenario 3: Legacy with grafana-image-renderer

✅ **Deploy grafana-image-renderer service**
- Set mode to "image-renderer"
- Keep existing infrastructure
- Backward compatible

## Troubleshooting

### PDF Generation Fails

**Error**: "failed to create PDF generator"

**Solution**: Install wkhtmltopdf
```bash
# Check if installed
which wkhtmltopdf

# Install if missing
sudo apt-get install -y wkhtmltopdf
```

### Charts Not Rendering in PDF

**Error**: Blank or missing charts in PDF

**Solution**: Increase `JavascriptDelay` (already set to 3000ms by default)

### Permission Errors

**Error**: "failed to write temp HTML file"

**Solution**: Ensure Grafana has write access to `/tmp`
```bash
sudo chown -R grafana:grafana /tmp
```

### wkhtmltopdf Not Found in Docker

**Error**: "exec: wkhtmltopdf: executable file not found in $PATH"

**Solution**: Add installation to Dockerfile:
```dockerfile
USER root
RUN apt-get update && apt-get install -y wkhtmltopdf
USER grafana
```

## Performance Considerations

### HTML Export
- **Fast**: ~1-2 seconds per dashboard
- **Lightweight**: No external processes
- **Scalable**: CPU-bound only

### PDF Export with wkhtmltopdf
- **Medium**: ~3-5 seconds per dashboard
- **Moderate**: Spawns wkhtmltopdf process
- **JavaScript Delay**: 3 seconds for Chart.js rendering

### Legacy Image Renderer
- **Slow**: ~5-10 seconds per dashboard
- **Heavy**: Full Chromium browser process
- **Resource Intensive**: High memory usage

## Security Notes

1. **wkhtmltopdf** runs as subprocesses with Grafana user permissions
2. Temporary HTML files are created in `/tmp` and cleaned up immediately
3. Service account tokens are used for Grafana API authentication
4. No data is sent to external services (everything runs locally)

## Monitoring

Check plugin health:
```bash
# Verify wkhtmltopdf is available
wkhtmltopdf --version

# Check Grafana logs for plugin errors
tail -f /var/log/grafana/grafana.log | grep "reporting"

# Monitor temp file cleanup
ls -la /tmp/report-*.html
```

## Upgrade Path

### From grafana-image-renderer to Native

1. Update plugin to latest version
2. Install wkhtmltopdf (if PDF export needed)
3. Change renderer mode in settings:
   ```json
   {"renderer_config": {"mode": "native"}}
   ```
4. Test reports
5. Remove grafana-image-renderer service (optional)

### No Breaking Changes

Both rendering modes are supported simultaneously. You can switch between them without data loss.
