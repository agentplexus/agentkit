# Platforms

AgentKit supports multiple deployment platforms. Your agent code is platform-agnostic - only the server bootstrap differs.

## Supported Platforms

### [Kubernetes](kubernetes.md)

Traditional container-based deployment with Helm charts.

- Any K8s distribution (EKS, GKE, AKS, Minikube, kind, k3s)
- Container orchestration with HPA
- Helm-based configuration
- Always-on pricing

### [AWS AgentCore](agentcore.md)

AWS Bedrock's serverless agent runtime.

- Firecracker microVM isolation
- Automatic scaling from zero
- Pay-per-use pricing
- Built-in session management

## Platform Comparison

| Aspect | Kubernetes | AWS AgentCore |
|--------|------------|---------------|
| Infrastructure | K8s manifests | AWS serverless |
| Config tool | Helm | CDK / Terraform |
| Scaling | HPA | Automatic |
| Isolation | Containers | Firecracker microVMs |
| Pricing | Always-on | Pay-per-use |
| Session handling | Application-managed | Built-in |

## Same Code, Different Runtimes

The key benefit: your agent implementation is runtime-agnostic.

```go
// Agent implementation - works on any platform
executor := orchestration.NewExecutor(graph, "stats")
```

### Kubernetes Deployment

```go
import "github.com/agentplexus/agentkit/httpserver"

httpServer, _ := httpserver.NewBuilder("stats", 8001).
    WithHandler("/stats", orchestration.NewHTTPHandler(executor)).
    Build()

httpServer.Start()
```

### AWS AgentCore Deployment

```go
import "github.com/agentplexus/agentkit/platforms/agentcore"

acServer := agentcore.NewBuilder().
    WithAgent(agentcore.WrapExecutor("stats", executor)).
    MustBuild(ctx)

acServer.Start()
```

## Choosing a Platform

| Scenario | Recommendation |
|----------|----------------|
| Existing K8s infrastructure | Kubernetes |
| AWS-native, pay-per-use | AgentCore |
| Need microVM isolation | AgentCore |
| Long-running agents | Kubernetes |
| Bursty, session-based workloads | AgentCore |
| Multi-cloud | Kubernetes |

## Local Development

Both platforms support the same local development workflow:

```bash
go run main.go
curl localhost:8080/ping
curl -X POST localhost:8080/invocations -d '{"prompt":"test"}'
```

The code runs locally as a regular Go process. Only production deployment differs.

## Next Steps

- [Kubernetes Deployment](kubernetes.md)
- [AWS AgentCore Deployment](agentcore.md)
