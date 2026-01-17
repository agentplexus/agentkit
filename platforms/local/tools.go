package local

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Tool represents a capability available to agents.
type Tool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, args map[string]any) (any, error)
}

// ToolSet provides filesystem and shell tools scoped to a workspace.
type ToolSet struct {
	workspace   string
	maxFileSize int64
}

// NewToolSet creates a new tool set for the given workspace.
func NewToolSet(workspace string) *ToolSet {
	return &ToolSet{
		workspace:   workspace,
		maxFileSize: 10 * 1024 * 1024, // 10MB default
	}
}

// SetMaxFileSize sets the maximum file size for read operations.
func (ts *ToolSet) SetMaxFileSize(size int64) {
	ts.maxFileSize = size
}

// validatePath ensures a path is within the workspace.
func (ts *ToolSet) validatePath(path string) (string, error) {
	// Handle relative paths
	if !filepath.IsAbs(path) {
		path = filepath.Join(ts.workspace, path)
	}

	// Clean and resolve the path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// Ensure path is within workspace
	relPath, err := filepath.Rel(ts.workspace, absPath)
	if err != nil {
		return "", fmt.Errorf("path outside workspace: %w", err)
	}
	if strings.HasPrefix(relPath, "..") {
		return "", fmt.Errorf("path outside workspace: %s", path)
	}

	return absPath, nil
}

// ReadFile reads the contents of a file within the workspace.
func (ts *ToolSet) ReadFile(ctx context.Context, path string) (string, error) {
	absPath, err := ts.validatePath(path)
	if err != nil {
		return "", err
	}

	// Check file size
	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("cannot access file: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("path is a directory: %s", path)
	}
	if info.Size() > ts.maxFileSize {
		return "", fmt.Errorf("file too large: %d bytes (max %d)", info.Size(), ts.maxFileSize)
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

// WriteFile writes content to a file within the workspace.
func (ts *ToolSet) WriteFile(ctx context.Context, path, content string) error {
	absPath, err := ts.validatePath(path)
	if err != nil {
		return err
	}

	// Create parent directories if needed
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(absPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// GlobFiles finds files matching a glob pattern within the workspace.
func (ts *ToolSet) GlobFiles(ctx context.Context, pattern string) ([]string, error) {
	// Handle relative patterns
	if !filepath.IsAbs(pattern) {
		pattern = filepath.Join(ts.workspace, pattern)
	}

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid glob pattern: %w", err)
	}

	// Filter to workspace and convert to relative paths
	var result []string
	for _, match := range matches {
		relPath, err := filepath.Rel(ts.workspace, match)
		if err != nil || strings.HasPrefix(relPath, "..") {
			continue // Skip paths outside workspace
		}
		result = append(result, relPath)
	}

	return result, nil
}

// GrepFiles searches for a pattern in files within the workspace.
func (ts *ToolSet) GrepFiles(ctx context.Context, pattern, filePattern string) ([]GrepMatch, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	var matches []GrepMatch

	err = filepath.WalkDir(ts.workspace, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip directories and hidden files
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") && d.Name() != "." {
				return filepath.SkipDir
			}
			return nil
		}

		// Check file pattern if specified
		if filePattern != "" {
			matched, _ := filepath.Match(filePattern, d.Name())
			if !matched {
				return nil
			}
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip unreadable files
		}

		// Search for matches
		lines := strings.Split(string(content), "\n")
		relPath, _ := filepath.Rel(ts.workspace, path)

		for lineNum, line := range lines {
			if regex.MatchString(line) {
				matches = append(matches, GrepMatch{
					File:    relPath,
					Line:    lineNum + 1,
					Content: strings.TrimSpace(line),
				})
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	return matches, nil
}

// GrepMatch represents a single grep match.
type GrepMatch struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Content string `json:"content"`
}

// RunCommand executes a shell command within the workspace.
func (ts *ToolSet) RunCommand(ctx context.Context, command string, args []string) (*CommandResult, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = ts.workspace

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &CommandResult{
		Command:  command,
		Args:     args,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: 0,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("command execution failed: %w", err)
		}
	}

	return result, nil
}

// RunShell executes a shell command string within the workspace.
func (ts *ToolSet) RunShell(ctx context.Context, shellCmd string) (*CommandResult, error) {
	// Use sh -c for shell command execution
	return ts.RunCommand(ctx, "sh", []string{"-c", shellCmd})
}

// CommandResult holds the result of a command execution.
type CommandResult struct {
	Command  string   `json:"command"`
	Args     []string `json:"args,omitempty"`
	Stdout   string   `json:"stdout"`
	Stderr   string   `json:"stderr"`
	ExitCode int      `json:"exit_code"`
}

// Success returns true if the command exited successfully.
func (r *CommandResult) Success() bool {
	return r.ExitCode == 0
}

// ListDirectory lists the contents of a directory within the workspace.
func (ts *ToolSet) ListDirectory(ctx context.Context, path string) ([]FileInfo, error) {
	absPath, err := ts.validatePath(path)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, FileInfo{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
			Size:  info.Size(),
		})
	}

	return files, nil
}

// FileInfo holds basic file information.
type FileInfo struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
}

// ReadTool wraps ReadFile as a Tool interface.
type ReadTool struct {
	ts *ToolSet
}

func (t *ReadTool) Name() string        { return "read" }
func (t *ReadTool) Description() string { return "Read the contents of a file" }
func (t *ReadTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path argument required")
	}
	return t.ts.ReadFile(ctx, path)
}

// WriteTool wraps WriteFile as a Tool interface.
type WriteTool struct {
	ts *ToolSet
}

func (t *WriteTool) Name() string        { return "write" }
func (t *WriteTool) Description() string { return "Write content to a file" }
func (t *WriteTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path argument required")
	}
	content, ok := args["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content argument required")
	}
	return nil, t.ts.WriteFile(ctx, path, content)
}

// GlobTool wraps GlobFiles as a Tool interface.
type GlobTool struct {
	ts *ToolSet
}

func (t *GlobTool) Name() string        { return "glob" }
func (t *GlobTool) Description() string { return "Find files matching a glob pattern" }
func (t *GlobTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	pattern, ok := args["pattern"].(string)
	if !ok {
		return nil, fmt.Errorf("pattern argument required")
	}
	return t.ts.GlobFiles(ctx, pattern)
}

// GrepTool wraps GrepFiles as a Tool interface.
type GrepTool struct {
	ts *ToolSet
}

func (t *GrepTool) Name() string        { return "grep" }
func (t *GrepTool) Description() string { return "Search for a pattern in files" }
func (t *GrepTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	pattern, ok := args["pattern"].(string)
	if !ok {
		return nil, fmt.Errorf("pattern argument required")
	}
	filePattern, _ := args["file_pattern"].(string)
	return t.ts.GrepFiles(ctx, pattern, filePattern)
}

// ShellTool wraps RunShell as a Tool interface.
type ShellTool struct {
	ts *ToolSet
}

func (t *ShellTool) Name() string        { return "shell" }
func (t *ShellTool) Description() string { return "Execute a shell command" }
func (t *ShellTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	command, ok := args["command"].(string)
	if !ok {
		return nil, fmt.Errorf("command argument required")
	}
	return t.ts.RunShell(ctx, command)
}

// CreateTools creates Tool instances for the specified tool names.
func (ts *ToolSet) CreateTools(names []string) ([]Tool, error) {
	var tools []Tool
	for _, name := range names {
		switch name {
		case "read":
			tools = append(tools, &ReadTool{ts: ts})
		case "write":
			tools = append(tools, &WriteTool{ts: ts})
		case "glob":
			tools = append(tools, &GlobTool{ts: ts})
		case "grep":
			tools = append(tools, &GrepTool{ts: ts})
		case "shell":
			tools = append(tools, &ShellTool{ts: ts})
		default:
			return nil, fmt.Errorf("unknown tool: %s", name)
		}
	}
	return tools, nil
}
