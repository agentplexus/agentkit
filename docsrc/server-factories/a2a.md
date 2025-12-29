# A2A Server

The A2A (Agent-to-Agent) server factory provides a complete A2A protocol server with minimal configuration.

## Basic Usage

```go
import "github.com/agentplexus/agentkit/a2a"

server, err := a2a.NewServer(a2a.Config{
    Agent:       myAgent,
    Port:        "9001",
    Description: "Research agent for web search",
})
if err != nil {
    log.Fatal(err)
}

server.Start(ctx)  // Blocking
```

## Configuration Options

```go
server, _ := a2a.NewServer(a2a.Config{
    // Required
    Agent: myAgent,  // Google ADK agent

    // Optional
    Port:              "9001",            // Empty = random port
    Description:       "My agent",        // Override agent description
    InvokePath:        "/invoke",         // Default: /invoke
    ReadHeaderTimeout: 10 * time.Second,  // Default: 10s
    SessionService:    customService,     // Default: in-memory
})
```

## Server Methods

```go
// URLs
server.URL()          // "http://localhost:9001"
server.AgentCardURL() // "http://localhost:9001/.well-known/agent.json"
server.InvokeURL()    // "http://localhost:9001/invoke"
server.Addr()         // net.Addr

// Lifecycle
server.Start(ctx)      // Blocking start
server.StartAsync(ctx) // Non-blocking start
server.Stop(ctx)       // Graceful shutdown
```

## Endpoints

The server provides these endpoints:

| Endpoint | Description |
|----------|-------------|
| `/.well-known/agent.json` | Agent card (capabilities, skills) |
| `/invoke` | JSON-RPC invocation endpoint |
| `/health` | Health check |

## Before: Manual Setup (~70 lines)

```go
func startA2AServer(agent agent.Agent, port string) error {
    listener, err := net.Listen("tcp", "0.0.0.0:"+port)
    if err != nil {
        return fmt.Errorf("failed to create listener: %w", err)
    }

    baseURL, _ := url.Parse(fmt.Sprintf("http://localhost:%s", port))
    agentPath := "/invoke"

    agentCard := &a2a.AgentCard{
        Name:               agent.Name(),
        Description:        "Agent description here",
        Skills:             adka2a.BuildAgentSkills(agent),
        PreferredTransport: a2a.TransportProtocolJSONRPC,
        URL:                baseURL.JoinPath(agentPath).String(),
        Capabilities:       a2a.AgentCapabilities{Streaming: true},
    }

    mux := http.NewServeMux()
    mux.Handle(a2asrv.WellKnownAgentCardPath,
        a2asrv.NewStaticAgentCardHandler(agentCard))

    executor := adka2a.NewExecutor(adka2a.ExecutorConfig{
        RunnerConfig: runner.Config{
            AppName:        agent.Name(),
            Agent:          agent,
            SessionService: session.InMemoryService(),
        },
    })

    requestHandler := a2asrv.NewHandler(executor)
    mux.Handle(agentPath, a2asrv.NewJSONRPCHandler(requestHandler))

    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })

    server := &http.Server{
        Handler:           mux,
        ReadHeaderTimeout: 10 * time.Second,
    }

    log.Printf("A2A server starting on port %s", port)
    return server.Serve(listener)
}
```

## After: With AgentKit (~5 lines)

```go
server, _ := a2a.NewServer(a2a.Config{
    Agent:       myAgent,
    Port:        "9001",
    Description: "Research agent for web search",
})
server.Start(ctx)
```

## Testing

Use a random port for tests:

```go
func TestA2AServer(t *testing.T) {
    server, _ := a2a.NewServer(a2a.Config{
        Agent: myAgent,
        Port:  "", // Random port
    })

    server.StartAsync(context.Background())
    defer server.Stop(context.Background())

    // Fetch agent card
    resp, err := http.Get(server.AgentCardURL())
    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)

    var card map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&card)
    assert.Equal(t, "my-agent", card["name"])
}
```

## Custom Session Service

For production, you may want persistent sessions:

```go
// Custom session service
sessionService := myCustomSessionService()

server, _ := a2a.NewServer(a2a.Config{
    Agent:          myAgent,
    Port:           "9001",
    SessionService: sessionService,
})
```
