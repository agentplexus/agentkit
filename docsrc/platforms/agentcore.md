# AWS AgentCore Deployment

Deploy AgentKit agents to AWS Bedrock AgentCore - a serverless agent runtime powered by Firecracker microVMs.

## Overview

AWS AgentCore provides:

- **Firecracker microVM isolation** - Each session runs in its own microVM
- **Serverless scaling** - Automatic scaling from zero
- **Pay-per-use pricing** - Only pay for active CPU time
- **Built-in session management** - 8-hour sessions with automatic cleanup

## Basic Setup

```go
import "github.com/agentplexus/agentkit/platforms/agentcore"

server := agentcore.NewBuilder().
    WithPort(8080).
    WithAgent(myAgent).
    MustBuild(ctx)

server.Start()
```

## Builder Pattern

```go
server := agentcore.NewBuilder().
    WithPort(8080).                      // Default: 8080
    WithAgent(researchAgent).            // Add agent
    WithAgent(synthesisAgent).           // Add another
    WithDefaultAgent("research").        // Default for routing
    WithRequestLogging(true).            // Enable logging
    WithSessionTracking(true).           // Enable sessions
    MustBuild(ctx)
```

## Configuration Options

```go
server := agentcore.NewServer(agentcore.Config{
    Port:                  8080,
    ReadTimeout:           30 * time.Second,
    WriteTimeout:          300 * time.Second,  // 5 min for long operations
    DefaultAgent:          "research",
    EnableRequestLogging:  true,
    EnableSessionTracking: true,
})
```

## Endpoints

AgentCore servers provide:

| Endpoint | Description |
|----------|-------------|
| `/ping` | Health check |
| `/invocations` | Agent invocation |

## Wrapping Eino Executors

Wrap your Eino workflow executors for AgentCore:

```go
import (
    "github.com/agentplexus/agentkit/orchestration"
    "github.com/agentplexus/agentkit/platforms/agentcore"
)

// Build Eino workflow
graph := buildOrchestrationGraph()
executor := orchestration.NewExecutor(graph, "stats-workflow")

// Simple wrap - uses JSON marshaling
agent := agentcore.WrapExecutor("stats", executor)

// Custom I/O transformation
agent := agentcore.WrapExecutorWithPrompt("stats", executor,
    func(prompt string) StatsRequest {
        return StatsRequest{Topic: prompt}
    },
    func(output StatsResponse) string {
        return output.Summary
    },
)
```

## Multi-Agent Routing

Register multiple agents with automatic routing:

```go
server := agentcore.NewBuilder().
    WithAgent(researchAgent).
    WithAgent(synthesisAgent).
    WithAgent(verificationAgent).
    WithDefaultAgent("research").
    MustBuild(ctx)
```

Invoke specific agents:

```bash
# Default agent
curl -X POST localhost:8080/invocations \
  -d '{"prompt": "Find AI statistics"}'

# Specific agent
curl -X POST localhost:8080/invocations \
  -d '{"prompt": "Verify this claim", "agent": "verification"}'
```

## Request/Response Format

### Request

```json
{
  "prompt": "Find statistics about AI adoption",
  "session_id": "optional-session-id",
  "agent": "optional-agent-name",
  "metadata": {
    "key": "value"
  }
}
```

### Response

```json
{
  "output": "Response from the agent...",
  "session_id": "session-123",
  "agent": "research",
  "metadata": {}
}
```

## Session Management

AgentCore provides built-in session isolation:

```go
// Access session in your agent
func (a *MyAgent) Invoke(ctx context.Context, req agentcore.Request) (agentcore.Response, error) {
    session := agentcore.SessionFromContext(ctx)

    // Session info
    sessionID := session.ID
    startTime := session.StartTime

    // Your logic...
}
```

## Local Development

The same code runs locally - no AWS required for development:

```bash
# Run locally
go run main.go

# Test
curl localhost:8080/ping
curl -X POST localhost:8080/invocations -d '{"prompt":"test"}'
```

| Aspect | Local | AWS AgentCore |
|--------|-------|---------------|
| Process | Go binary | Firecracker microVM |
| Sessions | In-memory | Isolated per microVM |
| Scaling | Manual | Automatic |
| Startup | Instant | ~100ms cold start |

## AWS Deployment

!!! note "Infrastructure as Code"
    Helm does **NOT** apply to AgentCore. Use AWS CDK or Terraform instead.

### Dockerfile

```dockerfile
FROM golang:1.21 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /agent ./cmd/agent

FROM gcr.io/distroless/static
COPY --from=builder /agent /agent
EXPOSE 8080
CMD ["/agent"]
```

### CDK Example

```typescript
import * as agentcore from '@aws-cdk/aws-bedrock-agentcore';

const agent = new agentcore.Agent(this, 'StatsAgent', {
  runtime: agentcore.Runtime.GO_1_21,
  code: agentcore.Code.fromAsset('./'),
  handler: 'main',
  memory: 512,
  timeout: Duration.minutes(5),
});
```

## Kubernetes vs AgentCore

| Aspect | Kubernetes | AgentCore |
|--------|------------|-----------|
| Infrastructure | K8s manifests | AWS-managed |
| Config tool | Helm | CDK / Terraform |
| Scaling | HPA | Automatic |
| Isolation | Containers | Firecracker microVMs |
| Pricing | Always-on | Pay-per-use |
| Session handling | Application | Built-in |

## Best Practices

1. **Use WrapExecutorWithPrompt** for type-safe I/O transformation
2. **Set appropriate timeouts** - AgentCore sessions can last up to 8 hours
3. **Use session tracking** for multi-turn conversations
4. **Test locally first** - Same code, same endpoints

## Next Steps

- [Kubernetes Deployment](kubernetes.md) - Alternative container-based deployment
- [Local Development](../getting-started/local-development.md) - Test before deploying
