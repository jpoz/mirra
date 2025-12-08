package grouping

import (
	"time"

	"github.com/jpoz/mirra/internal/recorder"
)

// SimilarityLevel represents different levels of similarity matching
type SimilarityLevel int

const (
	LevelSession SimilarityLevel = iota
	LevelSystemPrompt
	LevelTools
	LevelContent
)

// SessionGroup represents a group of recordings from the same session
type SessionGroup struct {
	TraceID        string    `json:"trace_id"`
	SessionID      string    `json:"session_id"`
	RecordingIDs   []string  `json:"recording_ids"`
	FirstTimestamp time.Time `json:"first_timestamp"`
	LastTimestamp  time.Time `json:"last_timestamp"`
	RequestCount   int       `json:"request_count"`
	Providers      []string  `json:"providers"`
	HasErrors      bool      `json:"has_errors"`
}

// SimilarityOptions configures similarity search behavior
type SimilarityOptions struct {
	Threshold float64  // For content similarity (0.0-1.0)
	Limit     int      // Max results to return
	Providers []string // Filter by providers
}

// SimilarMatch represents a recording similar to the query
type SimilarMatch struct {
	RecordingID      string
	SimilarityScore  float64
	SimilarityReason string
	Recording        *recorder.Recording
}
