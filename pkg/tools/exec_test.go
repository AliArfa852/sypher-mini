package tools

import (
	"context"
	"runtime"
	"strings"
	"testing"

	"github.com/sypherexx/sypher-mini/pkg/audit"
	"github.com/sypherexx/sypher-mini/pkg/config"
	"github.com/sypherexx/sypher-mini/pkg/process"
)

func TestExecTool_SafeMode(t *testing.T) {
	cfg := config.DefaultConfig()
	auditDir := t.TempDir()
	cfg.Audit.Dir = auditDir
	al := audit.New(auditDir)
	pt := process.New()
	exec := NewExecTool(cfg, al, pt, true)

	resp := exec.Execute(context.Background(), Request{
		ToolCallID: "tc1",
		TaskID:    "t1",
		AgentID:   "main",
		Name:      "exec",
		Args:      map[string]interface{}{"command": "echo hi"},
	})
	if !resp.IsError {
		t.Error("expected error in safe mode")
	}
}

func TestExecTool_DenyPattern(t *testing.T) {
	cfg := config.DefaultConfig()
	auditDir := t.TempDir()
	cfg.Audit.Dir = auditDir
	al := audit.New(auditDir)
	pt := process.New()
	exec := NewExecTool(cfg, al, pt, false)

	resp := exec.Execute(context.Background(), Request{
		ToolCallID: "tc1",
		TaskID:    "t1",
		AgentID:   "main",
		Name:      "exec",
		Args:      map[string]interface{}{"command": "rm -rf /"},
	})
	if !resp.IsError {
		t.Error("expected error for denied command")
	}
	if resp.Code != CodeSafetyBlocked {
		t.Errorf("expected SAFETY_BLOCKED, got %s", resp.Code)
	}
}

func TestExecTool_Success(t *testing.T) {
	cfg := config.DefaultConfig()
	auditDir := t.TempDir()
	cfg.Audit.Dir = auditDir
	workspace := t.TempDir()
	cfg.Agents.Defaults.Workspace = workspace
	al := audit.New(auditDir)
	pt := process.New()
	exec := NewExecTool(cfg, al, pt, false)

	cmd := "echo hello"
	if runtime.GOOS == "windows" {
		cmd = "echo hello"
	}
	resp := exec.Execute(context.Background(), Request{
		ToolCallID: "tc1",
		TaskID:    "t1",
		AgentID:   "main",
		Name:      "exec",
		Args:      map[string]interface{}{"command": cmd, "working_dir": workspace},
	})
	if resp.IsError {
		t.Errorf("unexpected error: %s", resp.ForLLM)
	}
}

func TestResolveCommandAlias(t *testing.T) {
	tests := []struct {
		cmd     string
		wantSub string // substring that should appear in result
	}{
		{"dir", "ls"},
		{"ls", "ls"},
		{"top", "top"},
		{"ps", "ps"},
	}
	if runtime.GOOS == "windows" {
		tests = []struct {
			cmd     string
			wantSub string
		}{
			{"dir", "dir"},
			{"ls", "dir"},
			{"top", "tasklist"},
			{"ps", "tasklist"},
		}
	}
	for _, tt := range tests {
		got := resolveCommandAlias(tt.cmd)
		if !strings.Contains(got, tt.wantSub) {
			t.Errorf("resolveCommandAlias(%q) = %q, want containing %q", tt.cmd, got, tt.wantSub)
		}
	}
}

func TestExecTool_CommandAlias_DirOrLs(t *testing.T) {
	cfg := config.DefaultConfig()
	auditDir := t.TempDir()
	workspace := t.TempDir()
	cfg.Audit.Dir = auditDir
	cfg.Agents.Defaults.Workspace = workspace
	al := audit.New(auditDir)
	pt := process.New()
	exec := NewExecTool(cfg, al, pt, false)

	cmd := "ls"
	if runtime.GOOS == "windows" {
		cmd = "dir"
	}
	resp := exec.Execute(context.Background(), Request{
		ToolCallID: "tc1",
		TaskID:    "t1",
		AgentID:   "main",
		Name:      "exec",
		Args:      map[string]interface{}{"command": cmd, "working_dir": workspace},
	})
	if resp.IsError {
		t.Errorf("dir/ls alias failed: %s", resp.ForLLM)
	}
	if len(resp.ForLLM) == 0 {
		t.Error("expected non-empty output from dir/ls")
	}
}

func TestExecTool_CommandAlias_ProcessList(t *testing.T) {
	cfg := config.DefaultConfig()
	auditDir := t.TempDir()
	workspace := t.TempDir()
	cfg.Audit.Dir = auditDir
	cfg.Agents.Defaults.Workspace = workspace
	al := audit.New(auditDir)
	pt := process.New()
	exec := NewExecTool(cfg, al, pt, false)

	// ps (Unix) or tasklist (Windows) - both should work
	cmd := "ps"
	if runtime.GOOS == "windows" {
		cmd = "tasklist"
	}
	resp := exec.Execute(context.Background(), Request{
		ToolCallID: "tc1",
		TaskID:    "t1",
		AgentID:   "main",
		Name:      "exec",
		Args:      map[string]interface{}{"command": cmd, "working_dir": workspace},
	})
	if resp.IsError {
		t.Errorf("ps/tasklist failed: %s", resp.ForLLM)
	}
	if len(resp.ForLLM) == 0 {
		t.Error("expected non-empty output from ps/tasklist")
	}
}
