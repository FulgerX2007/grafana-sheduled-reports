# Docker Setup Guide

This guide explains how to run the Grafana Scheduled Reports plugin with Docker Compose using native rendering with Chromium.

## Quick Start

1. **Build the plugin:**
   ```bash
   # Build frontend
   npm install
   npm run build

   # Build backend
   env CGO_ENABLED=1 go build -o dist/gpx_reporting ./cmd/backend
   ```

2. **Start the services:**
   ```bash
   docker-compose up -d
   ```

3. **Access Grafana:**
   - URL: http://localhost:3000
   - Default credentials: admin/admin
   - The plugin will be available at http://localhost:3000/a/sheduled-reports-app

## Architecture

The Docker Compose setup includes:

- **Grafana** (with Chromium): Main service with the reporting plugin
  - Native rendering mode (no grafana-image-renderer needed)
  - Chromium installed automatically on first run
  - Plugin provisioned with default configuration

## What's Included

### Docker Compose Services

**grafana**
- Image: `grafana/grafana:11.0.0`
- Ports: `3000:3000`
- Features:
  - Native HTML/PDF rendering with Chromium
  - Managed service accounts enabled
  - Plugin auto-provisioned with optimal defaults
  - SMTP configuration via environment variables

### Automatic Chromium Installation

The `docker-entrypoint.sh` script automatically:
1. Installs Chromium and required dependencies
2. Verifies Chromium is available
3. Starts Grafana with the grafana user

**Dependencies installed:**
- chromium
- chromium-driver
- fonts-liberation
- libnss3
- libatk-bridge2.0-0
- libxcomposite1
- libxdamage1
- libxrandr2
- libgbm1
- libasound2

## Configuration

### Environment Variables

Configure via environment variables or `.env` file:

```bash
# SMTP Configuration (optional)
GF_SMTP_HOST=smtp.gmail.com:587
GF_SMTP_USER=your-email@gmail.com
GF_SMTP_PASSWORD=your-app-password
GF_SMTP_FROM_ADDRESS=noreply@example.com
GF_SMTP_FROM_NAME=Grafana Reports

# Chromium Configuration
CHROME_BIN=/usr/bin/chromium-browser
CHROMEDP_DISABLE_LOGGING=1
```

### Plugin Configuration

The plugin is provisioned with default settings in `provisioning/plugins.yaml`:

```yaml
renderer_config:
  mode: native           # Native rendering (no image-renderer)
  pdf_engine: chromium   # Use Chromium for PDF generation
  timeout_ms: 30000      # 30 second timeout
  delay_ms: 3000         # 3 second delay for JS execution
```

## Building for Production

### Create Production Image

Create a `Dockerfile` for production deployment:

```dockerfile
FROM grafana/grafana:11.0.0

USER root

# Install Chromium and dependencies
RUN apt-get update && \
    apt-get install -y \
        chromium \
        chromium-driver \
        fonts-liberation \
        libnss3 \
        libatk-bridge2.0-0 \
        libxcomposite1 \
        libxdamage1 \
        libxrandr2 \
        libgbm1 \
        libasound2 && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Copy plugin
COPY dist /var/lib/grafana/plugins/sheduled-reports-app
COPY provisioning /etc/grafana/provisioning

# Chromium environment
ENV CHROME_BIN=/usr/bin/chromium-browser
ENV CHROMEDP_DISABLE_LOGGING=1

USER grafana
```

Build and run:

```bash
docker build -t grafana-with-reports:latest .
docker run -d -p 3000:3000 grafana-with-reports:latest
```

## Development Workflow

### Hot Reload Development

For frontend development with hot reload:

```bash
# Terminal 1: Run Grafana with docker-compose
docker-compose up

# Terminal 2: Watch and rebuild frontend
npm run dev

# Terminal 3: Watch and rebuild backend (if needed)
go build -o dist/gpx_reporting ./cmd/backend && docker-compose restart grafana
```

### Viewing Logs

```bash
# All services
docker-compose logs -f

# Just Grafana
docker-compose logs -f grafana

# Check Chromium installation
docker-compose exec grafana chromium-browser --version
```

## Troubleshooting

### Chromium Not Found

**Symptom**: PDF generation fails with "Chromium not available"

**Solution**:
```bash
# Check if Chromium is installed
docker-compose exec grafana which chromium-browser

# If not, rebuild the container
docker-compose down
docker-compose up -d --force-recreate
```

### Permission Issues

**Symptom**: "Permission denied" errors

**Solution**:
The container runs as root during entrypoint to install Chromium, then switches to grafana user. If you see permission errors:

```bash
# Check ownership of plugin files
docker-compose exec grafana ls -la /var/lib/grafana/plugins/sheduled-reports-app

# Fix if needed
docker-compose exec grafana chown -R grafana:grafana /var/lib/grafana/plugins/sheduled-reports-app
```

### Memory Issues

**Symptom**: Container crashes or OOM errors

**Solution**:
Chromium requires more memory. Increase Docker limits:

```yaml
services:
  grafana:
    deploy:
      resources:
        limits:
          memory: 2G
        reservations:
          memory: 1G
```

### Plugin Not Loading

**Symptom**: Plugin doesn't appear in Grafana

**Solution**:
```bash
# Verify plugin is mounted correctly
docker-compose exec grafana ls -la /var/lib/grafana/plugins/sheduled-reports-app

# Check plugin.json
docker-compose exec grafana cat /var/lib/grafana/plugins/sheduled-reports-app/plugin.json

# Check Grafana logs
docker-compose logs grafana | grep -i plugin
```

## Cleanup

Remove all containers and volumes:

```bash
docker-compose down -v
```

## Performance Tuning

### Chromium Flags

For better performance in containerized environments, add to `docker-entrypoint.sh`:

```bash
export CHROME_FLAGS="--no-sandbox --disable-dev-shm-usage --disable-gpu"
```

### Concurrent Rendering Limit

Adjust in `provisioning/plugins.yaml`:

```yaml
limits:
  max_concurrent_renders: 3  # Reduce for lower memory systems
```

## Advanced Configuration

### Using External Database

Replace SQLite with PostgreSQL:

```yaml
services:
  grafana:
    environment:
      GF_DATABASE_TYPE: postgres
      GF_DATABASE_HOST: postgres:5432
      GF_DATABASE_NAME: grafana
      GF_DATABASE_USER: grafana
      GF_DATABASE_PASSWORD: grafana

  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: grafana
      POSTGRES_USER: grafana
      POSTGRES_PASSWORD: grafana
    volumes:
      - postgres-data:/var/lib/postgresql/data

volumes:
  postgres-data:
```

### Custom Fonts

Add custom fonts for PDF rendering:

```dockerfile
COPY fonts/* /usr/share/fonts/truetype/custom/
RUN fc-cache -f -v
```

### SSL/TLS

Enable HTTPS with certificates:

```yaml
services:
  grafana:
    environment:
      GF_SERVER_PROTOCOL: https
      GF_SERVER_CERT_FILE: /etc/grafana/ssl/cert.pem
      GF_SERVER_CERT_KEY: /etc/grafana/ssl/key.pem
    volumes:
      - ./ssl:/etc/grafana/ssl:ro
```

## Monitoring

### Health Check

Add health check to docker-compose.yml:

```yaml
services:
  grafana:
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3000/api/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

### Metrics

Expose Grafana metrics:

```yaml
environment:
  GF_METRICS_ENABLED: true
  GF_METRICS_BASIC_AUTH_USERNAME: metrics
  GF_METRICS_BASIC_AUTH_PASSWORD: secret
```

Access metrics at: http://localhost:3000/metrics

## Security

### Non-Root User

The entrypoint script switches to grafana user after installing Chromium. For stricter security:

1. Pre-build an image with Chromium
2. Run container as non-root
3. Use read-only root filesystem where possible

### Secrets Management

Use Docker secrets for sensitive data:

```yaml
services:
  grafana:
    secrets:
      - smtp_password
    environment:
      GF_SMTP_PASSWORD__FILE: /run/secrets/smtp_password

secrets:
  smtp_password:
    file: ./secrets/smtp_password.txt
```

## References

- [Grafana Docker Documentation](https://grafana.com/docs/grafana/latest/setup-grafana/installation/docker/)
- [Chromium in Docker](https://github.com/chromedp/chromedp#running-in-docker)
- [Plugin Development Guide](https://grafana.com/docs/grafana/latest/developers/plugins/)
