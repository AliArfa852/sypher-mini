package tools

import (
	"context"
	"testing"
	"time"

	"github.com/sypherexx/sypher-mini/pkg/config"
	"github.com/sypherexx/sypher-mini/pkg/policy"
)

func TestBrowserSurfTool_SafeMode(t *testing.T) {
	cfg := config.DefaultConfig()
	policyEval := policy.NewEvaluator(cfg)
	tool := NewBrowserSurfTool(cfg, policyEval, true)

	resp := tool.Execute(context.Background(), Request{
		ToolCallID: "tc1",
		TaskID:     "t1",
		AgentID:    "main",
		Name:       "browser_surf",
		Args:       map[string]interface{}{"url": "https://example.com", "action": "navigate"},
	})
	if !resp.IsError {
		t.Error("expected error in safe mode")
	}
	if resp.Code != CodePermissionDenied {
		t.Errorf("expected CodePermissionDenied, got %s", resp.Code)
	}
	if resp.ForLLM != "browser_surf disabled in safe mode" {
		t.Errorf("unexpected ForLLM: %s", resp.ForLLM)
	}
}

func TestBrowserSurfTool_MissingURL(t *testing.T) {
	cfg := config.DefaultConfig()
	policyEval := policy.NewEvaluator(cfg)
	tool := NewBrowserSurfTool(cfg, policyEval, false)

	resp := tool.Execute(context.Background(), Request{
		ToolCallID: "tc1",
		TaskID:     "t1",
		AgentID:    "main",
		Name:       "browser_surf",
		Args:       map[string]interface{}{"action": "navigate"},
	})
	if !resp.IsError {
		t.Error("expected error for missing url")
	}
	if resp.ForLLM != "Missing 'url' argument" {
		t.Errorf("unexpected ForLLM: %s", resp.ForLLM)
	}
}

func TestBrowserSurfTool_SSRFBlocked(t *testing.T) {
	cfg := config.DefaultConfig()
	policyEval := policy.NewEvaluator(cfg)
	tool := NewBrowserSurfTool(cfg, policyEval, false)

	resp := tool.Execute(context.Background(), Request{
		ToolCallID: "tc1",
		TaskID:     "t1",
		AgentID:    "main",
		Name:       "browser_surf",
		Args:       map[string]interface{}{"url": "http://localhost:8080/test", "action": "navigate"},
	})
	if !resp.IsError {
		t.Error("expected error for blocked host (localhost)")
	}
	if resp.ForLLM != "URL host not allowed (internal/private addresses blocked)" {
		t.Errorf("unexpected ForLLM: %s", resp.ForLLM)
	}
}

func TestBrowserSurfTool_InvalidAction(t *testing.T) {
	cfg := config.DefaultConfig()
	policyEval := policy.NewEvaluator(cfg)
	tool := NewBrowserSurfTool(cfg, policyEval, false)

	resp := tool.Execute(context.Background(), Request{
		ToolCallID: "tc1",
		TaskID:     "t1",
		AgentID:    "main",
		Name:       "browser_surf",
		Args:       map[string]interface{}{"url": "https://example.com", "action": "invalid_action"},
	})
	if !resp.IsError {
		t.Error("expected error for invalid action")
	}
	if resp.ForLLM != "Invalid action: invalid_action" {
		t.Errorf("unexpected ForLLM: %s", resp.ForLLM)
	}
}

func TestBrowserSurfTool_ClickMissingSelector(t *testing.T) {
	cfg := config.DefaultConfig()
	policyEval := policy.NewEvaluator(cfg)
	tool := NewBrowserSurfTool(cfg, policyEval, false)

	resp := tool.Execute(context.Background(), Request{
		ToolCallID: "tc1",
		TaskID:     "t1",
		AgentID:    "main",
		Name:       "browser_surf",
		Args:       map[string]interface{}{"url": "https://example.com", "action": "click"},
	})
	if !resp.IsError {
		t.Error("expected error for click without selector")
	}
	if resp.ForLLM != "Missing 'selector' for click action" {
		t.Errorf("unexpected ForLLM: %s", resp.ForLLM)
	}
}

func TestBrowserSurfTool_NewBrowserSurfTool_Config(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Tools.BrowserSurf.TimeoutSec = 60
	policyEval := policy.NewEvaluator(cfg)
	tool := NewBrowserSurfTool(cfg, policyEval, false)

	if tool.timeout != 60*time.Second {
		t.Errorf("expected 60s timeout, got %v", tool.timeout)
	}
}
