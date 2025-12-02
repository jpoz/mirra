package recorder

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// IndexEntry represents metadata about a recording's location in a JSONL file
type IndexEntry struct {
	ID        string    `json:"id"`
	Filename  string    `json:"filename"`
	Offset    int64     `json:"offset"`
	Length    int64     `json:"length"`
	Timestamp time.Time `json:"timestamp"`
	Provider  string    `json:"provider"`
}

// Index manages an in-memory map of recording IDs to their file locations
type Index struct {
	entries map[string]IndexEntry
	mu      sync.RWMutex
	path    string // path to recordings directory
	dirty   bool   // tracks if index needs to be persisted
}

// NewIndex creates a new index instance
func NewIndex(recordingsPath string) *Index {
	return &Index{
		entries: make(map[string]IndexEntry),
		path:    recordingsPath,
	}
}

// Load reads the index from disk
func (idx *Index) Load() error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	indexPath := filepath.Join(idx.path, "index.json")

	file, err := os.Open(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Index doesn't exist yet, will be built on first use
			return nil
		}
		return fmt.Errorf("failed to open index file: %w", err)
	}
	defer file.Close()

	var entries []IndexEntry
	if err := json.NewDecoder(file).Decode(&entries); err != nil {
		return fmt.Errorf("failed to decode index: %w", err)
	}

	// Convert slice to map
	idx.entries = make(map[string]IndexEntry, len(entries))
	for _, entry := range entries {
		idx.entries[entry.ID] = entry
	}

	slog.Info("Loaded recording index", "count", len(idx.entries))
	return nil
}

// Save writes the index to disk
func (idx *Index) Save() error {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if !idx.dirty {
		return nil // No changes to persist
	}

	indexPath := filepath.Join(idx.path, "index.json")

	// Convert map to slice for JSON encoding
	entries := make([]IndexEntry, 0, len(idx.entries))
	for _, entry := range idx.entries {
		entries = append(entries, entry)
	}

	// Write to temporary file first
	tmpPath := indexPath + ".tmp"
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp index file: %w", err)
	}

	encoder := json.NewEncoder(tmpFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(entries); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to encode index: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to close temp index file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, indexPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp index file: %w", err)
	}

	slog.Info("Saved recording index", "count", len(entries))
	return nil
}

// Add adds or updates an entry in the index
func (idx *Index) Add(entry IndexEntry) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.entries[entry.ID] = entry
	idx.dirty = true
}

// Get retrieves an entry from the index
func (idx *Index) Get(id string) (IndexEntry, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	entry, found := idx.entries[id]
	return entry, found
}

// GetByPrefix finds the first entry matching the ID prefix
func (idx *Index) GetByPrefix(prefix string) (IndexEntry, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	// Try exact match first
	if entry, found := idx.entries[prefix]; found {
		return entry, true
	}

	// Search for prefix match
	for id, entry := range idx.entries {
		if strings.HasPrefix(id, prefix) {
			return entry, true
		}
	}

	return IndexEntry{}, false
}

// Rebuild scans all JSONL files and rebuilds the index from scratch
func (idx *Index) Rebuild() error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	slog.Info("Rebuilding recording index", "path", idx.path)

	entries, err := os.ReadDir(idx.path)
	if err != nil {
		return fmt.Errorf("failed to read recordings directory: %w", err)
	}

	newIndex := make(map[string]IndexEntry)
	var totalRecordings int

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}

		filePath := filepath.Join(idx.path, entry.Name())
		file, err := os.Open(filePath)
		if err != nil {
			slog.Error("Failed to open file for indexing", "file", entry.Name(), "error", err)
			continue
		}

		var offset int64
		scanner := bufio.NewScanner(file)
		// Increase buffer size for large recordings
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 10*1024*1024)

		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				offset += int64(len(line)) + 1 // +1 for newline
				continue
			}

			// Parse just enough to get ID, timestamp, and provider
			var partial struct {
				ID        string    `json:"id"`
				Timestamp time.Time `json:"timestamp"`
				Provider  string    `json:"provider"`
			}

			if err := json.Unmarshal(line, &partial); err != nil {
				slog.Error("Failed to parse recording for indexing", "file", entry.Name(), "error", err)
				offset += int64(len(line)) + 1
				continue
			}

			indexEntry := IndexEntry{
				ID:        partial.ID,
				Filename:  entry.Name(),
				Offset:    offset,
				Length:    int64(len(line)),
				Timestamp: partial.Timestamp,
				Provider:  partial.Provider,
			}

			newIndex[partial.ID] = indexEntry
			totalRecordings++

			offset += int64(len(line)) + 1 // +1 for newline
		}

		file.Close()

		if err := scanner.Err(); err != nil {
			slog.Error("Scanner error while indexing", "file", entry.Name(), "error", err)
			continue
		}
	}

	idx.entries = newIndex
	idx.dirty = true

	slog.Info("Rebuilt recording index", "files", len(entries), "recordings", totalRecordings)
	return nil
}

// ReadRecording reads a specific recording from disk using the index
func (idx *Index) ReadRecording(id string) (*Recording, error) {
	// Get index entry
	entry, found := idx.GetByPrefix(id)
	if !found {
		return nil, fmt.Errorf("recording not found in index")
	}

	// Open the file
	filePath := filepath.Join(idx.path, entry.Filename)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Seek to the offset
	if _, err := file.Seek(entry.Offset, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek to offset: %w", err)
	}

	// Read the line
	reader := bufio.NewReaderSize(file, int(entry.Length)+1024)
	line, err := reader.ReadBytes('\n')
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read line: %w", err)
	}

	// Parse the recording
	var rec Recording
	if err := json.Unmarshal(line, &rec); err != nil {
		return nil, fmt.Errorf("failed to parse recording: %w", err)
	}

	return &rec, nil
}

// Size returns the number of entries in the index
func (idx *Index) Size() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.entries)
}
