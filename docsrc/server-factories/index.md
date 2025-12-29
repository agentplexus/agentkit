# Server Factories

AgentKit provides server factories that eliminate boilerplate code for setting up agent servers.

## The Problem

Every agent project repeats the same patterns:

| Pattern | Lines Duplicated |
|---------|------------------|
| A2A server setup | ~350 lines |
| HTTP server setup | ~125 lines |
| HTTP handler setup | ~100 lines |
| **Total** | **~575 lines** |

## The Solution

Server factories reduce this to ~10 lines:

| Factory | Before | After | Savings |
|---------|--------|-------|---------|
| `a2a.NewServer()` | ~70 lines | ~5 lines | 93% |
| `httpserver.New()` | ~25 lines | ~5 lines | 80% |

## Available Factories

### [A2A Server](a2a.md)

For agent-to-agent communication using the A2A protocol:

```go
server, _ := a2a.NewServer(a2a.Config{
    Agent:       myAgent,
    Port:        "9001",
    Description: "My agent",
})
server.Start(ctx)
```

### [HTTP Server](httpserver.md)

For REST API endpoints:

```go
server, _ := httpserver.NewBuilder("my-agent", 8001).
    WithHandlerFunc("/process", agent.HandleProcess).
    Build()
server.Start()
```

## Combining Servers

Run both HTTP and A2A servers in the same application:

```go
func main() {
    ctx := context.Background()

    // HTTP server for REST API
    httpServer, _ := httpserver.NewBuilder("my-agent", 8001).
        WithHandlerFunc("/process", agent.HandleProcess).
        Build()

    // A2A server for agent communication
    a2aServer, _ := a2a.NewServer(a2a.Config{
        Agent: myAgent,
        Port:  "9001",
    })

    // Start both
    a2aServer.StartAsync(ctx)  // Non-blocking
    httpServer.Start()          // Blocking
}
```

## Built-in Features

All server factories include:

- Health check endpoints (`/health` or `/ping`)
- Graceful shutdown
- Configurable timeouts
- Request logging options
- Async/sync start modes

## Next Steps

- [A2A Server Documentation](a2a.md)
- [HTTP Server Documentation](httpserver.md)
