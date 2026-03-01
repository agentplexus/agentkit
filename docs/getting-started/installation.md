# Installation

## Requirements

- Go 1.21 or later
- Access to at least one LLM provider (Gemini, Claude, OpenAI, xAI, or Ollama)

## Install AgentKit

```bash
go get github.com/plexusone/agentkit
```

## Import Packages

Import the packages you need:

```go
import (
    "github.com/plexusone/agentkit/agent"
    "github.com/plexusone/agentkit/config"
    "github.com/plexusone/agentkit/llm"
    "github.com/plexusone/agentkit/orchestration"
    "github.com/plexusone/agentkit/a2a"
    "github.com/plexusone/agentkit/httpserver"
)
```

For platform-specific packages:

```go
// Kubernetes + Helm
import "github.com/plexusone/agentkit/platforms/kubernetes"

// AWS AgentCore
import "github.com/plexusone/agentkit/platforms/agentcore"
```

## Environment Variables

Configure your LLM provider:

| Variable | Description | Default |
|----------|-------------|---------|
| `LLM_PROVIDER` | Provider (gemini, claude, openai, xai, ollama) | gemini |
| `LLM_MODEL` | Model name | Provider default |
| `GEMINI_API_KEY` | Gemini API key | - |
| `CLAUDE_API_KEY` | Claude/Anthropic API key | - |
| `OPENAI_API_KEY` | OpenAI API key | - |
| `XAI_API_KEY` | xAI API key | - |
| `OLLAMA_URL` | Ollama server URL | http://localhost:11434 |

### Optional Observability

| Variable | Description | Default |
|----------|-------------|---------|
| `OBSERVABILITY_ENABLED` | Enable LLM observability | false |
| `OBSERVABILITY_PROVIDER` | Provider (opik, langfuse, phoenix) | opik |

## Verify Installation

Create a simple test file:

```go
package main

import (
    "context"
    "log"

    "github.com/plexusone/agentkit/agent"
    "github.com/plexusone/agentkit/config"
)

func main() {
    cfg := config.LoadConfig()

    ba, err := agent.NewBaseAgent(cfg, "test-agent", 30)
    if err != nil {
        log.Fatal(err)
    }
    defer ba.Close()

    log.Printf("AgentKit installed successfully!")
    log.Printf("Provider: %s", ba.GetProviderInfo())
}
```

Run it:

```bash
export LLM_PROVIDER=gemini
export GEMINI_API_KEY=your-api-key
go run main.go
```

## Next Steps

- [Quick Start](quick-start.md) - Build your first agent
- [Local Development](local-development.md) - Set up your development environment
