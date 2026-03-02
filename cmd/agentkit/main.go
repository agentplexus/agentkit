// Package main provides the agentkit CLI tool.
// This CLI supports running agents and generating code from multi-agent-spec.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/plexusone/agentkit/platforms/local/generate"
)

const (
	programName    = "agentkit"
	programVersion = "0.1.0"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "generate", "gen":
		if err := runGenerate(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "run":
		if err := runAgent(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "version", "-v", "--version":
		fmt.Printf("%s version %s\n", programName, programVersion)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1]) //nolint:gosec // G705: CLI output to stderr
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf(`%s - Multi-agent orchestration toolkit

Usage:
  %s <command> [options]

Commands:
  generate    Generate Go code from multi-agent-spec
  run         Run agents directly from spec (interpreted mode)
  version     Show version information
  help        Show this help message

Examples:
  # Generate CLI from spec
  %s generate --spec ./release-agent-team --output ./cmd/release-cli

  # Run workflow directly (interpreted mode)
  %s run --spec ./agent-team-prd --input "Review this feature"

Use "%s <command> --help" for more information about a command.
`, programName, programName, programName, programName, programName)
}

func runGenerate(args []string) error {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)

	var (
		specDir     = fs.String("spec", "", "Path to multi-agent-spec directory (required)")
		outputDir   = fs.String("output", "", "Output directory for generated code (required)")
		programName = fs.String("name", "", "CLI program name (defaults to team name)")
		modulePath  = fs.String("module", "", "Go module path (defaults to generated/<name>)")
	)

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Generate Go code from multi-agent-spec.

Usage:
  agentkit generate [options]

Options:
`)
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  # Generate from release-agent-team
  agentkit generate --spec ./release-agent-team --output ./cmd/release-cli

  # Generate with custom module path
  agentkit generate --spec ./my-team --output ./cmd/my-cli \
    --name my-cli --module github.com/myorg/my-cli

Generated Files:
  main.go         - CLI entry point
  agents.go       - Agent definitions with embedded instructions
  workflow.go     - DAG workflow definition
  tools.go        - Custom tool stubs (editable, not overwritten)
  instructions/   - Embedded instruction markdown files
  go.mod          - Go module file (not overwritten if exists)
`)
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *specDir == "" {
		return fmt.Errorf("--spec is required")
	}
	if *outputDir == "" {
		return fmt.Errorf("--output is required")
	}

	// Resolve paths
	absSpecDir, err := filepath.Abs(*specDir)
	if err != nil {
		return fmt.Errorf("invalid spec path: %w", err)
	}
	absOutputDir, err := filepath.Abs(*outputDir)
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}

	// Default program name from spec directory
	name := *programName
	if name == "" {
		name = filepath.Base(absSpecDir)
	}

	generator := &generate.Generator{
		SpecDir:     absSpecDir,
		OutputDir:   absOutputDir,
		ProgramName: name,
		ModulePath:  *modulePath,
	}

	fmt.Printf("Generating code from %s...\n", absSpecDir)
	if err := generator.Generate(); err != nil {
		return err
	}

	fmt.Printf("Generated code in %s\n", absOutputDir)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  cd %s\n", absOutputDir)
	fmt.Printf("  go mod tidy\n")
	fmt.Printf("  go build -o %s\n", name)
	fmt.Printf("  ./%s --help\n", name)

	return nil
}

func runAgent(args []string) error {
	fs := flag.NewFlagSet("run", flag.ExitOnError)

	var (
		specDir   = fs.String("spec", "", "Path to multi-agent-spec directory (required)")
		agent     = fs.String("agent", "", "Run a specific agent by name")
		input     = fs.String("input", "", "Input for the agent/workflow")
		inputFile = fs.String("input-file", "", "Read input from file")
		provider  = fs.String("provider", "anthropic", "LLM provider")
		model     = fs.String("model", "sonnet", "Model name or tier")
		workspace = fs.String("workspace", ".", "Workspace directory")
	)

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Run agents directly from spec (interpreted mode).

Usage:
  agentkit run [options]

Options:
`)
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  # Run entire workflow
  agentkit run --spec ./agent-team-prd --input "Review this PRD"

  # Run specific agent
  agentkit run --spec ./agent-team-prd --agent pm --input "Validate requirements"
`)
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *specDir == "" {
		return fmt.Errorf("--spec is required")
	}

	// Get input
	inputText := *input
	if inputText == "" && *inputFile != "" {
		data, err := os.ReadFile(*inputFile)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}
		inputText = string(data)
	}

	if inputText == "" {
		return fmt.Errorf("--input or --input-file is required")
	}

	// This would call into the runtime execution path
	// For now, suggest using the generated CLI
	fmt.Printf("Running from spec: %s\n", *specDir)
	fmt.Printf("Provider: %s, Model: %s\n", *provider, *model)
	fmt.Printf("Workspace: %s\n", *workspace)
	if *agent != "" {
		fmt.Printf("Agent: %s\n", *agent)
	}
	fmt.Printf("Input: %s\n", inputText)

	fmt.Println("\nNote: For production use, generate a CLI with 'agentkit generate'")

	return nil
}
