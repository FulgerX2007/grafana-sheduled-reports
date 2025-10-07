package cron

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/yourusername/sheduled-reports-app/pkg/mail"
	"github.com/yourusername/sheduled-reports-app/pkg/model"
	"github.com/yourusername/sheduled-reports-app/pkg/render"
	"github.com/yourusername/sheduled-reports-app/pkg/store"
)

// Scheduler handles report scheduling
type Scheduler struct {
	store         *store.Store
	cron          *cron.Cron
	grafanaURL    string
	artifactsPath string
	workerPool    chan struct{}
	baseCtx       context.Context // Context with Grafana config for background jobs
}

// NewScheduler creates a new scheduler instance
func NewScheduler(st *store.Store, grafanaURL, artifactsPath string, maxConcurrent int) *Scheduler {
	return &Scheduler{
		store:         st,
		cron:          cron.New(cron.WithSeconds()),
		grafanaURL:    grafanaURL,
		artifactsPath: artifactsPath,
		workerPool:    make(chan struct{}, maxConcurrent),
		baseCtx:       context.Background(), // Will be updated when plugin starts
	}
}

// SetContext sets the base context for the scheduler (should be called on plugin initialization)
func (s *Scheduler) SetContext(ctx context.Context) {
	s.baseCtx = ctx
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	// Add a job that runs every minute to check for due schedules
	_, err := s.cron.AddFunc("0 * * * * *", s.checkDueSchedules)
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	s.cron.Start()
	log.Println("Scheduler started")

	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.cron.Stop()
	log.Println("Scheduler stopped")
}

// checkDueSchedules checks for schedules that are due and executes them
func (s *Scheduler) checkDueSchedules() {
	schedules, err := s.store.GetDueSchedules()
	if err != nil {
		log.Printf("Failed to get due schedules: %v", err)
		return
	}

	for _, schedule := range schedules {
		// Update next run time immediately to prevent duplicate execution
		nextRun := s.calculateNextRun(schedule)
		schedule.NextRunAt = &nextRun
		if err := s.store.UpdateSchedule(schedule); err != nil {
			log.Printf("Failed to update schedule %d next run time: %v", schedule.ID, err)
			continue
		}

		// Execute in worker pool
		go s.executeSchedule(schedule)
	}
}

// ExecuteSchedule executes a schedule immediately (for manual runs)
func (s *Scheduler) ExecuteSchedule(schedule *model.Schedule) {
	go s.executeSchedule(schedule)
}

// executeSchedule executes a single schedule
func (s *Scheduler) executeSchedule(schedule *model.Schedule) {
	// Acquire worker slot
	s.workerPool <- struct{}{}
	defer func() { <-s.workerPool }()

	// Create run record
	run := &model.Run{
		ScheduleID: schedule.ID,
		OrgID:      schedule.OrgID,
		StartedAt:  time.Now(),
		Status:     "running",
	}

	if err := s.store.CreateRun(run); err != nil {
		log.Printf("Failed to create run record: %v", err)
		return
	}

	// Execute with retries
	err := s.executeWithRetry(schedule, run, 3)

	// Update run record
	now := time.Now()
	run.FinishedAt = &now

	if err != nil {
		run.Status = "failed"
		run.ErrorText = err.Error()
		log.Printf("Schedule %d execution failed: %v", schedule.ID, err)
	} else {
		run.Status = "completed"
	}

	if err := s.store.UpdateRun(run); err != nil {
		log.Printf("Failed to update run record: %v", err)
	}

	// Update schedule last run time
	schedule.LastRunAt = &run.StartedAt
	if err := s.store.UpdateSchedule(schedule); err != nil {
		log.Printf("Failed to update schedule last run time: %v", err)
	}
}

// executeWithRetry executes a schedule with retry logic
func (s *Scheduler) executeWithRetry(schedule *model.Schedule, run *model.Run, maxRetries int) error {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(attempt*attempt) * time.Second
			log.Printf("Retrying schedule %d (attempt %d/%d) after %v", schedule.ID, attempt+1, maxRetries, backoff)
			time.Sleep(backoff)
		}

		err := s.executeScheduleOnce(schedule, run)
		if err == nil {
			return nil
		}

		lastErr = err
		log.Printf("Schedule %d execution attempt %d failed: %v", schedule.ID, attempt+1, err)
	}

	return fmt.Errorf("all %d attempts failed: %w", maxRetries, lastErr)
}

// executeScheduleOnce executes a schedule once
func (s *Scheduler) executeScheduleOnce(schedule *model.Schedule, run *model.Run) error {
	// Use the base context which has Grafana config
	ctx := s.baseCtx

	// Get settings
	settings, err := s.store.GetSettings(schedule.OrgID)
	if err != nil {
		return fmt.Errorf("failed to get settings: %w", err)
	}
	if settings == nil {
		return fmt.Errorf("no settings configured for org %d", schedule.OrgID)
	}

	log.Printf("DEBUG: Rendering with grafanaURL=%s, mode=%s", s.grafanaURL, settings.RendererConfig.Mode)

	// Create renderer based on configuration
	renderer, err := render.NewDashboardRenderer(s.grafanaURL, settings.RendererConfig)
	if err != nil {
		return fmt.Errorf("failed to create renderer: %w", err)
	}

	// Generate report data based on format
	var reportData []byte
	var filename string

	if schedule.Format == "pdf" {
		reportData, err = renderer.RenderToPDF(ctx, schedule)
		if err != nil {
			return fmt.Errorf("failed to render PDF: %w", err)
		}
		filename = fmt.Sprintf("%s-%s.pdf", schedule.Name, time.Now().Format("2006-01-02-150405"))
	} else if schedule.Format == "html" {
		reportData, err = renderer.RenderToHTML(ctx, schedule)
		if err != nil {
			return fmt.Errorf("failed to render HTML: %w", err)
		}
		filename = fmt.Sprintf("%s-%s.html", schedule.Name, time.Now().Format("2006-01-02-150405"))
	} else {
		// Default to PDF for backwards compatibility
		reportData, err = renderer.RenderToPDF(ctx, schedule)
		if err != nil {
			return fmt.Errorf("failed to render PDF: %w", err)
		}
		filename = fmt.Sprintf("%s-%s.pdf", schedule.Name, time.Now().Format("2006-01-02-150405"))
	}

	run.RenderedPages = 1

	run.Bytes = int64(len(reportData))

	// Calculate checksum
	checksum := fmt.Sprintf("%x", sha256.Sum256(reportData))
	run.Checksum = checksum

	// Save artifact
	artifactPath := filepath.Join(s.artifactsPath, fmt.Sprintf("org_%d", schedule.OrgID), filename)
	if err := os.MkdirAll(filepath.Dir(artifactPath), 0755); err != nil {
		return fmt.Errorf("failed to create artifacts directory: %w", err)
	}

	if err := os.WriteFile(artifactPath, reportData, 0644); err != nil {
		return fmt.Errorf("failed to save artifact: %w", err)
	}

	run.ArtifactPath = artifactPath

	// Send email
	var smtpConfig model.SMTPConfig
	if settings.UseGrafanaSMTP {
		// Load Grafana SMTP config from environment variables
		smtpHost := os.Getenv("GF_SMTP_HOST")
		smtpUser := os.Getenv("GF_SMTP_USER")
		smtpPassword := os.Getenv("GF_SMTP_PASSWORD")
		smtpFrom := os.Getenv("GF_SMTP_FROM_ADDRESS")

		if smtpHost == "" {
			// Fall back to plugin SMTP config
			if settings.SMTPConfig != nil {
				smtpConfig = *settings.SMTPConfig
			} else {
				return fmt.Errorf("no SMTP configuration available")
			}
		} else {
			// Parse host:port
			smtpPort := 25
			if parts := strings.Split(smtpHost, ":"); len(parts) == 2 {
				smtpHost = parts[0]
				if port, err := strconv.Atoi(parts[1]); err == nil {
					smtpPort = port
				}
			}

			smtpConfig = model.SMTPConfig{
				Host:     smtpHost,
				Port:     smtpPort,
				Username: smtpUser,
				Password: smtpPassword,
				From:     smtpFrom,
				UseTLS:   false, // Set based on GF_SMTP_SKIP_VERIFY if needed
			}
		}
	} else {
		if settings.SMTPConfig == nil {
			return fmt.Errorf("SMTP configuration not set")
		}
		smtpConfig = *settings.SMTPConfig
	}

	mailer := mail.NewMailer(smtpConfig)

	// Interpolate template variables
	vars := map[string]string{
		"schedule.name":    schedule.Name,
		"dashboard.title":  schedule.DashboardTitle,
		"timerange":        fmt.Sprintf("%s to %s", schedule.RangeFrom, schedule.RangeTo),
		"run.started_at":   run.StartedAt.Format(time.RFC1123),
	}

	subject := mail.InterpolateTemplate(schedule.EmailSubject, vars)
	body := mail.InterpolateTemplate(schedule.EmailBody, vars)

	if err := mailer.SendReport(schedule.Recipients, subject, body, reportData, filename); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// CalculateNextRun calculates the next run time for a schedule (exported for use in handlers)
func (s *Scheduler) CalculateNextRun(schedule *model.Schedule) time.Time {
	return s.calculateNextRun(schedule)
}

// calculateNextRun calculates the next run time for a schedule
func (s *Scheduler) calculateNextRun(schedule *model.Schedule) time.Time {
	now := time.Now()

	switch schedule.IntervalType {
	case "daily":
		return now.Add(24 * time.Hour)
	case "weekly":
		return now.Add(7 * 24 * time.Hour)
	case "monthly":
		return now.AddDate(0, 1, 0)
	case "cron":
		if schedule.CronExpr != "" {
			parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
			sched, err := parser.Parse(schedule.CronExpr)
			if err != nil {
				log.Printf("Failed to parse cron expression %s: %v", schedule.CronExpr, err)
				return now.Add(1 * time.Hour)
			}
			return sched.Next(now)
		}
	}

	return now.Add(1 * time.Hour)
}
