package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/sypherexx/sypher-mini/pkg/config"
)

func TestTailOutputTool_Execute(t *testing.T) {
	dir := t.TempDir()
	fpath := filepath.Join(dir, "test.log")
	content := "line1\nline2\nline3\nline4\nline5\n"
	if err := os.WriteFile(fpath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.Agents.Defaults.Workspace = dir
	cfg.Agents.Defaults.RestrictToWorkspace = true
	tool := NewTailOutputTool(cfg, false)

	req := Request{
		ToolCallID: "tc1",
		TaskID:     "t1",
		AgentID:    "main",
		Name:       "tail_output",
		Args:       map[string]interface{}{"path": fpath, "lines": float64(3)},
	}
	resp := tool.Execute(context.Background(), req)
	if resp.IsError {
		t.Fatalf("unexpected error: %s", resp.ForLLM)
	}
	if resp.ForLLM != "line3\nline4\nline5" {
		t.Errorf("expected line3..line5, got %q", resp.ForLLM)
	}
}

func TestTailOutputTool_OutsideWorkspace(t *testing.T) {
	dir := t.TempDir()
	ws := filepath.Join(dir, "workspace")
	if err := os.MkdirAll(ws, 0755); err != nil {
		t.Fatal(err)
	}
	// File outside workspace
	outside := filepath.Join(dir, "outside.txt")
	if err := os.WriteFile(outside, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.Agents.Defaults.Workspace = ws
	cfg.Agents.Defaults.RestrictToWorkspace = true
	tool := NewTailOutputTool(cfg, false)

	req := Request{
		ToolCallID: "tc1",
		TaskID:     "t1",
		AgentID:    "main",
		Name:       "tail_output",
		Args:       map[string]interface{}{"path": outside},
	}
	resp := tool.Execute(context.Background(), req)
	if !resp.IsError {
		t.Fatal("expected error for path outside workspace")
	}
}
