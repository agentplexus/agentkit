---
marp: true
theme: agentplexus
paginate: true
---

<!-- _paginate: false -->

# AgentKit 🛠️

## A Go Library for Building AI Agent Applications

Reusable components for LLM integration, workflow orchestration, and multi-agent systems

---

# What is AgentKit? 🤔

A **foundational library** providing:

- 🧠 **LLM Abstraction** - Gemini, Claude, OpenAI, xAI, Ollama
- ⚙️ **Configuration Management** - Environment-based with optional security
- 🔀 **Workflow Orchestration** - Type-safe graph-based execution
- 🏭 **Server Factories** - A2A, HTTP, AgentCore setup in 5 lines
- ☁️ **Multi-Runtime** - Kubernetes (Helm) or AWS AgentCore

---

# Architecture Overview 🏗️

```
agentkit/
├── # Core (platform-agnostic)
├── a2a/             # 🔗 A2A protocol server
├── agent/           # 🤖 Base agent
├── config/          # ⚙️ Configuration
├── http/            # 🌐 HTTP utilities
├── httpserver/      # 🏭 HTTP server factory
├── llm/             # 🧠 LLM abstraction
├── orchestration/   # 🔀 Eino workflows
│
├── # Platform-specific
└── platforms/
    ├── agentcore/   # ☁️ AWS Bedrock AgentCore
    └── kubernetes/  # ⎈ Kubernetes + Helm
```

---

<!-- _paginate: false -->

# Server Factories ⚙️

Eliminating boilerplate with A2A and HTTP server factories

---

# The Problem: Boilerplate 😤

Every agent project repeats the same patterns:

| Pattern | Lines Duplicated |
|---------|------------------|
| 🔗 A2A server setup | ~350 lines |
| 🌐 HTTP server setup | ~125 lines |
| 📨 HTTP handler setup | ~100 lines |
| 🧠 LLM factory | ~200 lines |
| ⚙️ Config management | ~140 lines |
| **Total** | **~915 lines** |

---

# Case Study: stats-agent-team 📊

A multi-agent system for finding and verifying statistics

```
Total codebase:     5,226 lines
├── Domain logic:   ~3,500 lines (agents, models, CLI)
├── Shared pkg/:    ~930 lines (config, llm, http)
└── Boilerplate:    ~790 lines (server setup)
```

⚠️ **15% of the code is pure boilerplate**

---

# What AgentKit Provides 📦

### Core Packages
| Component | Benefit |
|-----------|---------|
| `config.Config` | ⚙️ Centralized configuration |
| `config.SecureConfig` | 🔒 VaultGuard security integration |
| `llm.ModelFactory` | 🧠 Multi-provider LLM abstraction |
| `orchestration.GraphBuilder[I,O]` | 🔀 Type-safe workflow graphs |
| `orchestration.HTTPHandler[I,O]` | 📨 Generic JSON handlers |
| `orchestration.AgentCaller` | 🔗 Inter-agent HTTP calls |

---

# Server Factories 🏭

### The Big Win 🎯

| Component | Lines Saved | Reduction |
|-----------|-------------|-----------|
| `a2a.NewServer()` | ~350 lines | 70 → 5 |
| `httpserver.New()` | ~125 lines | 25 → 5 |
| `httpserver.NewBuilder()` | Fluent API | - |

✅ **Total: ~475 lines of boilerplate eliminated per project**

---

# A2A Server: Before ❌

```go
// Every agent repeats this pattern (~70 lines)
func startA2AServer(agent Agent, port string) error {
    listener, _ := net.Listen("tcp", "0.0.0.0:"+port)

    agentCard := &a2a.AgentCard{
        Name:        agent.Name(),
        Description: "...",
        Skills:      adka2a.BuildAgentSkills(agent),
        // ... 10 more lines
    }

    mux := http.NewServeMux()
    mux.Handle(a2asrv.WellKnownAgentCardPath,
        a2asrv.NewStaticAgentCardHandler(agentCard))

    executor := adka2a.NewExecutor(/* 10 lines of config */)
    // ... 20 more lines of setup
}
```

---

# A2A Server: After ✅

```go
import "github.com/agentplexus/agentkit/a2a"

server, _ := a2a.NewServer(a2a.Config{
    Agent:       myAgent,
    Port:        "9001",
    Description: "Research agent for web search",
})

server.Start(ctx)
```

🎉 **70 lines → 5 lines**

---

# A2A Server Features ⚡

```go
server, _ := a2a.NewServer(a2a.Config{
    Agent:             myAgent,
    Port:              "9001",           // Empty = random port
    Description:       "My agent",
    InvokePath:        "/invoke",        // Default: /invoke
    ReadHeaderTimeout: 10 * time.Second,
    SessionService:    customService,    // Default: in-memory
})

// Useful methods
server.URL()          // "http://localhost:9001"
server.AgentCardURL() // "http://localhost:9001/.well-known/agent.json"
server.InvokeURL()    // "http://localhost:9001/invoke"
server.StartAsync(ctx) // 🚀 Non-blocking
server.Stop(ctx)       // 🛑 Graceful shutdown
```

---

# HTTP Server: Before ❌

```go
// Every agent repeats this (~25 lines)
server := &http.Server{
    Addr:         ":8001",
    ReadTimeout:  30 * time.Second,
    WriteTimeout: 120 * time.Second,
    IdleTimeout:  60 * time.Second,
}

http.HandleFunc("/research", agent.HandleResearchRequest)
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
})

log.Println("Server starting on :8001")
server.ListenAndServe()
```

---

# HTTP Server: After (Config) ✅

```go
import "github.com/agentplexus/agentkit/httpserver"

server, _ := httpserver.New(httpserver.Config{
    Name: "research-agent",
    Port: 8001,
    HandlerFuncs: map[string]http.HandlerFunc{
        "/research": agent.HandleResearchRequest,
    },
    EnableDualModeLog: true,
})

server.Start()
```

🎉 **25 lines → 8 lines**

---

# HTTP Server: After (Builder) 🔨

```go
server, _ := httpserver.NewBuilder("research-agent", 8001).
    WithHandlerFunc("/research", agent.HandleResearch).
    WithHandlerFunc("/synthesize", agent.HandleSynthesize).
    WithHandler("/orchestrate", orchestration.NewHTTPHandler(exec)).
    WithTimeouts(30*time.Second, 120*time.Second, 60*time.Second).
    WithDualModeLog().
    Build()

server.Start()
```

✨ **Fluent API for clean, readable configuration**

---

# Complete Agent: Before 📝

```go
func main() {
    cfg := config.LoadConfig()
    base, _ := agentbase.NewBaseAgent(cfg, 30)
    agent := NewResearchAgent(base, cfg)

    // HTTP server setup (~25 lines)
    server := &http.Server{...}
    http.HandleFunc("/research", agent.HandleResearchRequest)
    http.HandleFunc("/health", ...)

    // A2A server setup (~70 lines)
    go func() {
        listener, _ := net.Listen(...)
        agentCard := &a2a.AgentCard{...}
        // ... 50 more lines
    }()

    server.ListenAndServe()
}
```

---

# Complete Agent: After ✨

```go
func main() {
    cfg := config.LoadConfig()
    base, _ := agent.NewBaseAgent(cfg, 30)
    researchAgent := NewResearchAgent(base, cfg)

    // 🌐 HTTP server - 5 lines
    httpServer, _ := httpserver.NewBuilder("research-agent", 8001).
        WithHandlerFunc("/research", researchAgent.HandleResearch).
        Build()

    // 🔗 A2A server - 5 lines
    a2aServer, _ := a2a.NewServer(a2a.Config{
        Agent: researchAgent.ADKAgent(),
        Port:  "9001",
    })

    a2aServer.StartAsync(ctx)
    httpServer.Start()
}
```

---

<!-- _paginate: false -->

# Benefits 📈

Quantifying the impact across projects

---

# Benefit Analysis: Single Project 📊

```
stats-agent-team with AgentKit:

Before:  5,226 lines
After:   ~3,700 lines

Savings: ~1,500 lines (29%)
```

### Breakdown
- 📦 Replace `pkg/` with imports: ~930 lines
- 🔗 A2A server factory: ~350 lines
- 🌐 HTTP server factory: ~125 lines
- 📨 HTTPHandler generic: ~100 lines

---

# Benefit Analysis: Multiple Projects 📈

| Projects | Lines Saved | Maintenance Benefit |
|----------|-------------|---------------------|
| 1 | 1,500 | Single project |
| 2 | 3,000 | 🔄 Bug fixes shared |
| 5 | 7,500 | 📐 Consistent patterns |
| 10 | 15,000 | 🏢 Platform-level reuse |

🚀 **Each new project starts with 1,500 fewer lines to write**

---

# Beyond Line Count 🌟

### 📐 Consistency
- Same patterns across all agent projects
- Easier code reviews and onboarding

### 🔒 Security
- VaultGuard integration built-in
- Secure credential management

### 👁️ Observability
- OmniObserve hooks standardized
- Opik, Langfuse, Phoenix support

### 🚀 Deployment
- Helm validation and templates
- Kubernetes-ready from day one

---

<!-- _paginate: false -->

# Platform Deployment 🚀

Kubernetes with Helm and AWS AgentCore

---

# Helm Chart Support ⎈

### Reusable templates in `platforms/kubernetes/`

```
platforms/kubernetes/
├── values.go               # ✅ Go structs with validation
└── templates/
    ├── _helpers.tpl        # 🔧 Common template functions
    └── deployment.yaml.tpl # 📄 Generic deployment template
```

### Usage in your chart
```yaml
# In your deployment.yaml
{{- include "agentkit.deployment" (dict "agent" .Values.research "name" "research" "values" .) }}
```

⚠️ **Note:** Dockerfile not included - project-specific

---

# AWS AgentCore Runtime ☁️

AgentKit now supports **AWS Bedrock AgentCore**:

- 🔥 **Firecracker microVM** isolation per session
- 📈 **Serverless scaling** from zero
- 💰 **Pay-per-use** (only active CPU time)
- 🧠 **Built-in session memory** and identity

⚠️ **Note:** Helm does NOT apply - use AWS CDK or Terraform

---

# Kubernetes vs AgentCore ⚖️

| Aspect | Kubernetes | AgentCore |
|--------|------------|-----------|
| Distributions | ⎈ EKS, GKE, AKS, Minikube, kind | ☁️ AWS only |
| Config tool | 📄 Helm | 🏗️ CDK / Terraform |
| Scaling | 📊 HPA | 🚀 Automatic |
| Isolation | 📦 Containers | 🔥 Firecracker microVMs |
| Pricing | 💵 Always-on | 💰 Pay-per-use |

---

# AgentCore: Simple Setup 🚀

```go
import "github.com/agentplexus/agentkit/platforms/agentcore"

server := agentcore.NewBuilder().
    WithPort(8080).
    WithAgent(researchAgent).
    WithAgent(synthesisAgent).
    WithDefaultAgent("research").
    MustBuild(ctx)

server.Start()
```

📡 **Endpoints: /ping, /invocations**

---

# Wrap Eino Executors for AgentCore 🔄

```go
// Build Eino workflow (same as before)
graph := buildOrchestrationGraph()
executor := orchestration.NewExecutor(graph, "stats-workflow")

// 📦 Wrap for AgentCore
agent := agentcore.WrapExecutor("stats", executor)

// 🔧 Or with custom I/O
agent := agentcore.WrapExecutorWithPrompt("stats", executor,
    func(prompt string) StatsReq { return StatsReq{Topic: prompt} },
    func(out StatsResp) string { return out.Summary },
)
```

---

# Same Code, Different Runtimes ♻️

```go
// Agent implementation - runtime agnostic
executor := orchestration.NewExecutor(graph, "stats")

// ⎈ Runtime 1: Kubernetes
httpServer, _ := httpserver.NewBuilder("stats", 8001).
    WithHandler("/stats", orchestration.NewHTTPHandler(executor)).
    Build()

// ☁️ Runtime 2: AWS AgentCore
acServer := agentcore.NewBuilder().
    WithAgent(agentcore.WrapExecutor("stats", executor)).
    MustBuild(ctx)
```

✨ **Write once, deploy anywhere**

---

# Local Development 💻

AgentCore code runs locally - same binary, different infrastructure:

```bash
# Run locally
go run main.go

# Test endpoints
curl localhost:8080/ping
curl -X POST localhost:8080/invocations -d '{"prompt":"test"}'
```

| Aspect | 🖥️ Local | ☁️ AWS AgentCore |
|--------|----------|------------------|
| Process | Regular Go binary | Same binary in Firecracker |
| Sessions | In-memory | Isolated per microVM |
| Scaling | Manual | Automatic |

🚀 **No code changes between dev and production**

---

<!-- _paginate: false -->

# Getting Started 🏁

When to use AgentKit and how to migrate

---

# When to Use AgentKit 💡

| Scenario | Recommendation |
|----------|----------------|
| Single simple agent | 🤔 Maybe overkill |
| Single complex agent system | ✅ **Good fit** |
| 2-3 agent projects | ✅ **Strong fit** |
| Platform of agent teams | 🎯 **Essential** |

---

# Architecture Vision 🎯

```
                ┌─────────────────────────────┐
                │         agentkit            │
                │   (shared foundation)       │
                └─────────────┬───────────────┘
                              │
      ┌───────────────────────┼───────────────────────┐
      │                       │                       │
      ▼                       ▼                       ▼
┌───────────┐         ┌───────────┐         ┌───────────┐
│  stats-   │         │  docs-    │         │  code-    │
│  agent-   │         │  agent-   │         │  agent-   │
│  team     │         │  team     │         │  team     │
└───────────┘         └───────────┘         └───────────┘
```

---

# Migration Path 🛤️

### 1️⃣ Add dependency
```bash
go get github.com/agentplexus/agentkit
```

### 2️⃣ Replace imports
```go
// Before
import "github.com/myproject/pkg/config"

// After
import "github.com/agentplexus/agentkit/config"
```

### 3️⃣ Use factories
```go
server, _ := a2a.NewServer(a2a.Config{...})
httpServer, _ := httpserver.NewBuilder(...).Build()
```

---

# Summary ✅

### AgentKit eliminates boilerplate so you can focus on domain logic

- 📉 **~1,500 lines saved** per project (29%)
- 🏭 **Server factories** reduce setup from 100 → 10 lines
- ☁️ **Multi-runtime** - Kubernetes or AWS AgentCore
- ♻️ **Write once, deploy anywhere**
- 🔒 **Security & observability** built-in

### Get started:
```go
import "github.com/agentplexus/agentkit"
```

---

# Questions? ❓

GitHub: `github.com/agentplexus/agentkit`
