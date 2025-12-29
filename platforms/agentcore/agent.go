package agentcore

import (
	"context"
	"encoding/json"
)

// Request represents an AgentCore invocation request.
// This is passed to agents when /invocations is called.
type Request struct {
	// Prompt is the user's input/query to the agent.
	Prompt string `json:"prompt"`

	// SessionID is the AgentCore session identifier.
	// AgentCore provides session isolation via Firecracker microVMs.
	SessionID string `json:"session_id,omitempty"`

	// Agent specifies which agent to invoke (for multi-agent setups).
	// If empty, the server's DefaultAgent is used.
	Agent string `json:"agent,omitempty"`

	// Metadata contains additional context passed to the agent.
	Metadata map[string]string `json:"metadata,omitempty"`

	// RawInput contains the full raw JSON input for custom parsing.
	// Use this when your agent needs access to fields beyond the standard ones.
	RawInput json.RawMessage `json:"-"`
}

// Response represents an AgentCore invocation response.
type Response struct {
	// Output is the agent's response text.
	Output string `json:"output"`

	// Metadata contains additional response metadata.
	Metadata map[string]string `json:"metadata,omitempty"`

	// Error contains error information if the invocation failed.
	// This is separate from HTTP errors for partial failure scenarios.
	Error string `json:"error,omitempty"`
}

// Agent is the interface that AgentCore-compatible agents must implement.
// This interface is designed to be simple and runtime-agnostic.
type Agent interface {
	// Name returns the unique identifier for this agent.
	// Used for routing in multi-agent setups.
	Name() string

	// Invoke processes a request and returns a response.
	// The context carries session information and cancellation signals.
	Invoke(ctx context.Context, req Request) (Response, error)
}

// AgentFunc is a function type that implements the Agent interface.
// Useful for simple agents that don't need state.
type AgentFunc struct {
	name   string
	invoke func(ctx context.Context, req Request) (Response, error)
}

// NewAgentFunc creates an Agent from a function.
func NewAgentFunc(name string, fn func(ctx context.Context, req Request) (Response, error)) *AgentFunc {
	return &AgentFunc{name: name, invoke: fn}
}

// Name returns the agent name.
func (a *AgentFunc) Name() string {
	return a.name
}

// Invoke calls the underlying function.
func (a *AgentFunc) Invoke(ctx context.Context, req Request) (Response, error) {
	return a.invoke(ctx, req)
}

// HealthChecker is an optional interface for agents that support health checks.
// If an agent implements this, the server will call it for /ping requests.
type HealthChecker interface {
	// HealthCheck returns nil if the agent is healthy, error otherwise.
	HealthCheck(ctx context.Context) error
}

// Initializer is an optional interface for agents that need initialization.
// Called once when the agent is registered with the server.
type Initializer interface {
	// Initialize is called when the agent is registered.
	Initialize(ctx context.Context) error
}

// Closer is an optional interface for agents that need cleanup.
// Called when the server is shutting down.
type Closer interface {
	// Close releases resources held by the agent.
	Close() error
}
