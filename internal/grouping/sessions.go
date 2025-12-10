package grouping

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jpoz/mirra/internal/recorder"
)

const (
	SessionIndexVersion   = "1.0"
	MaxRecordingsPerGroup = 1000
	SessionIndexFilename  = "sessions.json"
)

// SessionGroupIndex manages session-based grouping of recordings
type SessionGroupIndex struct {
	Version     string                   `json:"version"`
	GeneratedAt time.Time                `json:"generated_at"`
	TotalGroups int                      `json:"total_groups"`
	Groups      map[string]*SessionGroup `json:"groups"` // key: trace_id or session_id

	// In-memory indexes (not persisted)
	bySessionID   map[string]*SessionGroup
	byRecordingID map[string]string // recording_id -> group_key

	path      string
	mu        sync.RWMutex
	dirty     bool
	lastSave  time.Time
	saveCount int
}

// NewSessionGroupIndex creates a new session group index
func NewSessionGroupIndex(recordingsPath string) *SessionGroupIndex {
	groupsPath := filepath.Join(recordingsPath, "groups")
	if err := os.MkdirAll(groupsPath, 0755); err != nil {
		slog.Error("failed to create groups directory", "error", err, "path", groupsPath)
	}

	return &SessionGroupIndex{
		Version:       SessionIndexVersion,
		GeneratedAt:   time.Now(),
		Groups:        make(map[string]*SessionGroup),
		bySessionID:   make(map[string]*SessionGroup),
		byRecordingID: make(map[string]string),
		path:          filepath.Join(groupsPath, SessionIndexFilename),
		lastSave:      time.Now(),
	}
}

// Load reads the session index from disk
func (idx *SessionGroupIndex) Load() error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	data, err := os.ReadFile(idx.path)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Info("session index not found, starting fresh")
			return nil
		}
		return fmt.Errorf("failed to read session index: %w", err)
	}

	var loaded SessionGroupIndex
	if err := json.Unmarshal(data, &loaded); err != nil {
		slog.Error("failed to parse session index, starting fresh", "error", err)
		return nil
	}

	if loaded.Version != SessionIndexVersion {
		slog.Warn("session index version mismatch, starting fresh",
			"current", SessionIndexVersion,
			"found", loaded.Version)
		return nil
	}

	// Copy loaded data
	idx.Version = loaded.Version
	idx.GeneratedAt = loaded.GeneratedAt
	idx.TotalGroups = loaded.TotalGroups
	idx.Groups = loaded.Groups

	// Rebuild in-memory indexes
	idx.bySessionID = make(map[string]*SessionGroup)
	idx.byRecordingID = make(map[string]string)
	for key, group := range idx.Groups {
		if group.SessionID != "" {
			idx.bySessionID[group.SessionID] = group
		}
		for _, recID := range group.RecordingIDs {
			idx.byRecordingID[recID] = key
		}
	}

	slog.Info("session index loaded",
		"groups", idx.TotalGroups,
		"recordings", len(idx.byRecordingID))

	return nil
}

// Save persists the session index to disk
func (idx *SessionGroupIndex) Save() error {
	idx.mu.RLock()
	if !idx.dirty {
		idx.mu.RUnlock()
		return nil
	}
	idx.mu.RUnlock()

	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.GeneratedAt = time.Now()

	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session index: %w", err)
	}

	// Atomic write: write to temp file, then rename
	tempPath := idx.path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write session index: %w", err)
	}

	if err := os.Rename(tempPath, idx.path); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename session index: %w", err)
	}

	idx.dirty = false
	idx.lastSave = time.Now()

	return nil
}

// AddRecording adds a recording to the appropriate session group
func (idx *SessionGroupIndex) AddRecording(rec *recorder.Recording) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Extract group key (trace ID or session ID)
	groupKey, isTraceID := extractGroupKey(rec)
	if groupKey == "" {
		// No session info available
		return nil
	}

	// Find or create group
	group, exists := idx.Groups[groupKey]
	if !exists {
		group = &SessionGroup{
			TraceID:        "",
			SessionID:      "",
			RecordingIDs:   []string{},
			FirstTimestamp: rec.Timestamp,
			Providers:      []string{},
		}

		// Set the appropriate ID field
		if isTraceID {
			group.TraceID = groupKey
			// Try to extract session ID as well
			sessionID := extractSessionID(rec)
			if sessionID != "" {
				group.SessionID = sessionID
				idx.bySessionID[sessionID] = group
			}
		} else {
			group.SessionID = groupKey
			idx.bySessionID[groupKey] = group
		}

		idx.Groups[groupKey] = group
		idx.TotalGroups++
	}

	// Check group size limit
	if len(group.RecordingIDs) >= MaxRecordingsPerGroup {
		return fmt.Errorf("group %s size limit exceeded (%d recordings)", groupKey, MaxRecordingsPerGroup)
	}

	// Update group
	group.RecordingIDs = append(group.RecordingIDs, rec.ID)
	group.LastTimestamp = rec.Timestamp
	group.RequestCount++

	if !containsString(group.Providers, rec.Provider) {
		group.Providers = append(group.Providers, rec.Provider)
	}

	if rec.Error != "" || rec.Response.Status >= 400 {
		group.HasErrors = true
	}

	// Update lookup map
	idx.byRecordingID[rec.ID] = groupKey

	idx.dirty = true
	idx.saveCount++

	return nil
}

// GetGroupByTraceID returns a session group by trace ID
func (idx *SessionGroupIndex) GetGroupByTraceID(traceID string) (*SessionGroup, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	group, exists := idx.Groups[traceID]
	if !exists {
		return nil, fmt.Errorf("group not found: %s", traceID)
	}

	return group, nil
}

// GetGroupBySessionID returns a session group by session ID
func (idx *SessionGroupIndex) GetGroupBySessionID(sessionID string) (*SessionGroup, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	group, exists := idx.bySessionID[sessionID]
	if !exists {
		return nil, fmt.Errorf("group not found for session: %s", sessionID)
	}

	return group, nil
}

// GetGroupByRecordingID returns the session group containing a recording
func (idx *SessionGroupIndex) GetGroupByRecordingID(recordingID string) (*SessionGroup, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	groupKey, exists := idx.byRecordingID[recordingID]
	if !exists {
		return nil, fmt.Errorf("recording not in any group: %s", recordingID)
	}

	group := idx.Groups[groupKey]
	return group, nil
}

// ListGroups returns all session groups with optional filtering
func (idx *SessionGroupIndex) ListGroups(opts *ListGroupsOptions) ([]*SessionGroup, int) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if opts == nil {
		opts = &ListGroupsOptions{}
	}

	// Collect and filter groups
	var filtered []*SessionGroup
	for _, group := range idx.Groups {
		// Apply filters
		if opts.FromDate != nil && group.LastTimestamp.Before(*opts.FromDate) {
			continue
		}
		if opts.ToDate != nil && group.FirstTimestamp.After(*opts.ToDate) {
			continue
		}
		if opts.Provider != "" && !containsString(group.Providers, opts.Provider) {
			continue
		}
		if opts.HasErrors != nil && group.HasErrors != *opts.HasErrors {
			continue
		}

		filtered = append(filtered, group)
	}

	total := len(filtered)

	// Sort by last timestamp (newest first)
	for i := 0; i < len(filtered); i++ {
		for j := i + 1; j < len(filtered); j++ {
			if filtered[i].LastTimestamp.Before(filtered[j].LastTimestamp) {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}

	// Apply pagination
	page := opts.Page
	if page < 1 {
		page = 1
	}
	limit := opts.Limit
	if limit < 1 {
		limit = 50
	}

	start := (page - 1) * limit
	end := start + limit

	if start >= len(filtered) {
		return []*SessionGroup{}, total
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end], total
}

// ShouldSave returns true if the index should be persisted
func (idx *SessionGroupIndex) ShouldSave() bool {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if !idx.dirty {
		return false
	}

	// Save every 100 recordings or every 60 seconds
	return idx.saveCount >= 100 || time.Since(idx.lastSave) >= 60*time.Second
}

// ListGroupsOptions configures group listing behavior
type ListGroupsOptions struct {
	Page      int
	Limit     int
	FromDate  *time.Time
	ToDate    *time.Time
	Provider  string
	HasErrors *bool
}
