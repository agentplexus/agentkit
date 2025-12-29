# llm

Multi-provider LLM abstraction via OmniLLM.

## Model Factory

```go
import "github.com/agentplexus/agentkit/llm"

cfg := config.LoadConfig()

factory := llm.NewModelFactory(cfg)
model, err := factory.CreateModel(ctx)
if err != nil {
    log.Fatal(err)
}
defer factory.Close()
```

## Supported Providers

| Provider | Environment Variables |
|----------|----------------------|
| Gemini | `GEMINI_API_KEY`, `LLM_MODEL` |
| Claude | `CLAUDE_API_KEY`, `LLM_MODEL` |
| OpenAI | `OPENAI_API_KEY`, `LLM_MODEL` |
| xAI | `XAI_API_KEY`, `LLM_MODEL` |
| Ollama | `OLLAMA_URL`, `LLM_MODEL` |

## Provider Selection

Set via environment:

```bash
export LLM_PROVIDER=gemini
export GEMINI_API_KEY=your-key
export LLM_MODEL=gemini-2.0-flash-exp
```

Or in code:

```go
cfg := config.LoadConfig()
cfg.LLMProvider = "claude"
cfg.LLMModel = "claude-3-5-sonnet-20241022"

factory := llm.NewModelFactory(cfg)
```

## Model Interface

The factory returns an OmniLLM model that implements a standard interface:

```go
model, _ := factory.CreateModel(ctx)

// Generate text
response, err := model.Generate(ctx, prompt)

// Chat completion
messages := []Message{
    {Role: "user", Content: "Hello"},
}
response, err := model.Chat(ctx, messages)
```

## Observability

Enable LLM observability with OmniObserve:

```bash
export OBSERVABILITY_ENABLED=true
export OBSERVABILITY_PROVIDER=opik  # or langfuse, phoenix
```

```go
cfg := config.LoadConfig()
// Observability is automatically configured
factory := llm.NewModelFactory(cfg)
```

## Adapters

The `llm/adapters` package provides adapters for specific frameworks:

```go
import "github.com/agentplexus/agentkit/llm/adapters"

// Eino adapter
einoModel := adapters.NewEinoAdapter(model)

// ADK adapter
adkModel := adapters.NewADKAdapter(model)
```

## Usage with BaseAgent

The BaseAgent automatically creates and manages the LLM:

```go
ba, _ := agent.NewBaseAgent(cfg, "my-agent", 30)

// LLM is available via base agent
response, err := ba.Generate(ctx, prompt)
```
