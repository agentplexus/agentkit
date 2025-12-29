# Migration Guide

Migrate existing agent projects to AgentKit.

## Overview

Migration typically involves:

1. Adding the AgentKit dependency
2. Replacing local `pkg/` imports with AgentKit imports
3. Using server factories instead of manual setup
4. Removing redundant local packages

## Step 1: Add Dependency

```bash
go get github.com/agentplexus/agentkit
```

## Step 2: Replace Imports

### Configuration

```go
// Before
import "github.com/myproject/pkg/config"

// After
import "github.com/agentplexus/agentkit/config"
```

### LLM

```go
// Before
import "github.com/myproject/pkg/llm"

// After
import "github.com/agentplexus/agentkit/llm"
```

### Agent

```go
// Before
import "github.com/myproject/pkg/agent"

// After
import "github.com/agentplexus/agentkit/agent"
```

### HTTP Client

```go
// Before
import "github.com/myproject/pkg/httpclient"

// After
import "github.com/agentplexus/agentkit/http"
```

## Step 3: Use Server Factories

### A2A Server

```go
// Before (~70 lines)
func startA2AServer(agent agent.Agent, port string) error {
    listener, _ := net.Listen("tcp", "0.0.0.0:"+port)
    agentCard := &a2a.AgentCard{...}
    mux := http.NewServeMux()
    // ... 50+ more lines
}

// After (~5 lines)
import "github.com/agentplexus/agentkit/a2a"

server, _ := a2a.NewServer(a2a.Config{
    Agent: myAgent,
    Port:  "9001",
})
server.Start(ctx)
```

### HTTP Server

```go
// Before (~25 lines)
server := &http.Server{
    Addr:         ":8001",
    ReadTimeout:  30 * time.Second,
    // ...
}
http.HandleFunc("/research", handler)
http.HandleFunc("/health", healthHandler)
server.ListenAndServe()

// After (~5 lines)
import "github.com/agentplexus/agentkit/httpserver"

server, _ := httpserver.NewBuilder("my-agent", 8001).
    WithHandlerFunc("/research", handler).
    Build()
server.Start()
```

## Step 4: Remove Redundant Code

Delete your local packages that are now provided by AgentKit:

```bash
rm -rf pkg/config/
rm -rf pkg/llm/
rm -rf pkg/agent/
rm -rf pkg/httpclient/
```

## Complete Example

### Before

```go
package main

import (
    "github.com/myproject/pkg/config"
    "github.com/myproject/pkg/agent"
)

func main() {
    cfg := config.LoadConfig()

    base, _ := agent.NewBaseAgent(cfg, 30)
    defer base.Close()

    researchAgent := NewResearchAgent(base, cfg)

    // Manual HTTP server setup (~25 lines)
    server := &http.Server{...}
    http.HandleFunc("/research", researchAgent.HandleResearch)
    http.HandleFunc("/health", ...)

    // Manual A2A server setup (~70 lines)
    go func() {
        listener, _ := net.Listen(...)
        agentCard := &a2a.AgentCard{...}
        // ... many more lines
    }()

    server.ListenAndServe()
}
```

### After

```go
package main

import (
    "github.com/agentplexus/agentkit/agent"
    "github.com/agentplexus/agentkit/config"
    "github.com/agentplexus/agentkit/a2a"
    "github.com/agentplexus/agentkit/httpserver"
)

func main() {
    ctx := context.Background()
    cfg := config.LoadConfig()

    base, _ := agent.NewBaseAgent(cfg, "research-agent", 30)
    defer base.Close()

    researchAgent := NewResearchAgent(base, cfg)

    // HTTP server - 5 lines
    httpServer, _ := httpserver.NewBuilder("research-agent", 8001).
        WithHandlerFunc("/research", researchAgent.HandleResearch).
        Build()

    // A2A server - 5 lines
    a2aServer, _ := a2a.NewServer(a2a.Config{
        Agent: researchAgent.ADKAgent(),
        Port:  "9001",
    })

    a2aServer.StartAsync(ctx)
    httpServer.Start()
}
```

## Migration Checklist

- [ ] Add AgentKit dependency
- [ ] Replace `pkg/config` with `agentkit/config`
- [ ] Replace `pkg/llm` with `agentkit/llm`
- [ ] Replace `pkg/agent` with `agentkit/agent`
- [ ] Replace `pkg/httpclient` with `agentkit/http`
- [ ] Replace manual A2A server with `a2a.NewServer()`
- [ ] Replace manual HTTP server with `httpserver.NewBuilder()`
- [ ] Remove redundant `pkg/` directories
- [ ] Update tests
- [ ] Verify all endpoints work

## Estimated Savings

| Component | Lines Removed |
|-----------|---------------|
| pkg/config | ~140 lines |
| pkg/llm | ~200 lines |
| pkg/agent | ~140 lines |
| pkg/httpclient | ~90 lines |
| A2A server boilerplate | ~350 lines |
| HTTP server boilerplate | ~125 lines |
| **Total** | **~1,045 lines** |

## Gradual Migration

You can migrate incrementally:

1. **Week 1**: Replace `pkg/config` and `pkg/llm`
2. **Week 2**: Replace server boilerplate with factories
3. **Week 3**: Replace remaining packages, remove old code

Each step is independently testable.
