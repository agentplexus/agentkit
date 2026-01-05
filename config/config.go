// Package config provides configuration management for agent applications.
// It supports loading from config files (JSON/YAML), environment variables,
// and integrates with OmniVault for unified secret management.
//
// Configuration sources (in order of precedence):
//  1. Environment variables (highest)
//  2. Config file (config.json or config.yaml)
//  3. Defaults (lowest)
//
// Secrets are loaded separately via OmniVault providers.
package config

import (
	"context"
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

	// Secrets Configuration (OmniVault)
	secrets *SecretsClient
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

// LoadConfigWithSecrets loads configuration using OmniVault for secrets.
// This is the recommended way to load configuration in production as it
// supports multiple secret backends (env, AWS Secrets Manager, etc.).
func LoadConfigWithSecrets(ctx context.Context, secretsCfg SecretsConfig) (*Config, error) {
	// Create secrets client
	secrets, err := NewSecretsClient(secretsCfg)
	if err != nil {
		return nil, err
	}

	provider := getEnv("LLM_PROVIDER", "gemini")

	cfg := &Config{
		// LLM settings
		LLMProvider: provider,
		LLMModel:    getEnv("LLM_MODEL", GetDefaultModel(provider)),
		LLMBaseURL:  getEnv("LLM_BASE_URL", ""),

		// Search settings
		SearchProvider: getEnv("SEARCH_PROVIDER", "serper"),

		// Agent URLs
		AgentURLs: make(map[string]string),

		// A2A Protocol
		A2AEnabled:   getEnv("A2A_ENABLED", "true") == "true",
		A2AAuthType:  getEnv("A2A_AUTH_TYPE", "apikey"),
		A2AAuthToken: getEnv("A2A_AUTH_TOKEN", ""),

		// Observability
		ObservabilityEnabled:  getEnv("OBSERVABILITY_ENABLED", "false") == "true",
		ObservabilityProvider: getEnv("OBSERVABILITY_PROVIDER", "opik"),
		ObservabilityEndpoint: getEnv("OBSERVABILITY_ENDPOINT", ""),
		ObservabilityProject:  getEnv("OBSERVABILITY_PROJECT", "agentkit"),

		// Security
		SecurityEnabled:      getEnv("SECURITY_ENABLED", "false") == "true",
		SecurityMinScore:     50,
		SecurityRequireEncry: getEnv("SECURITY_REQUIRE_ENCRYPTION", "false") == "true",

		// Secrets client
		secrets: secrets,
	}

	// Load API keys from secrets provider
	cfg.loadSecretsFromProvider(ctx)

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

	return cfg, nil
}

// loadSecretsFromProvider loads API keys from the configured secrets provider.
func (c *Config) loadSecretsFromProvider(ctx context.Context) {
	if c.secrets == nil {
		return
	}

	// Load LLM API keys
	if key, err := c.secrets.Get(ctx, "LLM_API_KEY"); err == nil && key != "" {
		c.LLMAPIKey = key
	}
	if key, err := c.secrets.Get(ctx, "GEMINI_API_KEY"); err == nil && key != "" {
		c.GeminiAPIKey = key
	} else if key, err := c.secrets.Get(ctx, "GOOGLE_API_KEY"); err == nil && key != "" {
		c.GeminiAPIKey = key
	}
	if key, err := c.secrets.Get(ctx, "CLAUDE_API_KEY"); err == nil && key != "" {
		c.ClaudeAPIKey = key
	} else if key, err := c.secrets.Get(ctx, "ANTHROPIC_API_KEY"); err == nil && key != "" {
		c.ClaudeAPIKey = key
	}
	if key, err := c.secrets.Get(ctx, "OPENAI_API_KEY"); err == nil && key != "" {
		c.OpenAIAPIKey = key
	}
	if key, err := c.secrets.Get(ctx, "XAI_API_KEY"); err == nil && key != "" {
		c.XAIAPIKey = key
	}

	// Load search API keys
	if key, err := c.secrets.Get(ctx, "SERPER_API_KEY"); err == nil && key != "" {
		c.SerperAPIKey = key
	}
	if key, err := c.secrets.Get(ctx, "SERPAPI_API_KEY"); err == nil && key != "" {
		c.SerpAPIKey = key
	}

	// Load observability API key
	if key, err := c.secrets.Get(ctx, "OBSERVABILITY_API_KEY"); err == nil && key != "" {
		c.ObservabilityAPIKey = key
	} else if key, err := c.secrets.Get(ctx, "OPIK_API_KEY"); err == nil && key != "" {
		c.ObservabilityAPIKey = key
	}

	// Load Ollama URL
	if url, err := c.secrets.Get(ctx, "OLLAMA_URL"); err == nil && url != "" {
		c.OllamaURL = url
	} else {
		c.OllamaURL = "http://localhost:11434"
	}
}

// GetSecret retrieves a secret from the configured secrets provider.
// Falls back to environment variables if no secrets provider is configured
// or if the secret is not found.
func (c *Config) GetSecret(ctx context.Context, name string) (string, error) {
	if c.secrets != nil {
		return c.secrets.Get(ctx, name)
	}
	// Fallback to environment variable
	if value := os.Getenv(name); value != "" {
		return value, nil
	}
	return "", nil
}

// SecretsProvider returns the configured secrets provider name.
// Returns "env" if no secrets client is configured.
func (c *Config) SecretsProvider() SecretsProvider {
	if c.secrets != nil {
		return c.secrets.Provider()
	}
	return SecretsProviderEnv
}

// Close releases resources held by the config (e.g., secrets client).
func (c *Config) Close() error {
	if c.secrets != nil {
		return c.secrets.Close()
	}
	return nil
}

// LoadOptions configures how configuration is loaded.
type LoadOptions struct {
	// ConfigFile is the path to config.json/config.yaml.
	// If empty, searches in standard locations.
	ConfigFile string

	// ProjectName is used for project-specific config lookup.
	// If empty, auto-detected from config.json stackName or directory name.
	ProjectName string

	// SecretsProvider specifies the secrets backend.
	// If empty, auto-detected based on environment.
	SecretsProvider SecretsProvider

	// SecretsPrefix is prepended to secret paths (e.g., "stats-agent/").
	SecretsPrefix string

	// SecretsRegion is the AWS region for aws-sm/aws-ssm providers.
	SecretsRegion string
}

// Load loads configuration from config file, environment variables, and secrets.
// This is the recommended way to load configuration as it:
//   - Reads settings from config.json (LLM_PROVIDER, SEARCH_PROVIDER, etc.)
//   - Allows environment variable overrides
//   - Loads secrets from OmniVault (API keys)
//
// Example:
//
//	cfg, err := config.Load(ctx, config.LoadOptions{
//	    ConfigFile: "config.json",
//	})
func Load(ctx context.Context, opts LoadOptions) (*Config, error) {
	// Detect project name if not provided
	projectName := opts.ProjectName
	if projectName == "" {
		projectName = GetProjectName()
	}

	// Load config file
	fileCfg, err := LoadConfigFile(opts.ConfigFile, projectName)
	if err != nil {
		return nil, err
	}

	// Apply defaults and merge environment overrides
	fileCfg.Defaults().MergeEnv()

	// Determine secrets configuration
	secretsCfg := SecretsConfig{
		Provider:      SecretsProvider(fileCfg.Secrets.Provider),
		Prefix:        fileCfg.Secrets.Prefix,
		Region:        fileCfg.Secrets.Region,
		FallbackToEnv: true,
	}

	// Override with explicit options
	if opts.SecretsProvider != "" {
		secretsCfg.Provider = opts.SecretsProvider
	}
	if opts.SecretsPrefix != "" {
		secretsCfg.Prefix = opts.SecretsPrefix
	}
	if opts.SecretsRegion != "" {
		secretsCfg.Region = opts.SecretsRegion
	}

	// Create secrets client
	secrets, err := NewSecretsClient(secretsCfg)
	if err != nil {
		return nil, err
	}

	// Build Config from file config
	cfg := &Config{
		// LLM settings from file
		LLMProvider: fileCfg.LLM.Provider,
		LLMModel:    fileCfg.LLM.Model,
		LLMBaseURL:  fileCfg.LLM.BaseURL,

		// Search settings from file
		SearchProvider: fileCfg.Search.Provider,

		// Agent URLs from file
		AgentURLs: make(map[string]string),

		// A2A Protocol from file
		A2AEnabled:  fileCfg.A2A.Enabled,
		A2AAuthType: fileCfg.A2A.AuthType,

		// Observability from file
		ObservabilityEnabled:  fileCfg.Observability.Enabled,
		ObservabilityProvider: fileCfg.Observability.Provider,
		ObservabilityEndpoint: fileCfg.Observability.Endpoint,
		ObservabilityProject:  fileCfg.Observability.Project,

		// Security from file
		SecurityEnabled:      fileCfg.Security.Enabled,
		SecurityMinScore:     fileCfg.Security.MinScore,
		SecurityRequireEncry: fileCfg.Security.RequireEncryption,

		// Secrets client
		secrets: secrets,
	}

	// Copy agent URLs from file
	for name, agent := range fileCfg.Agents {
		cfg.AgentURLs[name] = agent.URL
	}

	// Load API keys from secrets provider
	cfg.loadSecretsFromProvider(ctx)

	// Set LLMAPIKey based on provider if not explicitly set
	if cfg.LLMAPIKey == "" {
		switch cfg.LLMProvider {
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
	if cfg.LLMBaseURL == "" && cfg.LLMProvider == "ollama" {
		cfg.LLMBaseURL = cfg.OllamaURL
	}

	return cfg, nil
}
