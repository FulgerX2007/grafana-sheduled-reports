# Grafana Reporting Plugin

A Grafana app plugin for scheduling and emailing dashboard reports as PDF or HTML files.

## Features

- 📅 **Scheduled Reports**: Create recurring reports with cron expressions or simple presets (daily, weekly, monthly)
- 📊 **Dashboard Rendering**: Render full dashboards or specific panels to PDF or HTML
- 📧 **Email Delivery**: Send reports via email with customizable subjects and bodies
- 🔄 **Run History**: Track all report executions with status, duration, and downloadable artifacts
- ⚙️ **Flexible Configuration**: Configure SMTP, renderer settings, and usage limits per organization
- 🎨 **Template Support**: Customize PDF layouts with headers, footers, and logos
- 🔒 **Multi-tenancy**: Full organization isolation for all data and settings

## Prerequisites

- Grafana 10.0 or higher
- Grafana Image Renderer plugin or service
- Go 1.21+ (for building)
- Node.js 18+ (for building)

## Quick Start

### 1. Clone and Build

```bash
git clone https://github.com/yourusername/grafana-app-reporting.git
cd grafana-app-reporting
make install
make build
```

### 2. Configure Environment Variables

Copy the example environment file and configure your settings:

```bash
cp .env.example .env
```

Edit `.env` and set:
- `GF_PLUGIN_SA_TOKEN`: Your Grafana service account token (see step 3)
- `GF_SMTP_HOST`, `GF_SMTP_USER`, `GF_SMTP_PASSWORD`: Your SMTP settings (optional)
- `GF_SMTP_FROM_ADDRESS`, `GF_SMTP_FROM_NAME`: Email sender details (optional)

### 3. Start with Docker Compose

```bash
docker compose up -d
```

This will start:
- Grafana on http://localhost:3000 (admin/admin)
- Grafana Image Renderer on http://localhost:8081

### 4. Create Service Account

In Grafana:
1. Go to Administration → Service Accounts
2. Create a new service account named "reporting-plugin"
3. Set role to "Viewer" (or higher if you need access to restricted dashboards)
4. Generate a token and copy it
5. Add the token to your `.env` file: `GF_PLUGIN_SA_TOKEN=your-token-here`
6. Restart containers: `docker compose restart`

### 5. Enable Plugin

In Grafana:
1. Go to Administration → Plugins
2. Find "Reporting" in the list
3. Click "Enable"

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
GF_PLUGIN_SA_TOKEN=your-service-account-token

# Plugin Data Path (where SQLite DB and artifacts are stored)
GF_PLUGIN_APP_DATA_PATH=./data

# SMTP Configuration (optional if using Grafana's SMTP)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@example.com
SMTP_USE_TLS=true
```

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
   - **Format**: PDF or HTML
   - **Time Range**: Dashboard time range (e.g., "now-7d" to "now")
   - **Schedule**: Daily, Weekly, Monthly, or Custom cron
   - **Variables**: Dashboard variable values
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
- Check service account token has proper permissions
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
- Ensure service account has access to dashboard folders
- Upgrade service account role if needed
- Check organization ID matches

## Production Deployment

### Building for Production

```bash
make build
make sign  # If you have a plugin signing key
make dist  # Creates distribution zip
```

### Installation

1. Copy plugin to Grafana plugins directory:
   ```bash
   cp -r dist /var/lib/grafana/plugins/grafana-app-reporting
   ```

2. Restart Grafana:
   ```bash
   systemctl restart grafana-server
   ```

3. Allow unsigned plugin (if not signed):
   ```ini
   [plugins]
   allow_loading_unsigned_plugins = grafana-app-reporting
   ```

### Recommended Settings

- Use a dedicated service account with minimal permissions
- Configure rate limits to prevent abuse
- Set up artifact retention to manage disk space
- Use TLS for SMTP connections
- Monitor plugin logs and metrics

## API Reference

### Schedules

```bash
# List schedules
GET /api/plugins/grafana-app-reporting/resources/api/schedules

# Create schedule
POST /api/plugins/grafana-app-reporting/resources/api/schedules

# Get schedule
GET /api/plugins/grafana-app-reporting/resources/api/schedules/{id}

# Update schedule
PUT /api/plugins/grafana-app-reporting/resources/api/schedules/{id}

# Delete schedule
DELETE /api/plugins/grafana-app-reporting/resources/api/schedules/{id}

# Run now
POST /api/plugins/grafana-app-reporting/resources/api/schedules/{id}/run

# Get runs
GET /api/plugins/grafana-app-reporting/resources/api/schedules/{id}/runs
```

### Settings

```bash
# Get settings
GET /api/plugins/grafana-app-reporting/resources/api/settings

# Update settings
POST /api/plugins/grafana-app-reporting/resources/api/settings
```

### Artifacts

```bash
# Download artifact
GET /api/plugins/grafana-app-reporting/resources/api/runs/{id}/artifact
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
- Issues: https://github.com/yourusername/grafana-app-reporting/issues
- Discussions: https://github.com/yourusername/grafana-app-reporting/discussions
