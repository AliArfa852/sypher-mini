package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/sypherexx/sypher-mini/pkg/config"
)

// InvokeCliAgentTool runs a configured CLI agent (e.g. gemini) with a task.
type InvokeCliAgentTool struct {
	cfg       *config.Config
	workspace string
	safeMode  bool
}

// NewInvokeCliAgentTool creates an invoke_cli_agent tool.
func NewInvokeCliAgentTool(cfg *config.Config, safeMode bool) *InvokeCliAgentTool {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	workspace := config.ExpandPath(cfg.Agents.Defaults.Workspace)
	if workspace == "" {
		workspace, _ = filepath.Abs(".")
	}
	return &InvokeCliAgentTool{
		cfg:       cfg,
		workspace: workspace,
		safeMode:  safeMode,
	}
}

// Execute runs the CLI agent with the given task.
func (t *InvokeCliAgentTool) Execute(ctx context.Context, req Request) Response {
	if t.safeMode {
		return ErrorResponse(req.ToolCallID,
			"invoke_cli_agent disabled in safe mode",
			"CLI agent invocation is disabled in safe mode.",
			CodePermissionDenied, false)
	}

	taskStr, _ := req.Args["task"].(string)
	if taskStr == "" {
		return ErrorResponse(req.ToolCallID,
			"Missing 'task' argument",
			"Task is required.",
			CodePermissionDenied, false)
	}

	agentID, _ := req.Args["agent_id"].(string)
	workingDir, _ := req.Args["working_dir"].(string)
	if workingDir == "" {
		workingDir = t.workspace
	}
	workingDir = config.ExpandPath(workingDir)

	command, args := t.resolveCliAgent(agentID)
	if command == "" {
		return ErrorResponse(req.ToolCallID,
			"No CLI agent configured",
			"No agent with command/args found. Add an agent with command (e.g. gemini) and args in config.",
			CodePermissionDenied, false)
	}

	// Build args: existing args + task as final arg (gemini takes prompt as arg)
	allArgs := make([]string, 0, len(args)+1)
	allArgs = append(allArgs, args...)
	allArgs = append(allArgs, taskStr)

	runCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	cmd := exec.CommandContext(runCtx, command, allArgs...)
	cmd.Dir = workingDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	out := stdout.String() + stderr.String()
	if err != nil {
		if runCtx.Err() == context.DeadlineExceeded {
			return ErrorResponse(req.ToolCallID,
				"CLI agent timed out",
				"Command timed out.",
				CodeTimeout, true)
		}
		out += fmt.Sprintf("\nExit code: %v", err)
		return SuccessResponse(req.ToolCallID, out, fmt.Sprintf("CLI agent exited with error: %v", err), "")
	}

	if len(out) > 8192 {
		out = out[:8192] + "\n\n... (truncated)"
	}
	return SuccessResponse(req.ToolCallID, out, "CLI agent completed", "")
}

// resolveCliAgent returns command and args for the given agent_id, or first agent with command/args.
func (t *InvokeCliAgentTool) resolveCliAgent(agentID string) (string, []string) {
	for _, a := range t.cfg.Agents.List {
		if agentID != "" && a.ID != agentID {
			continue
		}
		if a.Command != "" {
			args := a.Args
			if args == nil {
				args = []string{}
			}
			return a.Command, args
		}
	}
	return "", nil
}
