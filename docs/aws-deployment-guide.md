# AWS Deployment Guide

This guide covers deployment options for agentkit-based agents on AWS.

## Deployment Targets

| Target | Type | Best For | Status |
|--------|------|----------|--------|
| [Bedrock AgentCore](#bedrock-agentcore) | Serverless microVMs | Pay-per-invoke, zero ops | Preview |
| [EKS](#eks-kubernetes) | Kubernetes | Long-running, high control | GA |
| [ECS/Fargate](#ecsfargate) | Serverless containers | Serverless with container flexibility | GA |

---

## Bedrock AgentCore

AWS Bedrock AgentCore runs agents in Firecracker microVMs with automatic scaling and pay-per-invocation pricing.

### IaC Options

| Approach | Module | Dependencies | Best For |
|----------|--------|--------------|----------|
| **CDK** | `agentkit-aws` | 21 | AWS-native teams, CDK users |
| **Pulumi** | `agentkit-pulumi-aws` | 340 | Multi-cloud teams, Pulumi users |
| **CloudFormation** | `agentkit` only | 0 (just yaml.v3) | No IaC runtime, AWS CLI only |

### CDK Deployment

```go
import "github.com/agentplexus/agentkit-aws/agentcore"

func main() {
    app := agentcore.NewApp()

    agentcore.NewStackBuilder("my-agents").
        WithAgents(research, orchestration).
        WithOpik("my-project", "arn:aws:secretsmanager:...").
        Build(app)

    agentcore.Synth(app)
}
```

```bash
cdk deploy
```

### Pulumi Deployment

```go
import "github.com/agentplexus/agentkit-pulumi-aws/agentcore"

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        _, err := agentcore.NewStackBuilder("my-agents").
            WithAgents(research, orchestration).
            WithOpik("my-project", "arn:aws:secretsmanager:...").
            Build(ctx)
        return err
    })
}
```

```bash
pulumi up
```

### Pure CloudFormation (No CDK/Pulumi)

```go
import "github.com/agentplexus/agentkit/platforms/agentcore/iac"

func main() {
    config, _ := iac.LoadStackConfigFromFile("config.yaml")
    iac.GenerateCloudFormationFile(config, "template.yaml")
}
```

```bash
go run generate.go
aws cloudformation deploy \
    --template-file template.yaml \
    --stack-name my-agents \
    --capabilities CAPABILITY_IAM CAPABILITY_NAMED_IAM
```

---

## EKS (Kubernetes)

For EKS deployment, use standard Kubernetes tooling (Helm). agentkit provides Helm values structs in `agentkit/platforms/kubernetes/`.

### Why Helm Instead of CDK/Pulumi?

- Helm is the standard for Kubernetes app deployment
- CDK/Pulumi for K8s apps adds unnecessary complexity
- EKS cluster provisioning is typically done once by platform teams
- Helm integrates with GitOps (ArgoCD, Flux)

### Deployment

```bash
# Using Helm
helm install my-agents ./charts/agentkit \
    -f values.yaml \
    --namespace agents

# Or kubectl with manifests
kubectl apply -f manifests/
```

### When to Use EKS vs AgentCore

| Choose EKS When | Choose AgentCore When |
|-----------------|----------------------|
| Need persistent connections | Bursty, event-driven workloads |
| Custom networking requirements | Minimal ops desired |
| Existing K8s infrastructure | Pay-per-invocation preferred |
| Need GPU access | Cold start latency acceptable |
| Long-running agents | Short-lived agent tasks |

---

## ECS/Fargate

ECS with Fargate provides serverless containers without Kubernetes complexity. Support is planned but not yet implemented.

### Status

- **Not yet implemented** in agentkit-aws or agentkit-pulumi-aws
- Can be added to `agentkit-aws/ecs/` and `agentkit-pulumi-aws/ecs/` if needed

### When to Consider ECS

- Want serverless without AgentCore preview limitations
- Need more container configuration than AgentCore allows
- Existing ECS infrastructure

---

## Module Architecture

```
agentkit/                              # Core library
├── platforms/
│   ├── agentcore/
│   │   ├── iac/                       # Shared IaC config
│   │   │   ├── config.go              # StackConfig, AgentConfig, etc.
│   │   │   ├── loader.go              # JSON/YAML loading
│   │   │   └── cloudformation.go      # Pure CF generator
│   │   └── *.go                       # AgentCore runtime
│   └── kubernetes/
│       └── values.go                  # Helm values structs

agentkit-aws/                          # AWS CDK (21 deps)
└── agentcore/                         # AgentCore via CDK

agentkit-pulumi-aws/                   # Pulumi AWS (340 deps)
└── agentcore/                         # AgentCore via Pulumi
```

### Dependency Strategy

| Module | Dependencies | Rationale |
|--------|--------------|-----------|
| `agentkit` | ~40 | Core stays lean |
| `agentkit-aws` | +21 | CDK uses jsii (lightweight) |
| `agentkit-pulumi-aws` | +340 | Native Go Pulumi SDK |

Separate modules ensure users only pull dependencies for their chosen IaC tool.

---

## Configuration Sharing

All IaC approaches share the same configuration schema:

```yaml
# config.yaml - works with CDK, Pulumi, and CloudFormation
stackName: my-agents
description: My agent deployment

agents:
  - name: research
    containerImage: ghcr.io/example/research:latest
    memoryMB: 512
    timeoutSeconds: 30

  - name: orchestration
    containerImage: ghcr.io/example/orchestration:latest
    memoryMB: 1024
    timeoutSeconds: 300
    isDefault: true

vpc:
  createVPC: true
  vpcCidr: 10.0.0.0/16

observability:
  provider: opik
  project: my-agents
  enableCloudWatchLogs: true

tags:
  Environment: production
```

---

## Decision Tree

```
Need to deploy agentkit agents to AWS?
│
├─ Want serverless (no cluster management)?
│   │
│   ├─ OK with AgentCore preview? → Use AgentCore (CDK/Pulumi/CF)
│   │
│   └─ Need GA serverless? → Use ECS/Fargate (coming soon)
│
└─ Have/want Kubernetes?
    │
    └─ Use EKS + Helm
```

---

## Related Documentation

- [agentkit-aws README](https://github.com/agentplexus/agentkit-aws)
- [agentkit-pulumi-aws README](https://github.com/agentplexus/agentkit-pulumi-aws)
- [AWS Bedrock AgentCore Documentation](https://docs.aws.amazon.com/bedrock/latest/userguide/agentcore.html)
