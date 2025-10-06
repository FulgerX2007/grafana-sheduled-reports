package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/yourusername/grafana-app-reporting/pkg/model"
)

// Store handles database operations
type Store struct {
	db *sql.DB
}

// NewStore creates a new store instance
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &Store{db: db}
	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return store, nil
}

// migrate runs database migrations
func (s *Store) migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS schedules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			org_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			dashboard_uid TEXT NOT NULL,
			dashboard_title TEXT,
			panel_ids TEXT,
			range_from TEXT NOT NULL,
			range_to TEXT NOT NULL,
			interval_type TEXT NOT NULL,
			cron_expr TEXT,
			timezone TEXT NOT NULL,
			format TEXT NOT NULL,
			variables TEXT,
			recipients TEXT NOT NULL,
			email_subject TEXT NOT NULL,
			email_body TEXT NOT NULL,
			template_id INTEGER,
			enabled INTEGER NOT NULL DEFAULT 1,
			last_run_at DATETIME,
			next_run_at DATETIME,
			owner_user_id INTEGER NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_schedules_org_id ON schedules(org_id)`,
		`CREATE INDEX IF NOT EXISTS idx_schedules_enabled ON schedules(enabled)`,
		`CREATE INDEX IF NOT EXISTS idx_schedules_next_run_at ON schedules(next_run_at)`,
		`CREATE TABLE IF NOT EXISTS runs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			schedule_id INTEGER NOT NULL,
			org_id INTEGER NOT NULL,
			started_at DATETIME NOT NULL,
			finished_at DATETIME,
			status TEXT NOT NULL,
			error_text TEXT,
			artifact_path TEXT,
			rendered_pages INTEGER NOT NULL DEFAULT 0,
			bytes INTEGER NOT NULL DEFAULT 0,
			checksum TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (schedule_id) REFERENCES schedules(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_runs_schedule_id ON runs(schedule_id)`,
		`CREATE INDEX IF NOT EXISTS idx_runs_org_id ON runs(org_id)`,
		`CREATE TABLE IF NOT EXISTS templates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			org_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			kind TEXT NOT NULL,
			config TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_templates_org_id ON templates(org_id)`,
		`CREATE TABLE IF NOT EXISTS settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			org_id INTEGER NOT NULL UNIQUE,
			use_grafana_smtp INTEGER NOT NULL DEFAULT 1,
			smtp_config TEXT,
			renderer_config TEXT NOT NULL,
			limits TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, migration := range migrations {
		if _, err := s.db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

// CreateSchedule creates a new schedule
func (s *Store) CreateSchedule(schedule *model.Schedule) error {
	now := time.Now()
	schedule.CreatedAt = now
	schedule.UpdatedAt = now

	result, err := s.db.Exec(`
		INSERT INTO schedules (
			org_id, name, dashboard_uid, dashboard_title, panel_ids, range_from, range_to,
			interval_type, cron_expr, timezone, format, variables, recipients,
			email_subject, email_body, template_id, enabled, owner_user_id,
			next_run_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		schedule.OrgID, schedule.Name, schedule.DashboardUID, schedule.DashboardTitle,
		schedule.PanelIDs, schedule.RangeFrom, schedule.RangeTo, schedule.IntervalType,
		schedule.CronExpr, schedule.Timezone, schedule.Format, schedule.Variables,
		schedule.Recipients, schedule.EmailSubject, schedule.EmailBody, schedule.TemplateID,
		schedule.Enabled, schedule.OwnerUserID, schedule.NextRunAt, now, now,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	schedule.ID = id

	return nil
}

// GetSchedule retrieves a schedule by ID
func (s *Store) GetSchedule(orgID, id int64) (*model.Schedule, error) {
	schedule := &model.Schedule{}
	err := s.db.QueryRow(`
		SELECT id, org_id, name, dashboard_uid, dashboard_title, panel_ids, range_from, range_to,
		       interval_type, cron_expr, timezone, format, variables, recipients,
		       email_subject, email_body, template_id, enabled, last_run_at, next_run_at,
		       owner_user_id, created_at, updated_at
		FROM schedules WHERE id = ? AND org_id = ?`,
		id, orgID,
	).Scan(
		&schedule.ID, &schedule.OrgID, &schedule.Name, &schedule.DashboardUID,
		&schedule.DashboardTitle, &schedule.PanelIDs, &schedule.RangeFrom, &schedule.RangeTo,
		&schedule.IntervalType, &schedule.CronExpr, &schedule.Timezone, &schedule.Format,
		&schedule.Variables, &schedule.Recipients, &schedule.EmailSubject, &schedule.EmailBody,
		&schedule.TemplateID, &schedule.Enabled, &schedule.LastRunAt, &schedule.NextRunAt,
		&schedule.OwnerUserID, &schedule.CreatedAt, &schedule.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("schedule not found")
	}
	return schedule, err
}

// ListSchedules retrieves all schedules for an organization
func (s *Store) ListSchedules(orgID int64) ([]*model.Schedule, error) {
	rows, err := s.db.Query(`
		SELECT id, org_id, name, dashboard_uid, dashboard_title, panel_ids, range_from, range_to,
		       interval_type, cron_expr, timezone, format, variables, recipients,
		       email_subject, email_body, template_id, enabled, last_run_at, next_run_at,
		       owner_user_id, created_at, updated_at
		FROM schedules WHERE org_id = ? ORDER BY created_at DESC`,
		orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schedules := make([]*model.Schedule, 0)
	for rows.Next() {
		schedule := &model.Schedule{}
		err := rows.Scan(
			&schedule.ID, &schedule.OrgID, &schedule.Name, &schedule.DashboardUID,
			&schedule.DashboardTitle, &schedule.PanelIDs, &schedule.RangeFrom, &schedule.RangeTo,
			&schedule.IntervalType, &schedule.CronExpr, &schedule.Timezone, &schedule.Format,
			&schedule.Variables, &schedule.Recipients, &schedule.EmailSubject, &schedule.EmailBody,
			&schedule.TemplateID, &schedule.Enabled, &schedule.LastRunAt, &schedule.NextRunAt,
			&schedule.OwnerUserID, &schedule.CreatedAt, &schedule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

// UpdateSchedule updates an existing schedule
func (s *Store) UpdateSchedule(schedule *model.Schedule) error {
	schedule.UpdatedAt = time.Now()

	_, err := s.db.Exec(`
		UPDATE schedules SET
			name = ?, dashboard_uid = ?, dashboard_title = ?, panel_ids = ?,
			range_from = ?, range_to = ?, interval_type = ?, cron_expr = ?,
			timezone = ?, format = ?, variables = ?, recipients = ?,
			email_subject = ?, email_body = ?, template_id = ?, enabled = ?,
			next_run_at = ?, updated_at = ?
		WHERE id = ? AND org_id = ?`,
		schedule.Name, schedule.DashboardUID, schedule.DashboardTitle, schedule.PanelIDs,
		schedule.RangeFrom, schedule.RangeTo, schedule.IntervalType, schedule.CronExpr,
		schedule.Timezone, schedule.Format, schedule.Variables, schedule.Recipients,
		schedule.EmailSubject, schedule.EmailBody, schedule.TemplateID, schedule.Enabled,
		schedule.NextRunAt, schedule.UpdatedAt, schedule.ID, schedule.OrgID,
	)
	return err
}

// DeleteSchedule deletes a schedule
func (s *Store) DeleteSchedule(orgID, id int64) error {
	_, err := s.db.Exec("DELETE FROM schedules WHERE id = ? AND org_id = ?", id, orgID)
	return err
}

// CreateRun creates a new run record
func (s *Store) CreateRun(run *model.Run) error {
	run.CreatedAt = time.Now()

	result, err := s.db.Exec(`
		INSERT INTO runs (schedule_id, org_id, started_at, status, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		run.ScheduleID, run.OrgID, run.StartedAt, run.Status, run.CreatedAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	run.ID = id

	return nil
}

// UpdateRun updates a run record
func (s *Store) UpdateRun(run *model.Run) error {
	_, err := s.db.Exec(`
		UPDATE runs SET
			finished_at = ?, status = ?, error_text = ?, artifact_path = ?,
			rendered_pages = ?, bytes = ?, checksum = ?
		WHERE id = ?`,
		run.FinishedAt, run.Status, run.ErrorText, run.ArtifactPath,
		run.RenderedPages, run.Bytes, run.Checksum, run.ID,
	)
	return err
}

// GetRun retrieves a run by ID
func (s *Store) GetRun(orgID, id int64) (*model.Run, error) {
	run := &model.Run{}
	var finishedAt sql.NullTime
	var errorText, artifactPath, checksum sql.NullString

	err := s.db.QueryRow(`
		SELECT id, schedule_id, org_id, started_at, finished_at, status, error_text,
		       artifact_path, rendered_pages, bytes, checksum, created_at
		FROM runs WHERE id = ? AND org_id = ?`,
		id, orgID,
	).Scan(
		&run.ID, &run.ScheduleID, &run.OrgID, &run.StartedAt, &finishedAt,
		&run.Status, &errorText, &artifactPath, &run.RenderedPages,
		&run.Bytes, &checksum, &run.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("run not found")
	}
	if err != nil {
		return nil, err
	}

	// Convert nullable fields
	if finishedAt.Valid {
		run.FinishedAt = &finishedAt.Time
	}
	if errorText.Valid {
		run.ErrorText = errorText.String
	}
	if artifactPath.Valid {
		run.ArtifactPath = artifactPath.String
	}
	if checksum.Valid {
		run.Checksum = checksum.String
	}

	return run, nil
}

// ListRuns retrieves runs for a schedule
func (s *Store) ListRuns(orgID, scheduleID int64) ([]*model.Run, error) {
	rows, err := s.db.Query(`
		SELECT id, schedule_id, org_id, started_at, finished_at, status, error_text,
		       artifact_path, rendered_pages, bytes, checksum, created_at
		FROM runs WHERE schedule_id = ? AND org_id = ? ORDER BY started_at DESC LIMIT 50`,
		scheduleID, orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	runs := make([]*model.Run, 0)
	for rows.Next() {
		run := &model.Run{}
		var finishedAt sql.NullTime
		var errorText, artifactPath, checksum sql.NullString

		err := rows.Scan(
			&run.ID, &run.ScheduleID, &run.OrgID, &run.StartedAt, &finishedAt,
			&run.Status, &errorText, &artifactPath, &run.RenderedPages,
			&run.Bytes, &checksum, &run.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Convert nullable fields
		if finishedAt.Valid {
			run.FinishedAt = &finishedAt.Time
		}
		if errorText.Valid {
			run.ErrorText = errorText.String
		}
		if artifactPath.Valid {
			run.ArtifactPath = artifactPath.String
		}
		if checksum.Valid {
			run.Checksum = checksum.String
		}

		runs = append(runs, run)
	}

	return runs, nil
}

// GetSettings retrieves settings for an organization
func (s *Store) GetSettings(orgID int64) (*model.Settings, error) {
	settings := &model.Settings{}
	err := s.db.QueryRow(`
		SELECT id, org_id, use_grafana_smtp, smtp_config, renderer_config, limits, created_at, updated_at
		FROM settings WHERE org_id = ?`,
		orgID,
	).Scan(
		&settings.ID, &settings.OrgID, &settings.UseGrafanaSMTP, &settings.SMTPConfig,
		&settings.RendererConfig, &settings.Limits, &settings.CreatedAt, &settings.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return settings, err
}

// UpsertSettings creates or updates settings
func (s *Store) UpsertSettings(settings *model.Settings) error {
	now := time.Now()
	settings.UpdatedAt = now

	existing, err := s.GetSettings(settings.OrgID)
	if err != nil {
		return err
	}

	if existing == nil {
		settings.CreatedAt = now
		result, err := s.db.Exec(`
			INSERT INTO settings (org_id, use_grafana_smtp, smtp_config, renderer_config, limits, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			settings.OrgID, settings.UseGrafanaSMTP, settings.SMTPConfig, settings.RendererConfig,
			settings.Limits, settings.CreatedAt, settings.UpdatedAt,
		)
		if err != nil {
			return err
		}
		id, _ := result.LastInsertId()
		settings.ID = id
	} else {
		_, err := s.db.Exec(`
			UPDATE settings SET
				use_grafana_smtp = ?, smtp_config = ?, renderer_config = ?, limits = ?, updated_at = ?
			WHERE org_id = ?`,
			settings.UseGrafanaSMTP, settings.SMTPConfig, settings.RendererConfig,
			settings.Limits, settings.UpdatedAt, settings.OrgID,
		)
		return err
	}

	return nil
}

// GetDueSchedules retrieves schedules that are due to run
func (s *Store) GetDueSchedules() ([]*model.Schedule, error) {
	rows, err := s.db.Query(`
		SELECT id, org_id, name, dashboard_uid, dashboard_title, panel_ids, range_from, range_to,
		       interval_type, cron_expr, timezone, format, variables, recipients,
		       email_subject, email_body, template_id, enabled, last_run_at, next_run_at,
		       owner_user_id, created_at, updated_at
		FROM schedules
		WHERE enabled = 1 AND (next_run_at IS NULL OR next_run_at <= datetime('now'))
		ORDER BY next_run_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schedules := make([]*model.Schedule, 0)
	for rows.Next() {
		schedule := &model.Schedule{}
		err := rows.Scan(
			&schedule.ID, &schedule.OrgID, &schedule.Name, &schedule.DashboardUID,
			&schedule.DashboardTitle, &schedule.PanelIDs, &schedule.RangeFrom, &schedule.RangeTo,
			&schedule.IntervalType, &schedule.CronExpr, &schedule.Timezone, &schedule.Format,
			&schedule.Variables, &schedule.Recipients, &schedule.EmailSubject, &schedule.EmailBody,
			&schedule.TemplateID, &schedule.Enabled, &schedule.LastRunAt, &schedule.NextRunAt,
			&schedule.OwnerUserID, &schedule.CreatedAt, &schedule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}
