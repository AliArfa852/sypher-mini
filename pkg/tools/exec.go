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
	allowGitPush        bool
	allowDirs           []string
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
	for _, re := range defaultDenyPatterns {
		// Skip git push/force deny when allow_git_push is enabled
		if cfg.Tools.Exec.AllowGitPush && (re.String() == `\bgit\s+push\b` || re.String() == `\bgit\s+force\b`) {
			continue
		}
		denyPatterns = append(denyPatterns, re)
	}

	allowDirs := make([]string, 0, len(cfg.Tools.Exec.AllowDirs))
	for _, d := range cfg.Tools.Exec.AllowDirs {
		allowDirs = append(allowDirs, config.ExpandPath(d))
	}

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
		allowGitPush:        cfg.Tools.Exec.AllowGitPush,
		allowDirs:           allowDirs,
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
	if len(cmdStr) > 32*1024 {
		return ErrorResponse(req.ToolCallID,
			"Command too long (max 32KB)",
			"Command exceeds maximum length.",
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

	// Workspace restriction (hardened: reject roots, use Rel, validate command paths)
	if t.restrictToWorkspace {
		if errMsg := t.guardWorkspaceAndCommand(workingDir, cmdStr); errMsg != "" {
			return ErrorResponse(req.ToolCallID,
				errMsg,
				"Command blocked: path or working directory outside allowed workspace.",
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

// guardWorkspaceAndCommand validates working directory and paths in the command string.
// Returns non-empty error message if validation fails.
func (t *ExecTool) guardWorkspaceAndCommand(workingDir, cmdStr string) string {
	wsAbs, err := filepath.Abs(t.workingDir)
	if err != nil {
		wsAbs = t.workingDir
	}
	wsClean := filepath.Clean(wsAbs)

	// Reject workspace roots (E:\, C:\, /) - allows access to entire drive/filesystem
	if isPathRoot(wsClean) {
		return "Workspace cannot be a filesystem root (security)"
	}

	abs, err := filepath.Abs(workingDir)
	if err != nil {
		abs = workingDir
	}
	absClean := filepath.Clean(abs)

	// Allow working_dir in tools.exec.allow_dirs
	if t.isInAllowDirs(absClean) {
		// Path is allowed; continue to command path validation
	} else {
		// Validate working directory is within workspace using Rel (handles Windows/Unix)
		rel, err := filepath.Rel(wsClean, absClean)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return "Working directory outside workspace"
		}
	}

	// Path traversal in command string
	if strings.Contains(cmdStr, ".."+string(filepath.Separator)) || strings.Contains(cmdStr, "..\\") {
		return "Command blocked by safety guard (path traversal detected)"
	}

	// Extract and validate absolute paths in command (port from picoclaw guardCommand)
	pathPattern := regexp.MustCompile(`[A-Za-z]:\\[^\\"']+|/[^\s"']+`)
	matches := pathPattern.FindAllString(cmdStr, -1)
	cwdAbs := absClean
	if cwdAbs == "" {
		cwdAbs = wsClean
	}

	for _, raw := range matches {
		// Skip common benign paths
		if raw == "/dev/null" || strings.HasPrefix(raw, "/dev/") {
			continue
		}
		p, err := filepath.Abs(raw)
		if err != nil {
			continue
		}
		relPath, err := filepath.Rel(cwdAbs, p)
		if err != nil {
			continue
		}
		if strings.HasPrefix(relPath, "..") {
			return "Command blocked by safety guard (path outside working dir)"
		}
	}

	return ""
}

// isInAllowDirs returns true if path is within any of the allowed directories.
func (t *ExecTool) isInAllowDirs(absPath string) bool {
	for _, allowed := range t.allowDirs {
		if allowed == "" {
			continue
		}
		allowedAbs, err := filepath.Abs(allowed)
		if err != nil {
			continue
		}
		allowedClean := filepath.Clean(allowedAbs)
		if absPath == allowedClean || strings.HasPrefix(absPath, allowedClean+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// isPathRoot returns true if path is a filesystem root (e.g. /, C:\, E:\).
func isPathRoot(p string) bool {
	p = filepath.Clean(p)
	if p == "" {
		return false
	}
	if runtime.GOOS == "windows" {
		// C:\, D:\, etc.
		if len(p) == 3 && p[1] == ':' && (p[2] == '\\' || p[2] == '/') {
			return true
		}
		return false
	}
	return p == "/"
}
