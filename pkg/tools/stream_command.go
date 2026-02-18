package tools

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/sypherexx/sypher-mini/pkg/bus"
	"github.com/sypherexx/sypher-mini/pkg/config"
)

// StreamCommandTool runs a command and streams output to the user via the message bus.
type StreamCommandTool struct {
	msgBus             *bus.MessageBus
	messageTool        *MessageTool
	workspace          string
	restrictToWorkspace bool
	allowedCommands    []string // allowlist; empty = none allowed
	safeMode           bool
}

// NewStreamCommandTool creates a stream_command tool.
func NewStreamCommandTool(cfg *config.Config, msgBus *bus.MessageBus, messageTool *MessageTool, safeMode bool) *StreamCommandTool {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	workspace := config.ExpandPath(cfg.Agents.Defaults.Workspace)
	if workspace == "" {
		workspace, _ = filepath.Abs(".")
	}
	var allowed []string
	if cfg.Tools.LiveMonitoring.AllowedCommands != nil {
		allowed = cfg.Tools.LiveMonitoring.AllowedCommands
	}
	return &StreamCommandTool{
		msgBus:             msgBus,
		messageTool:       messageTool,
		workspace:         workspace,
		restrictToWorkspace: cfg.Agents.Defaults.RestrictToWorkspace,
		allowedCommands:   allowed,
		safeMode:          safeMode,
	}
}

// Execute runs the command and streams output to the user.
func (t *StreamCommandTool) Execute(ctx context.Context, req Request) Response {
	if t.safeMode {
		return ErrorResponse(req.ToolCallID,
			"stream_command disabled in safe mode",
			"Stream command is disabled in safe mode.",
			CodePermissionDenied, false)
	}

	cmdStr, _ := req.Args["command"].(string)
	if cmdStr == "" {
		return ErrorResponse(req.ToolCallID,
			"Missing 'command' argument",
			"Command is required.",
			CodePermissionDenied, false)
	}

	// Check allowlist
	if !t.isCommandAllowed(cmdStr) {
		return ErrorResponse(req.ToolCallID,
			"Command not in live_monitoring allowed list",
			"Command is not allowed for live streaming.",
			CodePermissionDenied, false)
	}

	// Basic deny patterns (subset of exec)
	deny := []*regexp.Regexp{
		regexp.MustCompile(`\brm\s+-[rf]{1,2}\b`),
		regexp.MustCompile(`\bsudo\b`),
		regexp.MustCompile(`\|\s*sh\b`),
		regexp.MustCompile(`\|\s*bash\b`),
	}
	for _, re := range deny {
		if re.MatchString(cmdStr) {
			return ErrorResponse(req.ToolCallID,
				"Command blocked by safety guard",
				"Command was blocked for safety.",
				CodeSafetyBlocked, false)
		}
	}

	workingDir, _ := req.Args["working_dir"].(string)
	if workingDir == "" {
		workingDir = t.workspace
	}
	workingDir = config.ExpandPath(workingDir)

	if t.restrictToWorkspace {
		abs, _ := filepath.Abs(workingDir)
		wsAbs, _ := filepath.Abs(t.workspace)
		if !strings.HasPrefix(abs, wsAbs) {
			return ErrorResponse(req.ToolCallID,
				"Working directory outside workspace",
				"Command blocked: working directory outside allowed workspace.",
				CodePermissionDenied, false)
		}
	}

	channel, chatID := t.messageTool.GetReplyTarget(req.TaskID)
	if channel == "" {
		channel = "cli"
	}
	if chatID == "" {
		chatID = "default"
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/c", cmdStr)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", cmdStr)
	}
	cmd.Dir = workingDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return ErrorResponse(req.ToolCallID,
			fmt.Sprintf("Failed to get stdout: %v", err),
			"Command failed to start.",
			CodePermissionDenied, false)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return ErrorResponse(req.ToolCallID,
			fmt.Sprintf("Failed to get stderr: %v", err),
			"Command failed to start.",
			CodePermissionDenied, false)
	}

	if err := cmd.Start(); err != nil {
		return ErrorResponse(req.ToolCallID,
			fmt.Sprintf("Failed to start: %v", err),
			"Command failed to start.",
			CodePermissionDenied, false)
	}

	var buf bytes.Buffer
	var mu sync.Mutex
	sendChunk := func(text string) {
		if text == "" {
			return
		}
		t.msgBus.PublishOutbound(bus.OutboundMessage{
			Channel: channel,
			ChatID:  chatID,
			Content: text,
		})
	}

	// Stream stdout
	go func() {
		sc := bufio.NewScanner(stdout)
		for sc.Scan() {
			line := sc.Text() + "\n"
			mu.Lock()
			buf.WriteString(line)
			mu.Unlock()
			sendChunk(line)
		}
	}()

	// Stream stderr
	go func() {
		sc := bufio.NewScanner(stderr)
		for sc.Scan() {
			line := "[stderr] " + sc.Text() + "\n"
			mu.Lock()
			buf.WriteString(line)
			mu.Unlock()
			sendChunk(line)
		}
	}()

	err = cmd.Wait()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			if ctx.Err() != nil {
				sendChunk("\n[Command cancelled]")
				return ErrorResponse(req.ToolCallID,
					"Command cancelled",
					"Command was cancelled.",
					CodeTimeout, true)
			}
			exitCode = -1
		}
	}

	// Brief pause so last chunks arrive before summary
	time.Sleep(50 * time.Millisecond)

	summary := fmt.Sprintf("Stream completed (exit %d)", exitCode)
	if exitCode != 0 {
		summary = fmt.Sprintf("Stream failed (exit %d)", exitCode)
	}
	out := buf.String()
	if len(out) > 4096 {
		out = out[len(out)-4096:] + "\n\n... (truncated)"
	}

	return SuccessResponse(req.ToolCallID, out, summary, "")
}

func (t *StreamCommandTool) isCommandAllowed(cmd string) bool {
	if len(t.allowedCommands) == 0 {
		return false
	}
	cmd = strings.TrimSpace(cmd)
	for _, allowed := range t.allowedCommands {
		if allowed == "*" {
			return true
		}
		if strings.HasPrefix(cmd, allowed) {
			return true
		}
	}
	return false
}
