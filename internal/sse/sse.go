package sse

import "time"

// Parser defines the interface for parsing provider-specific SSE streams
type Parser interface {
	// Parse parses a raw SSE body string into a structured ParsedStream
	Parse(body string) (*ParsedStream, error)
}

// ParsedStream represents a parsed SSE stream with reconstructed content
type ParsedStream struct {
	// Provider is the LLM provider (claude, openai, gemini)
	Provider string

	// Events contains all parsed events in order
	Events []Event

	// Text is the reconstructed full text output
	Text string

	// Metadata contains provider-specific metadata
	Metadata map[string]interface{}
}

// Event represents a single SSE event
type Event struct {
	// Type is the event type (e.g., message_start, content_block_delta)
	Type string

	// Data is the parsed JSON data for this event
	Data map[string]interface{}

	// Timestamp when the event was processed (if available)
	Timestamp time.Time
}

// NewParser creates a Parser for the specified provider
func NewParser(provider string) Parser {
	switch provider {
	case "claude":
		return &ClaudeParser{}
	case "openai":
		return &OpenAIParser{}
	case "gemini":
		return &GeminiParser{}
	default:
		return nil
	}
}
