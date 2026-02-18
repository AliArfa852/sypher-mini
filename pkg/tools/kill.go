package tools

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/sypherexx/sypher-mini/pkg/process"
)

// KillTool kills processes started by Sypher-mini for the current task.
type KillTool struct {
	procTracker *process.Tracker
	safeMode    bool
}

// NewKillTool creates a kill tool.
func NewKillTool(procTracker *process.Tracker, safeMode bool) *KillTool {
	return &KillTool{
		procTracker: procTracker,
		safeMode:    safeMode,
	}
}

// Execute kills a process if it belongs to the task.
func (t *KillTool) Execute(ctx context.Context, req Request) Response {
	if t.safeMode {
		return ErrorResponse(req.ToolCallID,
			"Kill disabled in safe mode",
			"Process killing is disabled in safe mode.",
			CodePermissionDenied, false)
	}

	if t.procTracker == nil {
		return ErrorResponse(req.ToolCallID,
			"Process tracker not configured",
			"Cannot kill: process tracker not available.",
			CodePermissionDenied, false)
	}

	pidVal, ok := req.Args["pid"]
	if !ok {
		return ErrorResponse(req.ToolCallID,
			"Missing 'pid' argument",
			"PID is required.",
			CodePermissionDenied, false)
	}

	var pid int
	switch v := pidVal.(type) {
	case float64:
		pid = int(v)
	case int:
		pid = v
	case string:
		var err error
		pid, err = strconv.Atoi(v)
		if err != nil {
			return ErrorResponse(req.ToolCallID,
				fmt.Sprintf("Invalid pid: %v", v),
				"Invalid PID.",
				CodePermissionDenied, false)
		}
	default:
		return ErrorResponse(req.ToolCallID,
			"Invalid pid type",
			"PID must be a number.",
			CodePermissionDenied, false)
	}

	if !t.procTracker.CanKill(req.TaskID, pid) {
		return ErrorResponse(req.ToolCallID,
			"PID not owned by this task - cannot kill",
			"Process not owned by Sypher-mini for this task.",
			CodePermissionDenied, false)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return ErrorResponse(req.ToolCallID,
			fmt.Sprintf("FindProcess: %v", err),
			"Could not find process.",
			CodePermissionDenied, false)
	}

	if err := proc.Kill(); err != nil {
		return ErrorResponse(req.ToolCallID,
			fmt.Sprintf("Kill failed: %v", err),
			"Kill failed.",
			CodePermissionDenied, false)
	}

	return SuccessResponse(req.ToolCallID,
		fmt.Sprintf("Process %d killed", pid),
		fmt.Sprintf("Process %d killed.", pid),
		"")
}
