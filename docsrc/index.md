# AgentKit

A Go library for building AI agent applications. Provides server factories, LLM abstractions, workflow orchestration, and multi-runtime deployment support.

## Features

- **Server Factories** - A2A and HTTP servers in 5 lines (saves ~475 lines per project)
- **Multi-Provider LLM** - Gemini, Claude, OpenAI, xAI, Ollama via OmniLLM
- **Workflow Orchestration** - Type-safe graph-based execution with Eino
- **Multi-Runtime Deployment** - Kubernetes (Helm) or AWS AgentCore
- **VaultGuard Integration** - Security-gated credential access

## Architecture

```
agentkit/
├── # Core (platform-agnostic)
├── a2a/             # A2A protocol server factory
├── agent/           # Base agent framework
├── config/          # Configuration management
├── http/            # HTTP client utilities
├── httpserver/      # HTTP server factory
├── llm/             # Multi-provider LLM abstraction
├── orchestration/   # Eino workflow orchestration
│
├── # Platform-specific
└── platforms/
    ├── agentcore/   # AWS Bedrock AgentCore runtime
    └── kubernetes/  # Kubernetes + Helm deployment
```

## Quick Example

```go
package main

import (
    "context"

    "github.com/agentplexus/agentkit/a2a"
    "github.com/agentplexus/agentkit/agent"
    "github.com/agentplexus/agentkit/config"
    "github.com/agentplexus/agentkit/httpserver"
)

func main() {
    ctx := context.Background()
    cfg := config.LoadConfig()

    // Create agent
    ba, _ := agent.NewBaseAgent(cfg, "research-agent", 30)
    researchAgent := NewResearchAgent(ba, cfg)

    // HTTP server - 5 lines
    httpServer, _ := httpserver.NewBuilder("research-agent", 8001).
        WithHandlerFunc("/research", researchAgent.HandleResearch).
        Build()

    // A2A server - 5 lines
    a2aServer, _ := a2a.NewServer(a2a.Config{
        Agent:       researchAgent.ADKAgent(),
        Port:        "9001",
        Description: "Research agent for web search",
    })

    // Start servers
    a2aServer.StartAsync(ctx)
    httpServer.Start()
}
```

## Benefits

AgentKit eliminates ~1,500 lines of boilerplate per project:

| Component | Lines Saved |
|-----------|-------------|
| A2A server factory | ~350 lines |
| HTTP server factory | ~125 lines |
| Shared pkg/ code | ~930 lines |

See the [Benefits](benefits.md) page for detailed analysis.

## Getting Started

1. [Installation](getting-started/installation.md) - Add AgentKit to your project
2. [Quick Start](getting-started/quick-start.md) - Build your first agent
3. [Local Development](getting-started/local-development.md) - Test locally before deploying
