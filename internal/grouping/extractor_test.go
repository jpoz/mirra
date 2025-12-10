package grouping

import (
	"testing"
	"time"

	"github.com/jpoz/mirra/internal/recorder"
)

func TestExtractTraceID(t *testing.T) {
	tests := []struct {
		name string
		rec  *recorder.Recording
		want string
	}{
		{
			name: "valid sentry-trace header",
			rec: &recorder.Recording{
				Request: recorder.RequestData{
					Headers: map[string][]string{
						"Sentry-Trace": {"41cb435ca2a6434b913b733d81c463ae-span123"},
					},
				},
			},
			want: "41cb435ca2a6434b913b733d81c463ae",
		},
		{
			name: "case insensitive header",
			rec: &recorder.Recording{
				Request: recorder.RequestData{
					Headers: map[string][]string{
						"sentry-trace": {"abc123def456-span789"},
					},
				},
			},
			want: "abc123def456",
		},
		{
			name: "missing header",
			rec: &recorder.Recording{
				Request: recorder.RequestData{
					Headers: map[string][]string{},
				},
			},
			want: "",
		},
		{
			name: "empty trace id",
			rec: &recorder.Recording{
				Request: recorder.RequestData{
					Headers: map[string][]string{
						"Sentry-Trace": {"-span123"},
					},
				},
			},
			want: "",
		},
		{
			name: "nil recording",
			rec:  nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTraceID(tt.rec)
			if got != tt.want {
				t.Errorf("extractTraceID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractSessionID(t *testing.T) {
	tests := []struct {
		name string
		rec  *recorder.Recording
		want string
	}{
		{
			name: "valid session id",
			rec: &recorder.Recording{
				Request: recorder.RequestData{
					Body: map[string]interface{}{
						"metadata": map[string]interface{}{
							"user_id": "user_abc123_account_def456_session_c593e22f-34d1-4dee-9937-d718f1e95aec",
						},
					},
				},
			},
			want: "c593e22f-34d1-4dee-9937-d718f1e95aec",
		},
		{
			name: "missing metadata",
			rec: &recorder.Recording{
				Request: recorder.RequestData{
					Body: map[string]interface{}{},
				},
			},
			want: "",
		},
		{
			name: "missing user_id",
			rec: &recorder.Recording{
				Request: recorder.RequestData{
					Body: map[string]interface{}{
						"metadata": map[string]interface{}{},
					},
				},
			},
			want: "",
		},
		{
			name: "malformed user_id",
			rec: &recorder.Recording{
				Request: recorder.RequestData{
					Body: map[string]interface{}{
						"metadata": map[string]interface{}{
							"user_id": "user_abc123",
						},
					},
				},
			},
			want: "",
		},
		{
			name: "nil body",
			rec: &recorder.Recording{
				Request: recorder.RequestData{
					Body: nil,
				},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractSessionID(tt.rec)
			if got != tt.want {
				t.Errorf("extractSessionID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractGroupKey(t *testing.T) {
	tests := []struct {
		name          string
		rec           *recorder.Recording
		wantKey       string
		wantIsTraceID bool
	}{
		{
			name: "trace id takes precedence",
			rec: &recorder.Recording{
				Request: recorder.RequestData{
					Headers: map[string][]string{
						"Sentry-Trace": {"trace123-span456"},
					},
					Body: map[string]interface{}{
						"metadata": map[string]interface{}{
							"user_id": "user_abc_account_def_session_session123",
						},
					},
				},
			},
			wantKey:       "trace123",
			wantIsTraceID: true,
		},
		{
			name: "fallback to session id",
			rec: &recorder.Recording{
				Request: recorder.RequestData{
					Headers: map[string][]string{},
					Body: map[string]interface{}{
						"metadata": map[string]interface{}{
							"user_id": "user_abc_account_def_session_session456",
						},
					},
				},
			},
			wantKey:       "session456",
			wantIsTraceID: false,
		},
		{
			name: "no identifiers",
			rec: &recorder.Recording{
				Request: recorder.RequestData{
					Headers: map[string][]string{},
					Body:    map[string]interface{}{},
				},
			},
			wantKey:       "",
			wantIsTraceID: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKey, gotIsTraceID := extractGroupKey(tt.rec)
			if gotKey != tt.wantKey {
				t.Errorf("extractGroupKey() key = %v, want %v", gotKey, tt.wantKey)
			}
			if gotIsTraceID != tt.wantIsTraceID {
				t.Errorf("extractGroupKey() isTraceID = %v, want %v", gotIsTraceID, tt.wantIsTraceID)
			}
		})
	}
}

func TestSessionGroupIndex_AddRecording(t *testing.T) {
	idx := &SessionGroupIndex{
		Version:       SessionIndexVersion,
		Groups:        make(map[string]*SessionGroup),
		bySessionID:   make(map[string]*SessionGroup),
		byRecordingID: make(map[string]string),
		path:          "/tmp/test-sessions.json",
	}

	// Test adding recording with trace ID
	rec1 := &recorder.Recording{
		ID:        "20251203-rec1",
		Timestamp: time.Now(),
		Provider:  "claude",
		Request: recorder.RequestData{
			Headers: map[string][]string{
				"Sentry-Trace": {"trace123-span1"},
			},
			Body: map[string]interface{}{
				"metadata": map[string]interface{}{
					"user_id": "user_abc_account_def_session_sess123",
				},
			},
		},
		Response: recorder.ResponseData{
			Status: 200,
		},
	}

	if err := idx.AddRecording(rec1); err != nil {
		t.Fatalf("AddRecording() error = %v", err)
	}

	// Verify group was created
	if idx.TotalGroups != 1 {
		t.Errorf("TotalGroups = %v, want 1", idx.TotalGroups)
	}

	group, exists := idx.Groups["trace123"]
	if !exists {
		t.Fatal("Group not found for trace123")
	}

	if group.TraceID != "trace123" {
		t.Errorf("TraceID = %v, want trace123", group.TraceID)
	}

	if group.SessionID != "sess123" {
		t.Errorf("SessionID = %v, want sess123", group.SessionID)
	}

	if len(group.RecordingIDs) != 1 || group.RecordingIDs[0] != "20251203-rec1" {
		t.Errorf("RecordingIDs = %v, want [20251203-rec1]", group.RecordingIDs)
	}

	// Add another recording to same session
	rec2 := &recorder.Recording{
		ID:        "20251203-rec2",
		Timestamp: time.Now().Add(1 * time.Minute),
		Provider:  "claude",
		Request: recorder.RequestData{
			Headers: map[string][]string{
				"Sentry-Trace": {"trace123-span2"},
			},
		},
		Response: recorder.ResponseData{
			Status: 200,
		},
	}

	if err := idx.AddRecording(rec2); err != nil {
		t.Fatalf("AddRecording() error = %v", err)
	}

	// Verify same group was updated
	if idx.TotalGroups != 1 {
		t.Errorf("TotalGroups = %v, want 1", idx.TotalGroups)
	}

	if len(group.RecordingIDs) != 2 {
		t.Errorf("len(RecordingIDs) = %v, want 2", len(group.RecordingIDs))
	}

	if group.RequestCount != 2 {
		t.Errorf("RequestCount = %v, want 2", group.RequestCount)
	}
}
