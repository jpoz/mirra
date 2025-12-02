package proxy

import (
	"testing"

	"github.com/jpoz/mirra/internal/config"
	"github.com/jpoz/mirra/internal/recorder"
	"github.com/stretchr/testify/assert"
)

func TestIdentifyProvider(t *testing.T) {
	tests := []struct {
		name             string
		path             string
		expectedProvider string
	}{
		// Claude endpoints
		{
			name:             "Claude_Messages",
			path:             "/v1/messages",
			expectedProvider: "claude",
		},
		{
			name:             "Claude_Complete",
			path:             "/v1/complete",
			expectedProvider: "claude",
		},
		{
			name:             "Claude_MessagesWithID",
			path:             "/v1/messages/123",
			expectedProvider: "claude",
		},

		// OpenAI endpoints
		{
			name:             "OpenAI_ChatCompletions",
			path:             "/v1/chat/completions",
			expectedProvider: "openai",
		},
		{
			name:             "OpenAI_Completions",
			path:             "/v1/completions",
			expectedProvider: "openai",
		},
		{
			name:             "OpenAI_Embeddings",
			path:             "/v1/embeddings",
			expectedProvider: "openai",
		},
		{
			name:             "OpenAI_ModelsList",
			path:             "/v1/models",
			expectedProvider: "openai",
		},
		{
			name:             "OpenAI_ModelGet",
			path:             "/v1/models/gpt-4",
			expectedProvider: "openai",
		},
		{
			name:             "OpenAI_Responses",
			path:             "/v1/responses",
			expectedProvider: "openai",
		},

		// Gemini endpoints - v1
		{
			name:             "Gemini_GenerateContent_v1",
			path:             "/v1/models/gemini-pro:generateContent",
			expectedProvider: "gemini",
		},
		{
			name:             "Gemini_StreamGenerateContent_v1",
			path:             "/v1/models/gemini-pro:streamGenerateContent",
			expectedProvider: "gemini",
		},
		{
			name:             "Gemini_EmbedContent_v1",
			path:             "/v1/models/text-embedding:embedContent",
			expectedProvider: "gemini",
		},
		{
			name:             "Gemini_CountTokens_v1",
			path:             "/v1/models/gemini-pro:countTokens",
			expectedProvider: "gemini",
		},
		{
			name:             "Gemini_Files_v1",
			path:             "/v1/files",
			expectedProvider: "gemini",
		},
		{
			name:             "Gemini_FileGet_v1",
			path:             "/v1/files/abc123",
			expectedProvider: "gemini",
		},
		{
			name:             "Gemini_FileUpload_v1",
			path:             "/upload/v1/files",
			expectedProvider: "gemini",
		},
		{
			name:             "Gemini_CachedContents_v1",
			path:             "/v1/cachedContents",
			expectedProvider: "gemini",
		},
		{
			name:             "Gemini_CachedContent_v1",
			path:             "/v1/cachedContents/abc123",
			expectedProvider: "gemini",
		},
		{
			name:             "Gemini_Corpora_v1",
			path:             "/v1/corpora",
			expectedProvider: "gemini",
		},
		{
			name:             "Gemini_Corpus_v1",
			path:             "/v1/corpora/my-corpus",
			expectedProvider: "gemini",
		},
		{
			name:             "Gemini_Documents_v1",
			path:             "/v1/corpora/my-corpus/documents",
			expectedProvider: "gemini",
		},
		{
			name:             "Gemini_Chunks_v1",
			path:             "/v1/corpora/my-corpus/documents/my-doc/chunks",
			expectedProvider: "gemini",
		},
		{
			name:             "Gemini_TunedModels_v1",
			path:             "/v1/tunedModels",
			expectedProvider: "gemini",
		},
		{
			name:             "Gemini_TunedModel_v1",
			path:             "/v1/tunedModels/my-model",
			expectedProvider: "gemini",
		},
		{
			name:             "Gemini_TunedModelOperations_v1",
			path:             "/v1/tunedModels/my-model/operations",
			expectedProvider: "gemini",
		},
		{
			name:             "Gemini_Batches_v1",
			path:             "/v1/batches",
			expectedProvider: "gemini",
		},
		{
			name:             "Gemini_Batch_v1",
			path:             "/v1/batches/batch123",
			expectedProvider: "gemini",
		},

		// Gemini endpoints - v1beta
		{
			name:             "Gemini_GenerateContent_v1beta",
			path:             "/v1beta/models/gemini-2.5-pro:generateContent",
			expectedProvider: "gemini",
		},
		{
			name:             "Gemini_Files_v1beta",
			path:             "/v1beta/files",
			expectedProvider: "gemini",
		},
		{
			name:             "Gemini_FileUpload_v1beta",
			path:             "/upload/v1beta/files",
			expectedProvider: "gemini",
		},

		// Gemini endpoints - v1alpha
		{
			name:             "Gemini_GenerateContent_v1alpha",
			path:             "/v1alpha/models/gemini-exp:generateContent",
			expectedProvider: "gemini",
		},
		{
			name:             "Gemini_Files_v1alpha",
			path:             "/v1alpha/files/xyz789",
			expectedProvider: "gemini",
		},
		{
			name:             "Gemini_FileUpload_v1alpha",
			path:             "/upload/v1alpha/files",
			expectedProvider: "gemini",
		},

		// Gemini with slash notation (models/)
		// Note: /v1/models/ and /v1/models/{id} without colons are ambiguous
		// and will match OpenAI since they don't have Gemini-specific colon operations
		{
			name:             "Models_List_Ambiguous",
			path:             "/v1/models/",
			expectedProvider: "openai", // Ambiguous, matches OpenAI
		},
		{
			name:             "Models_Get_Ambiguous",
			path:             "/v1/models/gemini-pro",
			expectedProvider: "openai", // Ambiguous, matches OpenAI without colon
		},

		// Unknown endpoints
		{
			name:             "Unknown_NoVersion",
			path:             "/api/endpoint",
			expectedProvider: "",
		},
		{
			name:             "Unknown_WrongVersion",
			path:             "/v2/messages",
			expectedProvider: "",
		},
		{
			name:             "Unknown_Empty",
			path:             "",
			expectedProvider: "",
		},
		{
			name:             "Unknown_Root",
			path:             "/",
			expectedProvider: "",
		},
		{
			name:             "Unknown_Random",
			path:             "/random/path/here",
			expectedProvider: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			p := New(cfg, nil)

			provider := p.identifyProvider(tt.path)
			assert.Equal(t, tt.expectedProvider, provider,
				"identifyProvider(%q) = %q, want %q", tt.path, provider, tt.expectedProvider)
		})
	}
}

func TestIsGeminiPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		// Valid Gemini paths - v1
		{name: "v1_models_generateContent", path: "/v1/models/gemini-pro:generateContent", expected: true},
		{name: "v1_models_streamGenerateContent", path: "/v1/models/gemini-pro:streamGenerateContent", expected: true},
		{name: "v1_models_embedContent", path: "/v1/models/text-embedding:embedContent", expected: true},
		{name: "v1_models_countTokens", path: "/v1/models/gemini-pro:countTokens", expected: true},
		{name: "v1_models_slash_no_colon", path: "/v1/models/gemini-pro", expected: false}, // No colon, not Gemini
		{name: "v1_models_list_no_colon", path: "/v1/models/", expected: false},            // No colon, not Gemini
		{name: "v1_files", path: "/v1/files", expected: true},
		{name: "v1_files_id", path: "/v1/files/abc123", expected: true},
		{name: "v1_files_nested", path: "/v1/files/abc123/metadata", expected: true},
		{name: "v1_cachedContents", path: "/v1/cachedContents", expected: true},
		{name: "v1_cachedContents_id", path: "/v1/cachedContents/cache123", expected: true},
		{name: "v1_corpora", path: "/v1/corpora", expected: true},
		{name: "v1_corpora_documents", path: "/v1/corpora/my-corpus/documents", expected: true},
		{name: "v1_corpora_chunks", path: "/v1/corpora/my-corpus/documents/doc1/chunks", expected: true},
		{name: "v1_tunedModels", path: "/v1/tunedModels", expected: true},
		{name: "v1_tunedModels_operations", path: "/v1/tunedModels/model1/operations", expected: true},
		{name: "v1_batches", path: "/v1/batches", expected: true},
		{name: "v1_batches_id", path: "/v1/batches/batch123", expected: true},

		// Valid Gemini paths - v1beta
		{name: "v1beta_models", path: "/v1beta/models/gemini-2.5-pro:generateContent", expected: true},
		{name: "v1beta_files", path: "/v1beta/files", expected: true},
		{name: "v1beta_cachedContents", path: "/v1beta/cachedContents/cache1", expected: true},

		// Valid Gemini paths - v1alpha
		{name: "v1alpha_models", path: "/v1alpha/models/gemini-exp:generateContent", expected: true},
		{name: "v1alpha_files", path: "/v1alpha/files/xyz", expected: true},

		// Valid Gemini paths - file uploads
		{name: "upload_v1_files", path: "/upload/v1/files", expected: true},
		{name: "upload_v1beta_files", path: "/upload/v1beta/files", expected: true},
		{name: "upload_v1alpha_files", path: "/upload/v1alpha/files", expected: true},
		{name: "upload_v1_files_id", path: "/upload/v1/files/file123", expected: true},

		// Invalid Gemini paths
		{name: "no_version", path: "/models/gemini-pro", expected: false},
		{name: "wrong_version", path: "/v2/models/gemini-pro", expected: false},
		{name: "v1_without_resource", path: "/v1/", expected: false},
		{name: "v1_unknown_resource", path: "/v1/unknown", expected: false},
		{name: "v1beta_unknown", path: "/v1beta/something", expected: false},
		{name: "upload_without_files", path: "/upload/v1/other", expected: false},
		{name: "upload_no_version", path: "/upload/files", expected: false},
		{name: "empty", path: "", expected: false},
		{name: "root", path: "/", expected: false},
		{name: "no_leading_slash", path: "v1/models/gemini", expected: false},

		// Edge cases
		{name: "models_with_colon", path: "/v1/models:batchPredict", expected: true},
		{name: "files_query_params", path: "/v1/files?pageSize=10", expected: true},
		{name: "multiple_slashes_no_colon", path: "/v1//models//gemini", expected: false}, // No colon, not Gemini
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGeminiPath(tt.path)
			assert.Equal(t, tt.expected, result,
				"isGeminiPath(%q) = %v, want %v", tt.path, result, tt.expected)
		})
	}
}

func TestNew(t *testing.T) {
	cfg := &config.Config{
		Port: 4567,
		Providers: map[string]config.Provider{
			"claude": {UpstreamURL: "https://api.anthropic.com"},
		},
	}

	rec := recorder.New(true, "./test-recordings")
	defer func() {
		_ = rec.Close()
	}()

	proxy := New(cfg, rec)

	assert.NotNil(t, proxy)
	assert.Equal(t, cfg, proxy.cfg)
	assert.NotNil(t, proxy.client)
	assert.NotNil(t, proxy.recorder)
	assert.Equal(t, rec, proxy.recorder)
}

func TestIdentifyProvider_GeminiBeforeOpenAI(t *testing.T) {
	// This test demonstrates the distinction between Gemini and OpenAI model endpoints
	// Gemini operations use colons (:generateContent), while OpenAI uses simple paths
	cfg := &config.Config{}
	p := New(cfg, nil)

	tests := []struct {
		path             string
		expectedProvider string
		reason           string
	}{
		{
			path:             "/v1/models/gemini-pro:generateContent",
			expectedProvider: "gemini",
			reason:           "Gemini model with operation (has colon)",
		},
		{
			path:             "/v1/models/gemini-pro",
			expectedProvider: "openai",
			reason:           "Ambiguous path without colon, matches OpenAI",
		},
		{
			path:             "/v1/models/",
			expectedProvider: "openai",
			reason:           "Ambiguous path without colon, matches OpenAI",
		},
		{
			path:             "/v1/models",
			expectedProvider: "openai",
			reason:           "OpenAI models endpoint (no trailing slash, no colon)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			provider := p.identifyProvider(tt.path)
			assert.Equal(t, tt.expectedProvider, provider,
				"Expected %s for path %s: %s", tt.expectedProvider, tt.path, tt.reason)
		})
	}
}
