package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

var (
	builtBinOnce sync.Once
	builtBinPath string
)

func getBuiltBin(t *testing.T) string {
	t.Helper()
	builtBinOnce.Do(func() {
		tmp, err := os.MkdirTemp("", "sypher-cli-test-*")
		if err != nil {
			t.Fatalf("temp dir: %v", err)
		}
		bin := filepath.Join(tmp, "sypher-test")
		if runtime.GOOS == "windows" {
			bin += ".exe"
		}
		build := exec.Command("go", "build", "-o", bin, "github.com/sypherexx/sypher-mini/cmd/sypher")
		if err := build.Run(); err != nil {
			t.Fatalf("build failed: %v", err)
		}
		builtBinPath = bin
	})
	return builtBinPath
}

// runSypher runs the sypher binary with args. Uses a fresh temp dir for HOME.
func runSypher(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	return runSypherWithHome(t, t.TempDir(), args...)
}

// runSypherWithHome runs the sypher binary with the given HOME (for tests that need shared config).
func runSypherWithHome(t *testing.T, home string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	env := append(os.Environ(), "HOME="+home, "USERPROFILE="+home)
	bin := getBuiltBin(t)
	cmd := exec.Command(bin, args...)
	cmd.Env = env
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			t.Fatalf("run failed: %v", err)
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}

func TestCLI_Version(t *testing.T) {
	out, _, code := runSypher(t, "version")
	if code != 0 {
		t.Errorf("version exit code = %d", code)
	}
	if !strings.Contains(out, "sypher-mini") {
		t.Errorf("version output: %s", out)
	}
}

func TestCLI_VersionAliases(t *testing.T) {
	for _, arg := range []string{"-v", "--version"} {
		out, _, code := runSypher(t, arg)
		if code != 0 {
			t.Errorf("%s exit code = %d", arg, code)
		}
		if !strings.Contains(out, "sypher-mini") {
			t.Errorf("%s output: %s", arg, out)
		}
	}
}

func TestCLI_UnknownCommand(t *testing.T) {
	stdout, stderr, code := runSypher(t, "unknown")
	if code == 0 {
		t.Error("unknown command should exit non-zero")
	}
	combined := stdout + stderr
	if !strings.Contains(combined, "Unknown command") {
		t.Errorf("expected 'Unknown command' in output; stdout: %q stderr: %q", stdout, stderr)
	}
}

func TestCLI_NoArgs(t *testing.T) {
	tmp := t.TempDir()
	env := append(os.Environ(), "HOME="+tmp, "USERPROFILE="+tmp)
	bin := getBuiltBin(t)
	cmd := exec.Command(bin)
	cmd.Env = env
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	cmd.Run()
	if !strings.Contains(out.String(), "Usage") {
		t.Errorf("no args should print help: %s", out.String())
	}
}

func TestCLI_Onboard(t *testing.T) {
	out, stderr, code := runSypher(t, "onboard")
	if code != 0 {
		t.Fatalf("onboard failed: %s %s", out, stderr)
	}
	if !strings.Contains(out, "Onboard complete") && !strings.Contains(stderr, "Onboard complete") {
		t.Errorf("onboard output: %s %s", out, stderr)
	}
}

func TestCLI_Status(t *testing.T) {
	home := t.TempDir()
	runSypherWithHome(t, home, "onboard")
	out, _, code := runSypherWithHome(t, home, "status")
	if code != 0 {
		t.Errorf("status exit code = %d", code)
	}
	if !strings.Contains(out, "Sypher-mini status") {
		t.Errorf("status output: %s", out)
	}
	if !strings.Contains(out, "Config:") {
		t.Errorf("status should show config: %s", out)
	}
}

func TestCLI_ConfigGet(t *testing.T) {
	home := t.TempDir()
	runSypherWithHome(t, home, "onboard")
	out, _, code := runSypherWithHome(t, home, "config", "get", "task.timeout_sec")
	if code != 0 {
		t.Errorf("config get exit code = %d", code)
	}
	if !strings.Contains(out, "300") {
		t.Errorf("config get task.timeout_sec should show 300 (default): %s", out)
	}
}

func TestCLI_ConfigSet(t *testing.T) {
	home := t.TempDir()
	runSypherWithHome(t, home, "onboard")
	out, _, code := runSypherWithHome(t, home, "config", "set", "task.timeout_sec", "120")
	if code != 0 {
		t.Errorf("config set exit code = %d", code)
	}
	if !strings.Contains(out, "Config updated") {
		t.Errorf("config set output: %s", out)
	}
	// Verify the value was set
	getOut, _, _ := runSypherWithHome(t, home, "config", "get", "task.timeout_sec")
	if !strings.Contains(getOut, "120") {
		t.Errorf("config get after set should show 120: %s", getOut)
	}
}

func TestCLI_ConfigUsage(t *testing.T) {
	stdout, stderr, _ := runSypher(t, "config")
	combined := stdout + stderr
	if !strings.Contains(combined, "Usage") && !strings.Contains(combined, "config") {
		t.Errorf("config usage: %s", combined)
	}
}

func TestCLI_AgentsList(t *testing.T) {
	home := t.TempDir()
	runSypherWithHome(t, home, "onboard")
	out, _, code := runSypherWithHome(t, home, "agents", "list")
	if code != 0 {
		t.Errorf("agents list exit code = %d", code)
	}
	if !strings.Contains(out, "Agents") {
		t.Errorf("agents output: %s", out)
	}
}

func TestCLI_AgentsUsage(t *testing.T) {
	stdout, stderr, _ := runSypher(t, "agents")
	combined := stdout + stderr
	if !strings.Contains(combined, "Usage") && !strings.Contains(combined, "list") {
		t.Errorf("agents usage: %s", combined)
	}
}

func TestCLI_MonitorsList(t *testing.T) {
	home := t.TempDir()
	runSypherWithHome(t, home, "onboard")
	out, _, code := runSypherWithHome(t, home, "monitors", "list")
	if code != 0 {
		t.Errorf("monitors list exit code = %d", code)
	}
	if !strings.Contains(out, "No monitors") && !strings.Contains(out, "monitors") {
		t.Errorf("monitors output: %s", out)
	}
}

func TestCLI_AuditShow(t *testing.T) {
	home := t.TempDir()
	runSypherWithHome(t, home, "onboard")
	auditDir := filepath.Join(home, ".sypher-mini", "audit")
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		t.Fatalf("create audit dir: %v", err)
	}
	taskLog := filepath.Join(auditDir, "test-task-123.log")
	if err := os.WriteFile(taskLog, []byte("2024-01-01 12:00:00 [INFO] test log line\n"), 0644); err != nil {
		t.Fatalf("write audit log: %v", err)
	}
	out, _, code := runSypherWithHome(t, home, "audit", "show", "test-task-123")
	if code != 0 {
		t.Errorf("audit show exit code = %d", code)
	}
	if !strings.Contains(out, "test log line") {
		t.Errorf("audit show output: %s", out)
	}
}

func TestCLI_AuditUsage(t *testing.T) {
	stdout, stderr, _ := runSypher(t, "audit")
	combined := stdout + stderr
	if !strings.Contains(combined, "Usage") && !strings.Contains(combined, "show") {
		t.Errorf("audit usage: %s", combined)
	}
}

func TestCLI_Replay(t *testing.T) {
	home := t.TempDir()
	runSypherWithHome(t, home, "onboard")
	replayDir := filepath.Join(home, ".sypher-mini", "replay")
	if err := os.MkdirAll(replayDir, 0755); err != nil {
		t.Fatalf("create replay dir: %v", err)
	}
	replayJSON := `{"input":"hello","tool_calls":[],"tool_results":[],"llm_responses":[]}`
	if err := os.WriteFile(filepath.Join(replayDir, "replay-task-456.json"), []byte(replayJSON), 0644); err != nil {
		t.Fatalf("write replay file: %v", err)
	}
	out, _, code := runSypherWithHome(t, home, "replay", "replay-task-456")
	if code != 0 {
		t.Errorf("replay exit code = %d", code)
	}
	if !strings.Contains(out, "Replay") && !strings.Contains(out, "hello") {
		t.Errorf("replay output: %s", out)
	}
}

func TestCLI_ReplayUsage(t *testing.T) {
	stdout, stderr, _ := runSypher(t, "replay")
	combined := stdout + stderr
	if !strings.Contains(combined, "Usage") {
		t.Errorf("replay usage: %s", combined)
	}
}

func TestCLI_CancelUsage(t *testing.T) {
	out, _, _ := runSypher(t, "cancel")
	if !strings.Contains(out, "Usage") && !strings.Contains(out, "task_id") {
		t.Errorf("cancel usage: %s", out)
	}
}

func TestCLI_WhatsAppUsage(t *testing.T) {
	out, _, _ := runSypher(t, "whatsapp")
	if !strings.Contains(out, "Usage") && !strings.Contains(out, "connect") {
		t.Errorf("whatsapp usage: %s", out)
	}
}

func TestCLI_InstallService(t *testing.T) {
	out, _, code := runSypher(t, "install-service")
	if code != 0 {
		t.Errorf("install-service exit code = %d", code)
	}
	if !strings.Contains(out, "install-service") {
		t.Errorf("install-service output: %s", out)
	}
	if !strings.Contains(out, "systemd") && !strings.Contains(out, "LaunchAgents") && !strings.Contains(out, "Task Scheduler") {
		t.Errorf("install-service should mention platform options: %s", out)
	}
}

func TestCLI_Extensions(t *testing.T) {
	out, _, code := runSypher(t, "extensions")
	if code != 0 {
		t.Errorf("extensions exit code = %d", code)
	}
	if !strings.Contains(out, "extensions") && !strings.Contains(out, "No extensions") && !strings.Contains(out, "whatsapp") {
		t.Logf("extensions output: %s", out)
	}
}

func TestCLI_CommandsList(t *testing.T) {
	home := t.TempDir()
	runSypherWithHome(t, home, "onboard")
	out, _, code := runSypherWithHome(t, home, "commands", "list")
	if code != 0 {
		t.Errorf("commands list exit code = %d", code)
	}
	if !strings.Contains(out, "No command") && !strings.Contains(out, "Available") && !strings.Contains(out, "commands") {
		t.Errorf("commands output: %s", out)
	}
}

func TestCLI_CommandsUsage(t *testing.T) {
	stdout, stderr, _ := runSypher(t, "commands")
	combined := stdout + stderr
	if !strings.Contains(combined, "Usage") && !strings.Contains(combined, "list") {
		t.Errorf("commands usage: %s", combined)
	}
}

func TestCLI_SafeFlag(t *testing.T) {
	home := t.TempDir()
	runSypherWithHome(t, home, "onboard")
	// --safe must be after the command, not before: sypher status --safe
	out, _, code := runSypherWithHome(t, home, "status", "--safe")
	if code != 0 {
		t.Errorf("status --safe exit code = %d", code)
	}
	if !strings.Contains(out, "Sypher-mini") {
		t.Errorf("status with --safe: %s", out)
	}
}

// TestIsAllowedSender verifies WhatsApp allow_from normalization (Baileys JID vs config +123, LID format).
func TestIsAllowedSender(t *testing.T) {
	tests := []struct {
		from      string
		allowFrom []string
		want      bool
	}{
		{"1234567890@s.whatsapp.net", []string{"+1234567890"}, true},
		{"1234567890@s.whatsapp.net", []string{"1234567890"}, true},
		{"+1234567890", []string{"1234567890@s.whatsapp.net"}, true},
		{"1234567890@s.whatsapp.net", []string{}, true},
		{"1234567890@s.whatsapp.net", []string{"+9999999999"}, false},
		{"other@s.whatsapp.net", []string{"+1234567890"}, false},
		{"60838547296357@lid", []string{"60838547296357"}, true},
		{"60838547296357@lid", []string{"+60838547296357"}, true},
		{"202383321759875@lid", []string{"202383321759875"}, true},
	}
	for _, tt := range tests {
		got := isAllowedSender(tt.from, tt.allowFrom)
		if got != tt.want {
			t.Errorf("isAllowedSender(%q, %v) = %v, want %v", tt.from, tt.allowFrom, got, tt.want)
		}
	}
}
