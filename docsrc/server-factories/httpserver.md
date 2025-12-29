# HTTP Server

The HTTP server factory provides a production-ready HTTP server with minimal configuration.

## Builder Pattern (Recommended)

```go
import "github.com/agentplexus/agentkit/httpserver"

server, err := httpserver.NewBuilder("my-agent", 8001).
    WithHandlerFunc("/process", agent.HandleProcess).
    WithHandlerFunc("/analyze", agent.HandleAnalyze).
    WithDualModeLog().
    Build()

if err != nil {
    log.Fatal(err)
}

server.Start()
```

## Config-Based

```go
server, err := httpserver.New(httpserver.Config{
    Name: "my-agent",
    Port: 8001,
    HandlerFuncs: map[string]http.HandlerFunc{
        "/process": agent.HandleProcess,
        "/analyze": agent.HandleAnalyze,
    },
    EnableDualModeLog: true,
})
```

## Builder Methods

```go
httpserver.NewBuilder("name", port).
    // Handlers
    WithHandlerFunc("/path", handlerFunc).  // http.HandlerFunc
    WithHandler("/path", handler).           // http.Handler

    // Timeouts
    WithTimeouts(read, write, idle).

    // Health check
    WithHealthPath("/health").              // Default: /health
    WithHealthHandler(customHealthFunc).

    // Logging
    WithDualModeLog().                      // Log startup info

    Build()
```

## Configuration Options

```go
httpserver.Config{
    Name:              "my-agent",
    Port:              8001,

    // Handlers
    Handlers:          map[string]http.Handler{...},
    HandlerFuncs:      map[string]http.HandlerFunc{...},

    // Timeouts
    ReadTimeout:       30 * time.Second,   // Default: 30s
    WriteTimeout:      120 * time.Second,  // Default: 120s
    IdleTimeout:       60 * time.Second,   // Default: 60s

    // Health
    HealthPath:        "/health",          // Default: /health
    HealthHandler:     customHealthFunc,   // Default: returns "OK"

    // Logging
    EnableDualModeLog: true,
}
```

## Server Methods

```go
// Lifecycle
server.Start()              // Blocking start
server.StartAsync()         // Non-blocking start
server.Stop(ctx)            // Graceful shutdown

// Info
server.URL()                // "http://localhost:8001"
server.Addr()               // ":8001"
```

## With Orchestration

Use with the orchestration package for workflow-based handlers:

```go
import (
    "github.com/agentplexus/agentkit/httpserver"
    "github.com/agentplexus/agentkit/orchestration"
)

// Create executor from workflow
executor := orchestration.NewExecutor(graph, "my-workflow")

// Wrap as HTTP handler
handler := orchestration.NewHTTPHandler(executor)

// Add to server
server, _ := httpserver.NewBuilder("my-agent", 8001).
    WithHandler("/workflow", handler).
    Build()
```

## Before: Manual Setup (~25 lines)

```go
func startHTTPServer(agent *ResearchAgent) error {
    server := &http.Server{
        Addr:         ":8001",
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 120 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    http.HandleFunc("/research", agent.HandleResearchRequest)
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        if _, err := w.Write([]byte("OK")); err != nil {
            log.Printf("Failed to write health response: %v", err)
        }
    })

    log.Println("Research Agent HTTP server starting on :8001")
    log.Println("(Dual mode: HTTP for security/observability, A2A for interoperability)")

    if err := server.ListenAndServe(); err != nil {
        log.Fatalf("HTTP server failed: %v", err)
    }
    return nil
}
```

## After: With AgentKit (~5 lines)

```go
server, _ := httpserver.NewBuilder("research-agent", 8001).
    WithHandlerFunc("/research", agent.HandleResearchRequest).
    WithDualModeLog().
    Build()

server.Start()
```

## Testing

Use port 0 for random port assignment in tests:

```go
func TestHTTPServer(t *testing.T) {
    server, _ := httpserver.NewBuilder("test-agent", 0).
        WithHandlerFunc("/test", testHandler).
        Build()

    server.StartAsync()
    defer server.Stop(context.Background())

    resp, err := http.Get(server.URL() + "/test")
    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
}
```

## Multiple Handlers

```go
server, _ := httpserver.NewBuilder("multi-agent", 8001).
    WithHandlerFunc("/research", agent.HandleResearch).
    WithHandlerFunc("/synthesize", agent.HandleSynthesize).
    WithHandlerFunc("/verify", agent.HandleVerify).
    WithHandler("/workflow", orchestration.NewHTTPHandler(executor)).
    Build()
```
