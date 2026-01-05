# Credentials and Configuration

This guide covers how to configure credentials and settings for agentplexus projects using the `~/.agentplexus` directory.

## Overview

The `~/.agentplexus` directory provides a centralized location for storing credentials and configuration across all your agent projects. This approach:

- Keeps API keys out of project directories (and version control)
- Supports multiple projects with project-specific overrides
- Provides global defaults that apply to all projects

## Directory Structure

```
~/.agentplexus/
├── .env                           # Global credentials (shared across all projects)
├── config.yaml                    # Optional: global defaults (region, prefix)
└── projects/
    ├── stats-agent-team/
    │   ├── .env                   # Project-specific credentials/overrides
    │   └── config.json            # Optional: project CDK config
    ├── code-review-agent/
    │   └── .env
    └── my-custom-agent/
        └── .env
```

## Credential Lookup Order

The deployment tools (`deploy`, `push-secrets`) search for `.env` files in this order:

1. `.env` (current directory)
2. `../.env` (parent directory)
3. `~/.agentplexus/projects/{project}/.env` (project-specific)
4. `~/.agentplexus/.env` (global fallback)

The first file found is used.

## Project Detection

Project name is determined in this order:

1. `--project` flag (explicit)
2. `stackName` from `config.json` in the CDK directory
3. Current directory name (fallback)

## Setup

### 1. Create the Directory Structure

```bash
mkdir -p ~/.agentplexus/projects
chmod 700 ~/.agentplexus
```

### 2. Create Global Credentials

Add API keys that are shared across all projects:

```bash
cat > ~/.agentplexus/.env << 'EOF'
# LLM Providers (choose one or more)
GOOGLE_API_KEY=your-gemini-api-key
# ANTHROPIC_API_KEY=your-anthropic-key
# OPENAI_API_KEY=your-openai-key

# Search Providers
SERPER_API_KEY=your-serper-key

# Default Configuration
LLM_PROVIDER=gemini
SEARCH_PROVIDER=serper
EOF

chmod 600 ~/.agentplexus/.env
```

### 3. Create Project-Specific Overrides (Optional)

For projects that need different settings:

```bash
mkdir -p ~/.agentplexus/projects/stats-agent-team

cat > ~/.agentplexus/projects/stats-agent-team/.env << 'EOF'
# Override LLM provider for this project
LLM_PROVIDER=claude
ANTHROPIC_API_KEY=your-anthropic-key

# Project-specific observability
OPIK_API_KEY=your-opik-key
OPIK_PROJECT=stats-research
EOF

chmod 600 ~/.agentplexus/projects/stats-agent-team/.env
```

## Environment Variables

### LLM Providers

| Variable | Description |
|----------|-------------|
| `GOOGLE_API_KEY` | Google AI / Gemini API key |
| `ANTHROPIC_API_KEY` | Anthropic / Claude API key |
| `OPENAI_API_KEY` | OpenAI API key |
| `XAI_API_KEY` | xAI / Grok API key |

### Search Providers

| Variable | Description |
|----------|-------------|
| `SERPER_API_KEY` | Serper.dev API key |
| `SERPAPI_API_KEY` | SerpAPI key |

### Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `LLM_PROVIDER` | LLM provider: `gemini`, `claude`, `openai` | `gemini` |
| `LLM_MODEL` | Model override | Provider default |
| `SEARCH_PROVIDER` | Search provider: `serper`, `serpapi` | `serper` |

### Observability

| Variable | Description |
|----------|-------------|
| `OBSERVABILITY_ENABLED` | Enable LLM observability |
| `OBSERVABILITY_PROVIDER` | Provider: `opik`, `langfuse`, `phoenix` |
| `OPIK_API_KEY` | Opik API key |
| `OPIK_WORKSPACE` | Opik workspace |
| `OPIK_PROJECT` | Opik project name |

## Usage with Deploy Tools

### deploy

```bash
# From your CDK directory - auto-detects project from config.json
cd myproject/cdk
deploy

# Explicit project
deploy --project stats-agent-team

# Use specific env file instead
deploy --env /path/to/.env
```

### push-secrets

```bash
# Auto-detect project and env file
push-secrets --dry-run

# Explicit project
push-secrets --project stats-agent-team

# Use specific env file
push-secrets .env
```

## Security Best Practices

1. **Protect the directory**:
   ```bash
   chmod 700 ~/.agentplexus
   chmod 600 ~/.agentplexus/.env
   chmod 600 ~/.agentplexus/projects/*/.env
   ```

2. **Never commit credentials**: Add to global gitignore:
   ```bash
   echo ".agentplexus" >> ~/.gitignore_global
   git config --global core.excludesfile ~/.gitignore_global
   ```

3. **Use project-specific keys** for production to limit blast radius if keys are compromised.

4. **Rotate keys regularly** and update both local files and AWS Secrets Manager.

## Troubleshooting

### Tool can't find credentials

Check the search order:

```bash
# See what the tool would find
push-secrets --dry-run --verbose
```

### Wrong credentials being used

Use explicit paths:

```bash
deploy --env ~/.agentplexus/projects/myproject/.env
```

### Project not detected correctly

Specify explicitly:

```bash
deploy --project my-agent-team
```

## Related Documentation

- [AWS AgentCore Deployment](../platforms/aws-agentcore.md)
- [agentkit-aws-cdk Repository](https://github.com/agentplexus/agentkit-aws-cdk)
