Fine. You want a real, UI-driven reporting plugin for Grafana OSS with both frontend and backend, built “the Grafana way.” Here’s a concrete plan you can hand to your repo and start coding instead of doom-scrolling docs.

# Scope

* App plugin (not just panel) so you get a left-nav entry, pages, backend routes, secure settings, and background jobs.
* Features: schedule reports (weekly, daily, cron), render dashboard(s) to PDF/HTML, email them, keep history and audit, all managed from UI. Works on Grafana OSS.

# Architecture

* **Frontend**: App plugin in React/TypeScript using Grafana UI and app routing.
* **Backend**: Go backend plugin using Grafana Plugin SDK for HTTP routes, secure storage, org scoping, background workers.
* **Renderer**: Use the official grafana-image-renderer HTTP service. Plugin calls Grafana’s `/render` endpoints to get PNGs or a full dashboard image. Assemble PDFs in backend.
* **Email**: SMTP client in backend. Prefer reusing Grafana SMTP settings if present; otherwise plugin’s own secure SMTP settings.
* **Storage**: SQLite (embedded) or BoltDB in plugin data dir for schedules, runs, history. Everything scoped by OrgID.
* **Auth**: Use Grafana service account token (stored securely in plugin settings) to call render endpoints. All plugin UI actions are permission-gated.

# UX Pages (App plugin)

1. **Schedules**

   * Table: name, target dashboard(s), recipients, format, time range, variables, next run, status.
   * Actions: Create, Edit, Duplicate, Pause/Resume, Run now, View history.
2. **Create/Edit Schedule**

   * Dashboard picker (UID), panel selection (optional), time range preset or absolute, dashboard variables (KV), layout options, format (PDF/HTML), email subject/body, recipients (To/CC/BCC), attachments vs inline, file name template, timezone.
   * Schedule: presets (Every Monday 08:00) or cron expression. Toggle “skip if same as last.”
3. **Run History**

   * Per schedule: runs with status, duration, pages rendered, email result, link to stored artifact. Download button.
4. **Templates**

   * PDF/HTML layout templates, email templates, header/footer, logo upload, watermark, page size and margins.
5. **Settings**

   * SMTP mode: use Grafana SMTP or custom SMTP (host, port, TLS, user/pass in secure JSON).
   * Rendering: URL to image-renderer, timeouts, retries, viewport, delay before capture.
   * Security: allowed org roles (Viewer/Editor/Admin) for managing schedules, max recipients, attachment size limits.
   * Service Account token (secure), default org-wide headers, rate limits.

# Permissions model

* **View schedules**: Editor and above by default.
* **Create/Edit/Delete**: Editor and above.
* **Settings**: Admin only.
* All records stored with `OrgID`; cross-org access is denied. Enforce on every route.

# Data model (SQLite suggested)

Tables, all with `org_id`, `created_at`, `updated_at`:

* `schedules(id, name, dashboard_uid, panel_ids_json, range_from, range_to, interval_type, cron_expr, tz, format, vars_json, recipients_json, email_subject, email_body, template_id, enabled, last_run_at, next_run_at, owner_user_id)`
* `runs(id, schedule_id, started_at, finished_at, status, error_text, artifact_path, rendered_pages, bytes, checksum)`
* `templates(id, name, kind, config_json)`  // kind: pdf, html, email
* `settings(id, use_grafana_smtp, smtp_json_secure, renderer_json, limits_json)`  // singleton per org
* `audit(id, actor_user_id, action, target_type, target_id, details_json, at)`

# Backend components (Go)

* **HTTP API** under `/api/<your-app-id>`:

  * `GET /schedules`, `POST /schedules`, `PUT /schedules/:id`, `DELETE /schedules/:id`
  * `POST /schedules/:id/run`  // run now
  * `GET /schedules/:id/runs`, `GET /runs/:id/artifact`
  * `GET/POST /settings`, `GET/POST /templates`
  * `GET /dashboards/search?query=...`  // proxy to Grafana API to help the UI pick dashboards
* **Scheduler**:

  * Goroutine with cron parser (support cron and presets). On tick, select due schedules per org, enqueue jobs.
  * Worker pool with backoff retry (e.g., 3 tries, exponential).
  * Idempotency with run tokens to avoid double send after restarts.
* **Renderer client**:

  * Build render URLs to `/render/d/<uid>` or `/render/d-solo/<uid>` with `from/to`, `var-xyz`, `theme`, `orgId`, `kiosk`, `tz`.
  * Wait for panel data: respect `rendering_delay_ms`, `timeout_ms`.
  * Collect images, then:

    * **PDF**: assemble with gofpdf or unidoc, add header/footer, page numbers, logo, optional watermark.
    * **HTML**: assemble a static HTML with embedded images (base64) or referential links; include CSS template.
* **Email sender**:

  * If `use_grafana_smtp=true`, read Grafana’s cfg via env injection you provide at deployment time or expose a minimal “use same host/port/from” option replicated in plugin’s settings UI. If off, use plugin’s own SMTP.
  * Compose MIME with attachments; support inline images and HTML body.
* **Secrets**:

  * Store tokens and SMTP secrets in plugin secure settings (secureJSON). Never in plaintext tables.
* **Org & user context**:

  * Read from request context; stamp `org_id` and `actor_user_id` in audit.
* **Validation**:

  * Validate cron, recipients, dashboard existence, variable names, range bounds, size limits.

# Frontend (React + Grafana UI)

* Use `@grafana/runtime`, `@grafana/data`, `@grafana/ui`.
* Routing: `/schedules`, `/schedules/new`, `/schedules/:id`, `/runs/:id`, `/settings`, `/templates`.
* Components:

  * DashboardPicker with search by title, UID preview.
  * VariableEditor: dynamic key/value rows, supports multi-value lists.
  * TimeRangePicker: common presets and custom absolute.
  * CronEditor: preset buttons + advanced text with live next-5 preview.
  * TemplateDesigner: live PDF preview stub that fetches a quick render of first page.
  * AccessGuard: hide actions based on user role from `/api/user`.
* State:

  * Query hooks for schedules/runs/settings via backend routes.
  * Forms with zod/yup validation, show red flags before save.
* Polish:

  * “Run now” button with progress toast; link to new run entry.
  * Copy schedule as new template.
  * Diff view when editing running schedules.

# Repo structure

```
grafana-app-reporting/
├─ src/                      # frontend
│  ├─ pages/
│  │  ├─ Schedules/
│  │  ├─ ScheduleEdit/
│  │  ├─ Settings/
│  │  └─ Templates/
│  ├─ components/
│  ├─ types/
│  ├─ plugin.json
│  └─ module.ts
├─ pkg/
│  ├─ api/                   # HTTP handlers
│  ├─ cron/                  # scheduler + workers
│  ├─ render/                # renderer client
│  ├─ pdf/                   # pdf assembler
│  ├─ mail/                  # smtp sender
│  ├─ store/                 # sqlite/bolt db, migrations
│  ├─ model/                 # DTOs, validation
│  └─ auth/                  # SA token, org/user context
├─ cmd/
│  └─ backend/               # main.go for backend plugin
├─ dist/                     # build outputs
├─ provisioning/             # optional app provisioning
└─ Makefile / magefile.go
```

# plugin.json essentials

* `type: "app"`
* Backend:

  * `backend: true`
  * `executable: "backend"` (built binary path)
* Routes:

  * mark `includes` for your pages
* Secure settings schema for SMTP and SA token
* Sign plugin (developer cert) for loading in Grafana without unsigned flag in prod.

# Scheduling semantics

* Support:

  * **Presets**: daily/weekly/monthly at HH:MM.
  * **Cron**: full five-field with seconds optional.
  * Timezone per schedule.
  * “Do not run if identical to previous week” option by checksumming last run artifact.
* Backpressure:

  * Global limit: max concurrent renders; queue the rest.
  * Per org quotas to avoid one tenant consuming whole renderer.

# Rendering details that save your sanity

* Force `kiosk=tv` to hide Grafana chrome.
* Set `timezone`, `theme=light|dark` matching your template.
* Add a configurable `render_delay_ms` to let queries finish for heavy dashboards.
* For variable values with spaces, URL-encode strictly.
* Use `panelId` loop for panel-level pages or full dashboard for one-shot.
* For multi-page PDF, single dashboard render per page is fine; or stitch multiple panel images into a grid per page.

# Email & deliverability

* DKIM/DMARC aren’t your plugin’s job, but add:

  * optional custom “From” and “Reply-To”
  * rate limit per minute
  * attachment size guard, fallback to download link if too large
* HTML email body supports a mini-template with placeholders:

  * `{{schedule.name}}`, `{{timerange}}`, `{{dashboard.title}}`, `{{run.started_at}}`

# History & observability

* Store artifacts on disk under plugin data dir, rotated by retention days.
* Expose `/metrics` for plugin: render time, success rate, queue length, email failures.
* Health check route for container orchestration.

# Security

* Enforce org scoping on every query.
* Mask secrets in every GET.
* Only Admins can change SMTP or SA token.
* Use a dedicated Grafana Service Account with Viewer role limited by folder permissions to avoid leaking restricted dashboards in reports.

# Deployment & ops

* Docker envs to set:

  * `GF_INSTALL_PLUGINS=grafana-image-renderer`
  * `GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=<your-app-id>` during dev
  * `GF_RENDERING_SERVER_URL` and `GF_RENDERING_CALLBACK_URL`
* Migrations: simple goose or embed SQL migrations on backend start.
* CI:

  * Lint TS/Go, run unit tests for PDF/mail building, build signed plugin zip.
  * E2E smoke: spin Grafana + renderer, create a schedule, assert a PDF file appears.

# MVP cut (2 weeks of focused work if you don’t get lost in tabs)

1. App shell with Schedules list, Create form, minimal Settings.
2. Backend routes + SQLite store.
3. Basic cron scheduler with one worker.
4. Render full dashboard to PNG then PDF with logo header/footer.
5. SMTP send with attachment.
6. Run history list and artifact download.

# Hard edges to plan for

* Very large dashboards or slow panels: increase timeouts, allow per-schedule render delay.
* Auth failures: rotate SA token via Settings and store securely.
* Variables with dynamic queries: pre-resolve them via Grafana API if needed before render call.
* Multi-org: test. Seriously test. Then test again.

# Tiny code teasers

## Backend route sketch (Go)

```go
r.Group("/api/your-app", func(gr *mux.Router) {
  gr.Get("/schedules", h.ListSchedules)
  gr.Post("/schedules", h.CreateSchedule)
  gr.Put("/schedules/{id}", h.UpdateSchedule)
  gr.Delete("/schedules/{id}", h.DeleteSchedule)
  gr.Post("/schedules/{id}/run", h.RunNow)
  gr.Get("/schedules/{id}/runs", h.ListRuns)
  gr.Get("/runs/{id}/artifact", h.DownloadArtifact)
  gr.Get("/settings", h.GetSettings)
  gr.Post("/settings", h.SaveSettings)
})
```

## Render call outline

```go
func (c *Renderer) RenderDashboard(ctx context.Context, orgID int64, uid string, rng TimeRange, vars map[string]string) ([]byte, error) {
  url := fmt.Sprintf("%s/render/d/%s", c.GrafanaURL, uid)
  q := urlpkg.Values{}
  q.Set("from", rng.From)
  q.Set("to", rng.To)
  q.Set("kiosk", "tv")
  q.Set("orgId", strconv.FormatInt(orgID, 10))
  for k,v := range vars {
    q.Set("var-"+k, v)
  }
  req, _ := http.NewRequestWithContext(ctx, "GET", url+"?"+q.Encode(), nil)
  req.Header.Set("Authorization", "Bearer "+c.SAToken)
  // http GET -> bytes (PNG)
}
```

## PDF assemble (one image per page)

```go
pdf := gofpdf.New("L", "mm", "A4", "")
pdf.SetHeaderFunc(func() { /* logo, title, date */ })
for _, img := range images {
  pdf.AddPage()
  pdf.ImageOptions(img.Path, 10, 20, 277, 0, false, gofpdf.ImageOptions{ImageType:"PNG"}, 0, "")
}
buf := new(bytes.Buffer)
_ = pdf.Output(buf)
```

That’s the plan. It checks all the “Grafana plugin” boxes, lets you manage everything in UI, and doesn’t rely on Enterprise magic. If you follow it, you’ll ship something your future self won’t hate.
