package tools

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/sypherexx/sypher-mini/pkg/audit"
	"github.com/sypherexx/sypher-mini/pkg/config"
	"github.com/sypherexx/sypher-mini/pkg/process"
)

var defaultDenyPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\brm\s+-[rf]{1,2}\b`),
	regexp.MustCompile(`\bdel\s+/[fq]\b`),
	regexp.MustCompile(`\brmdir\s+/s\b`),
	regexp.MustCompile(`\b(format|mkfs|diskpart)\b\s`),
	regexp.MustCompile(`\bdd\s+if=`),
	regexp.MustCompile(`>\s*/dev/sd[a-z]\b`),
	regexp.MustCompile(`\b(shutdown|reboot|poweroff)\b`),
	regexp.MustCompile(`:\(\)\s*\{.*\};\s*:`),
	regexp.MustCompile(`\$\([^)]+\)`),
	regexp.MustCompile(`\$\{[^}]+\}`),
	regexp.MustCompile("`[^`]+`"),
	regexp.MustCompile(`\|\s*sh\b`),
	regexp.MustCompile(`\|\s*bash\b`),
	regexp.MustCompile(`;\s*rm\s+-[rf]`),
	regexp.MustCompile(`&&\s*rm\s+-[rf]`),
	regexp.MustCompile(`\|\|\s*rm\s+-[rf]`),
	regexp.MustCompile(`>\s*/dev/null\s*>&?\s*\d?`),
	regexp.MustCompile(`<<\s*EOF`),
	regexp.MustCompile(`\$\(\s*cat\s+`),
	regexp.MustCompile(`\$\(\s*curl\s+`),
	regexp.MustCompile(`\$\(\s*wget\s+`),
	regexp.MustCompile(`\$\(\s*which\s+`),
	regexp.MustCompile(`\bsudo\b`),
	regexp.MustCompile(`\bchmod\s+[0-7]{3,4}\b`),
	regexp.MustCompile(`\bchown\b`),
	regexp.MustCompile(`\bpkill\b`),
	regexp.MustCompile(`\bkillall\b`),
	regexp.MustCompile(`\bkill\s+-[9]\b`),
	regexp.MustCompile(`\bcurl\b.*\|\s*(sh|bash)`),
	regexp.MustCompile(`\bwget\b.*\|\s*(sh|bash)`),
	regexp.MustCompile(`\bnpm\s+install\s+-g\b`),
	regexp.MustCompile(`\bpip\s+install\s+--user\b`),
	regexp.MustCompile(`\bapt\s+(install|remove|purge)\b`),
	regexp.MustCompile(`\byum\s+(install|remove)\b`),
	regexp.MustCompile(`\bdnf\s+(install|remove)\b`),
	regexp.MustCompile(`\bdocker\s+run\b`),
	regexp.MustCompile(`\bdocker\s+exec\b`),
	regexp.MustCompile(`\bgit\s+push\b`),
	regexp.MustCompile(`\bgit\s+force\b`),
	regexp.MustCompile(`\bssh\b.*@`),
	regexp.MustCompile(`\beval\b`),
	regexp.MustCompile(`\bsource\s+.*\.sh\b`),
}

// ExecTool executes shell commands with safety checks.
type ExecTool struct {
	workingDir          string
	timeout             time.Duration
	denyPatterns        []*regexp.Regexp
	restrictToWorkspace bool
	auditLogger         *audit.Logger
	procTracker         *process.Tracker
	authorizedTerms     []string
	safeMode            bool
}

// NewExecTool creates an exec tool.
func NewExecTool(cfg *config.Config, auditLogger *audit.Logger, procTracker *process.Tracker, safeMode bool) *ExecTool {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	denyPatterns := make([]*regexp.Regexp, 0, len(defaultDenyPatterns)+10)
	if len(cfg.Tools.Exec.CustomDenyPatterns) > 0 {
		for _, p := range cfg.Tools.Exec.CustomDenyPatterns {
			if re, err := regexp.Compile(p); err == nil {
				denyPatterns = append(denyPatterns, re)
			}
		}
	}
	denyPatterns = append(denyPatterns, defaultDenyPatterns...)

	timeout := 60 * time.Second
	if cfg != nil && cfg.Tools.Exec.TimeoutSec > 0 {
		timeout = time.Duration(cfg.Tools.Exec.TimeoutSec) * time.Second
	}

	workspace := config.ExpandPath(cfg.Agents.Defaults.Workspace)
	if workspace == "" {
		workspace, _ = os.Getwd()
	}
	terms := cfg.AuthorizedTerminals
	if len(terms) == 0 {
		terms = []string{"default"}
	}

	return &ExecTool{
		workingDir:          workspace,
		timeout:             timeout,
		denyPatterns:        denyPatterns,
		restrictToWorkspace: cfg.Agents.Defaults.RestrictToWorkspace,
		auditLogger:         auditLogger,
		procTracker:         procTracker,
		authorizedTerms:     terms,
		safeMode:            safeMode,
	}
}

// Execute runs a command and returns a tool response.
func (t *ExecTool) Execute(ctx context.Context, req Request) Response {
	if t.safeMode {
		return ErrorResponse(req.ToolCallID,
			"Exec disabled in safe mode",
			"Command execution is disabled in safe mode.",
			CodePermissionDenied, false)
	}

	cmdStr, _ := req.Args["command"].(string)
	if cmdStr == "" {
		return ErrorResponse(req.ToolCallID,
			"Missing 'command' argument",
			"Command is required.",
			CodePermissionDenied, false)
	}

	workingDir, _ := req.Args["working_dir"].(string)
	if workingDir == "" {
		workingDir = t.workingDir
	}
	workingDir = config.ExpandPath(workingDir)

	// Deny pattern check
	for _, re := range t.denyPatterns {
		if re.MatchString(cmdStr) {
			return ErrorResponse(req.ToolCallID,
				"Command blocked by safety guard (dangerous pattern detected)",
				"Command was blocked for safety.",
				CodeSafetyBlocked, false)
		}
	}

	// Workspace restriction
	if t.restrictToWorkspace {
		abs, err := filepath.Abs(workingDir)
		if err != nil {
			abs = workingDir
		}
		wsAbs, _ := filepath.Abs(t.workingDir)
		if !strings.HasPrefix(abs, wsAbs) {
			return ErrorResponse(req.ToolCallID,
				"Working directory outside workspace",
				"Command blocked: working directory outside allowed workspace.",
				CodePermissionDenied, false)
		}
	}

	// Build command with timeout
	runCtx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(runCtx, "cmd", "/c", cmdStr)
	} else {
		cmd = exec.CommandContext(runCtx, "sh", "-c", cmdStr)
	}
	cmd.Dir = workingDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Start()
	if err != nil {
		return ErrorResponse(req.ToolCallID,
			fmt.Sprintf("Failed to start: %v", err),
			"Command failed to start.",
			CodePermissionDenied, false)
	}

	// Record PID
	pid := cmd.Process.Pid
	if t.procTracker != nil {
		t.procTracker.Record(req.TaskID, pid)
	}

	err = cmd.Wait()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			if runCtx.Err() == context.DeadlineExceeded {
				return ErrorResponse(req.ToolCallID,
					"Command timed out",
					"Command timed out.",
					CodeTimeout, true)
			}
			exitCode = -1
		}
	}

	// Audit log
	out := stdout.String() + stderr.String()
	if t.auditLogger != nil {
		_ = t.auditLogger.LogCommand(req.TaskID, req.ToolCallID, cmdStr, workingDir, exitCode, out)
	}

	auditRef := fmt.Sprintf("audit/%s.log", req.TaskID)
	forLLM := out
	if len(forLLM) > 4096 {
		forLLM = forLLM[:4096] + "\n\n... (truncated)"
	}
	forUser := fmt.Sprintf("Exit code: %d", exitCode)
	if exitCode != 0 {
		forUser = fmt.Sprintf("Command failed (exit %d)", exitCode)
	}

	return SuccessResponse(req.ToolCallID, forLLM, forUser, auditRef)
}
