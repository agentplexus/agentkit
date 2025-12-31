# AgentKit v0.2.0 Release Notes

**Release Date:** December 31, 2025

This release adds Infrastructure-as-Code (IaC) support for AWS Bedrock AgentCore deployments, completing key roadmap items from v0.1.0.

## Highlights

- **IaC configuration in core** - Shared config structs for CDK, Pulumi, and CloudFormation
- **Pure CloudFormation generation** - Deploy without CDK/Pulumi runtime dependencies
- **New companion modules** - `agentkit-aws-cdk` (CDK) and `agentkit-aws-pulumi` (Pulumi)
- **AWS deployment guide** - Comprehensive documentation for AWS deployment options

## New Features

### IaC Configuration Package (`platforms/agentcore/iac/`)

Shared infrastructure configuration that works across all IaC tools:

```go
import "github.com/agentplexus/agentkit/platforms/agentcore/iac"

// Load from YAML or JSON
config, _ := iac.LoadStackConfigFromFile("config.yaml")

// Or build programmatically
config := &iac.StackConfig{
    StackName: "my-agents",
    Agents: []iac.AgentConfig{
        {Name: "research", ContainerImage: "ghcr.io/example/research:latest"},
        {Name: "orchestration", ContainerImage: "ghcr.io/example/orchestration:latest", IsDefault: true},
    },
    VPC: iac.DefaultVPCConfig(),
    Observability: &iac.ObservabilityConfig{Provider: "opik", Project: "my-agents"},
}
```

#### Configuration Types

| Type | Description |
|------|-------------|
| `StackConfig` | Complete deployment configuration |
| `AgentConfig` | Individual agent settings (memory, timeout, env vars) |
| `VPCConfig` | VPC/networking (create new or use existing) |
| `SecretsConfig` | AWS Secrets Manager integration |
| `ObservabilityConfig` | Opik, Langfuse, Phoenix, or CloudWatch |
| `IAMConfig` | Execution role and Bedrock access |

#### Config Loading

```go
// From file (auto-detects JSON/YAML)
config, _ := iac.LoadStackConfigFromFile("config.yaml")

// From raw data
config, _ := iac.LoadStackConfigFromJSON(jsonBytes)
config, _ := iac.LoadStackConfigFromYAML(yamlBytes)

// Write example config
iac.WriteExampleConfig("example.yaml")
```

### Pure CloudFormation Generation

Generate CloudFormation templates without CDK or Pulumi:

```go
import "github.com/agentplexus/agentkit/platforms/agentcore/iac"

config, _ := iac.LoadStackConfigFromFile("config.yaml")
iac.GenerateCloudFormationFile(config, "template.yaml")
```

```bash
aws cloudformation deploy \
    --template-file template.yaml \
    --stack-name my-agents \
    --capabilities CAPABILITY_IAM CAPABILITY_NAMED_IAM
```

**Zero additional dependencies** - uses only `gopkg.in/yaml.v3` which is already in agentkit.

### AWS Deployment Documentation

New comprehensive guide at `docs/aws-deployment-guide.md` covering:

- Deployment target comparison (AgentCore vs EKS vs ECS)
- IaC tool selection (CDK vs Pulumi vs CloudFormation)
- Decision tree for choosing deployment approach
- Module architecture and dependency strategy

## Companion Modules

### agentkit-aws-cdk (CDK)

AWS CDK constructs for AgentCore deployment:

```bash
go get github.com/agentplexus/agentkit-aws-cdk
```

```go
import "github.com/agentplexus/agentkit-aws-cdk/agentcore"

app := agentcore.NewApp()
agentcore.NewStackBuilder("my-agents").
    WithAgents(research, orchestration).
    WithOpik("my-project", secretARN).
    Build(app)
agentcore.Synth(app)
```

**Dependencies:** 21 transitive packages (CDK uses lightweight jsii bindings)

### agentkit-aws-pulumi (Pulumi)

Pulumi components for AgentCore deployment:

```bash
go get github.com/agentplexus/agentkit-aws-pulumi
```

```go
import "github.com/agentplexus/agentkit-aws-pulumi/agentcore"

pulumi.Run(func(ctx *pulumi.Context) error {
    _, err := agentcore.NewStackBuilder("my-agents").
        WithAgents(research, orchestration).
        WithOpik("my-project", secretARN).
        Build(ctx)
    return err
})
```

**Dependencies:** 340 transitive packages (native Go Pulumi SDK)

## IaC Approach Comparison

| Approach | Module | Dependencies | Runtime Required |
|----------|--------|--------------|------------------|
| **Pure CloudFormation** | `agentkit` only | 0 extra | AWS CLI only |
| **CDK** | `agentkit-aws-cdk` | +21 | Node.js (jsii) |
| **Pulumi** | `agentkit-aws-pulumi` | +340 | Pulumi CLI |

All approaches share the same YAML/JSON configuration schema.

## Package Structure

```
agentkit/
├── platforms/
│   ├── agentcore/
│   │   ├── iac/                    # NEW: Shared IaC config
│   │   │   ├── config.go           # Configuration structs
│   │   │   ├── loader.go           # JSON/YAML loading
│   │   │   └── cloudformation.go   # CF template generation
│   │   └── *.go                    # Runtime (unchanged)
│   └── kubernetes/                 # Unchanged
└── docs/
    └── aws-deployment-guide.md     # NEW: Deployment documentation
```

## Roadmap Progress

### Completed in v0.2.0

- [x] AWS CDK constructs for AgentCore deployment
- [x] Pulumi components for AgentCore deployment
- [x] Pure CloudFormation template generation

### Remaining

- [ ] Terraform modules for AgentCore
- [ ] ECS/Fargate deployment support
- [ ] Extended Helm chart library
- [ ] Additional LLM provider adapters
- [ ] Streaming response support

## Breaking Changes

None. This release is fully backward compatible with v0.1.0.

## Migration Guide

No migration required. Existing v0.1.0 code continues to work unchanged.

To use the new IaC features:

```go
// Add import
import "github.com/agentplexus/agentkit/platforms/agentcore/iac"

// Use config loading and CF generation
config, _ := iac.LoadStackConfigFromFile("config.yaml")
iac.GenerateCloudFormationFile(config, "template.yaml")
```

## Installation

```bash
# Core library (includes IaC config and CF generation)
go get github.com/agentplexus/agentkit

# Optional: CDK support
go get github.com/agentplexus/agentkit-aws-cdk

# Optional: Pulumi support
go get github.com/agentplexus/agentkit-aws-pulumi
```

## License

MIT License
