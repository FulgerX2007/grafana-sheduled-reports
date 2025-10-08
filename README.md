<div align="center">
  <img src="src/img/logo.png" alt="Scheduled Reports Logo" width="200"/>

  # Scheduled Reports

  A Grafana app plugin for scheduling and emailing dashboard reports as PDF or HTML files.
</div>

## Features

- 📅 **Scheduled Reports**: Create recurring reports with cron expressions or simple presets (daily, weekly, monthly)
- 📊 **Dashboard Rendering**: Render full dashboards or specific panels to PDF or HTML
- 📧 **Email Delivery**: Send reports via email with customizable subjects and bodies
- 🔄 **Run History**: Track all report executions with status, duration, and downloadable artifacts
- ⚙️ **Flexible Configuration**: Configure SMTP, renderer settings, and usage limits per organization
- 🎨 **Template Support**: Customize PDF layouts with headers, footers, and logos
- 🔒 **Multi-tenancy**: Full organization isolation for all data and settings

## Prerequisites

- Grafana 10.3 or higher (for managed service accounts)
- Chromium/Chrome browser (for native PDF rendering)
- Go 1.21+ (for building)
- Node.js 18+ (for building)

> **Note**: The plugin now uses **native rendering with Chromium** by default. No external image-renderer service is required!

## Quick Start

### 1. Clone and Build

```bash
git clone https://github.com/yourusername/sheduled-reports-app.git
cd sheduled-reports-app
make install
make build
```

### 2. Configure Environment Variables (Optional)

Copy the example environment file and configure SMTP settings if needed:

```bash
cp .env.example .env
```

Edit `.env` and set (optional):
- `GF_SMTP_HOST`, `GF_SMTP_USER`, `GF_SMTP_PASSWORD`: Your SMTP settings
- `GF_SMTP_FROM_ADDRESS`, `GF_SMTP_FROM_NAME`: Email sender details

### 3. Start with Docker Compose

```bash
docker compose up -d
```

This will start:
- Grafana on http://localhost:3000 (admin/admin)
- Grafana Image Renderer on http://localhost:8081

### 4. Enable Plugin

In Grafana:
1. Go to Administration → Plugins
2. Find "Scheduled Reports" in the list
3. Click "Enable"

The plugin automatically uses Grafana's managed service accounts for authentication (Grafana 10.3+). No manual token configuration is required.

## Development

### Building

```bash
# Install dependencies
make install

# Build both frontend and backend
make build

# Build only backend
make build-backend

# Build only frontend
make build-frontend
```

### Running in Development Mode

```bash
# Start with file watching
make dev
```

### Testing

```bash
# Run all tests
make test

# Run Go tests only
go test -v ./...

# Run frontend tests only
npm test
```

## Configuration

### Environment Variables

Create a `.env` file based on `.env.example`:

```bash
# Grafana Configuration
GF_GRAFANA_URL=http://localhost:3000

# Plugin Data Path (where SQLite DB and artifacts are stored)
GF_PLUGIN_APP_DATA_PATH=./data

# SMTP Configuration (optional if using Grafana's SMTP)
GF_SMTP_HOST=smtp.gmail.com:587
GF_SMTP_USER=your-email@gmail.com
GF_SMTP_PASSWORD=your-app-password
GF_SMTP_FROM_ADDRESS=noreply@example.com
GF_SMTP_FROM_NAME=Grafana Reports
```

**Authentication**: The plugin uses Grafana's managed service accounts (Grafana 10.3+). No manual token configuration required.

### Plugin Settings

After enabling the plugin, configure it in the Settings page:

**SMTP Configuration**
- Toggle between Grafana's SMTP or custom SMTP settings
- Configure host, port, credentials, and TLS settings

**Renderer Configuration**
- Renderer URL (default: http://renderer:8081/render)
- Timeout and delay settings for heavy dashboards
- Viewport dimensions

**Limits**
- Max recipients per email
- Max attachment size
- Max concurrent renders
- Artifact retention days

## Usage

### Creating a Schedule

1. Navigate to the plugin: Apps → Reporting
2. Click "New Schedule"
3. Configure:
   - **Name**: Descriptive name for the schedule
   - **Dashboard**: Select dashboard to render
     - *Note: When you select a dashboard, its template variables are automatically loaded*
   - **Format**: PDF or HTML
   - **Time Range**: Dashboard time range (e.g., "now-7d" to "now")
   - **Schedule**: Daily, Weekly, Monthly, or Custom cron
   - **Variables**: Dashboard variable values (auto-populated from selected dashboard)
   - **Recipients**: Email addresses (To, CC, BCC)
   - **Subject & Body**: Email template with placeholders

4. Click "Create"

### Template Variables

Use these placeholders in email subject and body:

- `{{schedule.name}}` - Schedule name
- `{{dashboard.title}}` - Dashboard title
- `{{timerange}}` - Time range
- `{{run.started_at}}` - Execution start time

### Running Reports Manually

- Click the ▶️ icon next to any schedule to run it immediately
- View execution status in the Run History

### Viewing Run History

- Click the 🕐 icon next to any schedule
- See all executions with status, duration, and errors
- Download generated PDFs/HTMLs

## Architecture

### Frontend (React + TypeScript)
- `src/pages/` - Page components (Schedules, Settings, etc.)
- `src/components/` - Reusable UI components
- `src/types/` - TypeScript type definitions

### Backend (Go)
- `pkg/api/` - HTTP API handlers
- `pkg/cron/` - Scheduler and job execution
- `pkg/render/` - Dashboard rendering client
- `pkg/pdf/` - PDF generation from images
- `pkg/mail/` - SMTP email sender
- `pkg/store/` - SQLite database operations
- `pkg/model/` - Data models

### Data Storage
- SQLite database in `$GF_PLUGIN_APP_DATA_PATH/reporting.db`
- Artifacts in `$GF_PLUGIN_APP_DATA_PATH/artifacts/`

## Troubleshooting

### Rendering Fails

**Problem**: Dashboard rendering returns errors

**Solutions**:
- Ensure grafana-image-renderer is running and accessible
- Verify the managed service account has proper dashboard permissions
- Increase render timeout in Settings
- Add render delay for heavy dashboards

### Email Not Sending

**Problem**: Emails aren't being delivered

**Solutions**:
- Verify SMTP configuration in Settings
- Check SMTP credentials and firewall rules
- Test with a simple SMTP service first
- Check Grafana logs for email errors

### Schedule Not Running

**Problem**: Scheduled reports don't execute

**Solutions**:
- Verify schedule is enabled
- Check cron expression is valid
- Review backend logs for scheduler errors
- Ensure `next_run_at` is in the future

### Permission Denied

**Problem**: Cannot access certain dashboards

**Solutions**:
- The managed service account automatically has permissions defined in plugin.json
- Ensure dashboards are not in restricted folders
- Check organization ID matches

## Production Deployment

### Building for Production

```bash
make build
make sign  # If you have a plugin signing key
make dist  # Creates distribution zip
```

### Installation from Release

Download the appropriate release archive for your platform from GitHub releases and extract it:

```bash
unzip sheduled-reports-app-X.Y.Z.linux-amd64.zip -d /var/lib/grafana/plugins/
systemctl restart grafana-server
```

Configure Grafana to allow unsigned plugins:
```ini
[plugins]
allow_loading_unsigned_plugins = sheduled-reports-app
```

### Local Development Installation

For local development, you need to ensure only the correct binary is present:

```bash
# Build the plugin
make build

# Copy to Grafana plugins directory (ensure gpx_reporting exists, not gpx_reporting_linux_amd64)
mkdir -p /var/lib/grafana/plugins/sheduled-reports-app
rsync -av --exclude='gpx_reporting_*' dist/ /var/lib/grafana/plugins/sheduled-reports-app/

# Restart Grafana
systemctl restart grafana-server
```

### Recommended Settings

- Enable managed service accounts feature in Grafana 10.3+
- Configure rate limits to prevent abuse
- Set up artifact retention to manage disk space
- Use TLS for SMTP connections
- Monitor plugin logs and metrics

## API Reference

### Schedules

```bash
# List schedules
GET /api/plugins/sheduled-reports-app/resources/api/schedules

# Create schedule
POST /api/plugins/sheduled-reports-app/resources/api/schedules

# Get schedule
GET /api/plugins/sheduled-reports-app/resources/api/schedules/{id}

# Update schedule
PUT /api/plugins/sheduled-reports-app/resources/api/schedules/{id}

# Delete schedule
DELETE /api/plugins/sheduled-reports-app/resources/api/schedules/{id}

# Run now
POST /api/plugins/sheduled-reports-app/resources/api/schedules/{id}/run

# Get runs
GET /api/plugins/sheduled-reports-app/resources/api/schedules/{id}/runs
```

### Settings

```bash
# Get settings
GET /api/plugins/sheduled-reports-app/resources/api/settings

# Update settings
POST /api/plugins/sheduled-reports-app/resources/api/settings
```

### Artifacts

```bash
# Download artifact
GET /api/plugins/sheduled-reports-app/resources/api/runs/{id}/artifact
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

Apache 2.0

## Support

- Documentation: See [CLAUDE.md](./CLAUDE.md) for development guidance
- Issues: https://github.com/yourusername/sheduled-reports-app/issues
- Discussions: https://github.com/yourusername/sheduled-reports-app/discussions
