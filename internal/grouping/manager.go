package grouping

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/jpoz/mirra/internal/recorder"
)

// Manager coordinates all group indexes and provides unified access
type Manager struct {
	sessions *SessionGroupIndex

	recordingsPath string
	enabled        bool
	mu             sync.RWMutex
}

// NewManager creates a new grouping manager
func NewManager(recordingsPath string, enabled bool) *Manager {
	m := &Manager{
		recordingsPath: recordingsPath,
		enabled:        enabled,
	}

	if enabled {
		m.sessions = NewSessionGroupIndex(recordingsPath)

		// Load existing index
		if err := m.sessions.Load(); err != nil {
			slog.Error("failed to load session index", "error", err)
		}
	}

	return m
}

// OnRecordingWrite is called by the recorder after writing a recording
// This is the main entry point for adding recordings to group indexes
func (m *Manager) OnRecordingWrite(rec *recorder.Recording) error {
	if !m.enabled {
		return nil
	}

	// Add to session index (synchronous - fast operation)
	if err := m.sessions.AddRecording(rec); err != nil {
		slog.Error("failed to add recording to session index",
			"recording_id", rec.ID,
			"error", err)
		// Don't fail the recording write if grouping fails
	}

	// Check if we should persist
	if m.sessions.ShouldSave() {
		if err := m.sessions.Save(); err != nil {
			slog.Error("failed to save session index", "error", err)
		}
	}

	return nil
}

// GetSessionGroup returns a session group by trace ID
func (m *Manager) GetSessionGroup(traceID string) (*SessionGroup, error) {
	if !m.enabled {
		return nil, fmt.Errorf("grouping is disabled")
	}

	return m.sessions.GetGroupByTraceID(traceID)
}

// GetSessionGroupBySessionID returns a session group by session ID
func (m *Manager) GetSessionGroupBySessionID(sessionID string) (*SessionGroup, error) {
	if !m.enabled {
		return nil, fmt.Errorf("grouping is disabled")
	}

	return m.sessions.GetGroupBySessionID(sessionID)
}

// GetSessionGroupForRecording returns the session group containing a recording
func (m *Manager) GetSessionGroupForRecording(recordingID string) (*SessionGroup, error) {
	if !m.enabled {
		return nil, fmt.Errorf("grouping is disabled")
	}

	return m.sessions.GetGroupByRecordingID(recordingID)
}

// ListSessionGroups returns a list of session groups with optional filtering
func (m *Manager) ListSessionGroups(opts *ListGroupsOptions) ([]*SessionGroup, int) {
	if !m.enabled {
		return []*SessionGroup{}, 0
	}

	return m.sessions.ListGroups(opts)
}

// Close saves all indexes and cleans up resources
func (m *Manager) Close() error {
	if !m.enabled {
		return nil
	}

	// Save session index
	if err := m.sessions.Save(); err != nil {
		return fmt.Errorf("failed to save session index: %w", err)
	}

	slog.Info("grouping manager closed")
	return nil
}

// Enabled returns whether grouping is enabled
func (m *Manager) Enabled() bool {
	return m.enabled
}
