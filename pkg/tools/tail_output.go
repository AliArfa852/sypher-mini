package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sypherexx/sypher-mini/pkg/config"
)

// TailOutputTool reads the last N lines from a file.
type TailOutputTool struct {
	workspace          string
	restrictToWorkspace bool
	safeMode           bool
}

// NewTailOutputTool creates a tail_output tool.
func NewTailOutputTool(cfg *config.Config, safeMode bool) *TailOutputTool {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	workspace := config.ExpandPath(cfg.Agents.Defaults.Workspace)
	if workspace == "" {
		workspace, _ = os.Getwd()
	}
	return &TailOutputTool{
		workspace:           workspace,
		restrictToWorkspace: cfg.Agents.Defaults.RestrictToWorkspace,
		safeMode:            safeMode,
	}
}

// Execute reads the last N lines from the given file path.
func (t *TailOutputTool) Execute(ctx context.Context, req Request) Response {
	if t.safeMode {
		return ErrorResponse(req.ToolCallID,
			"tail_output disabled in safe mode",
			"Tail output is disabled in safe mode.",
			CodePermissionDenied, false)
	}

	path, _ := req.Args["path"].(string)
	if path == "" {
		return ErrorResponse(req.ToolCallID,
			"Missing 'path' argument",
			"Path is required.",
			CodePermissionDenied, false)
	}

	n := 50
	if v, ok := req.Args["lines"].(float64); ok && v > 0 && v <= 1000 {
		n = int(v)
	}

	path = config.ExpandPath(path)
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}

	if t.restrictToWorkspace {
		wsAbs, _ := filepath.Abs(t.workspace)
		if !strings.HasPrefix(abs, wsAbs) {
			return ErrorResponse(req.ToolCallID,
				"Path outside workspace",
				"File path is outside the allowed workspace.",
				CodePermissionDenied, false)
		}
	}

	f, err := os.Open(path)
	if err != nil {
		return ErrorResponse(req.ToolCallID,
			fmt.Sprintf("Failed to open file: %v", err),
			"Could not read file.",
			CodePermissionDenied, false)
	}
	defer f.Close()

	// Read all lines, keep last N
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return ErrorResponse(req.ToolCallID,
			fmt.Sprintf("Read error: %v", err),
			"Could not read file.",
			CodePermissionDenied, false)
	}

	start := 0
	if len(lines) > n {
		start = len(lines) - n
	}
	result := strings.Join(lines[start:], "\n")
	if len(result) > 8192 {
		result = result[len(result)-8192:] + "\n\n... (truncated)"
	}

	return SuccessResponse(req.ToolCallID,
		result,
		fmt.Sprintf("Last %d lines from %s", len(lines[start:]), path),
		"")
}
