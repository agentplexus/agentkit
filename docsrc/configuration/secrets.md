# Secrets Management

AgentKit integrates with [OmniVault](https://github.com/agentplexus/omnivault) for unified secret management across all deployment environments.

## Overview

OmniVault provides a single API for accessing secrets regardless of where your agent runs:

- **Local development**: Environment variables
- **AWS AgentCore**: AWS Secrets Manager
- **EKS**: AWS Secrets Manager with IRSA
- **Docker Compose**: Mounted .env files

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        AgentKit                              │
│                                                              │
│  ┌──────────────┐    ┌──────────────────────────────────┐   │
│  │ ConfigLoader │    │         OmniVault                │   │
│  │              │    │  ┌────────────┐ ┌─────────────┐  │   │
│  │ config.json  │    │  │ env://     │ │ aws-sm://   │  │   │
│  │              │    │  │ file://    │ │ aws-ssm://  │  │   │
│  └──────────────┘    │  │ memory://  │ │             │  │   │
│                      │  └────────────┘ └─────────────┘  │   │
│                      └──────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## Providers

| Provider | Use Case | Authentication |
|----------|----------|----------------|
| `env` | Local development | Environment variables |
| `aws-sm` | AWS Secrets Manager | IAM role / IRSA |
| `aws-ssm` | AWS Parameter Store | IAM role / IRSA |
| `memory` | Testing | In-memory storage |

## Configuration

### Config File (config.json)

The `secrets` section in config.json configures the provider:

```json
{
  "llm": {
    "provider": "gemini",
    "model": "gemini-2.0-flash-exp"
  },
  "search": {
    "provider": "serper"
  },
  "secrets": {
    "provider": "env",
    "prefix": "",
    "region": ""
  }
}
```

For AWS deployment:

```json
{
  "secrets": {
    "provider": "aws-sm",
    "prefix": "stats-agent-team/",
    "region": "us-west-2"
  }
}
```

### Programmatic Configuration

```go
import "github.com/agentplexus/agentkit/config"

// Load config with automatic provider detection
cfg, err := config.Load(ctx, config.LoadOptions{})

// Or specify provider explicitly
cfg, err := config.Load(ctx, config.LoadOptions{
    SecretsProvider: "aws-sm",
    SecretsPrefix:   "stats-agent-team/",
    SecretsRegion:   "us-west-2",
})
```

## Credential Resolution Order

When retrieving a secret, AgentKit follows this order:

1. **OmniVault provider** (if configured)
2. **Environment variables** (fallback)
3. **VaultGuard validation** (security checks)
4. **Default values** (if allowed)

## Usage

### Basic Usage

```go
import "github.com/agentplexus/agentkit/config"

func main() {
    ctx := context.Background()

    // Load configuration (reads config.json + secrets)
    cfg, err := config.Load(ctx, config.LoadOptions{})
    if err != nil {
        log.Fatal(err)
    }
    defer cfg.Close()

    // API keys are automatically loaded from the configured provider
    fmt.Println("LLM Provider:", cfg.LLMProvider)
    fmt.Println("Has API Key:", cfg.LLMAPIKey != "")

    // Get a specific secret
    apiKey, err := cfg.GetSecret(ctx, "CUSTOM_API_KEY")
    if err != nil {
        log.Printf("Secret not found: %v", err)
    }
}
```

### With SecureConfig (VaultGuard Integration)

```go
import "github.com/agentplexus/agentkit/config"

func main() {
    ctx := context.Background()

    // Load with security checks + OmniVault
    cfg, err := config.LoadSecureConfig(ctx,
        config.WithAutoSecretsProvider(),  // Auto-detect environment
    )
    if err != nil {
        log.Fatal(err)
    }
    defer cfg.Close()

    // Check security status
    fmt.Println("Environment:", cfg.Environment())
    fmt.Println("Security Score:", cfg.SecurityResult().Score)
}
```

### AWS Secrets Manager

```go
import "github.com/agentplexus/agentkit/config"

func main() {
    ctx := context.Background()

    cfg, err := config.LoadSecureConfig(ctx,
        config.WithAWSSecretsManager("stats-agent-team/", "us-west-2"),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer cfg.Close()

    // Secrets are loaded from AWS Secrets Manager
    // with prefix "stats-agent-team/"
}
```

## Environment Detection

AgentKit automatically detects the runtime environment and selects the appropriate provider:

| Environment | Detection | Provider |
|-------------|-----------|----------|
| Local dev | No AWS indicators | `env` |
| Docker Compose | No AWS indicators | `env` |
| AWS ECS/Fargate | `ECS_CONTAINER_METADATA_URI_V4` | `aws-sm` |
| AWS Lambda | `AWS_LAMBDA_FUNCTION_NAME` | `aws-sm` |
| EC2 | `AWS_EXECUTION_ENV` | `aws-sm` |
| EKS with IRSA | AWS web identity token | `aws-sm` |

To use auto-detection:

```go
cfg, err := config.LoadSecureConfig(ctx,
    config.WithAutoSecretsProvider(),
)
```

## Config File Search Path

AgentKit searches for config files in this order:

1. Explicit path provided to `LoadOptions.ConfigFile`
2. `config.json` in current directory
3. `config.yaml` in current directory
4. `../config.json` (parent directory)
5. `~/.agentplexus/projects/{project}/config.json`
6. `~/.agentplexus/config.json`

## AWS Secrets Manager Setup

### 1. Create Secrets

Use the `push-secrets` tool from agentkit-aws-cdk:

```bash
push-secrets --project stats-agent-team --region us-west-2
```

This creates secrets like:
- `stats-agent-team/GOOGLE_API_KEY`
- `stats-agent-team/SERPER_API_KEY`

### 2. IAM Permissions

Your ECS task role needs:

```json
{
  "Effect": "Allow",
  "Action": [
    "secretsmanager:GetSecretValue"
  ],
  "Resource": "arn:aws:secretsmanager:us-west-2:*:secret:stats-agent-team/*"
}
```

### 3. Config File

Include in your container image:

```json
{
  "secrets": {
    "provider": "aws-sm",
    "prefix": "stats-agent-team/",
    "region": "us-west-2"
  }
}
```

## Migration from Environment Variables

If you're currently using `config.LoadConfig()` (env-only), migration is seamless:

```go
// Before: env-only
cfg := config.LoadConfig()

// After: config.json + OmniVault (backward compatible)
cfg := config.LoadConfig()  // Still works, now checks config.json first
```

Or use the new context-based API:

```go
// Recommended: explicit context and error handling
cfg, err := config.Load(ctx, config.LoadOptions{})
```

## Related Documentation

- [Credentials and Configuration](credentials.md)
- [AWS AgentCore Deployment](../platforms/agentcore.md)
