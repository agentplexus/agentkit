package local

import (
	"context"
	"fmt"
	"log"
	"sync"
)

// Runner orchestrates multiple embedded agents.
type Runner struct {
	config  *Config
	agents  map[string]*EmbeddedAgent
	toolSet *ToolSet
	llm     LLMClient
	mu      sync.RWMutex
}

// NewRunner creates a new agent runner.
func NewRunner(cfg *Config, llm LLMClient) (*Runner, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	toolSet := NewToolSet(cfg.Workspace)

	runner := &Runner{
		config:  cfg,
		agents:  make(map[string]*EmbeddedAgent),
		toolSet: toolSet,
		llm:     llm,
	}

	// Initialize all configured agents
	for _, agentCfg := range cfg.Agents {
		agent, err := NewEmbeddedAgent(agentCfg, toolSet, llm)
		if err != nil {
			return nil, fmt.Errorf("failed to create agent %s: %w", agentCfg.Name, err)
		}
		runner.agents[agentCfg.Name] = agent
		log.Printf("[Runner] Registered agent: %s", agentCfg.Name)
	}

	return runner, nil
}

// Invoke runs a single agent synchronously.
func (r *Runner) Invoke(ctx context.Context, agentName, input string) (*AgentResult, error) {
	r.mu.RLock()
	agent, ok := r.agents[agentName]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("agent not found: %s", agentName)
	}

	log.Printf("[Runner] Invoking agent: %s", agentName)
	result, err := agent.Invoke(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("agent invocation failed: %w", err)
	}

	log.Printf("[Runner] Agent %s completed: success=%v", agentName, result.Success)
	return result, nil
}

// AgentTask represents a task to be executed by an agent.
type AgentTask struct {
	Agent string `json:"agent"`
	Input string `json:"input"`
}

// InvokeParallel runs multiple agents concurrently.
func (r *Runner) InvokeParallel(ctx context.Context, tasks []AgentTask) ([]*AgentResult, error) {
	if len(tasks) == 0 {
		return nil, nil
	}

	log.Printf("[Runner] Starting parallel execution of %d agents", len(tasks))

	results := make([]*AgentResult, len(tasks))
	errors := make([]error, len(tasks))
	var wg sync.WaitGroup

	for i, task := range tasks {
		wg.Add(1)
		go func(idx int, t AgentTask) {
			defer wg.Done()

			result, err := r.Invoke(ctx, t.Agent, t.Input)
			if err != nil {
				errors[idx] = err
				results[idx] = &AgentResult{
					Agent:   t.Agent,
					Input:   t.Input,
					Success: false,
					Error:   err.Error(),
				}
			} else {
				results[idx] = result
			}
		}(i, task)
	}

	wg.Wait()

	// Check for errors
	var errCount int
	for _, err := range errors {
		if err != nil {
			errCount++
		}
	}

	log.Printf("[Runner] Parallel execution completed: %d/%d successful", len(tasks)-errCount, len(tasks))

	return results, nil
}

// InvokeSequential runs multiple agents in sequence, passing context between them.
func (r *Runner) InvokeSequential(ctx context.Context, tasks []AgentTask) ([]*AgentResult, error) {
	if len(tasks) == 0 {
		return nil, nil
	}

	log.Printf("[Runner] Starting sequential execution of %d agents", len(tasks))

	results := make([]*AgentResult, 0, len(tasks))
	var contextBuilder string

	for i, task := range tasks {
		// Build input with context from previous results
		input := task.Input
		if contextBuilder != "" && i > 0 {
			input = fmt.Sprintf("Previous context:\n%s\n\nCurrent task:\n%s", contextBuilder, task.Input)
		}

		result, err := r.Invoke(ctx, task.Agent, input)
		if err != nil {
			result = &AgentResult{
				Agent:   task.Agent,
				Input:   task.Input,
				Success: false,
				Error:   err.Error(),
			}
		}

		results = append(results, result)

		// Build context for next agent
		if result.Success {
			contextBuilder += fmt.Sprintf("\n[%s]: %s\n", task.Agent, result.Output)
		}
	}

	return results, nil
}

// ListAgents returns the names of all registered agents.
func (r *Runner) ListAgents() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.agents))
	for name := range r.agents {
		names = append(names, name)
	}
	return names
}

// GetAgent returns an agent by name.
func (r *Runner) GetAgent(name string) (*EmbeddedAgent, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, ok := r.agents[name]
	return agent, ok
}

// GetAgentInfo returns information about an agent.
func (r *Runner) GetAgentInfo(name string) (*AgentInfo, error) {
	r.mu.RLock()
	agent, ok := r.agents[name]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("agent not found: %s", name)
	}

	return &AgentInfo{
		Name:        agent.Name(),
		Description: agent.Description(),
	}, nil
}

// AgentInfo holds basic information about an agent.
type AgentInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ListAgentInfo returns information about all registered agents.
func (r *Runner) ListAgentInfo() []AgentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	infos := make([]AgentInfo, 0, len(r.agents))
	for _, agent := range r.agents {
		infos = append(infos, AgentInfo{
			Name:        agent.Name(),
			Description: agent.Description(),
		})
	}
	return infos
}

// Workspace returns the workspace path.
func (r *Runner) Workspace() string {
	return r.config.Workspace
}

// ToolSet returns the tool set.
func (r *Runner) ToolSet() *ToolSet {
	return r.toolSet
}

// Close cleans up resources.
func (r *Runner) Close() error {
	// No resources to clean up currently
	return nil
}

// OrchestratedTask represents a high-level task that may involve multiple agents.
type OrchestratedTask struct {
	// Name is a descriptive name for the task.
	Name string `json:"name"`

	// Agents lists the agents to involve, in order of execution for sequential,
	// or all at once for parallel.
	Agents []string `json:"agents"`

	// Input is the task description/prompt.
	Input string `json:"input"`

	// Mode is "parallel" or "sequential".
	Mode string `json:"mode"`
}

// ExecuteOrchestrated runs an orchestrated task involving multiple agents.
func (r *Runner) ExecuteOrchestrated(ctx context.Context, task OrchestratedTask) (*OrchestratedResult, error) {
	log.Printf("[Runner] Executing orchestrated task: %s (mode=%s, agents=%v)",
		task.Name, task.Mode, task.Agents)

	// Build agent tasks
	tasks := make([]AgentTask, len(task.Agents))
	for i, agentName := range task.Agents {
		tasks[i] = AgentTask{
			Agent: agentName,
			Input: task.Input,
		}
	}

	var results []*AgentResult
	var err error

	switch task.Mode {
	case "parallel":
		results, err = r.InvokeParallel(ctx, tasks)
	case "sequential":
		results, err = r.InvokeSequential(ctx, tasks)
	default:
		return nil, fmt.Errorf("unknown mode: %s", task.Mode)
	}

	if err != nil {
		return nil, err
	}

	// Aggregate results
	return &OrchestratedResult{
		Task:    task.Name,
		Mode:    task.Mode,
		Results: results,
	}, nil
}

// OrchestratedResult holds the results of an orchestrated task.
type OrchestratedResult struct {
	Task    string         `json:"task"`
	Mode    string         `json:"mode"`
	Results []*AgentResult `json:"results"`
}

// AllSuccessful returns true if all agent results were successful.
func (r *OrchestratedResult) AllSuccessful() bool {
	for _, result := range r.Results {
		if !result.Success {
			return false
		}
	}
	return true
}

// Summary returns a summary of all results.
func (r *OrchestratedResult) Summary() string {
	var summary string
	for _, result := range r.Results {
		status := "SUCCESS"
		if !result.Success {
			status = "FAILED"
		}
		summary += fmt.Sprintf("[%s] %s: %s\n", result.Agent, status, truncate(result.Output, 200))
	}
	return summary
}

// truncate truncates a string to the given length.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
