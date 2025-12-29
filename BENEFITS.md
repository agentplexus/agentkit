# AgentKit Benefit Assessment

## Executive Summary

AgentKit is a Go library that provides reusable components for building AI agent applications. This document quantifies its benefits using `stats-agent-team` as a reference implementation.

**Key Finding:** AgentKit saves ~1,500 lines (29%) per project and provides significant value when building multiple agent systems.

---

## Reference Project: stats-agent-team

A multi-agent system for finding, extracting, and verifying statistics from web sources.

### Codebase Metrics

| Metric | Value |
|--------|-------|
| Total Lines | 5,226 |
| Agent Binaries | 6 |
| Shared Packages | 11 |
| LLM Providers | 5 (Gemini, Claude, OpenAI, xAI, Ollama) |
| Orchestration Strategies | 2 (ADK, Eino) |

### Code Breakdown

```
stats-agent-team: 5,226 lines
├── Domain Logic:     ~3,500 lines (67%)
│   ├── Research agent logic
│   ├── Synthesis agent logic
│   ├── Verification agent logic
│   ├── Orchestration logic
│   ├── Data models
│   └── CLI tool
│
├── Shared pkg/:      ~930 lines (18%)
│   ├── config/       137 lines
│   ├── llm/          308 lines
│   ├── agent/        143 lines
│   ├── httpclient/   89 lines
│   ├── search/       80 lines
│   └── orchestration/ 173 lines
│
└── Boilerplate:      ~790 lines (15%)
    ├── HTTP handlers  100 lines
    ├── A2A server     350 lines
    ├── HTTP server    125 lines
    ├── Error handling 50 lines
    └── Misc           165 lines
```

---

## Component Overlap Analysis

AgentKit and stats-agent-team share nearly identical abstractions:

| Component | stats-agent-team | agentkit | Overlap |
|-----------|------------------|----------|---------|
| Config management | 137 lines | 151 lines | ~90% |
| LLM factory | 208 lines | 170 lines | ~95% |
| OmniLLM adapter | 100 lines | 100 lines | 100% |
| Base agent | 143 lines | 143 lines | ~90% |
| HTTP client | 89 lines | 89 lines | ~95% |
| Orchestration | 120 lines | 200 lines | ~80% |

**Implication:** AgentKit is essentially a generalized extraction of stats-agent-team's `pkg/` directory, plus new server factories.

---

## Quantified Benefits

### What AgentKit Provides

#### Core Packages

| Component | Description | Lines Saved |
|-----------|-------------|-------------|
| `config.Config` | Environment-based configuration | 137 |
| `config.SecureConfig` | VaultGuard security integration | NEW |
| `llm.ModelFactory` | Multi-provider LLM factory | 208 |
| `llm/adapters` | OmniLLM adapter | 100 |
| `agent.BaseAgent` | Common agent functionality | 143 |
| `http.PostJSON/GetJSON` | HTTP utilities | 89 |
| `orchestration.GraphBuilder[I,O]` | Type-safe Eino graphs | 120 |
| `orchestration.HTTPHandler[I,O]` | Generic JSON handlers | 100 |
| `orchestration.AgentCaller` | Inter-agent calls | 30 |
| `platforms/kubernetes` | Kubernetes + Helm support | varies |

#### Server Factories

| Component | Description | Lines Saved |
|-----------|-------------|-------------|
| `a2a.NewServer()` | A2A protocol server factory | 350 |
| `httpserver.New()` | Agent HTTP server factory | 125 |
| `httpserver.NewBuilder()` | Fluent builder for HTTP servers | included |

### Total Savings Per Project

```
Shared pkg/ replaced:        930 lines
HTTP handler boilerplate:    100 lines
A2A server boilerplate:      350 lines
HTTP server boilerplate:     125 lines
─────────────────────────────────────
Total:                     1,505 lines (29%)
```

---

## Benefit by Scale

### Single Project

| Metric | Value |
|--------|-------|
| Lines eliminated | ~1,500 |
| Percentage reduction | 29% |
| New features gained | VaultGuard security, better Helm |

**Verdict:** Moderate benefit. Worth it for consistency and security features.

### Multiple Projects

| Projects | Lines Saved | Cumulative Benefit |
|----------|-------------|-------------------|
| 1 | 1,500 | Single codebase cleanup |
| 2 | 3,000 | Shared maintenance |
| 3 | 4,500 | Pattern consistency |
| 5 | 7,500 | Platform-level reuse |
| 10 | 15,000 | Significant engineering savings |

**Verdict:** Strong benefit. Each new project avoids 1,500 lines of boilerplate.

---

## Beyond Line Count

### Consistency Benefits

- **Same patterns everywhere** - Developers moving between projects see familiar code
- **Easier code reviews** - Reviewers know the standard patterns
- **Reduced cognitive load** - Less unique code to understand per project

### Security Benefits

- **VaultGuard integration** - Secure credential management out of the box
- **SecureConfig** - Security assessment with scoring
- **Policy options** - Development, strict, and custom security policies

### Observability Benefits

- **Built-in hooks** - OmniObserve integration standardized
- **Multiple providers** - Opik, Langfuse, Phoenix support
- **Consistent tracing** - Same observability patterns across agents

### Deployment Benefits

- **Helm validation** - Type-safe Kubernetes configuration
- **Reusable templates** - Standard deployment patterns
- **Resource validation** - Kubernetes quantity and port validation

### Maintenance Benefits

- **Single point of fixes** - Bug fixes benefit all projects
- **Centralized updates** - Security patches in one place
- **Version management** - Clear dependency on agentkit version

---

## Code Examples

### A2A Server Factory

#### Before: 70 lines per agent

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

#### After: 6 lines with `a2a.NewServer()`

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

// Blocking start
server.Start(ctx)

// Or async start
server.StartAsync(ctx)
defer server.Stop(ctx)
```

#### A2A Server Features

```go
// Full configuration options
server, _ := a2a.NewServer(a2a.Config{
    Agent:             myAgent,
    Port:              "9001",           // Empty = random port
    Description:       "My agent",       // Override agent description
    InvokePath:        "/invoke",        // Default: /invoke
    ReadHeaderTimeout: 10 * time.Second, // Default: 10s
    SessionService:    customService,    // Default: in-memory
})

// Useful methods
server.URL()          // "http://localhost:9001"
server.AgentCardURL() // "http://localhost:9001/.well-known/agent.json"
server.InvokeURL()    // "http://localhost:9001/invoke"
server.Addr()         // net.Addr
```

---

### HTTP Server Factory

#### Before: 25 lines per agent

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

#### After: Config-based approach

```go
import "github.com/agentplexus/agentkit/httpserver"

server, err := httpserver.New(httpserver.Config{
    Name:              "research-agent",
    Port:              8001,
    HandlerFuncs: map[string]http.HandlerFunc{
        "/research": agent.HandleResearchRequest,
    },
    EnableDualModeLog: true,
})
if err != nil {
    log.Fatal(err)
}
server.Start()
```

#### After: Fluent builder approach

```go
server, err := httpserver.NewBuilder("research-agent", 8001).
    WithHandlerFunc("/research", agent.HandleResearchRequest).
    WithHandlerFunc("/synthesize", agent.HandleSynthesizeRequest).
    WithHandler("/orchestrate", orchestration.NewHTTPHandler(executor)).
    WithTimeouts(30*time.Second, 120*time.Second, 60*time.Second).
    WithDualModeLog().
    Build()

if err != nil {
    log.Fatal(err)
}

// Blocking start
server.Start()

// Or async start
server.StartAsync()
defer server.Stop(ctx)
```

#### HTTP Server Features

```go
// Full configuration options
server, _ := httpserver.New(httpserver.Config{
    Name:              "my-agent",
    Port:              8001,
    Handlers:          map[string]http.Handler{...},     // http.Handler
    HandlerFuncs:      map[string]http.HandlerFunc{...}, // http.HandlerFunc
    ReadTimeout:       30 * time.Second,   // Default: 30s
    WriteTimeout:      120 * time.Second,  // Default: 120s
    IdleTimeout:       60 * time.Second,   // Default: 60s
    HealthPath:        "/health",          // Default: /health
    HealthHandler:     customHealthFunc,   // Default: returns "OK"
    EnableDualModeLog: true,               // Log dual mode message
})
```

---

### HTTP Handler (Orchestration Package)

#### Before: 20 lines per handler

```go
func (a *ResearchAgent) HandleResearchRequest(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req models.ResearchRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
        return
    }

    if req.NumResults == 0 {
        req.NumResults = 10
    }

    resp, err := a.Research(r.Context(), &req)
    if err != nil {
        http.Error(w, fmt.Sprintf("Research failed: %v", err), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}
```

#### After: 1 line with generic HTTPHandler

```go
import "github.com/agentplexus/agentkit/orchestration"

// Create executor from graph
executor := orchestration.NewExecutor(graph, "research-workflow")

// Wrap as HTTP handler - handles JSON encode/decode, errors, method check
handler := orchestration.NewHTTPHandler(executor)

// Use with httpserver
server, _ := httpserver.NewBuilder("research-agent", 8001).
    WithHandler("/research", handler).
    Build()
```

---

## Complete Agent Example

### Before: ~150 lines

```go
func main() {
    cfg := config.LoadConfig()

    // Create base agent (~10 lines)
    base, err := agentbase.NewBaseAgent(cfg, 30)
    if err != nil { log.Fatal(err) }
    defer base.Close()

    agent := NewResearchAgent(base, cfg)

    // HTTP server setup (~25 lines)
    server := &http.Server{...}
    http.HandleFunc("/research", agent.HandleResearchRequest)
    http.HandleFunc("/health", ...)

    // A2A server setup (~70 lines)
    go func() {
        listener, _ := net.Listen(...)
        agentCard := &a2a.AgentCard{...}
        mux := http.NewServeMux()
        // ... 50 more lines
    }()

    server.ListenAndServe()
}
```

### After: ~40 lines

```go
import (
    "github.com/agentplexus/agentkit/agent"
    "github.com/agentplexus/agentkit/config"
    "github.com/agentplexus/agentkit/a2a"
    "github.com/agentplexus/agentkit/httpserver"
)

func main() {
    cfg := config.LoadConfig()

    base, err := agent.NewBaseAgent(cfg, 30)
    if err != nil { log.Fatal(err) }
    defer base.Close()

    researchAgent := NewResearchAgent(base, cfg)

    // HTTP server - 5 lines
    httpServer, _ := httpserver.NewBuilder("research-agent", 8001).
        WithHandlerFunc("/research", researchAgent.HandleResearch).
        WithDualModeLog().
        Build()

    // A2A server - 5 lines
    a2aServer, _ := a2a.NewServer(a2a.Config{
        Agent: researchAgent.ADKAgent(),
        Port:  "9001",
    })

    // Start both
    a2aServer.StartAsync(ctx)
    httpServer.Start() // blocks
}
```

**Reduction: 150 lines → 40 lines (73% savings in main.go)**

---

## Recommendation Matrix

| Scenario | AgentKit Benefit | Recommendation |
|----------|------------------|----------------|
| Single simple agent | Low | Optional |
| Single complex multi-agent system | Medium | Recommended |
| 2-3 agent team projects | High | Strongly Recommended |
| 4+ agent team projects | Very High | Essential |
| Enterprise platform | Critical | Required |

---

## Migration Path

### For Existing Projects (like stats-agent-team)

1. **Add agentkit dependency**
   ```bash
   go get github.com/agentplexus/agentkit
   ```

2. **Replace pkg/ imports**
   ```go
   // Before
   import "github.com/agentplexus/stats-agent-team/pkg/config"
   import "github.com/agentplexus/stats-agent-team/pkg/llm"
   import "github.com/agentplexus/stats-agent-team/pkg/agent"

   // After
   import "github.com/agentplexus/agentkit/config"
   import "github.com/agentplexus/agentkit/llm"
   import "github.com/agentplexus/agentkit/agent"
   ```

3. **Replace A2A server setup**
   ```go
   // Before: 70 lines in a2a.go

   // After
   server, _ := a2a.NewServer(a2a.Config{
       Agent: myAgent,
       Port:  os.Getenv("A2A_PORT"),
   })
   ```

4. **Replace HTTP server setup**
   ```go
   // Before: 25 lines

   // After
   server, _ := httpserver.NewBuilder("agent-name", 8001).
       WithHandlerFunc("/endpoint", handler).
       Build()
   ```

5. **Remove redundant pkg/ code**
   - Delete `pkg/config/`
   - Delete `pkg/llm/`
   - Delete `pkg/agent/`
   - Delete `pkg/httpclient/`

### For New Projects

1. Start with agentkit as foundation
2. Focus only on domain-specific logic
3. Use provided factories for infrastructure
4. Add custom models as needed

---

## Architecture

```
agentkit/
│
├── # ===== Core Library (platform-agnostic) =====
├── a2a/                 # A2A protocol server factory
├── agent/               # Base agent implementations
├── config/              # Configuration management
├── http/                # HTTP client utilities
├── httpserver/          # HTTP server factory
├── llm/                 # LLM abstraction
│   └── adapters/        # OmniLLM adapter
├── orchestration/       # Workflow orchestration (Eino)
│
├── # ===== Platform-Specific =====
└── platforms/
    ├── agentcore/       # AWS Bedrock AgentCore
    │   ├── server.go    # /ping, /invocations
    │   ├── adapter.go   # Wrap Executors
    │   ├── registry.go  # Agent routing
    │   └── deploy/      # CDK, Terraform (future)
    │
    └── kubernetes/      # Kubernetes + Helm
        ├── values.go    # Helm values validation
        ├── templates/   # Go templates for Helm
        └── deploy/      # Helm charts (future)
```

---

## AWS AgentCore Runtime

AgentKit now supports **AWS Bedrock AgentCore** as a deployment target. AgentCore provides:

- **Firecracker microVM isolation** per session
- **Serverless scaling** from zero
- **Pay-per-use pricing** (only active CPU time)
- **Built-in session memory** and identity

### Kubernetes vs AgentCore

| Aspect | Kubernetes + Helm | AWS AgentCore |
|--------|------------------|---------------|
| **Distributions** | EKS, GKE, AKS, Minikube, kind | AWS only |
| **Config tool** | Helm values.yaml | AWS CDK / CloudFormation |
| **Scaling** | HPA, pod autoscaling | Automatic, session-based |
| **Isolation** | Containers | Firecracker microVMs |
| **Deployment** | `helm install` | `aws bedrock-agent-runtime` |

**Note:** Helm does NOT apply to AgentCore - use AWS CDK or Terraform instead.

### Running on AgentCore

```go
import "github.com/agentplexus/agentkit/platforms/agentcore"

func main() {
    // Create server with builder pattern
    server := agentcore.NewBuilder().
        WithPort(8080).
        WithAgent(researchAgent).
        WithAgent(synthesisAgent).
        WithDefaultAgent("research").
        MustBuild(ctx)

    // Start (blocks until shutdown)
    server.Start()
}
```

### Wrapping Eino Executors for AgentCore

```go
import (
    "github.com/agentplexus/agentkit/orchestration"
    "github.com/agentplexus/agentkit/platforms/agentcore"
)

// Build Eino workflow (same as before)
graph := buildOrchestrationGraph()
executor := orchestration.NewExecutor(graph, "stats-workflow")

// Wrap for AgentCore
agent := agentcore.WrapExecutor("stats", executor)

// Or with custom input/output handling
agent := agentcore.WrapExecutorWithPrompt("stats", executor,
    func(prompt string) StatsRequest {
        return StatsRequest{Topic: prompt}
    },
    func(output StatsResponse) string {
        return output.Summary
    },
)
```

### AgentCore Server Features

```go
server := agentcore.NewServer(agentcore.Config{
    Port:                  8080,           // Default: 8080
    ReadTimeout:           30 * time.Second,
    WriteTimeout:          300 * time.Second, // 5 min for long operations
    DefaultAgent:          "research",
    EnableRequestLogging:  true,
    EnableSessionTracking: true,
})

// Register agents
server.Register(ctx, researchAgent)
server.Register(ctx, synthesisAgent)

// Endpoints provided:
// - /ping        (health check)
// - /invocations (agent invocation)
```

### Same Agents, Different Runtimes

The key benefit: **same agent code runs on both Kubernetes and AgentCore**:

```go
// Agent implementation - runtime agnostic
type StatsAgent struct {
    executor *orchestration.Executor[StatsReq, StatsResp]
}

// Runtime 1: Kubernetes
httpServer, _ := httpserver.NewBuilder("stats", 8001).
    WithHandler("/stats", orchestration.NewHTTPHandler(executor)).
    Build()
httpServer.Start()

// Runtime 2: AWS AgentCore
acServer := agentcore.NewBuilder().
    WithAgent(agentcore.WrapExecutor("stats", executor)).
    MustBuild(ctx)
acServer.Start()
```

### Local Development

AgentCore code runs locally without AWS - same binary, different infrastructure:

```bash
go run main.go
curl localhost:8080/ping
curl -X POST localhost:8080/invocations -d '{"prompt":"test"}'
```

| Aspect | Local | AWS AgentCore |
|--------|-------|---------------|
| Process | Go binary | Firecracker microVM |
| Sessions | In-memory | Isolated per microVM |
| Scaling | Manual | Automatic |

**No code changes needed between local development and production.**

---

## Conclusion

AgentKit provides measurable benefits:

- **29% code reduction** per project (~1,500 lines)
- **Multiplicative savings** across multiple projects
- **Consistency** through standardized patterns
- **Security** via VaultGuard integration
- **Observability** through OmniObserve hooks
- **Multi-runtime deployment** - Kubernetes (Helm) or AWS AgentCore

The server factories (`a2a.NewServer()`, `httpserver.New()`, `agentcore.NewBuilder()`) eliminate tedious boilerplate, reducing agent setup from ~100 lines to ~10 lines.

**Write once, deploy anywhere:** The same agent code runs on Kubernetes or AWS AgentCore without modification. This is the key architectural benefit of AgentKit.

The value proposition strengthens significantly with scale. For organizations building multiple agent systems, AgentKit is essential infrastructure.
