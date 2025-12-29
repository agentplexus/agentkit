# config

Configuration management with optional VaultGuard security integration.

## Basic Configuration

```go
import "github.com/agentplexus/agentkit/config"

cfg := config.LoadConfig()

// Access values
provider := cfg.LLMProvider
model := cfg.LLMModel
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `LLM_PROVIDER` | Provider (gemini, claude, openai, xai, ollama) | gemini |
| `LLM_MODEL` | Model name | Provider default |
| `GEMINI_API_KEY` | Gemini API key | - |
| `CLAUDE_API_KEY` | Claude/Anthropic API key | - |
| `OPENAI_API_KEY` | OpenAI API key | - |
| `XAI_API_KEY` | xAI API key | - |
| `OLLAMA_URL` | Ollama server URL | http://localhost:11434 |
| `OBSERVABILITY_ENABLED` | Enable observability | false |
| `OBSERVABILITY_PROVIDER` | Provider (opik, langfuse, phoenix) | opik |

## Secure Configuration

Use VaultGuard for production credential management:

```go
import "github.com/agentplexus/agentkit/config"

// Load with security checks
secCfg, err := config.LoadSecureConfig(ctx,
    config.WithPolicy(nil), // Default policy
)
if err != nil {
    log.Fatalf("Security check failed: %v", err)
}
defer secCfg.Close()

// Check security score
result := secCfg.SecurityResult()
log.Printf("Security score: %d", result.Score)
log.Printf("Environment: %s", secCfg.Environment())

// Get credentials securely
apiKey, err := secCfg.GetCredential(ctx, "GEMINI_API_KEY")
```

## Security Policies

```go
// Development policy (relaxed)
secCfg, _ := config.LoadSecureConfig(ctx, config.WithDevPolicy())

// Strict policy
secCfg, _ := config.LoadSecureConfig(ctx, config.WithStrictPolicy())

// Custom policy
policy := &vaultguard.Policy{
    MinSecurityScore:  70,
    RequireEncryption: true,
    RequireIAM:        true,
}
secCfg, _ := config.LoadSecureConfig(ctx, config.WithPolicy(policy))
```

## Config Struct

```go
type Config struct {
    // LLM
    LLMProvider string
    LLMModel    string

    // API Keys (from environment)
    GeminiAPIKey  string
    ClaudeAPIKey  string
    OpenAIAPIKey  string
    XAIAPIKey     string
    OllamaURL     string

    // Observability
    ObservabilityEnabled  bool
    ObservabilityProvider string

    // Search
    SearchProvider string
    SerperAPIKey   string
    SerpAPIKey     string
}
```

## Usage with BaseAgent

```go
cfg := config.LoadConfig()

// Config is passed to BaseAgent
ba, err := agent.NewBaseAgent(cfg, "my-agent", 30)
```

## Secure BaseAgent

```go
// Creates agent with security checks
ba, secCfg, err := agent.NewBaseAgentSecure(ctx, "my-agent", 30,
    config.WithPolicy(nil),
)
if err != nil {
    log.Fatalf("Security check failed: %v", err)
}
defer ba.Close()
defer secCfg.Close()
```
