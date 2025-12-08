package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jpoz/mirra/internal/grouping"
	"github.com/jpoz/mirra/internal/recorder"
)

// GroupHandlers contains the group-related API handler methods
type GroupHandlers struct {
	log          *slog.Logger
	rec          *recorder.Recorder
	groupManager *grouping.Manager
}

// NewGroupHandlers creates a new group handlers instance
func NewGroupHandlers(log *slog.Logger, rec *recorder.Recorder, groupMgr *grouping.Manager) *GroupHandlers {
	return &GroupHandlers{
		log:          log,
		rec:          rec,
		groupManager: groupMgr,
	}
}

// SessionGroupResponse represents a session group with summary info
type SessionGroupResponse struct {
	TraceID        string    `json:"trace_id"`
	SessionID      string    `json:"session_id"`
	RecordingIDs   []string  `json:"recording_ids"`
	FirstTimestamp time.Time `json:"first_timestamp"`
	LastTimestamp  time.Time `json:"last_timestamp"`
	RequestCount   int       `json:"request_count"`
	Providers      []string  `json:"providers"`
	HasErrors      bool      `json:"has_errors"`
}

// SessionGroupListResponse represents the API response for listing session groups
type SessionGroupListResponse struct {
	Groups  []SessionGroupResponse `json:"groups"`
	Total   int                    `json:"total"`
	Page    int                    `json:"page"`
	Limit   int                    `json:"limit"`
	HasMore bool                   `json:"hasMore"`
}

// SessionGroupDetailResponse represents a session group with full recording data
type SessionGroupDetailResponse struct {
	Group      SessionGroupResponse `json:"group"`
	Recordings []RecordingSummary   `json:"recordings"`
}

// ListSessionGroups handles GET /api/groups/sessions
func (h *GroupHandlers) ListSessionGroups(w http.ResponseWriter, r *http.Request) {
	if h.groupManager == nil || !h.groupManager.Enabled() {
		http.Error(w, "Grouping is not enabled", http.StatusNotImplemented)
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	page := parseInt(query.Get("page"), 1)
	limit := parseInt(query.Get("limit"), 50)
	provider := query.Get("provider")
	fromDateStr := query.Get("from")
	toDateStr := query.Get("to")
	hasErrorsStr := query.Get("has_errors")

	// Parse optional filters
	opts := &grouping.ListGroupsOptions{
		Page:     page,
		Limit:    limit,
		Provider: provider,
	}

	if fromDateStr != "" {
		if fromDate, err := time.Parse("2006-01-02", fromDateStr); err == nil {
			opts.FromDate = &fromDate
		}
	}

	if toDateStr != "" {
		if toDate, err := time.Parse("2006-01-02", toDateStr); err == nil {
			opts.ToDate = &toDate
		}
	}

	if hasErrorsStr != "" {
		hasErrors := hasErrorsStr == "true"
		opts.HasErrors = &hasErrors
	}

	// Get session groups
	groups, total := h.groupManager.ListSessionGroups(opts)

	// Convert to response format
	respGroups := make([]SessionGroupResponse, len(groups))
	for i, group := range groups {
		respGroups[i] = SessionGroupResponse{
			TraceID:        group.TraceID,
			SessionID:      group.SessionID,
			RecordingIDs:   group.RecordingIDs,
			FirstTimestamp: group.FirstTimestamp,
			LastTimestamp:  group.LastTimestamp,
			RequestCount:   group.RequestCount,
			Providers:      group.Providers,
			HasErrors:      group.HasErrors,
		}
	}

	resp := SessionGroupListResponse{
		Groups:  respGroups,
		Total:   total,
		Page:    page,
		Limit:   limit,
		HasMore: page*limit < total,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.log.Error("Failed to encode response", "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// GetSessionGroup handles GET /api/groups/sessions/{trace_id}
func (h *GroupHandlers) GetSessionGroup(w http.ResponseWriter, r *http.Request) {
	if h.groupManager == nil || !h.groupManager.Enabled() {
		http.Error(w, "Grouping is not enabled", http.StatusNotImplemented)
		return
	}

	// Extract trace_id from path
	traceID := strings.TrimPrefix(r.URL.Path, "/api/groups/sessions/")
	if traceID == "" {
		http.Error(w, "Trace ID is required", http.StatusBadRequest)
		return
	}

	// Get session group
	group, err := h.groupManager.GetSessionGroup(traceID)
	if err != nil {
		h.log.Error("Failed to get session group", "trace_id", traceID, "error", err)
		http.Error(w, fmt.Sprintf("Session group not found: %v", err), http.StatusNotFound)
		return
	}

	// Load full recordings
	recordings := make([]RecordingSummary, 0, len(group.RecordingIDs))
	index := h.rec.GetIndex()

	for _, recID := range group.RecordingIDs {
		rec, err := index.ReadRecording(recID)
		if err != nil {
			h.log.Error("Failed to read recording", "id", recID, "error", err)
			continue
		}

		recordings = append(recordings, RecordingSummary{
			ID:           rec.ID,
			Timestamp:    rec.Timestamp,
			Provider:     rec.Provider,
			Method:       rec.Request.Method,
			Path:         rec.Request.Path,
			Status:       rec.Response.Status,
			Duration:     rec.Timing.DurationMs,
			ResponseSize: rec.ResponseSize,
			Error:        rec.Error,
		})
	}

	resp := SessionGroupDetailResponse{
		Group: SessionGroupResponse{
			TraceID:        group.TraceID,
			SessionID:      group.SessionID,
			RecordingIDs:   group.RecordingIDs,
			FirstTimestamp: group.FirstTimestamp,
			LastTimestamp:  group.LastTimestamp,
			RequestCount:   group.RequestCount,
			Providers:      group.Providers,
			HasErrors:      group.HasErrors,
		},
		Recordings: recordings,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.log.Error("Failed to encode response", "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
