package sse

import (
	"encoding/json"
	"fmt"
	"strings"
)

// GeminiParser parses Google Gemini's SSE streaming format
type GeminiParser struct{}

// Parse parses a Gemini SSE stream into a ParsedStream
func (p *GeminiParser) Parse(body string) (*ParsedStream, error) {
	parsed := &ParsedStream{
		Provider: "gemini",
		Events:   make([]Event, 0),
		Metadata: make(map[string]interface{}),
	}

	// Parse SSE format line by line
	lines := strings.Split(body, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Gemini format: "data: {...}"
		if strings.HasPrefix(line, "data: ") {
			dataContent := strings.TrimPrefix(line, "data: ")

			// Parse JSON data
			if err := p.processChunk(dataContent, parsed); err != nil {
				return nil, fmt.Errorf("failed to process chunk: %w", err)
			}
		}
	}

	return parsed, nil
}

// processChunk processes a single Gemini streaming chunk
func (p *GeminiParser) processChunk(dataJSON string, parsed *ParsedStream) error {
	var chunk map[string]interface{}
	if err := json.Unmarshal([]byte(dataJSON), &chunk); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Store the event
	parsed.Events = append(parsed.Events, Event{
		Type: "chunk",
		Data: chunk,
	})

	// Extract metadata from chunks
	if modelVersion, ok := chunk["modelVersion"].(string); ok {
		parsed.Metadata["model"] = modelVersion
	}
	if responseID, ok := chunk["responseId"].(string); ok {
		parsed.Metadata["response_id"] = responseID
	}

	// Extract text from candidates
	if candidates, ok := chunk["candidates"].([]interface{}); ok && len(candidates) > 0 {
		if candidate, ok := candidates[0].(map[string]interface{}); ok {
			// Extract finish reason if present
			if finishReason, ok := candidate["finishReason"].(string); ok {
				parsed.Metadata["finish_reason"] = finishReason
			}

			// Extract safety ratings if present
			if safetyRatings, ok := candidate["safetyRatings"].([]interface{}); ok {
				parsed.Metadata["safety_ratings"] = safetyRatings
			}

			// Extract content
			if content, ok := candidate["content"].(map[string]interface{}); ok {
				// Extract role
				if role, ok := content["role"].(string); ok && parsed.Metadata["role"] == nil {
					parsed.Metadata["role"] = role
				}

				// Extract parts (text content)
				if parts, ok := content["parts"].([]interface{}); ok {
					for _, part := range parts {
						if partMap, ok := part.(map[string]interface{}); ok {
							// Handle text parts
							if text, ok := partMap["text"].(string); ok {
								parsed.Text += text
							}

							// Handle function calls if present
							if functionCall, ok := partMap["functionCall"].(map[string]interface{}); ok {
								p.extractFunctionCall(functionCall, parsed)
							}
						}
					}
				}
			}

			// Extract grounding metadata if present
			if groundingMetadata, ok := candidate["groundingMetadata"].(map[string]interface{}); ok {
				parsed.Metadata["grounding_metadata"] = groundingMetadata
			}
		}
	}

	// Extract usage metadata
	if usageMetadata, ok := chunk["usageMetadata"].(map[string]interface{}); ok {
		if promptTokenCount, ok := usageMetadata["promptTokenCount"].(float64); ok {
			parsed.Metadata["prompt_tokens"] = int(promptTokenCount)
		}
		if candidatesTokenCount, ok := usageMetadata["candidatesTokenCount"].(float64); ok {
			parsed.Metadata["completion_tokens"] = int(candidatesTokenCount)
		}
		if totalTokenCount, ok := usageMetadata["totalTokenCount"].(float64); ok {
			parsed.Metadata["total_tokens"] = int(totalTokenCount)
		}
		// Cache hit tokens if present
		if cachedContentTokenCount, ok := usageMetadata["cachedContentTokenCount"].(float64); ok {
			parsed.Metadata["cached_content_tokens"] = int(cachedContentTokenCount)
		}
	}

	return nil
}

// extractFunctionCall extracts function call information
func (p *GeminiParser) extractFunctionCall(functionCall map[string]interface{}, parsed *ParsedStream) {
	if parsed.Metadata["function_calls"] == nil {
		parsed.Metadata["function_calls"] = make([]map[string]interface{}, 0)
	}

	calls := parsed.Metadata["function_calls"].([]map[string]interface{})

	callData := make(map[string]interface{})
	if name, ok := functionCall["name"].(string); ok {
		callData["name"] = name
	}
	if args, ok := functionCall["args"].(map[string]interface{}); ok {
		callData["args"] = args
	}

	calls = append(calls, callData)
	parsed.Metadata["function_calls"] = calls
}
