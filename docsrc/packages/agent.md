# agent

Base agent implementation with LLM integration.

## BaseAgent

```go
import (
    "github.com/agentplexus/agentkit/agent"
    "github.com/agentplexus/agentkit/config"
)

cfg := config.LoadConfig()

ba, err := agent.NewBaseAgent(cfg, "my-agent", 30) // 30 second timeout
if err != nil {
    log.Fatal(err)
}
defer ba.Close()
```

## Secure Agent

Create an agent with VaultGuard security checks:

```go
ba, secCfg, err := agent.NewBaseAgentSecure(ctx, "my-agent", 30,
    config.WithPolicy(nil), // Default policy
)
if err != nil {
    log.Fatalf("Security check failed: %v", err)
}
defer ba.Close()
defer secCfg.Close()

// Security info
log.Printf("Environment: %s", secCfg.Environment())
log.Printf("Security score: %d", secCfg.SecurityResult().Score)
```

## Methods

### LLM Operations

```go
// Get provider info
info := ba.GetProviderInfo()

// Generate response
response, err := ba.Generate(ctx, prompt)
```

### HTTP Operations

```go
// Fetch URL content
content, err := ba.FetchURL(ctx, "https://example.com", 10) // 10MB max
```

### Logging

```go
ba.LogInfo("Processing request: %s", requestID)
ba.LogError("Failed to process: %v", err)
ba.LogDebug("Debug info: %v", data)
```

## Building Custom Agents

Embed BaseAgent in your custom agent:

```go
type ResearchAgent struct {
    base *agent.BaseAgent
    cfg  *config.Config
}

func NewResearchAgent(base *agent.BaseAgent, cfg *config.Config) *ResearchAgent {
    return &ResearchAgent{
        base: base,
        cfg:  cfg,
    }
}

func (a *ResearchAgent) Research(ctx context.Context, query string) (*Result, error) {
    // Use base agent's LLM
    response, err := a.base.Generate(ctx, query)
    if err != nil {
        a.base.LogError("Research failed: %v", err)
        return nil, err
    }

    a.base.LogInfo("Research completed for: %s", query)
    return &Result{Content: response}, nil
}
```

## HTTP Handler Pattern

```go
func (a *ResearchAgent) HandleResearch(w http.ResponseWriter, r *http.Request) {
    var req ResearchRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    result, err := a.Research(r.Context(), req.Query)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}
```

## Google ADK Integration

Wrap your agent for A2A protocol:

```go
func (a *ResearchAgent) ADKAgent() agent.Agent {
    // Return Google ADK agent implementation
    return a.adkAgent
}
```

Use with A2A server:

```go
a2aServer, _ := a2a.NewServer(a2a.Config{
    Agent: researchAgent.ADKAgent(),
    Port:  "9001",
})
```
