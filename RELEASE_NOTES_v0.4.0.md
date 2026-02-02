# Release Notes: v0.4.0

**Release Date:** 2026-02-02

## Highlights

This release adds two major features: a Model Context Protocol (MCP) server for exposing agent tools to AI assistants like Claude Code, and a local development platform for CLI-based agent testing without cloud infrastructure.

## New Features

### MCP Server

The new `mcp` package provides a Model Context Protocol server implementation for exposing agent tools to AI assistants:

```go
import "github.com/agentplexus/agentkit/mcp"

// Create MCP server with tools
server := mcp.NewServer()
server.RegisterTool("search", searchHandler)
server.RegisterTool("calculate", calcHandler)

// Run with stdio transport
server.ServeStdio(ctx)
```

Features:

- JSON-RPC 2.0 protocol with stdio transport
- Tool listing via `tools/list` method
- Tool execution via `tools/call` method
- Session lifecycle management (initialize, initialized, ping)
- Graceful shutdown handling
- Extensible tool registration

### Local Development Platform

The new `platforms/local` package enables running agents locally for development and testing:

```go
import "github.com/agentplexus/agentkit/platforms/local"

// Create local runner
runner := local.NewRunner(config)

// Run interactive CLI session
runner.Run(ctx)
```

Features:

- Interactive chat with agents via CLI
- Built-in file system tools (read, write, list)
- Bash command execution
- Conversation history management
- No cloud infrastructure required

## API Changes

### Changed

- LLM integration updated for new OmniLLM and OmniObserve APIs

## Dependencies

| Package | Previous | Current |
|---------|----------|---------|
| `google.golang.org/adk` | 0.3.0 | 0.4.0 |
| `github.com/cloudwego/eino` | 0.7.17 | 0.7.29 |
| `google.golang.org/genai` | 1.42.0 | 1.44.0 |
| `github.com/a2aproject/a2a-go` | 0.3.4 | 0.3.6 |
| `github.com/agentplexus/omnivault` | 0.2.0 | 0.2.1 |
| `github.com/agentplexus/omniobserve` | 0.5.0 | 0.5.1 |

## Installation

```bash
go get github.com/agentplexus/agentkit@v0.4.0
```

## Full Changelog

See [CHANGELOG.md](CHANGELOG.md) for the complete list of changes.
