package tools

import (
	"context"
	"runtime"
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
