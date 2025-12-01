package sse

import (
	"encoding/json"
	"fmt"
	"strings"
)

// OpenAIParser parses OpenAI's SSE streaming format
type OpenAIParser struct{}

// Parse parses an OpenAI SSE stream into a ParsedStream
func (p *OpenAIParser) Parse(body string) (*ParsedStream, error) {
	parsed := &ParsedStream{
		Provider: "openai",
		Events:   make([]Event, 0),
		Metadata: make(map[string]interface{}),
	}

	// Parse SSE format line by line
	lines := strings.Split(body, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and whitespace-only lines
		if line == "" {
			continue
		}

		// OpenAI format: "data: {...}" or "data: [DONE]"
		if strings.HasPrefix(line, "data: ") {
			dataContent := strings.TrimPrefix(line, "data: ")

			// Check for [DONE] marker
			if dataContent == "[DONE]" {
				parsed.Events = append(parsed.Events, Event{
					Type: "done",
					Data: map[string]interface{}{"marker": "[DONE]"},
				})
				continue
			}

			// Parse JSON data
			if err := p.processChunk(dataContent, parsed); err != nil {
				return nil, fmt.Errorf("failed to process chunk: %w", err)
			}
		}
	}

	return parsed, nil
}

// processChunk processes a single OpenAI streaming chunk
func (p *OpenAIParser) processChunk(dataJSON string, parsed *ParsedStream) error {
	var chunk map[string]interface{}
	if err := json.Unmarshal([]byte(dataJSON), &chunk); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Store the event
	parsed.Events = append(parsed.Events, Event{
		Type: "chunk",
		Data: chunk,
	})

	// Extract metadata from the first chunk
	if len(parsed.Events) == 1 {
		if id, ok := chunk["id"].(string); ok {
			parsed.Metadata["id"] = id
		}
		if model, ok := chunk["model"].(string); ok {
			parsed.Metadata["model"] = model
		}
		if created, ok := chunk["created"].(float64); ok {
			parsed.Metadata["created"] = int64(created)
		}
	}

	// Extract text from choices
	choices, ok := chunk["choices"].([]interface{})
	if !ok {
		// No choices field - this is fine, some chunks don't have it
		return nil
	}
	if len(choices) == 0 {
		// Empty choices array - this is fine
		return nil
	}

	if choice, ok := choices[0].(map[string]interface{}); ok {
		// Extract delta content
		delta, ok := choice["delta"].(map[string]interface{})
		if ok && len(delta) > 0 {
			// Handle text content
			if content, ok := delta["content"].(string); ok {
				parsed.Text += content
			}

			// Handle role (appears in first chunk)
			if role, ok := delta["role"].(string); ok {
				parsed.Metadata["role"] = role
			}

			// Handle tool calls if present
			if toolCalls, ok := delta["tool_calls"].([]interface{}); ok {
				p.extractToolCalls(toolCalls, parsed)
			}
		}
		// If delta is empty or missing, we skip it but continue processing other fields

		// Extract finish reason from final chunk
		if finishReason, ok := choice["finish_reason"]; ok && finishReason != nil {
			if reason, ok := finishReason.(string); ok {
				parsed.Metadata["finish_reason"] = reason
			}
		}

		// Extract index
		if index, ok := choice["index"].(float64); ok {
			parsed.Metadata["choice_index"] = int(index)
		}
	}

	// Extract usage information if present (appears in some streaming responses)
	if usage, ok := chunk["usage"].(map[string]interface{}); ok {
		if promptTokens, ok := usage["prompt_tokens"].(float64); ok {
			parsed.Metadata["prompt_tokens"] = int(promptTokens)
		}
		if completionTokens, ok := usage["completion_tokens"].(float64); ok {
			parsed.Metadata["completion_tokens"] = int(completionTokens)
		}
		if totalTokens, ok := usage["total_tokens"].(float64); ok {
			parsed.Metadata["total_tokens"] = int(totalTokens)
		}
	}

	return nil
}

// extractToolCalls extracts tool call information from deltas
func (p *OpenAIParser) extractToolCalls(toolCalls []interface{}, parsed *ParsedStream) {
	// Initialize tool_calls in metadata if not present
	if parsed.Metadata["tool_calls"] == nil {
		parsed.Metadata["tool_calls"] = make([]map[string]interface{}, 0)
	}

	existingCalls := parsed.Metadata["tool_calls"].([]map[string]interface{})

	for _, tc := range toolCalls {
		if toolCall, ok := tc.(map[string]interface{}); ok {
			index := int(toolCall["index"].(float64))

			// Ensure we have enough space in the array
			for len(existingCalls) <= index {
				existingCalls = append(existingCalls, make(map[string]interface{}))
			}

			// Accumulate tool call data
			if id, ok := toolCall["id"].(string); ok {
				existingCalls[index]["id"] = id
			}
			if tcType, ok := toolCall["type"].(string); ok {
				existingCalls[index]["type"] = tcType
			}
			if function, ok := toolCall["function"].(map[string]interface{}); ok {
				if existingCalls[index]["function"] == nil {
					existingCalls[index]["function"] = make(map[string]interface{})
				}
				fnData := existingCalls[index]["function"].(map[string]interface{})

				if name, ok := function["name"].(string); ok {
					fnData["name"] = name
				}
				if args, ok := function["arguments"].(string); ok {
					// Accumulate arguments
					if fnData["arguments"] == nil {
						fnData["arguments"] = args
					} else {
						fnData["arguments"] = fnData["arguments"].(string) + args
					}
				}
			}
		}
	}

	parsed.Metadata["tool_calls"] = existingCalls
}
