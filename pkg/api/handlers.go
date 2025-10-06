package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"
	"github.com/yourusername/sheduled-reports-app/pkg/cron"
	"github.com/yourusername/sheduled-reports-app/pkg/model"
	"github.com/yourusername/sheduled-reports-app/pkg/store"
)

// Handler handles HTTP API requests
type Handler struct {
	store         *store.Store
	scheduler     *cron.Scheduler
	mux           *http.ServeMux
	contextCached bool
}

// NewHandler creates a new API handler
func NewHandler(st *store.Store, scheduler *cron.Scheduler) *Handler {
	h := &Handler{
		store:         st,
		scheduler:     scheduler,
		mux:           http.NewServeMux(),
		contextCached: false,
	}

	h.registerRoutes()
	return h
}

// registerRoutes registers all HTTP routes
func (h *Handler) registerRoutes() {
	h.mux.HandleFunc("/api/schedules", h.handleSchedules)
	h.mux.HandleFunc("/api/schedules/", h.handleSchedule)
	h.mux.HandleFunc("/api/runs/", h.handleRun)
	h.mux.HandleFunc("/api/settings", h.handleSettings)
}

// CallResource implements backend.CallResourceHandler
func (h *Handler) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	// Cache the Grafana config context for background scheduler jobs
	// Only do this once on first request
	if !h.contextCached {
		h.scheduler.SetContext(ctx)
		h.contextCached = true
		log.Println("Cached Grafana config context for scheduler")
	}

	adapter := httpadapter.New(h.mux)
	return adapter.CallResource(ctx, req, sender)
}

// handleSchedules handles GET /api/schedules and POST /api/schedules
func (h *Handler) handleSchedules(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgID(r)

	switch r.Method {
	case http.MethodGet:
		schedules, err := h.store.ListSchedules(orgID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, map[string]interface{}{"schedules": schedules})

	case http.MethodPost:
		var schedule model.Schedule
		if err := json.NewDecoder(r.Body).Decode(&schedule); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		schedule.OrgID = orgID
		schedule.OwnerUserID = getUserID(r)

		// Calculate and set next run time
		nextRun := h.scheduler.CalculateNextRun(&schedule)
		schedule.NextRunAt = &nextRun

		if err := h.store.CreateSchedule(&schedule); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respondJSON(w, schedule)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleSchedule handles operations on a specific schedule
func (h *Handler) handleSchedule(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgID(r)
	path := r.URL.Path

	// Parse schedule ID and action from path
	var scheduleID int64
	var action string

	// Path format: /api/schedules/{id} or /api/schedules/{id}/runs or /api/schedules/{id}/run
	if _, err := fmt.Sscanf(path, "/api/schedules/%d/%s", &scheduleID, &action); err != nil {
		// Try without action
		if _, err := fmt.Sscanf(path, "/api/schedules/%d", &scheduleID); err != nil {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}
	}

	// Handle actions
	if action == "run" && r.Method == http.MethodPost {
		schedule, err := h.store.GetSchedule(orgID, scheduleID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		h.scheduler.ExecuteSchedule(schedule)
		respondJSON(w, map[string]string{"status": "started"})
		return
	}

	if action == "runs" && r.Method == http.MethodGet {
		runs, err := h.store.ListRuns(orgID, scheduleID)
		if err != nil {
			fmt.Printf("Error loading runs for schedule %d, org %d: %v\n", scheduleID, orgID, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, map[string]interface{}{"runs": runs})
		return
	}

	// Handle CRUD operations
	switch r.Method {
	case http.MethodGet:
		schedule, err := h.store.GetSchedule(orgID, scheduleID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		respondJSON(w, schedule)

	case http.MethodPut:
		var schedule model.Schedule
		if err := json.NewDecoder(r.Body).Decode(&schedule); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		schedule.ID = scheduleID
		schedule.OrgID = orgID

		// Recalculate next run time if interval or cron expression changed
		nextRun := h.scheduler.CalculateNextRun(&schedule)
		schedule.NextRunAt = &nextRun

		if err := h.store.UpdateSchedule(&schedule); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respondJSON(w, schedule)

	case http.MethodDelete:
		if err := h.store.DeleteSchedule(orgID, scheduleID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleRun handles run-related operations
func (h *Handler) handleRun(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgID(r)
	path := r.URL.Path

	var runID int64
	var action string

	// Path format: /api/runs/{id}/artifact
	if _, err := fmt.Sscanf(path, "/api/runs/%d/%s", &runID, &action); err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	if action == "artifact" && r.Method == http.MethodGet {
		run, err := h.store.GetRun(orgID, runID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if run.ArtifactPath == "" {
			http.Error(w, "Artifact not found", http.StatusNotFound)
			return
		}

		file, err := os.Open(run.ArtifactPath)
		if err != nil {
			http.Error(w, "Failed to open artifact", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// Set content type based on file extension
		contentType := "application/pdf"
		if run.ArtifactPath[len(run.ArtifactPath)-4:] == ".png" {
			contentType = "image/png"
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", run.ArtifactPath))
		io.Copy(w, file)
		return
	}

	http.Error(w, "Invalid action", http.StatusBadRequest)
}

// handleSettings handles settings operations
func (h *Handler) handleSettings(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgID(r)

	switch r.Method {
	case http.MethodGet:
		settings, err := h.store.GetSettings(orgID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if settings == nil {
			// Return default settings
			settings = &model.Settings{
				OrgID:          orgID,
				UseGrafanaSMTP: true,
				RendererConfig: model.RendererConfig{
					URL:            "http://renderer:8081/render",
					TimeoutMS:      60000,
					DelayMS:        1000,
					ViewportWidth:  1920,
					ViewportHeight: 1080,
				},
				Limits: model.Limits{
					MaxRecipients:        50,
					MaxAttachmentSizeMB:  25,
					MaxConcurrentRenders: 5,
					RetentionDays:        30,
				},
			}
		}
		respondJSON(w, settings)

	case http.MethodPost:
		var settings model.Settings
		if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		settings.OrgID = orgID

		if err := h.store.UpsertSettings(&settings); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respondJSON(w, settings)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Helper functions

func getOrgID(r *http.Request) int64 {
	// In a real plugin, this would come from the Grafana request context
	// For now, we'll try to get it from a header or default to 1
	orgIDStr := r.Header.Get("X-Grafana-Org-Id")
	if orgIDStr == "" {
		return 1
	}
	orgID, _ := strconv.ParseInt(orgIDStr, 10, 64)
	return orgID
}

func getUserID(r *http.Request) int64 {
	// In a real plugin, this would come from the Grafana request context
	userIDStr := r.Header.Get("X-Grafana-User-Id")
	if userIDStr == "" {
		return 1
	}
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)
	return userID
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
