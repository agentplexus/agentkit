# AgentKit v0.1.0 Release Notes

**Release Date:** December 29, 2025

AgentKit is a Go library for building AI agent applications. This initial release provides server factories, LLM abstractions, workflow orchestration, and multi-runtime deployment support.

## Highlights

- **~1,500 lines of boilerplate eliminated** per project (29% reduction)
- **Server factories** reduce setup from ~100 lines to ~10 lines
- **Multi-runtime deployment** - Kubernetes or AWS AgentCore
- **Write once, deploy anywhere** - same code runs locally and in production

## Features

### Server Factories

#### A2A Server (`a2a.NewServer()`)
- Complete A2A protocol server in 5 lines
- Automatic agent card generation
- Health check endpoint
- Async/sync start modes
- Graceful shutdown

```go
server, _ := a2a.NewServer(a2a.Config{
    Agent:       myAgent,
    Port:        "9001",
    Description: "My agent",
})
server.Start(ctx)
```

#### HTTP Server (`httpserver.NewBuilder()`)
- Fluent builder API
- Multiple handler support
- Configurable timeouts
- Built-in health checks

```go
server, _ := httpserver.NewBuilder("my-agent", 8001).
    WithHandlerFunc("/process", handler).
    WithDualModeLog().
    Build()
server.Start()
```

### Core Packages

| Package | Description |
|---------|-------------|
| `config` | Environment-based configuration with VaultGuard integration |
| `llm` | Multi-provider LLM factory (Gemini, Claude, OpenAI, xAI, Ollama) |
| `agent` | Base agent framework with LLM integration |
| `orchestration` | Eino-based workflow graphs with type-safe execution |
| `http` | HTTP client utilities for inter-agent communication |

### Platform Support

#### Kubernetes (`platforms/kubernetes`)
- Works with any K8s distribution: EKS, GKE, AKS, Minikube, kind, k3s
- Helm values validation with Go structs
- Reusable deployment templates
- Port conflict detection
- Resource quantity validation

#### AWS AgentCore (`platforms/agentcore`)
- Firecracker microVM isolation per session
- Serverless scaling from zero
- Pay-per-use pricing
- Built-in session management
- Eino executor wrapping

```go
server := agentcore.NewBuilder().
    WithAgent(myAgent).
    WithPort(8080).
    MustBuild(ctx)
server.Start()
```

### Local Development

AgentCore code runs locally without AWS:

```bash
go run main.go
curl localhost:8080/ping
curl -X POST localhost:8080/invocations -d '{"prompt":"test"}'
```

No code changes needed between local development and production.

## Package Structure

```
agentkit/
├── a2a/             # A2A protocol server factory
├── agent/           # Base agent framework
├── config/          # Configuration management
├── http/            # HTTP client utilities
├── httpserver/      # HTTP server factory
├── llm/             # Multi-provider LLM abstraction
├── orchestration/   # Eino workflow orchestration
└── platforms/
    ├── agentcore/   # AWS Bedrock AgentCore runtime
    └── kubernetes/  # Kubernetes + Helm deployment
```

## Dependencies

- [OmniLLM](https://github.com/agentplexus/omnillm) - Multi-provider LLM abstraction
- [VaultGuard](https://github.com/agentplexus/vaultguard) - Security-gated credentials
- [Eino](https://github.com/cloudwego/eino) - Graph-based orchestration
- [Google ADK](https://google.golang.org/adk) - Agent Development Kit
- [a2a-go](https://github.com/a2aserver/a2a-go) - A2A protocol implementation

## Documentation

- [README.md](README.md) - Quick start guide
- [BENEFITS.md](BENEFITS.md) - Detailed benefit analysis
- [PRESENTATION.md](PRESENTATION.md) - Marp slide deck
- [docsrc/](docsrc/) - MkDocs documentation site

## Installation

```bash
go get github.com/agentplexus/agentkit
```

## Getting Started

```go
import (
    "github.com/agentplexus/agentkit/agent"
    "github.com/agentplexus/agentkit/config"
    "github.com/agentplexus/agentkit/httpserver"
)

func main() {
    cfg := config.LoadConfig()
    ba, _ := agent.NewBaseAgent(cfg, "my-agent", 30)

    server, _ := httpserver.NewBuilder("my-agent", 8001).
        WithHandlerFunc("/process", myHandler).
        Build()

    server.Start()
}
```

## Known Limitations

- AgentCore IaC templates (CDK/Terraform) not yet included
- Helm chart templates are minimal; expand based on project needs
- Observability hooks require OmniObserve setup

## Future Roadmap

- [ ] AWS CDK constructs for AgentCore deployment
- [ ] Terraform modules for AgentCore
- [ ] Extended Helm chart library
- [ ] Additional LLM provider adapters
- [ ] Streaming response support

## License

MIT License
