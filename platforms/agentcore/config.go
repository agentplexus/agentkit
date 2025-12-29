// Package agentcore provides a runtime adapter for AWS Bedrock AgentCore.
//
// AgentCore is AWS's serverless runtime for AI agents, providing:
//   - Firecracker microVM isolation per session
//   - Automatic scaling from zero
//   - Built-in session memory and identity
//   - Pay-per-use pricing (only active CPU time)
//
// This package implements the AgentCore HTTP contract (/ping, /invocations)
// and allows AgentKit agents to run on AgentCore without code changes.
//
// Note: Helm/Kubernetes configuration does NOT apply to AgentCore.
// Use AWS CDK, CloudFormation, or Terraform for AgentCore deployment.
package agentcore

import (
	"os"
	"strconv"
	"time"
)

// Config holds configuration for an AgentCore runtime server.
type Config struct {
	// Port is the port to listen on. Default is 8080 (AgentCore standard).
	Port int

	// ReadTimeout is the maximum duration for reading requests.
	// Default is 30 seconds.
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration for writing responses.
	// Default is 300 seconds (5 minutes) for long-running agent operations.
	WriteTimeout time.Duration

	// IdleTimeout is the maximum time to wait for the next request.
	// Default is 60 seconds.
	IdleTimeout time.Duration

	// DefaultAgent is the agent to use when no agent is specified in the request.
	// If empty, the "agent" field is required in invocation requests.
	DefaultAgent string

	// EnableRequestLogging enables logging of incoming requests.
	// Default is true.
	EnableRequestLogging bool

	// EnableSessionTracking enables session ID tracking in logs.
	// Default is true.
	EnableSessionTracking bool
}

// DefaultConfig returns a Config with sensible defaults for AgentCore.
func DefaultConfig() Config {
	return Config{
		Port:                  8080,
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          300 * time.Second, // 5 min for long agent operations
		IdleTimeout:           60 * time.Second,
		EnableRequestLogging:  true,
		EnableSessionTracking: true,
	}
}

// LoadConfigFromEnv loads configuration from environment variables.
// Environment variables:
//   - AGENTCORE_PORT: Port to listen on (default: 8080)
//   - AGENTCORE_DEFAULT_AGENT: Default agent name
//   - AGENTCORE_READ_TIMEOUT_SECS: Read timeout in seconds
//   - AGENTCORE_WRITE_TIMEOUT_SECS: Write timeout in seconds
//   - AGENTCORE_ENABLE_REQUEST_LOGGING: Enable request logging (true/false)
func LoadConfigFromEnv() Config {
	cfg := DefaultConfig()

	if port := os.Getenv("AGENTCORE_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Port = p
		}
	}

	if agent := os.Getenv("AGENTCORE_DEFAULT_AGENT"); agent != "" {
		cfg.DefaultAgent = agent
	}

	if timeout := os.Getenv("AGENTCORE_READ_TIMEOUT_SECS"); timeout != "" {
		if t, err := strconv.Atoi(timeout); err == nil {
			cfg.ReadTimeout = time.Duration(t) * time.Second
		}
	}

	if timeout := os.Getenv("AGENTCORE_WRITE_TIMEOUT_SECS"); timeout != "" {
		if t, err := strconv.Atoi(timeout); err == nil {
			cfg.WriteTimeout = time.Duration(t) * time.Second
		}
	}

	if logging := os.Getenv("AGENTCORE_ENABLE_REQUEST_LOGGING"); logging != "" {
		cfg.EnableRequestLogging = logging == "true" || logging == "1"
	}

	if tracking := os.Getenv("AGENTCORE_ENABLE_SESSION_TRACKING"); tracking != "" {
		cfg.EnableSessionTracking = tracking == "true" || tracking == "1"
	}

	return cfg
}
