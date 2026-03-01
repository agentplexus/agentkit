# Local Development

AgentKit is designed for seamless local development. The same code runs locally and in production - only the infrastructure differs.

## Running Locally

### HTTP Server

```bash
go run main.go
```

Test your endpoints:

```bash
# Health check
curl http://localhost:8001/health

# Your endpoint
curl -X POST http://localhost:8001/research \
  -H "Content-Type: application/json" \
  -d '{"query": "test"}'
```

### A2A Server

```bash
go run main.go
```

Test the A2A endpoints:

```bash
# Agent card
curl http://localhost:9001/.well-known/agent.json

# Invoke
curl -X POST http://localhost:9001/invoke \
  -H "Content-Type: application/json" \
  -d '{"method": "agent/run", "params": {"prompt": "test"}}'
```

### AgentCore Runtime

The AgentCore server runs locally with the same `/ping` and `/invocations` endpoints as production:

```bash
go run main.go
```

Test the endpoints:

```bash
# Health check
curl http://localhost:8080/ping

# Invoke agent
curl -X POST http://localhost:8080/invocations \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Find statistics about AI adoption"}'
```

## Local vs Production

The code is identical - only the infrastructure changes:

| Aspect | Local | Kubernetes | AWS AgentCore |
|--------|-------|------------|---------------|
| Process | Go binary | Container | Firecracker microVM |
| Sessions | In-memory | Pod-based | Isolated per microVM |
| Scaling | Manual | HPA | Automatic |
| Startup | Instant | Container pull | Cold start ~100ms |

## Environment Setup

Create a `.env` file for local development:

```bash
# LLM Provider
LLM_PROVIDER=gemini
GEMINI_API_KEY=your-api-key

# Optional: Observability
OBSERVABILITY_ENABLED=false

# Optional: Search provider
SEARCH_PROVIDER=serper
SERPER_API_KEY=your-api-key
```

Load it:

```bash
source .env && go run main.go
```

Or use a tool like [direnv](https://direnv.net/).

## Hot Reload

For faster development iteration, use a tool like [air](https://github.com/cosmtrek/air):

```bash
# Install
go install github.com/cosmtrek/air@latest

# Run with hot reload
air
```

Create `.air.toml`:

```toml
[build]
cmd = "go build -o ./tmp/main ."
bin = "./tmp/main"
include_ext = ["go"]
exclude_dir = ["tmp", "vendor"]
```

## Testing

### Unit Tests

```bash
go test ./...
```

### Integration Tests

Test your agent endpoints:

```go
func TestResearchEndpoint(t *testing.T) {
    // Start server
    server, _ := httpserver.NewBuilder("test-agent", 0). // Port 0 = random
        WithHandlerFunc("/research", agent.HandleResearch).
        Build()

    server.StartAsync()
    defer server.Stop(context.Background())

    // Make request
    resp, err := http.Post(
        server.URL()+"/research",
        "application/json",
        strings.NewReader(`{"query": "test"}`),
    )

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
}
```

### A2A Server Testing

```go
func TestA2AServer(t *testing.T) {
    server, _ := a2a.NewServer(a2a.Config{
        Agent: myAgent,
        Port:  "", // Random port
    })

    server.StartAsync(context.Background())
    defer server.Stop(context.Background())

    // Fetch agent card
    resp, _ := http.Get(server.AgentCardURL())
    assert.Equal(t, 200, resp.StatusCode)
}
```

## Debugging

### Enable Request Logging

```go
server, _ := httpserver.NewBuilder("my-agent", 8001).
    WithHandlerFunc("/research", agent.HandleResearch).
    WithDualModeLog(). // Logs startup info
    Build()
```

### Verbose LLM Logging

Set the log level:

```bash
export LOG_LEVEL=debug
go run main.go
```

## Next Steps

- [Kubernetes Deployment](../platforms/kubernetes.md)
- [AWS AgentCore Deployment](../platforms/agentcore.md)
