# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Important Guidelines

**Git Commit Policy**:
- **NEVER** commit changes automatically
- **ONLY** create commits when explicitly requested by the user
- **DO NOT** add AI attribution or Claude AI references in commit messages
- Use semantic commit format when creating commits (feat:, fix:, refactor:, etc.)

## Repository Overview

This is a Grafana App Plugin for scheduled dashboard reporting in Grafana OSS. It provides scheduled rendering of dashboards to PDF/HTML, email delivery, run history, and audit tracking - all managed through a UI.

## Architecture

**Type**: Grafana App Plugin (frontend + backend)

**Core Components**:
- **Frontend**: React/TypeScript app using Grafana UI components and app routing
- **Backend**: Go plugin using Grafana Plugin SDK for HTTP routes, background workers, and secure storage
- **Renderer**: Integrates with grafana-image-renderer service via Grafana's `/render` endpoints
- **Email**: SMTP client supporting either Grafana's SMTP settings or plugin-specific configuration
- **Storage**: SQLite or BoltDB for schedules, runs, history, and templates (all scoped by OrgID)
- **Auth**: Uses Grafana service account tokens stored in secure plugin settings

## Project Structure

```
sheduled-reports-app/
├─ src/                      # Frontend (React/TypeScript)
│  ├─ pages/                 # App pages: Schedules, ScheduleEdit, Settings, Templates
│  ├─ components/            # UI components (DashboardPicker, CronEditor, etc.)
│  ├─ types/                 # TypeScript types
│  ├─ plugin.json            # Plugin manifest
│  └─ module.ts              # Plugin entry point
├─ pkg/                      # Backend Go packages
│  ├─ api/                   # HTTP handlers for schedules, runs, settings
│  ├─ cron/                  # Scheduler and worker pool
│  ├─ render/                # Grafana renderer client
│  ├─ pdf/                   # PDF assembly from rendered images
│  ├─ mail/                  # SMTP sender
│  ├─ store/                 # SQLite/BoltDB storage and migrations
│  ├─ model/                 # DTOs and validation
│  └─ auth/                  # Service account token and org/user context
├─ cmd/backend/              # Backend plugin main.go
├─ dist/                     # Build outputs
└─ provisioning/             # Optional app provisioning files
```

## Development Commands

### Frontend Development
```bash
npm install                  # Install dependencies
npm run dev                  # Development build with watch
npm run build                # Production build
npm run test                 # Run tests
```

### Backend Development
```bash
go mod download              # Download dependencies
go build -o dist/backend ./cmd/backend   # Build backend plugin
go test ./...                # Run all tests
go test -v ./pkg/cron        # Test specific package
```

### Plugin Development
```bash
# Run with Grafana (requires Docker)
docker-compose up -d         # Start Grafana with renderer

# Build both frontend and backend
npm run build && go build -o dist/backend ./cmd/backend

# Sign plugin (for production)
npx @grafana/sign-plugin
```

## Data Model

All tables include `org_id`, `created_at`, `updated_at` for multi-tenancy and auditing.

**Tables**:
- `schedules`: Report configurations (dashboard, time range, recipients, cron expression, variables)
- `runs`: Execution history (status, artifacts, duration, error details)
- `templates`: PDF/HTML/email layout templates
- `settings`: Per-org configuration (SMTP, renderer settings, limits)
- `audit`: Audit trail of all actions

## Backend HTTP API Routes

All routes under `/api/<app-id>/`:
- `GET/POST /schedules` - List and create schedules
- `PUT/DELETE /schedules/:id` - Update and delete schedules
- `POST /schedules/:id/run` - Trigger immediate execution
- `GET /schedules/:id/runs` - Get run history for a schedule
- `GET /runs/:id/artifact` - Download PDF/HTML artifact
- `GET/POST /settings` - Org-wide settings
- `GET/POST /templates` - Layout templates
- `GET /dashboards/search` - Proxy to Grafana API for dashboard picker

## Key Implementation Patterns

### Rendering Flow
1. Build render URL: `/render/d/<uid>?from=X&to=Y&var-foo=bar&kiosk=tv&orgId=N`
2. Add service account token in Authorization header
3. Respect `render_delay_ms` for heavy dashboards to let queries finish
4. Collect PNG images from rendered panels/dashboards
5. Assemble PDF with gofpdf or HTML with embedded images

### Scheduling
- Support cron expressions and presets (daily/weekly/monthly)
- Per-schedule timezone support
- Worker pool with backoff retry (3 attempts, exponential)
- Idempotency tokens to avoid duplicate sends on restart
- Global render concurrency limit and per-org quotas

### Security & Multi-tenancy
- Every query filtered by `org_id` from request context
- Service account tokens and SMTP secrets in secure plugin settings (never plaintext)
- Permission model: Viewer can view, Editor can manage schedules, Admin can modify settings
- Validate dashboard access via service account with appropriate folder permissions

### Email Delivery
- SMTP configuration: use Grafana's SMTP or plugin-specific settings
- Template support with placeholders: `{{schedule.name}}`, `{{timerange}}`, `{{dashboard.title}}`
- Attachment size limits with fallback to download links
- Rate limiting per minute to avoid abuse

## Frontend Architecture

**Routing**: `/schedules`, `/schedules/new`, `/schedules/:id`, `/runs/:id`, `/settings`, `/templates`

**Key Components**:
- **DashboardPicker**: Search dashboards by title with UID preview
- **VariableEditor**: Dynamic key/value rows for dashboard variables
- **TimeRangePicker**: Presets and custom absolute ranges
- **CronEditor**: Preset buttons with advanced cron expression and next-5 preview
- **TemplateDesigner**: Live PDF preview stub
- **AccessGuard**: Role-based action visibility

**State Management**: Query hooks for schedules/runs/settings via backend routes

## plugin.json Configuration

```json
{
  "type": "app",
  "backend": true,
  "executable": "backend",
  "includes": [/* page definitions */],
  "secureJsonData": {
    /* SA token, SMTP credentials */
  }
}
```

## Environment Variables for Development

```bash
GF_INSTALL_PLUGINS=grafana-image-renderer
GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=<your-app-id>
GF_RENDERING_SERVER_URL=http://renderer:8081/render
GF_RENDERING_CALLBACK_URL=http://grafana:3000/
```

## Testing Strategy

**Backend**:
- Unit tests for PDF assembly, email composition, cron parsing
- Integration tests for database operations
- Mock renderer and SMTP for isolated testing

**Frontend**:
- Component tests with React Testing Library
- E2E smoke tests: create schedule → assert artifact generation

**E2E Deployment Test**:
```bash
# Spin up Grafana + renderer + plugin
docker-compose up -d
# Create a test schedule via API
# Assert PDF artifact appears in storage
```

## MVP Feature Checklist

1. App shell with Schedules list and Create form
2. Backend HTTP routes with SQLite storage
3. Basic cron scheduler with single worker
4. Dashboard rendering to PNG then PDF assembly
5. SMTP email delivery with attachment
6. Run history list and artifact download
7. Settings page for SMTP and renderer configuration

## Performance Considerations

- **Large dashboards**: Increase timeouts, configure per-schedule render delay
- **Backpressure**: Global render concurrency limit to avoid overwhelming renderer
- **Artifact rotation**: Auto-delete old artifacts based on retention days
- **Metrics endpoint**: Expose render time, success rate, queue length for observability

## Common Pitfalls to Avoid

- **Auth failures**: Ensure service account token has proper permissions and refresh on rotation
- **Multi-org isolation**: Always filter by `org_id` in every database query
- **Variable encoding**: Strictly URL-encode dashboard variables with special characters
- **Renderer delays**: Heavy dashboards with complex queries need `render_delay_ms` adjustment
- **Timezone handling**: Store schedules with explicit timezone, convert for cron evaluation
