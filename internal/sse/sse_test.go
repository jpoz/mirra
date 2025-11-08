package sse

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClaudeParser(t *testing.T) {
	// Real Claude SSE sample from recordings
	sseBody := `event: message_start
data: {"type":"message_start","message":{"model":"claude-haiku-4-5-20251001","id":"msg_018CttprAoSqXdPFkmoKBpNS","type":"message","role":"assistant","content":[],"stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":706,"cache_creation_input_tokens":0,"cache_read_input_tokens":0,"cache_creation":{"ephemeral_5m_input_tokens":0,"ephemeral_1h_input_tokens":0},"output_tokens":1,"service_tier":"standard"}}    }

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}             }

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Building"}         }

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" Mirra"}         }

event: ping
data: {"type": "ping"}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" Recordings Table UI with"}  }

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" API"}             }

event: content_block_stop
data: {"type":"content_block_stop","index":0       }

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"input_tokens":706,"cache_creation_input_tokens":0,"cache_read_input_tokens":0,"output_tokens":13}  }

event: message_stop
data: {"type":"message_stop"           }
`

	parser := NewParser("claude")
	require.NotNil(t, parser)

	parsed, err := parser.Parse(sseBody)
	require.NoError(t, err)
	require.NotNil(t, parsed)

	// Verify provider
	assert.Equal(t, "claude", parsed.Provider)

	// Verify reconstructed text
	assert.Equal(t, "Building Mirra Recordings Table UI with API", parsed.Text)

	// Verify metadata
	assert.Equal(t, "claude-haiku-4-5-20251001", parsed.Metadata["model"])
	assert.Equal(t, "msg_018CttprAoSqXdPFkmoKBpNS", parsed.Metadata["message_id"])
	assert.Equal(t, "end_turn", parsed.Metadata["stop_reason"])

	// Verify events were captured
	assert.Greater(t, len(parsed.Events), 0)

	// Find specific events
	var hasMessageStart, hasContentDelta, hasMessageStop bool
	for _, event := range parsed.Events {
		if event.Type == "message_start" {
			hasMessageStart = true
		}
		if event.Type == "content_block_delta" {
			hasContentDelta = true
		}
		if event.Type == "message_stop" {
			hasMessageStop = true
		}
	}
	assert.True(t, hasMessageStart, "Should have message_start event")
	assert.True(t, hasContentDelta, "Should have content_block_delta events")
	assert.True(t, hasMessageStop, "Should have message_stop event")
}

func TestClaudeParserWithThinking(t *testing.T) {
	// Claude SSE with thinking block
	sseBody := `event: message_start
data: {"type":"message_start","message":{"model":"claude-sonnet-4-5-20250929","id":"msg_01ABC","type":"message","role":"assistant","content":[],"stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":12,"output_tokens":2}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"thinking","thinking":"","signature":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"Let me think"}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":" about this"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: content_block_start
data: {"type":"content_block_start","index":1,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"text_delta","text":"Here is"}}

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"text_delta","text":" the answer"}}

event: content_block_stop
data: {"type":"content_block_stop","index":1}

event: message_stop
data: {"type":"message_stop"}
`

	parser := NewParser("claude")
	parsed, err := parser.Parse(sseBody)
	require.NoError(t, err)

	// Verify text content (not thinking)
	assert.Equal(t, "Here is the answer", parsed.Text)

	// Verify thinking was captured separately
	assert.Equal(t, "Let me think about this", parsed.Metadata["thinking"])
}

func TestOpenAIParser(t *testing.T) {
	// OpenAI SSE sample
	sseBody := `data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1694268190,"model":"gpt-4","choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1694268190,"model":"gpt-4","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1694268190,"model":"gpt-4","choices":[{"index":0,"delta":{"content":" world"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1694268190,"model":"gpt-4","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: [DONE]
`

	parser := NewParser("openai")
	require.NotNil(t, parser)

	parsed, err := parser.Parse(sseBody)
	require.NoError(t, err)
	require.NotNil(t, parsed)

	// Verify provider
	assert.Equal(t, "openai", parsed.Provider)

	// Verify reconstructed text
	assert.Equal(t, "Hello world", parsed.Text)

	// Verify metadata
	assert.Equal(t, "chatcmpl-123", parsed.Metadata["id"])
	assert.Equal(t, "gpt-4", parsed.Metadata["model"])
	assert.Equal(t, "assistant", parsed.Metadata["role"])
	assert.Equal(t, "stop", parsed.Metadata["finish_reason"])

	// Verify [DONE] marker was captured
	lastEvent := parsed.Events[len(parsed.Events)-1]
	assert.Equal(t, "done", lastEvent.Type)
}

func TestGeminiParser(t *testing.T) {
	// Gemini SSE sample
	sseBody := `data: {"candidates":[{"content":{"parts":[{"text":"The sky"}],"role":"model"},"index":0}],"usageMetadata":{"promptTokenCount":5},"modelVersion":"gemini-2.5-flash-lite","responseId":"abc123"}

data: {"candidates":[{"content":{"parts":[{"text":" is blue"}],"role":"model"},"index":0}],"modelVersion":"gemini-2.5-flash-lite","responseId":"abc123"}

data: {"candidates":[{"content":{"parts":[{"text":"."}],"role":"model"},"finishReason":"STOP","index":0}],"usageMetadata":{"promptTokenCount":5,"candidatesTokenCount":4,"totalTokenCount":9},"modelVersion":"gemini-2.5-flash-lite","responseId":"abc123"}
`

	parser := NewParser("gemini")
	require.NotNil(t, parser)

	parsed, err := parser.Parse(sseBody)
	require.NoError(t, err)
	require.NotNil(t, parsed)

	// Verify provider
	assert.Equal(t, "gemini", parsed.Provider)

	// Verify reconstructed text
	assert.Equal(t, "The sky is blue.", parsed.Text)

	// Verify metadata
	assert.Equal(t, "gemini-2.5-flash-lite", parsed.Metadata["model"])
	assert.Equal(t, "abc123", parsed.Metadata["response_id"])
	assert.Equal(t, "model", parsed.Metadata["role"])
	assert.Equal(t, "STOP", parsed.Metadata["finish_reason"])
	assert.Equal(t, 5, parsed.Metadata["prompt_tokens"])
	assert.Equal(t, 4, parsed.Metadata["completion_tokens"])
	assert.Equal(t, 9, parsed.Metadata["total_tokens"])
}

func TestNewParser(t *testing.T) {
	tests := []struct {
		provider string
		wantNil  bool
	}{
		{"claude", false},
		{"openai", false},
		{"gemini", false},
		{"unknown", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			parser := NewParser(tt.provider)
			if tt.wantNil {
				assert.Nil(t, parser)
			} else {
				assert.NotNil(t, parser)
			}
		})
	}
}

func TestClaudeParserEmptyBody(t *testing.T) {
	parser := NewParser("claude")
	parsed, err := parser.Parse("")
	require.NoError(t, err)
	assert.Equal(t, "", parsed.Text)
	assert.Equal(t, 0, len(parsed.Events))
}

func TestOpenAIParserEmptyBody(t *testing.T) {
	parser := NewParser("openai")
	parsed, err := parser.Parse("")
	require.NoError(t, err)
	assert.Equal(t, "", parsed.Text)
	assert.Equal(t, 0, len(parsed.Events))
}

func TestGeminiParserEmptyBody(t *testing.T) {
	parser := NewParser("gemini")
	parsed, err := parser.Parse("")
	require.NoError(t, err)
	assert.Equal(t, "", parsed.Text)
	assert.Equal(t, 0, len(parsed.Events))
}
