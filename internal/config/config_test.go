package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/llmite-ai/mirra/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_DefaultConfig(t *testing.T) {
	cfg, err := Load("")
	require.NoError(t, err)

	assert.Equal(t, 4567, cfg.Port)
	assert.Equal(t, true, cfg.Recording.Enabled)
	assert.Equal(t, "file", cfg.Recording.Storage)
	assert.Equal(t, "./recordings", cfg.Recording.Path)
	assert.Equal(t, "jsonl", cfg.Recording.Format)
	assert.Equal(t, "pretty", cfg.Logging.Format)
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "https://api.anthropic.com", cfg.Providers["claude"].UpstreamURL)
	assert.Equal(t, "https://api.openai.com", cfg.Providers["openai"].UpstreamURL)
	assert.Equal(t, "https://generativelanguage.googleapis.com", cfg.Providers["gemini"].UpstreamURL)
}

func TestLoad_FromFile(t *testing.T) {
	tests := []struct {
		name           string
		configData     map[string]interface{}
		expectedPort   int
		expectedFormat string
	}{
		{
			name: "CustomPort",
			configData: map[string]interface{}{
				"port": 8080,
			},
			expectedPort:   8080,
			expectedFormat: "pretty", // default
		},
		{
			name: "CustomLogging",
			configData: map[string]interface{}{
				"logging": map[string]interface{}{
					"format": "json",
					"level":  "debug",
				},
			},
			expectedPort:   4567, // default
			expectedFormat: "json",
		},
		{
			name: "DisableRecording",
			configData: map[string]interface{}{
				"recording": map[string]interface{}{
					"enabled": false,
				},
			},
			expectedPort:   4567,
			expectedFormat: "pretty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, cleanup := testutil.TempDir(t)
			defer cleanup()

			configPath := filepath.Join(tempDir, "config.json")
			testutil.WriteJSONFile(t, configPath, tt.configData)

			cfg, err := Load(configPath)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedPort, cfg.Port)
			assert.Equal(t, tt.expectedFormat, cfg.Logging.Format)
		})
	}
}

func TestLoad_EnvironmentVariables(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		validate func(t *testing.T, cfg *Config)
	}{
		{
			name: "MIRRA_PORT",
			envVars: map[string]string{
				"MIRRA_PORT": "9999",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 9999, cfg.Port)
			},
		},
		{
			name: "MIRRA_RECORDING_ENABLED_True",
			envVars: map[string]string{
				"MIRRA_RECORDING_ENABLED": "true",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.True(t, cfg.Recording.Enabled)
			},
		},
		{
			name: "MIRRA_RECORDING_ENABLED_False",
			envVars: map[string]string{
				"MIRRA_RECORDING_ENABLED": "false",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.False(t, cfg.Recording.Enabled)
			},
		},
		{
			name: "MIRRA_RECORDING_PATH",
			envVars: map[string]string{
				"MIRRA_RECORDING_PATH": "/custom/path",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "/custom/path", cfg.Recording.Path)
			},
		},
		{
			name: "MIRRA_CLAUDE_UPSTREAM",
			envVars: map[string]string{
				"MIRRA_CLAUDE_UPSTREAM": "https://custom.claude.api",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "https://custom.claude.api", cfg.Providers["claude"].UpstreamURL)
			},
		},
		{
			name: "MIRRA_OPENAI_UPSTREAM",
			envVars: map[string]string{
				"MIRRA_OPENAI_UPSTREAM": "https://custom.openai.api",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "https://custom.openai.api", cfg.Providers["openai"].UpstreamURL)
			},
		},
		{
			name: "MIRRA_GEMINI_UPSTREAM",
			envVars: map[string]string{
				"MIRRA_GEMINI_UPSTREAM": "https://custom.gemini.api",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "https://custom.gemini.api", cfg.Providers["gemini"].UpstreamURL)
			},
		},
		{
			name: "MultipleEnvVars",
			envVars: map[string]string{
				"MIRRA_PORT":              "7777",
				"MIRRA_RECORDING_ENABLED": "false",
				"MIRRA_CLAUDE_UPSTREAM":   "https://custom.claude.api",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 7777, cfg.Port)
				assert.False(t, cfg.Recording.Enabled)
				assert.Equal(t, "https://custom.claude.api", cfg.Providers["claude"].UpstreamURL)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				testutil.SetEnv(t, key, value)
			}

			cfg, err := Load("")
			require.NoError(t, err)

			tt.validate(t, cfg)
		})
	}
}

func TestLoad_EnvOverridesFile(t *testing.T) {
	tempDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Create config file with port 8080
	configPath := filepath.Join(tempDir, "config.json")
	testutil.WriteJSONFile(t, configPath, map[string]interface{}{
		"port": 8080,
		"providers": map[string]interface{}{
			"claude": map[string]interface{}{
				"upstream_url": "https://file.claude.api",
			},
		},
	})

	// Set environment variable to override
	testutil.SetEnv(t, "MIRRA_PORT", "9090")
	testutil.SetEnv(t, "MIRRA_CLAUDE_UPSTREAM", "https://env.claude.api")

	cfg, err := Load(configPath)
	require.NoError(t, err)

	// Environment variable should override file
	assert.Equal(t, 9090, cfg.Port)
	assert.Equal(t, "https://env.claude.api", cfg.Providers["claude"].UpstreamURL)
}

func TestLoad_InvalidJSON(t *testing.T) {
	tempDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	configPath := filepath.Join(tempDir, "config.json")
	err := os.WriteFile(configPath, []byte("invalid json {{{"), 0644)
	require.NoError(t, err)

	cfg, err := Load(configPath)
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoad_NonExistentFile(t *testing.T) {
	cfg, err := Load("/path/that/does/not/exist.json")
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoad_InvalidPortEnvVar(t *testing.T) {
	// Set invalid port (not a number)
	testutil.SetEnv(t, "MIRRA_PORT", "not-a-number")

	cfg, err := Load("")
	require.NoError(t, err)

	// Should fall back to default port
	assert.Equal(t, 4567, cfg.Port)
}

func TestLoad_PartialConfig(t *testing.T) {
	tempDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Config file with only port specified
	configPath := filepath.Join(tempDir, "config.json")
	testutil.WriteJSONFile(t, configPath, map[string]interface{}{
		"port": 3000,
	})

	cfg, err := Load(configPath)
	require.NoError(t, err)

	// Should have custom port
	assert.Equal(t, 3000, cfg.Port)

	// Should have default values for other fields
	assert.True(t, cfg.Recording.Enabled)
	assert.Equal(t, "file", cfg.Recording.Storage)
	assert.Equal(t, "./recordings", cfg.Recording.Path)
	assert.Equal(t, "pretty", cfg.Logging.Format)
}

func TestLoad_CustomProviders(t *testing.T) {
	tempDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	configPath := filepath.Join(tempDir, "config.json")
	testutil.WriteJSONFile(t, configPath, map[string]interface{}{
		"providers": map[string]interface{}{
			"claude": map[string]interface{}{
				"upstream_url": "https://custom-claude.example.com",
			},
			"openai": map[string]interface{}{
				"upstream_url": "https://custom-openai.example.com",
			},
			"gemini": map[string]interface{}{
				"upstream_url": "https://custom-gemini.example.com",
			},
		},
	})

	cfg, err := Load(configPath)
	require.NoError(t, err)

	assert.Equal(t, "https://custom-claude.example.com", cfg.Providers["claude"].UpstreamURL)
	assert.Equal(t, "https://custom-openai.example.com", cfg.Providers["openai"].UpstreamURL)
	assert.Equal(t, "https://custom-gemini.example.com", cfg.Providers["gemini"].UpstreamURL)
}
