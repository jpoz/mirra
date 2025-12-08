package grouping

import (
	"strings"

	"github.com/jpoz/mirra/internal/recorder"
)

// extractTraceID extracts the trace ID from Sentry-Trace header
// Format: "trace_id-span_id" -> returns "trace_id"
func extractTraceID(rec *recorder.Recording) string {
	if rec == nil || rec.Request.Headers == nil {
		return ""
	}

	// Check for Sentry-Trace header (case-insensitive)
	for key, values := range rec.Request.Headers {
		if strings.EqualFold(key, "Sentry-Trace") && len(values) > 0 {
			// Format: "41cb435ca2a6434b913b733d81c463ae-span123"
			parts := strings.Split(values[0], "-")
			if len(parts) >= 1 && parts[0] != "" {
				return parts[0]
			}
		}
	}

	return ""
}

// extractSessionID extracts the session UUID from request body metadata
// Format: "user_{hash}_account_{uuid}_session_{uuid}" -> returns session UUID
func extractSessionID(rec *recorder.Recording) string {
	if rec == nil || rec.Request.Body == nil {
		return ""
	}

	// Navigate: body.metadata.user_id
	bodyMap, ok := rec.Request.Body.(map[string]interface{})
	if !ok {
		return ""
	}

	metadata, ok := bodyMap["metadata"].(map[string]interface{})
	if !ok {
		return ""
	}

	userID, ok := metadata["user_id"].(string)
	if !ok {
		return ""
	}

	// Parse: "user_..._session_{uuid}"
	parts := strings.Split(userID, "_session_")
	if len(parts) == 2 && parts[1] != "" {
		return parts[1]
	}

	return ""
}

// extractGroupKey returns the primary grouping key for a recording
// Priority: 1. Sentry-Trace ID, 2. Session ID
// Returns the key and a boolean indicating if it's a trace ID (true) or session ID (false)
func extractGroupKey(rec *recorder.Recording) (string, bool) {
	traceID := extractTraceID(rec)
	if traceID != "" {
		return traceID, true
	}

	sessionID := extractSessionID(rec)
	if sessionID != "" {
		return sessionID, false
	}

	return "", false
}

// containsString checks if a string slice contains a value
func containsString(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
