package config

import (
	"context"
	"fmt"

	"github.com/agentplexus/vaultguard"
)

// SecureConfig wraps Config with VaultGuard for secure credential access.
type SecureConfig struct {
	*Config
	vault *vaultguard.SecureVault
}

// LoadSecureConfig loads configuration with VaultGuard security checks.
// It enforces security policies based on the environment (local or cloud).
func LoadSecureConfig(ctx context.Context, opts ...SecureConfigOption) (*SecureConfig, error) {
	options := &secureConfigOptions{
		policy: nil, // Use default policy
	}
	for _, opt := range opts {
		opt(options)
	}

	// Create VaultGuard configuration
	vgConfig := &vaultguard.Config{
		Policy: options.policy,
	}

	// Create secure vault
	sv, err := vaultguard.New(vgConfig)
	if err != nil {
		return nil, fmt.Errorf("security check failed: %w", err)
	}

	// Load base config
	cfg := LoadConfig()

	sc := &SecureConfig{
		Config: cfg,
		vault:  sv,
	}

	// Load sensitive credentials from secure vault
	sc.loadSecureCredentials(ctx)

	return sc, nil
}

// loadSecureCredentials loads API keys from the secure vault.
// Missing credentials are silently skipped as they are optional.
func (sc *SecureConfig) loadSecureCredentials(ctx context.Context) {
	// Load LLM API key if not set
	if sc.LLMAPIKey == "" {
		key, err := sc.GetCredential(ctx, "LLM_API_KEY")
		if err == nil && key != "" {
			sc.LLMAPIKey = key
		}
	}

	// Load provider-specific keys
	if sc.GeminiAPIKey == "" {
		key, err := sc.GetCredential(ctx, "GEMINI_API_KEY")
		if err == nil && key != "" {
			sc.GeminiAPIKey = key
		}
	}

	if sc.ClaudeAPIKey == "" {
		key, err := sc.GetCredential(ctx, "CLAUDE_API_KEY")
		if err == nil && key != "" {
			sc.ClaudeAPIKey = key
		}
	}

	if sc.OpenAIAPIKey == "" {
		key, err := sc.GetCredential(ctx, "OPENAI_API_KEY")
		if err == nil && key != "" {
			sc.OpenAIAPIKey = key
		}
	}

	if sc.XAIAPIKey == "" {
		key, err := sc.GetCredential(ctx, "XAI_API_KEY")
		if err == nil && key != "" {
			sc.XAIAPIKey = key
		}
	}

	// Load search API keys
	if sc.SerperAPIKey == "" {
		key, err := sc.GetCredential(ctx, "SERPER_API_KEY")
		if err == nil && key != "" {
			sc.SerperAPIKey = key
		}
	}

	if sc.SerpAPIKey == "" {
		key, err := sc.GetCredential(ctx, "SERPAPI_API_KEY")
		if err == nil && key != "" {
			sc.SerpAPIKey = key
		}
	}

	// Update LLMAPIKey based on provider if still not set
	if sc.LLMAPIKey == "" {
		switch sc.LLMProvider {
		case "gemini":
			sc.LLMAPIKey = sc.GeminiAPIKey
		case "claude":
			sc.LLMAPIKey = sc.ClaudeAPIKey
		case "openai":
			sc.LLMAPIKey = sc.OpenAIAPIKey
		case "xai":
			sc.LLMAPIKey = sc.XAIAPIKey
		}
	}
}

// GetCredential retrieves a credential from the secure vault.
func (sc *SecureConfig) GetCredential(ctx context.Context, name string) (string, error) {
	return sc.vault.GetValue(ctx, name)
}

// GetRequiredCredentials retrieves multiple credentials, failing if any are missing.
func (sc *SecureConfig) GetRequiredCredentials(ctx context.Context, names ...string) (map[string]string, error) {
	result := make(map[string]string)
	for _, name := range names {
		value, err := sc.vault.GetValue(ctx, name)
		if err != nil {
			return nil, fmt.Errorf("required credential %s not found: %w", name, err)
		}
		if value == "" {
			return nil, fmt.Errorf("required credential %s is empty", name)
		}
		result[name] = value
	}
	return result, nil
}

// Environment returns the detected deployment environment.
func (sc *SecureConfig) Environment() vaultguard.Environment {
	return sc.vault.Environment()
}

// SecurityResult returns the security assessment result.
func (sc *SecureConfig) SecurityResult() *vaultguard.SecurityResult {
	return sc.vault.SecurityResult()
}

// Close cleans up resources.
func (sc *SecureConfig) Close() error {
	if sc.vault != nil {
		return sc.vault.Close()
	}
	return nil
}

// SecureConfigOption configures secure config loading.
type SecureConfigOption func(*secureConfigOptions)

type secureConfigOptions struct {
	policy *vaultguard.Policy
}

// WithPolicy sets a custom security policy.
func WithPolicy(policy *vaultguard.Policy) SecureConfigOption {
	return func(o *secureConfigOptions) {
		o.policy = policy
	}
}

// WithDevPolicy uses a permissive development policy.
func WithDevPolicy() SecureConfigOption {
	return func(o *secureConfigOptions) {
		o.policy = vaultguard.DevelopmentPolicy()
	}
}

// WithStrictPolicy uses a strict security policy.
func WithStrictPolicy() SecureConfigOption {
	return func(o *secureConfigOptions) {
		o.policy = vaultguard.StrictPolicy()
	}
}
