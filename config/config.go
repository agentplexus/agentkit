// Package config provides configuration management for agent applications.
// It supports loading from environment variables with sensible defaults.
package config

import (
	"os"
)

// Config holds the application configuration.
type Config struct {
	// LLM Configuration
	LLMProvider string // "gemini", "claude", "openai", "ollama", "xai"
	LLMAPIKey   string
	LLMModel    string
	LLMBaseURL  string // For Ollama or custom endpoints

	// Provider-specific API keys
	GeminiAPIKey string
	ClaudeAPIKey string
	OpenAIAPIKey string
	XAIAPIKey    string
	OllamaURL    string

	// Search Configuration
	SearchProvider string // "serper", "serpapi"
	SerperAPIKey   string
	SerpAPIKey     string

	// Agent URLs (for multi-agent systems)
	AgentURLs map[string]string

	// A2A Protocol Configuration
	A2AEnabled   bool
	A2AAuthType  string // "jwt", "apikey", "oauth2"
	A2AAuthToken string

	// Observability Configuration
	ObservabilityEnabled  bool   // Enable LLM observability
	ObservabilityProvider string // "opik", "langfuse", "phoenix"
	ObservabilityAPIKey   string
	ObservabilityEndpoint string // Custom endpoint (optional)
	ObservabilityProject  string // Project name for grouping traces

	// Security Configuration
	SecurityEnabled      bool // Enable VaultGuard security checks
	SecurityMinScore     int  // Minimum security score (0-100)
	SecurityRequireEncry bool // Require disk encryption
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() *Config {
	provider := getEnv("LLM_PROVIDER", "gemini")

	cfg := &Config{
		// LLM settings
		LLMProvider: provider,
		LLMAPIKey:   getEnv("LLM_API_KEY", ""),
		LLMModel:    getEnv("LLM_MODEL", GetDefaultModel(provider)),
		LLMBaseURL:  getEnv("LLM_BASE_URL", ""),

		// Provider-specific API keys
		GeminiAPIKey: getEnv("GEMINI_API_KEY", getEnv("GOOGLE_API_KEY", "")),
		ClaudeAPIKey: getEnv("CLAUDE_API_KEY", getEnv("ANTHROPIC_API_KEY", "")),
		OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),
		XAIAPIKey:    getEnv("XAI_API_KEY", ""),
		OllamaURL:    getEnv("OLLAMA_URL", "http://localhost:11434"),

		// Search settings
		SearchProvider: getEnv("SEARCH_PROVIDER", "serper"),
		SerperAPIKey:   getEnv("SERPER_API_KEY", ""),
		SerpAPIKey:     getEnv("SERPAPI_API_KEY", ""),

		// Agent URLs
		AgentURLs: make(map[string]string),

		// A2A Protocol
		A2AEnabled:   getEnv("A2A_ENABLED", "true") == "true",
		A2AAuthType:  getEnv("A2A_AUTH_TYPE", "apikey"),
		A2AAuthToken: getEnv("A2A_AUTH_TOKEN", ""),

		// Observability
		ObservabilityEnabled:  getEnv("OBSERVABILITY_ENABLED", "false") == "true",
		ObservabilityProvider: getEnv("OBSERVABILITY_PROVIDER", "opik"),
		ObservabilityAPIKey:   getEnv("OBSERVABILITY_API_KEY", getEnv("OPIK_API_KEY", "")),
		ObservabilityEndpoint: getEnv("OBSERVABILITY_ENDPOINT", ""),
		ObservabilityProject:  getEnv("OBSERVABILITY_PROJECT", "agentkit"),

		// Security
		SecurityEnabled:      getEnv("SECURITY_ENABLED", "false") == "true",
		SecurityMinScore:     50,
		SecurityRequireEncry: getEnv("SECURITY_REQUIRE_ENCRYPTION", "false") == "true",
	}

	// Set LLMAPIKey based on provider if not explicitly set
	if cfg.LLMAPIKey == "" {
		switch provider {
		case "gemini":
			cfg.LLMAPIKey = cfg.GeminiAPIKey
		case "claude":
			cfg.LLMAPIKey = cfg.ClaudeAPIKey
		case "openai":
			cfg.LLMAPIKey = cfg.OpenAIAPIKey
		case "xai":
			cfg.LLMAPIKey = cfg.XAIAPIKey
		}
	}

	// Set LLMBaseURL for Ollama if not explicitly set
	if cfg.LLMBaseURL == "" && provider == "ollama" {
		cfg.LLMBaseURL = cfg.OllamaURL
	}

	return cfg
}

// GetDefaultModel returns the default model for a given provider.
func GetDefaultModel(provider string) string {
	switch provider {
	case "gemini":
		return "gemini-2.0-flash-exp"
	case "claude":
		return "claude-sonnet-4-20250514"
	case "openai":
		return "gpt-4o"
	case "xai":
		return "grok-3"
	case "ollama":
		return "llama3.2:latest"
	default:
		return "gemini-2.0-flash-exp"
	}
}

// SetAgentURL sets a URL for a named agent.
func (c *Config) SetAgentURL(name, url string) {
	c.AgentURLs[name] = url
}

// GetAgentURL gets the URL for a named agent.
func (c *Config) GetAgentURL(name string) string {
	if url, ok := c.AgentURLs[name]; ok {
		return url
	}
	// Try environment variable fallback
	return getEnv(name+"_URL", "")
}

// getEnv gets an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
