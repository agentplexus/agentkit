package agentcore

import (
	"context"
	"fmt"
	"sync"
)

// Registry manages a collection of agents and routes requests to them.
type Registry struct {
	mu           sync.RWMutex
	agents       map[string]Agent
	defaultAgent string
}

// NewRegistry creates a new agent registry.
func NewRegistry() *Registry {
	return &Registry{
		agents: make(map[string]Agent),
	}
}

// Register adds an agent to the registry.
// If the agent implements Initializer, Initialize() is called.
// Returns an error if an agent with the same name already exists.
func (r *Registry) Register(ctx context.Context, agent Agent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := agent.Name()
	if _, exists := r.agents[name]; exists {
		return fmt.Errorf("agent already registered: %s", name)
	}

	// Call Initialize if the agent supports it
	if init, ok := agent.(Initializer); ok {
		if err := init.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize agent %s: %w", name, err)
		}
	}

	r.agents[name] = agent
	return nil
}

// MustRegister is like Register but panics on error.
// Useful for initialization code where registration should never fail.
func (r *Registry) MustRegister(ctx context.Context, agent Agent) {
	if err := r.Register(ctx, agent); err != nil {
		panic(err)
	}
}

// RegisterAll registers multiple agents.
// Stops on the first error.
func (r *Registry) RegisterAll(ctx context.Context, agents ...Agent) error {
	for _, agent := range agents {
		if err := r.Register(ctx, agent); err != nil {
			return err
		}
	}
	return nil
}

// SetDefault sets the default agent to use when no agent is specified.
func (r *Registry) SetDefault(name string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if _, exists := r.agents[name]; !exists {
		return fmt.Errorf("agent not found: %s", name)
	}
	r.defaultAgent = name
	return nil
}

// Get retrieves an agent by name.
// If name is empty, returns the default agent (if set).
func (r *Registry) Get(name string) (Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if name == "" {
		name = r.defaultAgent
	}

	if name == "" {
		return nil, fmt.Errorf("no agent specified and no default agent set")
	}

	agent, exists := r.agents[name]
	if !exists {
		return nil, fmt.Errorf("agent not found: %s", name)
	}
	return agent, nil
}

// List returns the names of all registered agents.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.agents))
	for name := range r.agents {
		names = append(names, name)
	}
	return names
}

// Count returns the number of registered agents.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.agents)
}

// HealthCheck checks the health of all agents that implement HealthChecker.
// Returns a map of agent names to their health status (nil = healthy).
func (r *Registry) HealthCheck(ctx context.Context) map[string]error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make(map[string]error)
	for name, agent := range r.agents {
		if hc, ok := agent.(HealthChecker); ok {
			results[name] = hc.HealthCheck(ctx)
		} else {
			results[name] = nil // Assume healthy if no health check
		}
	}
	return results
}

// Close closes all agents that implement Closer.
// Collects all errors and returns them as a combined error.
func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errs []error
	for name, agent := range r.agents {
		if closer, ok := agent.(Closer); ok {
			if err := closer.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close agent %s: %w", name, err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing agents: %v", errs)
	}
	return nil
}

// Invoke routes a request to the appropriate agent and invokes it.
// This is a convenience method that combines Get and Invoke.
func (r *Registry) Invoke(ctx context.Context, req Request) (Response, error) {
	agent, err := r.Get(req.Agent)
	if err != nil {
		return Response{}, err
	}
	return agent.Invoke(ctx, req)
}
