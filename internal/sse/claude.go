package sse

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ClaudeParser parses Claude's SSE streaming format
type ClaudeParser struct{}

// Parse parses a Claude SSE stream into a ParsedStream
func (p *ClaudeParser) Parse(body string) (*ParsedStream, error) {
	parsed := &ParsedStream{
		Provider: "claude",
		Events:   make([]Event, 0),
		Metadata: make(map[string]interface{}),
	}

	// Parse SSE format line by line
	lines := strings.Split(body, "\n")
	var currentEventType string
	var currentData strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "event: ") {
			// If we have a previous event, process it
			if currentEventType != "" && currentData.Len() > 0 {
				if err := p.processEvent(currentEventType, currentData.String(), parsed); err != nil {
					return nil, fmt.Errorf("failed to process event %s: %w", currentEventType, err)
				}
				currentData.Reset()
			}
			currentEventType = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "data: ") {
			dataContent := strings.TrimPrefix(line, "data: ")
			if currentData.Len() > 0 {
				currentData.WriteString("\n")
			}
			currentData.WriteString(dataContent)
		} else if line == "" && currentEventType != "" && currentData.Len() > 0 {
			// Empty line signals end of event
			if err := p.processEvent(currentEventType, currentData.String(), parsed); err != nil {
				return nil, fmt.Errorf("failed to process event %s: %w", currentEventType, err)
			}
			currentEventType = ""
			currentData.Reset()
		}
	}

	// Process last event if exists
	if currentEventType != "" && currentData.Len() > 0 {
		if err := p.processEvent(currentEventType, currentData.String(), parsed); err != nil {
			return nil, fmt.Errorf("failed to process event %s: %w", currentEventType, err)
		}
	}

	return parsed, nil
}

// processEvent processes a single SSE event and accumulates data
func (p *ClaudeParser) processEvent(eventType, dataJSON string, parsed *ParsedStream) error {
	// Parse the JSON data into a generic map for storage
	var eventData map[string]interface{}
	if err := json.Unmarshal([]byte(dataJSON), &eventData); err != nil {
		// Skip events that aren't valid JSON (like ping events)
		if eventType == "ping" {
			return nil
		}
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Store the event
	parsed.Events = append(parsed.Events, Event{
		Type: eventType,
		Data: eventData,
	})

	// Process specific event types to build the text output and metadata
	switch eventType {
	case "message_start":
		p.extractMessageStartMetadata(eventData, parsed)

	case "content_block_start":
		// Content block starting - we'll accumulate deltas next
		// No text to extract here

	case "content_block_delta":
		p.extractDeltaText(eventData, parsed)

	case "content_block_stop":
		// Content block finished
		// No text to extract here

	case "message_delta":
		p.extractMessageDeltaMetadata(eventData, parsed)

	case "message_stop":
		// Stream finished
		// No additional processing needed
	}

	return nil
}

// extractMessageStartMetadata extracts metadata from message_start event
func (p *ClaudeParser) extractMessageStartMetadata(data map[string]interface{}, parsed *ParsedStream) {
	if message, ok := data["message"].(map[string]interface{}); ok {
		if model, ok := message["model"].(string); ok {
			parsed.Metadata["model"] = model
		}
		if id, ok := message["id"].(string); ok {
			parsed.Metadata["message_id"] = id
		}
		if usage, ok := message["usage"].(map[string]interface{}); ok {
			parsed.Metadata["input_tokens"] = usage["input_tokens"]
			parsed.Metadata["cache_creation_input_tokens"] = usage["cache_creation_input_tokens"]
			parsed.Metadata["cache_read_input_tokens"] = usage["cache_read_input_tokens"]
		}
	}
}

// extractDeltaText extracts text from content_block_delta events
func (p *ClaudeParser) extractDeltaText(data map[string]interface{}, parsed *ParsedStream) {
	if delta, ok := data["delta"].(map[string]interface{}); ok {
		deltaType, _ := delta["type"].(string)

		// Handle different delta types
		switch deltaType {
		case "text_delta":
			if text, ok := delta["text"].(string); ok {
				parsed.Text += text
			}
		case "thinking_delta":
			if thinking, ok := delta["thinking"].(string); ok {
				// Store thinking separately if needed
				if parsed.Metadata["thinking"] == nil {
					parsed.Metadata["thinking"] = thinking
				} else {
					parsed.Metadata["thinking"] = parsed.Metadata["thinking"].(string) + thinking
				}
			}
		case "input_json_delta":
			if partialJSON, ok := delta["partial_json"].(string); ok {
				// Accumulate tool input JSON
				if parsed.Metadata["tool_input"] == nil {
					parsed.Metadata["tool_input"] = partialJSON
				} else {
					parsed.Metadata["tool_input"] = parsed.Metadata["tool_input"].(string) + partialJSON
				}
			}
		}
	}
}

// extractMessageDeltaMetadata extracts final metadata from message_delta event
func (p *ClaudeParser) extractMessageDeltaMetadata(data map[string]interface{}, parsed *ParsedStream) {
	if delta, ok := data["delta"].(map[string]interface{}); ok {
		if stopReason, ok := delta["stop_reason"].(string); ok {
			parsed.Metadata["stop_reason"] = stopReason
		}
	}
	if usage, ok := data["usage"].(map[string]interface{}); ok {
		if outputTokens, ok := usage["output_tokens"]; ok {
			parsed.Metadata["output_tokens"] = outputTokens
		}
	}
}
