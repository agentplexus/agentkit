# AgentKit v0.3.0 Release Notes

**Release Date:** January 4, 2026

This release adds AgentCore-specific configuration options for protocol selection, authorization, memory, and multi-agent gateway support.

## Highlights

- **Protocol configuration** - Choose HTTP, MCP, or A2A communication protocols
- **Authorization support** - Configure IAM or Lambda authorizers for inbound requests
- **Memory support** - Enable persistent memory for stateful agents
- **Gateway configuration** - Configure multi-agent routing gateways

## New Features

### Protocol Configuration

Agents can now specify their communication protocol:

```yaml
agents:
  - name: research
    containerImage: ghcr.io/example/research:latest
    protocol: HTTP  # or MCP, A2A
```

```go
agent := iac.AgentConfig{
    Name:           "research",
    ContainerImage: "ghcr.io/example/research:latest",
    Protocol:       "MCP",  // Model Context Protocol
}
```

Supported protocols:

| Protocol | Description |
|----------|-------------|
| `HTTP` | Standard HTTP/REST (default) |
| `MCP` | Model Context Protocol for tool integration |
| `A2A` | Agent-to-Agent Protocol for inter-agent communication |

### Authorization Configuration

Configure inbound authorization for agent endpoints:

```yaml
agents:
  - name: secure-agent
    containerImage: ghcr.io/example/agent:latest
    authorizer:
      type: IAM  # or LAMBDA, NONE
```

```go
agent := iac.AgentConfig{
    Name:           "secure-agent",
    ContainerImage: "ghcr.io/example/agent:latest",
    Authorizer: &iac.AuthorizerConfig{
        Type: "LAMBDA",
        LambdaARN: "arn:aws:lambda:us-east-1:123456789012:function:my-authorizer",
    },
}
```

Supported authorizer types:

| Type | Description |
|------|-------------|
| `NONE` | No authorization (default) |
| `IAM` | AWS IAM-based authorization |
| `LAMBDA` | Custom Lambda authorizer |

### Memory Support

Enable persistent memory for stateful agents:

```yaml
agents:
  - name: assistant
    containerImage: ghcr.io/example/assistant:latest
    enableMemory: true
```

### Gateway Configuration

Configure multi-agent gateways for routing requests:

```yaml
gateway:
  enabled: true
  name: my-gateway
  description: Multi-agent routing gateway
  targets:
    - research
    - synthesis
    - orchestration
```

```go
config := &iac.StackConfig{
    StackName: "my-agents",
    Gateway: &iac.GatewayConfig{
        Enabled:     true,
        Name:        "my-gateway",
        Description: "Multi-agent routing gateway",
        Targets:     []string{"research", "synthesis"},
    },
}
```

## New Types

### AuthorizerConfig

```go
type AuthorizerConfig struct {
    Type      string `json:"type" yaml:"type"`                           // IAM, LAMBDA, NONE
    LambdaARN string `json:"lambdaArn,omitempty" yaml:"lambdaArn,omitempty"`
}
```

### GatewayConfig

```go
type GatewayConfig struct {
    Enabled     bool     `json:"enabled,omitempty" yaml:"enabled,omitempty"`
    Name        string   `json:"name,omitempty" yaml:"name,omitempty"`
    Description string   `json:"description,omitempty" yaml:"description,omitempty"`
    Targets     []string `json:"targets,omitempty" yaml:"targets,omitempty"`
}
```

## New Helper Functions

```go
// ValidProtocols returns valid agent protocols
iac.ValidProtocols() // ["HTTP", "MCP", "A2A"]

// ValidAuthorizerTypes returns valid authorizer types
iac.ValidAuthorizerTypes() // ["IAM", "LAMBDA", "NONE"]
```

## Enhanced Validation

- Protocol values validated against allowed list
- Authorizer type validated against allowed list
- Lambda ARN required when authorizer type is LAMBDA
- Gateway targets validated against defined agent names

## Example Configuration

```yaml
stackName: stats-agent-team
description: Statistics research and verification system

agents:
  - name: research
    containerImage: ghcr.io/agentplexus/stats-agent-research:v0.5.1
    memoryMB: 512
    timeoutSeconds: 30
    protocol: HTTP
    description: Research agent

  - name: orchestration
    containerImage: ghcr.io/agentplexus/stats-agent-orchestration:v0.5.1
    memoryMB: 512
    timeoutSeconds: 300
    protocol: HTTP
    isDefault: true
    authorizer:
      type: IAM

gateway:
  enabled: true
  name: stats-gateway
  targets:
    - research
    - orchestration

vpc:
  createVPC: true

observability:
  provider: opik
  project: stats-agent-team
```

## Breaking Changes

None. This release is fully backward compatible with v0.2.0.

## Migration Guide

No migration required. New fields are optional with sensible defaults:

- `protocol` defaults to `"HTTP"`
- `authorizer` defaults to `nil` (no authorization)
- `enableMemory` defaults to `false`
- `gateway` defaults to `nil` (no gateway)

## Installation

```bash
go get github.com/agentplexus/agentkit@v0.3.0
```

## License

MIT License
