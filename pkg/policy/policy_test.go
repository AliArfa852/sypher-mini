package policy

import (
	"testing"

	"github.com/sypherexx/sypher-mini/pkg/config"
)

func TestEvaluator_CheckRateLimit(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Policies.RateLimits = []config.RateLimit{
		{AgentID: "*", ToolName: "exec", RequestsPerMinute: 2},
	}
	e := NewEvaluator(cfg)

	if !e.CheckRateLimit("main", "exec") {
		t.Error("first request should pass")
	}
	if !e.CheckRateLimit("main", "exec") {
		t.Error("second request should pass")
	}
	if e.CheckRateLimit("main", "exec") {
		t.Error("third request should be rate limited")
	}
}

func TestEvaluator_CanAccessFile(t *testing.T) {
	cfg := config.DefaultConfig()
	e := NewEvaluator(cfg)

	// Workspace should always be allowed
	if !e.CanAccessFile("main", cfg.Agents.Defaults.Workspace, "read") {
		t.Error("workspace should be accessible")
	}
}
