package testutil

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TempDir creates a temporary directory for testing and returns cleanup function
func TempDir(t *testing.T) (string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "mirra-test-*")
	require.NoError(t, err, "failed to create temp directory")

	cleanup := func() {
		_ = os.RemoveAll(dir)
	}

	return dir, cleanup
}

// WriteJSONFile writes JSON data to a file
func WriteJSONFile(t *testing.T, path string, data interface{}) {
	t.Helper()

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	require.NoError(t, err, "failed to create directory")

	// Marshal JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	require.NoError(t, err, "failed to marshal JSON")

	// Write to file
	err = os.WriteFile(path, jsonData, 0644)
	require.NoError(t, err, "failed to write file")
}

// ReadJSONFile reads and unmarshals JSON from a file
func ReadJSONFile(t *testing.T, path string, v interface{}) {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err, "failed to read file")

	err = json.Unmarshal(data, v)
	require.NoError(t, err, "failed to unmarshal JSON")
}

// MockHTTPServer creates a test HTTP server with custom handler
func MockHTTPServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// MockClaudeServer creates a mock Claude API server
func MockClaudeServer(response interface{}, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(response)
	}))
}

// MockOpenAIServer creates a mock OpenAI API server
func MockOpenAIServer(response interface{}, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(response)
	}))
}

// MockGeminiServer creates a mock Gemini API server
func MockGeminiServer(response interface{}, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(response)
	}))
}

// MockStreamingServer creates a mock server that returns SSE streaming responses
func MockStreamingServer(events []string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		for _, event := range events {
			_, _ = w.Write([]byte(event + "\n\n"))
			flusher.Flush()
		}
	}))
}

// GzipCompress compresses data using gzip
func GzipCompress(t *testing.T, data []byte) []byte {
	t.Helper()

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	_, err := gz.Write(data)
	require.NoError(t, err, "failed to write to gzip writer")

	err = gz.Close()
	require.NoError(t, err, "failed to close gzip writer")

	return buf.Bytes()
}

// GzipDecompress decompresses gzipped data
func GzipDecompress(t *testing.T, data []byte) []byte {
	t.Helper()

	gz, err := gzip.NewReader(bytes.NewReader(data))
	require.NoError(t, err, "failed to create gzip reader")
	defer func() {
		_ = gz.Close()
	}()

	decompressed, err := io.ReadAll(gz)
	require.NoError(t, err, "failed to read decompressed data")

	return decompressed
}

// MakeHTTPRequest creates an HTTP request for testing
func MakeHTTPRequest(t *testing.T, method, url string, body interface{}, headers map[string]string) *http.Request {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err, "failed to marshal request body")
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	require.NoError(t, err, "failed to create request")

	// Set default headers
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return req
}

// AssertJSONEqual compares two JSON objects for equality
func AssertJSONEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()

	expectedJSON, err := json.MarshalIndent(expected, "", "  ")
	require.NoError(t, err, "failed to marshal expected JSON")

	actualJSON, err := json.MarshalIndent(actual, "", "  ")
	require.NoError(t, err, "failed to marshal actual JSON")

	require.JSONEq(t, string(expectedJSON), string(actualJSON))
}

// LoadFixture loads a test fixture file
func LoadFixture(t *testing.T, name string) []byte {
	t.Helper()

	path := filepath.Join("../../_dev/fixtures", name)
	data, err := os.ReadFile(path)
	require.NoError(t, err, "failed to load fixture: %s", name)

	return data
}

// LoadJSONFixture loads and unmarshals a JSON fixture
func LoadJSONFixture(t *testing.T, name string, v interface{}) {
	t.Helper()

	data := LoadFixture(t, name)
	err := json.Unmarshal(data, v)
	require.NoError(t, err, "failed to unmarshal JSON fixture: %s", name)
}

// SetEnv sets an environment variable for the duration of the test
func SetEnv(t *testing.T, key, value string) {
	t.Helper()

	oldValue, exists := os.LookupEnv(key)
	_ = os.Setenv(key, value)

	t.Cleanup(func() {
		if exists {
			_ = os.Setenv(key, oldValue)
		} else {
			_ = os.Unsetenv(key)
		}
	})
}

// WaitForCondition waits for a condition to be true (useful for async operations)
func WaitForCondition(t *testing.T, condition func() bool, timeout int) {
	t.Helper()

	for i := 0; i < timeout; i++ {
		if condition() {
			return
		}
		// In production tests, this would use time.Sleep, but for simplicity we just check once
	}

	t.Fatal("condition not met within timeout")
}
