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
- **Renderer**: Multi-backend rendering system (Chromium via go-rod or wkhtmltopdf for lightweight deployments)
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
go build -o dist/gpx_reporting ./cmd/backend   # Build backend plugin
go test ./...                # Run all tests
go test -v ./pkg/render      # Test rendering backends
go test -v ./pkg/cron        # Test scheduler
```

### Plugin Development
```bash
# Run with Grafana (requires Docker with Chromium)
docker-compose up -d         # Start Grafana with embedded Chromium

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

## Rendering Backend Architecture

The plugin uses a **multi-backend rendering system** allowing users to choose between different rendering engines:

### Backend Interface Pattern

All rendering backends implement the `Backend` interface:

```go
type Backend interface {
    RenderDashboard(ctx context.Context, schedule *model.Schedule) ([]byte, error)
    Close() error
    Name() string
}
```

Factory function for runtime backend selection:
```go
func NewBackend(backendType BackendType, grafanaURL string, config model.RendererConfig) (Backend, error)
```

### Available Backends

**1. Chromium Backend** (default - `chromium_renderer.go`)
   - Uses `github.com/go-rod/rod` for browser automation
   - Full JavaScript support with modern Chromium engine
   - Per-organization browser instance reuse for performance
   - Lazy initialization - browser created on first render
   - Request hijacking for authentication header injection
   - Binary size: ~300MB (including Chromium)
   - Best for complex dashboards with heavy JavaScript

**2. wkhtmltopdf Backend** (lightweight - `wkhtmltopdf_renderer.go`)
   - Uses `github.com/SebastiaanKlippert/go-wkhtmltopdf` wrapper
   - Direct PDF generation (no intermediate PNG)
   - WebKit-based rendering engine
   - Minimal system dependencies
   - Binary size: ~12MB
   - Best for simple dashboards and Docker/containerized deployments

### Configuration

Backend selection configured per organization in `RendererConfig`:

```go
type RendererConfig struct {
    Backend           string  `json:"backend"`  // "chromium" or "wkhtmltopdf"

    // Common settings (both backends)
    TimeoutMS         int     `json:"timeout_ms"`
    DelayMS           int     `json:"delay_ms"`
    ViewportWidth     int     `json:"viewport_width"`
    ViewportHeight    int     `json:"viewport_height"`
    DeviceScaleFactor float64 `json:"device_scale_factor"`
    SkipTLSVerify     bool    `json:"skip_tls_verify"`

    // Chromium-specific
    ChromiumPath      string  `json:"chromium_path"`
    Headless          bool    `json:"headless"`
    DisableGPU        bool    `json:"disable_gpu"`
    NoSandbox         bool    `json:"no_sandbox"`

    // wkhtmltopdf-specific
    WkhtmltopdfPath   string  `json:"wkhtmltopdf_path"`
}
```

### Runtime Backend Switching

The scheduler (`pkg/cron/scheduler.go`) maintains per-org renderer instances:

```go
type Scheduler struct {
    renderers map[int64]render.Backend  // Per-org renderer cache
}

// Get or create renderer for this org
renderer, exists := s.renderers[schedule.OrgID]
if !exists || renderer.Name() != string(backendType) {
    // Create new renderer if doesn't exist or backend changed
    renderer, err = render.NewBackend(backendType, s.grafanaURL, settings.RendererConfig)
    s.renderers[schedule.OrgID] = renderer
}
```

When backend type changes in settings, old renderer is replaced on next execution.

### Authentication

Both backends inject service account tokens for dashboard access:
- **Chromium**: Via request hijacking with `rod.Hijack()`
- **wkhtmltopdf**: Via `page.CustomHeader.Set("Authorization", "Bearer "+token)`

### Adding New Backends

To add a new rendering backend:

1. Create `pkg/render/newbackend_renderer.go`
2. Implement the `Backend` interface
3. Add new `BackendType` constant to `interface.go`
4. Update `NewBackend()` factory function in `interface.go`
5. Add backend-specific config fields to `RendererConfig` in `pkg/model/models.go`
6. Add tests to `renderer_test.go`
7. Update documentation

### Testing

- **Unit tests**: `pkg/render/renderer_test.go` (11 test suites)
- **Integration tests**: `pkg/render/renderer_integration_test.go` (7 scenarios)
- Run with: `go test ./pkg/render/...`
- Integration tests require Chromium: `go test -tags=integration ./pkg/render/`

## Key Implementation Patterns

### Dashboard Variable Auto-loading
When a user selects a dashboard in the schedule editor:
1. Frontend fetches dashboard definition via `/api/dashboards/uid/{uid}`
2. Extracts template variables from `dashboard.templating.list`
3. Filters out constant and datasource variables
4. Auto-populates the variables editor with variable names and default values
5. User can then modify values as needed before saving

Implementation in `ScheduleEditPage.tsx:63-80`

### Rendering Flow

**Chromium Backend:**
1. Get or create browser instance for org (lazy initialization)
2. Build dashboard URL: `/d/<uid>?from=X&to=Y&var-foo=bar&kiosk=tv&orgId=N`
3. Create new page and set viewport dimensions
4. Inject service account token via request hijacking
5. Navigate to dashboard and wait for page load
6. Respect `render_delay_ms` for heavy dashboards to let queries finish
7. Capture full-page screenshot as PNG
8. Assemble PDF with gofpdf or use PNG directly for HTML format

**wkhtmltopdf Backend:**
1. Build dashboard URL with all parameters
2. Create PDF generator with page settings
3. Add custom Authorization header for service account token
4. Set JavaScript delay to allow queries to complete
5. Generate PDF directly (no intermediate PNG)
6. Return PDF bytes

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
GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=<your-app-id>
# Optional: Custom Chromium binary path
CHROMIUM_PATH=/usr/bin/chromium-browser
# Optional: Chromium flags (automatically configured for Docker)
CHROMIUM_NO_SANDBOX=true
CHROMIUM_DISABLE_GPU=true
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
- **Browser reuse**: Browser instances are reused per-org to reduce initialization overhead
- **Backpressure**: Global render concurrency limit to prevent memory exhaustion
- **Artifact rotation**: Auto-delete old artifacts based on retention days
- **Metrics endpoint**: Expose render time, success rate, queue length for observability
- **Memory management**: Each browser instance consumes ~100-200MB, managed via worker pool

## Common Pitfalls to Avoid

- **Auth failures**: Ensure service account token has proper permissions and refresh on rotation
- **Multi-org isolation**: Always filter by `org_id` in every database query
- **Variable encoding**: Strictly URL-encode dashboard variables with special characters
- **Renderer delays**: Heavy dashboards with complex queries need `render_delay_ms` adjustment
- **Timezone handling**: Store schedules with explicit timezone, convert for cron evaluation
- **Browser cleanup**: Always call renderer.Close() on shutdown to prevent zombie processes
- **Docker sandbox**: Must enable `no_sandbox` flag for Chromium in Docker containers
