# Quick Start

This guide walks you through creating a complete agent with both HTTP and A2A servers.

## Create Your Agent

First, define your agent logic:

```go
package main

import (
    "context"
    "encoding/json"
    "net/http"

    "github.com/agentplexus/agentkit/agent"
    "github.com/agentplexus/agentkit/config"
)

type ResearchAgent struct {
    base *agent.BaseAgent
}

type ResearchRequest struct {
    Query string `json:"query"`
}

type ResearchResponse struct {
    Results []string `json:"results"`
}

func NewResearchAgent(base *agent.BaseAgent) *ResearchAgent {
    return &ResearchAgent{base: base}
}

func (a *ResearchAgent) HandleResearch(w http.ResponseWriter, r *http.Request) {
    var req ResearchRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Your agent logic here
    results := []string{"Result 1", "Result 2"}

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(ResearchResponse{Results: results})
}
```

## Add HTTP Server

Use the httpserver factory to expose your agent:

```go
import "github.com/agentplexus/agentkit/httpserver"

func main() {
    cfg := config.LoadConfig()

    base, _ := agent.NewBaseAgent(cfg, "research-agent", 30)
    defer base.Close()

    researchAgent := NewResearchAgent(base)

    // Create HTTP server with builder pattern
    server, _ := httpserver.NewBuilder("research-agent", 8001).
        WithHandlerFunc("/research", researchAgent.HandleResearch).
        WithDualModeLog().
        Build()

    server.Start()
}
```

Test it:

```bash
curl -X POST http://localhost:8001/research \
  -H "Content-Type: application/json" \
  -d '{"query": "test"}'
```

## Add A2A Server

For agent-to-agent communication, add an A2A server:

```go
import (
    "github.com/agentplexus/agentkit/a2a"
    "github.com/agentplexus/agentkit/httpserver"
)

func main() {
    ctx := context.Background()
    cfg := config.LoadConfig()

    base, _ := agent.NewBaseAgent(cfg, "research-agent", 30)
    defer base.Close()

    researchAgent := NewResearchAgent(base)

    // HTTP server
    httpServer, _ := httpserver.NewBuilder("research-agent", 8001).
        WithHandlerFunc("/research", researchAgent.HandleResearch).
        Build()

    // A2A server
    a2aServer, _ := a2a.NewServer(a2a.Config{
        Agent:       researchAgent.ADKAgent(),
        Port:        "9001",
        Description: "Research agent for web search",
    })

    // Start both
    a2aServer.StartAsync(ctx)
    httpServer.Start()
}
```

## Use Workflow Orchestration

For complex multi-step workflows, use the orchestration package:

```go
import (
    "github.com/cloudwego/eino/compose"
    "github.com/agentplexus/agentkit/orchestration"
)

type WorkflowInput struct {
    Query string
}

type WorkflowOutput struct {
    Result string
}

func main() {
    // Build workflow graph
    builder := orchestration.NewGraphBuilder[*WorkflowInput, *WorkflowOutput]("research-workflow")
    graph := builder.Graph()

    // Add processing node
    processLambda := compose.InvokableLambda(func(ctx context.Context, input *WorkflowInput) (*WorkflowOutput, error) {
        return &WorkflowOutput{Result: "Processed: " + input.Query}, nil
    })
    graph.AddLambdaNode("process", processLambda)

    // Connect nodes
    builder.AddStartEdge("process")
    builder.AddEndEdge("process")

    // Create executor
    finalGraph := builder.Build()
    executor := orchestration.NewExecutor(finalGraph, "research-workflow")

    // Expose as HTTP handler
    handler := orchestration.NewHTTPHandler(executor)

    server, _ := httpserver.NewBuilder("research-agent", 8001).
        WithHandler("/research", handler).
        Build()

    server.Start()
}
```

## Next Steps

- [Local Development](local-development.md) - Test locally before deploying
- [A2A Server](../server-factories/a2a.md) - Full A2A server documentation
- [HTTP Server](../server-factories/httpserver.md) - HTTP server factory options
- [Platforms](../platforms/index.md) - Deploy to Kubernetes or AWS AgentCore
