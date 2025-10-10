package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Schedule represents a scheduled report
type Schedule struct {
	ID             int64             `json:"id"`
	OrgID          int64             `json:"org_id"`
	Name           string            `json:"name"`
	DashboardUID   string            `json:"dashboard_uid"`
	DashboardTitle string            `json:"dashboard_title,omitempty"`
	PanelIDs       IntSlice          `json:"panel_ids,omitempty"`
	RangeFrom      string            `json:"range_from"`
	RangeTo        string            `json:"range_to"`
	IntervalType   string            `json:"interval_type"`
	CronExpr       string            `json:"cron_expr,omitempty"`
	Timezone       string            `json:"timezone"`
	Format         string            `json:"format"`
	Variables      JSONMap           `json:"variables,omitempty"`
	Recipients     Recipients        `json:"recipients"`
	EmailSubject   string            `json:"email_subject"`
	EmailBody      string            `json:"email_body"`
	TemplateID     *int64            `json:"template_id,omitempty"`
	Enabled        bool              `json:"enabled"`
	LastRunAt      *time.Time        `json:"last_run_at,omitempty"`
	NextRunAt      *time.Time        `json:"next_run_at,omitempty"`
	OwnerUserID    int64             `json:"owner_user_id"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// Recipients holds email recipient information
type Recipients struct {
	To  []string `json:"to"`
	CC  []string `json:"cc,omitempty"`
	BCC []string `json:"bcc,omitempty"`
}

// Run represents a report execution
type Run struct {
	ID            int64      `json:"id"`
	ScheduleID    int64      `json:"schedule_id"`
	OrgID         int64      `json:"org_id"`
	StartedAt     time.Time  `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
	Status        string     `json:"status"`
	ErrorText     string     `json:"error_text,omitempty"`
	ArtifactPath  string     `json:"artifact_path,omitempty"`
	RenderedPages int        `json:"rendered_pages"`
	Bytes         int64      `json:"bytes"`
	Checksum      string     `json:"checksum,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// Template represents a report template
type Template struct {
	ID        int64          `json:"id"`
	OrgID     int64          `json:"org_id"`
	Name      string         `json:"name"`
	Kind      string         `json:"kind"`
	Config    TemplateConfig `json:"config"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// TemplateConfig holds template configuration
type TemplateConfig struct {
	Header      string  `json:"header,omitempty"`
	Footer      string  `json:"footer,omitempty"`
	LogoURL     string  `json:"logo_url,omitempty"`
	Watermark   string  `json:"watermark,omitempty"`
	PageSize    string  `json:"page_size,omitempty"`
	Orientation string  `json:"orientation,omitempty"`
	Margins     *Margins `json:"margins,omitempty"`
}

// Margins holds page margin configuration
type Margins struct {
	Top    float64 `json:"top"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
	Right  float64 `json:"right"`
}

// Settings holds plugin settings
type Settings struct {
	ID             int64          `json:"id"`
	OrgID          int64          `json:"org_id"`
	UseGrafanaSMTP bool           `json:"use_grafana_smtp"`
	SMTPConfig     *SMTPConfig    `json:"smtp_config,omitempty"`
	RendererConfig RendererConfig `json:"renderer_config"`
	Limits         Limits         `json:"limits"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
	UseTLS   bool   `json:"use_tls"`
}

// RendererConfig holds renderer configuration
type RendererConfig struct {
	URL               string  `json:"url"`                 // DEPRECATED: renderer service URL (kept for backward compatibility)
	TimeoutMS         int     `json:"timeout_ms"`
	DelayMS           int     `json:"delay_ms"`
	ViewportWidth     int     `json:"viewport_width"`
	ViewportHeight    int     `json:"viewport_height"`
	DeviceScaleFactor float64 `json:"device_scale_factor"` // Higher values (2-4) increase image quality
	SkipTLSVerify     bool    `json:"skip_tls_verify"`     // Skip TLS certificate verification

	// Chromium-specific configuration
	ChromiumPath      string  `json:"chromium_path"`       // Path to Chrome/Chromium binary (optional, auto-detect if empty)
	Headless          bool    `json:"headless"`            // Run in headless mode (default: true)
	DisableGPU        bool    `json:"disable_gpu"`         // Disable GPU acceleration for server environments
	NoSandbox         bool    `json:"no_sandbox"`          // Disable sandbox (needed for Docker)
}

// Limits holds usage limits
type Limits struct {
	MaxRecipients        int `json:"max_recipients"`
	MaxAttachmentSizeMB  int `json:"max_attachment_size_mb"`
	MaxConcurrentRenders int `json:"max_concurrent_renders"`
	RetentionDays        int `json:"retention_days"`
}

// JSONMap is a custom type for storing JSON key-value pairs in SQLite
type JSONMap map[string]string

// Scan implements sql.Scanner for JSONMap
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONMap)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// Value implements driver.Valuer for JSONMap
func (j JSONMap) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.Marshal(j)
}

// IntSlice is a custom type for storing integer slices in SQLite
type IntSlice []int64

// Scan implements sql.Scanner for IntSlice
func (i *IntSlice) Scan(value interface{}) error {
	if value == nil {
		*i = []int64{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, i)
}

// Value implements driver.Valuer for IntSlice
func (i IntSlice) Value() (driver.Value, error) {
	if len(i) == 0 {
		return nil, nil
	}
	return json.Marshal(i)
}

// Scan implements sql.Scanner for Recipients
func (r *Recipients) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, r)
}

// Value implements driver.Valuer for Recipients
func (r Recipients) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// Scan implements sql.Scanner for TemplateConfig
func (t *TemplateConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, t)
}

// Value implements driver.Valuer for TemplateConfig
func (t TemplateConfig) Value() (driver.Value, error) {
	return json.Marshal(t)
}

// Scan implements sql.Scanner for SMTPConfig
func (s *SMTPConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, s)
}

// Value implements driver.Valuer for SMTPConfig
func (s *SMTPConfig) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan implements sql.Scanner for RendererConfig
func (r *RendererConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, r)
}

// Value implements driver.Valuer for RendererConfig
func (r RendererConfig) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// Scan implements sql.Scanner for Limits
func (l *Limits) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, l)
}

// Value implements driver.Valuer for Limits
func (l Limits) Value() (driver.Value, error) {
	return json.Marshal(l)
}
