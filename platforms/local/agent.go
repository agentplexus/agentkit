package local

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EmbeddedAgent is a lightweight agent that runs in-process.
type EmbeddedAgent struct {
	name         string
	description  string
	instructions string
	tools        []Tool
	llm          LLMClient
	maxTokens    int
}

// LLMClient defines the interface for language model interactions.
type LLMClient interface {
	// Complete generates a completion for the given messages.
	Complete(ctx context.Context, messages []Message, tools []ToolDefinition) (*CompletionResponse, error)
}

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"` // "system", "user", "assistant", "tool"
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`    // For tool messages
	ToolID  string `json:"tool_id,omitempty"` // For tool messages
}

// ToolDefinition defines a tool for the LLM.
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolCall represents an LLM's request to call a tool.
type ToolCall struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// CompletionResponse holds the LLM response.
type CompletionResponse struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	Done      bool       `json:"done"`
}

// NewEmbeddedAgent creates a new embedded agent.
func NewEmbeddedAgent(cfg AgentConfig, toolSet *ToolSet, llm LLMClient) (*EmbeddedAgent, error) {
	// Load instructions
	instructions := cfg.Instructions
	if strings.HasSuffix(cfg.Instructions, ".md") {
		content, err := os.ReadFile(cfg.Instructions)
		if err != nil {
			// Try relative to workspace if absolute path fails
			content, err = os.ReadFile(filepath.Join(toolSet.workspace, cfg.Instructions))
			if err != nil {
				return nil, fmt.Errorf("failed to load instructions: %w", err)
			}
		}
		instructions = string(content)
	}

	// Create tools
	tools, err := toolSet.CreateTools(cfg.Tools)
	if err != nil {
		return nil, fmt.Errorf("failed to create tools: %w", err)
	}

	maxTokens := cfg.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	return &EmbeddedAgent{
		name:         cfg.Name,
		description:  cfg.Description,
		instructions: instructions,
		tools:        tools,
		llm:          llm,
		maxTokens:    maxTokens,
	}, nil
}

// Name returns the agent's name.
func (a *EmbeddedAgent) Name() string {
	return a.name
}

// Description returns the agent's description.
func (a *EmbeddedAgent) Description() string {
	return a.description
}

// Invoke runs the agent with the given input and returns the result.
func (a *EmbeddedAgent) Invoke(ctx context.Context, input string) (*AgentResult, error) {
	// Build initial messages
	messages := []Message{
		{Role: "system", Content: a.instructions},
		{Role: "user", Content: input},
	}

	// Build tool definitions
	toolDefs := a.buildToolDefinitions()

	// Agent loop - handle tool calls until done
	maxIterations := 10
	for i := 0; i < maxIterations; i++ {
		// Get completion from LLM
		resp, err := a.llm.Complete(ctx, messages, toolDefs)
		if err != nil {
			return nil, fmt.Errorf("LLM completion failed: %w", err)
		}

		// If no tool calls, we're done
		if len(resp.ToolCalls) == 0 || resp.Done {
			return &AgentResult{
				Agent:   a.name,
				Input:   input,
				Output:  resp.Content,
				Success: true,
			}, nil
		}

		// Add assistant message with tool calls
		messages = append(messages, Message{
			Role:    "assistant",
			Content: resp.Content,
		})

		// Execute tool calls
		for _, tc := range resp.ToolCalls {
			result, err := a.executeTool(ctx, tc)

			var resultContent string
			if err != nil {
				resultContent = fmt.Sprintf("Error: %v", err)
			} else {
				// Marshal result to JSON
				resultBytes, _ := json.Marshal(result)
				resultContent = string(resultBytes)
			}

			messages = append(messages, Message{
				Role:    "tool",
				Content: resultContent,
				Name:    tc.Name,
				ToolID:  tc.ID,
			})
		}
	}

	return &AgentResult{
		Agent:   a.name,
		Input:   input,
		Output:  "Max iterations reached",
		Success: false,
		Error:   "agent loop exceeded maximum iterations",
	}, nil
}

// buildToolDefinitions creates tool definitions for the LLM.
func (a *EmbeddedAgent) buildToolDefinitions() []ToolDefinition {
	var defs []ToolDefinition
	for _, tool := range a.tools {
		def := ToolDefinition{
			Name:        tool.Name(),
			Description: tool.Description(),
			Parameters:  a.getToolParameters(tool.Name()),
		}
		defs = append(defs, def)
	}
	return defs
}

// getToolParameters returns the parameter schema for a tool.
func (a *EmbeddedAgent) getToolParameters(name string) map[string]interface{} {
	switch name {
	case "read":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the file to read",
				},
			},
			"required": []string{"path"},
		}
	case "write":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the file to write",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "Content to write to the file",
				},
			},
			"required": []string{"path", "content"},
		}
	case "glob":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pattern": map[string]interface{}{
					"type":        "string",
					"description": "Glob pattern to match files",
				},
			},
			"required": []string{"pattern"},
		}
	case "grep":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pattern": map[string]interface{}{
					"type":        "string",
					"description": "Regex pattern to search for",
				},
				"file_pattern": map[string]interface{}{
					"type":        "string",
					"description": "Optional file name pattern to filter files",
				},
			},
			"required": []string{"pattern"},
		}
	case "shell":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "Shell command to execute",
				},
			},
			"required": []string{"command"},
		}
	default:
		return map[string]interface{}{"type": "object"}
	}
}

// executeTool executes a tool call and returns the result.
func (a *EmbeddedAgent) executeTool(ctx context.Context, tc ToolCall) (any, error) {
	for _, tool := range a.tools {
		if tool.Name() == tc.Name {
			return tool.Execute(ctx, tc.Arguments)
		}
	}
	return nil, fmt.Errorf("unknown tool: %s", tc.Name)
}

// AgentResult holds the result of an agent invocation.
type AgentResult struct {
	Agent   string `json:"agent"`
	Input   string `json:"input"`
	Output  string `json:"output"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
