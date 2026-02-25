package tools

import (
	"context"
	"runtime"
	"strings"
	"testing"

	"github.com/sypherexx/sypher-mini/pkg/config"
)

func TestInvokeCliAgentTool_SafeMode(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agents.List = []config.AgentConfig{
		{ID: "gemini-cli", Command: "echo", Args: []string{}},
	}
	tool := NewInvokeCliAgentTool(cfg, true)

	resp := tool.Execute(context.Background(), Request{
		ToolCallID: "tc1",
		TaskID:     "t1",
		AgentID:    "main",
		Name:       "invoke_cli_agent",
		Args:       map[string]interface{}{"task": "hello"},
	})
	if !resp.IsError {
		t.Error("expected error in safe mode")
	}
	if resp.ForLLM != "invoke_cli_agent disabled in safe mode" {
		t.Errorf("unexpected ForLLM: %s", resp.ForLLM)
	}
}

func TestInvokeCliAgentTool_MissingTask(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agents.List = []config.AgentConfig{
		{ID: "gemini-cli", Command: "echo", Args: []string{}},
	}
	tool := NewInvokeCliAgentTool(cfg, false)

	resp := tool.Execute(context.Background(), Request{
		ToolCallID: "tc1",
		TaskID:     "t1",
		AgentID:    "main",
		Name:       "invoke_cli_agent",
		Args:       map[string]interface{}{},
	})
	if !resp.IsError {
		t.Error("expected error for missing task")
	}
	if resp.ForLLM != "Missing 'task' argument" {
		t.Errorf("unexpected ForLLM: %s", resp.ForLLM)
	}
}

func TestInvokeCliAgentTool_NoCliAgentConfigured(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agents.List = []config.AgentConfig{
		{ID: "main", Default: true},
	}
	tool := NewInvokeCliAgentTool(cfg, false)

	resp := tool.Execute(context.Background(), Request{
		ToolCallID: "tc1",
		TaskID:     "t1",
		AgentID:    "main",
		Name:       "invoke_cli_agent",
		Args:       map[string]interface{}{"task": "hello"},
	})
	if !resp.IsError {
		t.Error("expected error when no CLI agent configured")
	}
	if resp.ForLLM != "No CLI agent configured" {
		t.Errorf("unexpected ForLLM: %s", resp.ForLLM)
	}
}

func TestInvokeCliAgentTool_ProjectResolver(t *testing.T) {
	cfg := config.DefaultConfig()
	// Use cmd (Windows) or sh (Unix) - both exist and can echo
	cmd, args := "sh", []string{"-c", "echo ok"}
	if runtime.GOOS == "windows" {
		cmd, args = "cmd", []string{"/c", "echo", "ok"}
	}
	cfg.Agents.List = []config.AgentConfig{
		{ID: "gemini-cli", Command: cmd, Args: args},
	}
	workspace := t.TempDir()
	cfg.Agents.Defaults.Workspace = workspace

	tool := NewInvokeCliAgentTool(cfg, false)
	projectDir := t.TempDir()
	tool.SetProjectResolver(func(projectID string) (string, bool) {
		if projectID == "test-proj" {
			return projectDir, true
		}
		return "", false
	})

	resp := tool.Execute(context.Background(), Request{
		ToolCallID: "tc1",
		TaskID:     "t1",
		AgentID:    "gemini-cli",
		Name:       "invoke_cli_agent",
		Args: map[string]interface{}{
			"task":    "hello",
			"project": "test-proj",
		},
	})
	if resp.IsError {
		t.Errorf("unexpected error: %s", resp.ForLLM)
	}
	// Tool runs with working_dir=projectDir; command should complete
	if resp.ForLLM != "CLI agent completed" && !strings.Contains(resp.ForLLM, "ok") {
		t.Logf("resp: %+v", resp)
	}
}

func TestInvokeCliAgentTool_ResolveCliAgent_ByID(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agents.List = []config.AgentConfig{
		{ID: "main", Default: true},
		{ID: "gemini-cli", Command: "echo", Args: []string{"gemini"}},
		{ID: "claude-cli", Command: "echo", Args: []string{"claude"}},
	}
	tool := NewInvokeCliAgentTool(cfg, false)

	cmd, args := tool.resolveCliAgent("claude-cli")
	if cmd != "echo" {
		t.Errorf("command = %q, want echo", cmd)
	}
	if len(args) != 1 || args[0] != "claude" {
		t.Errorf("args = %v, want [claude]", args)
	}
}

func TestInvokeCliAgentTool_ResolveCliAgent_FirstWithCommand(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agents.List = []config.AgentConfig{
		{ID: "main", Default: true},
		{ID: "gemini-cli", Command: "echo", Args: []string{"first"}},
	}
	tool := NewInvokeCliAgentTool(cfg, false)

	cmd, args := tool.resolveCliAgent("")
	if cmd != "echo" {
		t.Errorf("command = %q, want echo", cmd)
	}
	if len(args) != 1 || args[0] != "first" {
		t.Errorf("args = %v, want [first]", args)
	}
}
