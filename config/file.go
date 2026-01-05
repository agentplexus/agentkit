// Package config provides configuration file loading for agent applications.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ConfigFile represents the structure of config.json/config.yaml.
// This is the source of truth for non-secret configuration.
type ConfigFile struct {
	// LLM configuration
	LLM LLMConfig `json:"llm" yaml:"llm"`

	// Search configuration
	Search SearchConfig `json:"search" yaml:"search"`

	// Observability configuration
	Observability ObservabilityConfig `json:"observability" yaml:"observability"`

	// Agent URLs for multi-agent systems
	Agents map[string]AgentConfig `json:"agents" yaml:"agents"`

	// A2A Protocol configuration
	A2A A2AConfig `json:"a2a" yaml:"a2a"`

	// Security configuration
	Security SecurityConfig `json:"security" yaml:"security"`

	// Secrets configuration (provider settings, not actual secrets)
	Secrets SecretsFileConfig `json:"secrets" yaml:"secrets"`

	// Environment overrides (optional)
	Environment string `json:"environment" yaml:"environment"`
}

// LLMConfig holds LLM provider configuration.
type LLMConfig struct {
	Provider string `json:"provider" yaml:"provider"` // gemini, claude, openai, ollama, xai
	Model    string `json:"model" yaml:"model"`       // Model name override
	BaseURL  string `json:"baseUrl" yaml:"baseUrl"`   // Custom endpoint (for ollama)
}

// SearchConfig holds search provider configuration.
type SearchConfig struct {
	Provider string `json:"provider" yaml:"provider"` // serper, serpapi
}

// ObservabilityConfig holds observability settings.
type ObservabilityConfig struct {
	Enabled  bool   `json:"enabled" yaml:"enabled"`
	Provider string `json:"provider" yaml:"provider"` // opik, langfuse, phoenix
	Endpoint string `json:"endpoint" yaml:"endpoint"` // Custom endpoint
	Project  string `json:"project" yaml:"project"`   // Project name
}

// AgentConfig holds configuration for a single agent in multi-agent systems.
type AgentConfig struct {
	URL         string `json:"url" yaml:"url"`
	Description string `json:"description" yaml:"description"`
}

// A2AConfig holds A2A protocol configuration.
type A2AConfig struct {
	Enabled  bool   `json:"enabled" yaml:"enabled"`
	AuthType string `json:"authType" yaml:"authType"` // jwt, apikey, oauth2
}

// SecurityConfig holds security settings.
type SecurityConfig struct {
	Enabled           bool `json:"enabled" yaml:"enabled"`
	MinScore          int  `json:"minScore" yaml:"minScore"`
	RequireEncryption bool `json:"requireEncryption" yaml:"requireEncryption"`
}

// SecretsFileConfig holds secrets provider configuration (not actual secrets).
type SecretsFileConfig struct {
	Provider string `json:"provider" yaml:"provider"` // env, aws-sm, aws-ssm
	Prefix   string `json:"prefix" yaml:"prefix"`     // Secret path prefix
	Region   string `json:"region" yaml:"region"`     // AWS region
}

// LoadConfigFile loads configuration from a JSON or YAML file.
// It searches in the following order:
//  1. Explicit path provided
//  2. config.json in current directory
//  3. config.yaml in current directory
//  4. ../config.json (parent directory)
//  5. ~/.agentplexus/projects/{project}/config.json
func LoadConfigFile(path string, projectName string) (*ConfigFile, error) {
	var configPath string

	if path != "" {
		configPath = path
	} else {
		// Search for config file
		var err error
		configPath, err = findConfigFile(projectName)
		if err != nil {
			// Return empty config if no file found (use defaults)
			return &ConfigFile{}, nil
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg ConfigFile

	// Determine format based on extension
	ext := strings.ToLower(filepath.Ext(configPath))
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parsing JSON config: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parsing YAML config: %w", err)
		}
	default:
		// Try JSON first, then YAML
		if err := json.Unmarshal(data, &cfg); err != nil {
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return nil, fmt.Errorf("parsing config file (unknown format): %w", err)
			}
		}
	}

	return &cfg, nil
}

// findConfigFile searches for a config file in standard locations.
func findConfigFile(projectName string) (string, error) {
	candidates := []string{
		"config.json",
		"config.yaml",
		"config.yml",
		"../config.json",
		"../config.yaml",
	}

	// Add project-specific path
	if projectName != "" {
		if home, err := os.UserHomeDir(); err == nil {
			candidates = append(candidates,
				filepath.Join(home, ".agentplexus", "projects", projectName, "config.json"),
				filepath.Join(home, ".agentplexus", "projects", projectName, "config.yaml"),
			)
		}
	}

	// Add global agentplexus config
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates,
			filepath.Join(home, ".agentplexus", "config.json"),
			filepath.Join(home, ".agentplexus", "config.yaml"),
		)
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no config file found")
}

// GetProjectName attempts to detect the project name from config.json stackName
// or falls back to the current directory name.
func GetProjectName() string {
	// Try to read stackName from config.json (CDK config format)
	configPaths := []string{"config.json", "../config.json", "cdk/config.json"}
	for _, path := range configPaths {
		if data, err := os.ReadFile(path); err == nil {
			var cfg struct {
				StackName string `json:"stackName"`
			}
			if json.Unmarshal(data, &cfg) == nil && cfg.StackName != "" {
				return cfg.StackName
			}
		}
	}

	// Fall back to current directory name
	if wd, err := os.Getwd(); err == nil {
		return filepath.Base(wd)
	}

	return ""
}

// Defaults returns a ConfigFile with sensible defaults.
func (c *ConfigFile) Defaults() *ConfigFile {
	if c.LLM.Provider == "" {
		c.LLM.Provider = "gemini"
	}
	if c.LLM.Model == "" {
		c.LLM.Model = GetDefaultModel(c.LLM.Provider)
	}
	if c.Search.Provider == "" {
		c.Search.Provider = "serper"
	}
	if c.Observability.Provider == "" {
		c.Observability.Provider = "opik"
	}
	if c.Observability.Project == "" {
		c.Observability.Project = "agentkit"
	}
	if c.A2A.AuthType == "" {
		c.A2A.AuthType = "apikey"
	}
	if c.Security.MinScore == 0 {
		c.Security.MinScore = 50
	}
	if c.Secrets.Provider == "" {
		c.Secrets.Provider = "env"
	}
	return c
}

// MergeEnv merges environment variable overrides into the config.
// Environment variables take precedence over file values.
func (c *ConfigFile) MergeEnv() *ConfigFile {
	// LLM overrides
	if v := os.Getenv("LLM_PROVIDER"); v != "" {
		c.LLM.Provider = v
	}
	if v := os.Getenv("LLM_MODEL"); v != "" {
		c.LLM.Model = v
	}
	if v := os.Getenv("LLM_BASE_URL"); v != "" {
		c.LLM.BaseURL = v
	}

	// Search overrides
	if v := os.Getenv("SEARCH_PROVIDER"); v != "" {
		c.Search.Provider = v
	}

	// Observability overrides
	if v := os.Getenv("OBSERVABILITY_ENABLED"); v == "true" {
		c.Observability.Enabled = true
	}
	if v := os.Getenv("OBSERVABILITY_PROVIDER"); v != "" {
		c.Observability.Provider = v
	}
	if v := os.Getenv("OBSERVABILITY_ENDPOINT"); v != "" {
		c.Observability.Endpoint = v
	}
	if v := os.Getenv("OBSERVABILITY_PROJECT"); v != "" {
		c.Observability.Project = v
	}

	// A2A overrides
	if v := os.Getenv("A2A_ENABLED"); v != "" {
		c.A2A.Enabled = v == "true"
	}
	if v := os.Getenv("A2A_AUTH_TYPE"); v != "" {
		c.A2A.AuthType = v
	}

	// Security overrides
	if v := os.Getenv("SECURITY_ENABLED"); v == "true" {
		c.Security.Enabled = true
	}
	if v := os.Getenv("SECURITY_REQUIRE_ENCRYPTION"); v == "true" {
		c.Security.RequireEncryption = true
	}

	// Secrets provider overrides
	if v := os.Getenv("SECRETS_PROVIDER"); v != "" {
		c.Secrets.Provider = v
	}
	if v := os.Getenv("SECRETS_PREFIX"); v != "" {
		c.Secrets.Prefix = v
	}
	if v := os.Getenv("AWS_REGION"); v != "" && c.Secrets.Region == "" {
		c.Secrets.Region = v
	}

	return c
}
