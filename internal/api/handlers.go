package api

import (
	"bufio"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/llmite-ai/mirra/internal/config"
	"github.com/llmite-ai/mirra/internal/recorder"
)

// RecordingListResponse represents the API response for listing recordings
type RecordingListResponse struct {
	Recordings []RecordingSummary `json:"recordings"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	HasMore    bool               `json:"hasMore"`
}

// RecordingSummary represents a summary of a recording for list view
type RecordingSummary struct {
	ID           string    `json:"id"`
	Timestamp    time.Time `json:"timestamp"`
	Provider     string    `json:"provider"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	Status       int       `json:"status"`
	Duration     int64     `json:"duration"`
	ResponseSize int64     `json:"responseSize"`
	Error        string    `json:"error,omitempty"`
}

// Handlers contains the API handler methods
type Handlers struct {
	cfg *config.Config
	log *slog.Logger
}

// NewHandlers creates a new API handlers instance
func NewHandlers(cfg *config.Config, log *slog.Logger) *Handlers {
	return &Handlers{
		cfg: cfg,
		log: log,
	}
}

// ListRecordings handles GET /api/recordings
func (h *Handlers) ListRecordings(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	page := parseInt(query.Get("page"), 1)
	limit := parseInt(query.Get("limit"), 50)
	provider := query.Get("provider")
	fromDate := query.Get("from")
	toDate := query.Get("to")
	search := strings.TrimSpace(query.Get("search"))

	// Read all recordings
	recordings, err := h.readAllRecordings(fromDate, toDate)
	if err != nil {
		h.log.Error("Failed to read recordings", "error", err)
		http.Error(w, "Failed to read recordings", http.StatusInternalServerError)
		return
	}

	// Filter recordings
	filtered := h.filterRecordings(recordings, provider, search)

	// Sort by timestamp descending (newest first)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.After(filtered[j].Timestamp)
	})

	// Apply pagination
	total := len(filtered)
	start := (page - 1) * limit
	if start >= total {
		start = 0
	}
	end := start + limit
	if end > total {
		end = total
	}

	paginated := filtered[start:end]
	hasMore := end < total

	// Convert to summaries
	summaries := make([]RecordingSummary, len(paginated))
	for i, rec := range paginated {
		summaries[i] = h.recordingToSummary(rec)
	}

	response := RecordingListResponse{
		Recordings: summaries,
		Total:      total,
		Page:       page,
		Limit:      limit,
		HasMore:    hasMore,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetRecording handles GET /api/recordings/:id
func (h *Handlers) GetRecording(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/recordings/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Recording ID required", http.StatusBadRequest)
		return
	}
	id := parts[0]

	// Read all recordings to find the one with matching ID
	recordings, err := h.readAllRecordings("", "")
	if err != nil {
		h.log.Error("Failed to read recordings", "error", err)
		http.Error(w, "Failed to read recordings", http.StatusInternalServerError)
		return
	}

	// Find recording by ID (support prefix matching)
	var found *recorder.Recording
	for i := range recordings {
		if strings.HasPrefix(recordings[i].ID, id) {
			found = &recordings[i]
			break
		}
	}

	if found == nil {
		http.Error(w, "Recording not found", http.StatusNotFound)
		return
	}

	// Redact sensitive data
	redacted := h.redactRecording(*found)

	// Decompress gzipped response body if needed
	if body, ok := redacted.Response.Body.(string); ok {
		if strings.HasPrefix(body, "base64:") {
			decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(body, "base64:"))
			if err == nil {
				gr, err := gzip.NewReader(strings.NewReader(string(decoded)))
				if err == nil {
					decompressed, err := io.ReadAll(gr)
					gr.Close()
					if err == nil {
						// Try to parse as JSON
						var jsonBody interface{}
						if json.Unmarshal(decompressed, &jsonBody) == nil {
							redacted.Response.Body = jsonBody
						} else {
							redacted.Response.Body = string(decompressed)
						}
					}
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(redacted)
}

// readAllRecordings reads all recordings from JSONL files
func (h *Handlers) readAllRecordings(fromDate, toDate string) ([]recorder.Recording, error) {
	recordingsPath := h.cfg.Recording.Path
	if recordingsPath == "" {
		recordingsPath = "./recordings"
	}

	var recordings []recorder.Recording

	// Parse date range
	var from, to time.Time
	if fromDate != "" {
		if parsed, err := time.Parse("2006-01-02", fromDate); err == nil {
			from = parsed
		}
	}
	if toDate != "" {
		if parsed, err := time.Parse("2006-01-02", toDate); err == nil {
			to = parsed.Add(24 * time.Hour) // Include the entire day
		}
	}

	// Read directory
	entries, err := os.ReadDir(recordingsPath)
	if err != nil {
		return nil, err
	}

	// Process each JSONL file
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}

		// Parse date from filename (recordings-YYYY-MM-DD.jsonl)
		parts := strings.Split(strings.TrimSuffix(entry.Name(), ".jsonl"), "-")
		if len(parts) < 4 {
			continue
		}
		fileDate := strings.Join(parts[1:], "-")
		fileTime, err := time.Parse("2006-01-02", fileDate)
		if err != nil {
			continue
		}

		// Skip if outside date range
		if !from.IsZero() && fileTime.Before(from) {
			continue
		}
		if !to.IsZero() && fileTime.After(to) {
			continue
		}

		// Read file
		filePath := filepath.Join(recordingsPath, entry.Name())
		fileRecordings, err := h.readRecordingsFromFile(filePath)
		if err != nil {
			h.log.Error("Failed to read recordings file", "file", entry.Name(), "error", err)
			continue
		}

		recordings = append(recordings, fileRecordings...)
	}

	return recordings, nil
}

// readRecordingsFromFile reads recordings from a single JSONL file
func (h *Handlers) readRecordingsFromFile(path string) ([]recorder.Recording, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var recordings []recorder.Recording
	scanner := bufio.NewScanner(file)

	// Increase buffer size for large recordings (up to 10MB per line)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var rec recorder.Recording
		if err := json.Unmarshal(line, &rec); err != nil {
			h.log.Error("Failed to parse recording", "error", err)
			continue
		}

		recordings = append(recordings, rec)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return recordings, nil
}

// filterRecordings filters recordings by provider and search term
func (h *Handlers) filterRecordings(recordings []recorder.Recording, provider, search string) []recorder.Recording {
	filtered := make([]recorder.Recording, 0, len(recordings))

	for _, rec := range recordings {
		// Filter by provider
		if provider != "" && !strings.EqualFold(rec.Provider, provider) {
			continue
		}

		// Filter by search term (ID, path, or error)
		if search != "" {
			searchLower := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(rec.ID), searchLower) &&
				!strings.Contains(strings.ToLower(rec.Request.Path), searchLower) &&
				!strings.Contains(strings.ToLower(rec.Error), searchLower) {
				continue
			}
		}

		filtered = append(filtered, rec)
	}

	return filtered
}

// recordingToSummary converts a full recording to a summary
func (h *Handlers) recordingToSummary(rec recorder.Recording) RecordingSummary {
	return RecordingSummary{
		ID:           rec.ID,
		Timestamp:    rec.Timestamp,
		Provider:     rec.Provider,
		Method:       rec.Request.Method,
		Path:         rec.Request.Path,
		Status:       rec.Response.Status,
		Duration:     rec.Timing.DurationMs,
		ResponseSize: rec.ResponseSize,
		Error:        rec.Error,
	}
}

// redactRecording removes sensitive data from a recording
func (h *Handlers) redactRecording(rec recorder.Recording) recorder.Recording {
	// Clone headers to avoid modifying original
	reqHeaders := make(map[string][]string)
	for k, v := range rec.Request.Headers {
		if h.isSensitiveHeader(k) {
			reqHeaders[k] = []string{"[REDACTED]"}
		} else {
			reqHeaders[k] = v
		}
	}
	rec.Request.Headers = reqHeaders

	respHeaders := make(map[string][]string)
	for k, v := range rec.Response.Headers {
		if h.isSensitiveHeader(k) {
			respHeaders[k] = []string{"[REDACTED]"}
		} else {
			respHeaders[k] = v
		}
	}
	rec.Response.Headers = respHeaders

	// Redact API keys in request body
	if body, ok := rec.Request.Body.(map[string]interface{}); ok {
		if _, exists := body["api_key"]; exists {
			body["api_key"] = "[REDACTED]"
		}
	}

	return rec
}

// isSensitiveHeader checks if a header contains sensitive data
func (h *Handlers) isSensitiveHeader(name string) bool {
	name = strings.ToLower(name)
	sensitiveHeaders := []string{
		"authorization",
		"x-api-key",
		"api-key",
		"cookie",
		"set-cookie",
	}

	for _, sensitive := range sensitiveHeaders {
		if strings.Contains(name, sensitive) {
			return true
		}
	}

	return false
}

// parseInt safely parses an integer with a default fallback
func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	var val int
	if _, err := fmt.Sscanf(s, "%d", &val); err != nil {
		return defaultVal
	}
	if val < 1 {
		return defaultVal
	}
	return val
}
