package tools

import (
	"context"
	"testing"
	"time"

	"github.com/sypherexx/sypher-mini/pkg/config"
	"github.com/sypherexx/sypher-mini/pkg/policy"
)

func TestBrowserScrapeTool_SafeMode(t *testing.T) {
	cfg := config.DefaultConfig()
	policyEval := policy.NewEvaluator(cfg)
	tool := NewBrowserScrapeTool(cfg, policyEval, true)

	resp := tool.Execute(context.Background(), Request{
		ToolCallID: "tc1",
		TaskID:     "t1",
		AgentID:    "main",
		Name:       "browser_scrape",
		Args:       map[string]interface{}{"url": "https://example.com"},
	})
	if !resp.IsError {
		t.Error("expected error in safe mode")
	}
	if resp.ForLLM != "browser_scrape disabled in safe mode" {
		t.Errorf("unexpected ForLLM: %s", resp.ForLLM)
	}
}

func TestBrowserScrapeTool_MissingURL(t *testing.T) {
	cfg := config.DefaultConfig()
	policyEval := policy.NewEvaluator(cfg)
	tool := NewBrowserScrapeTool(cfg, policyEval, false)

	resp := tool.Execute(context.Background(), Request{
		ToolCallID: "tc1",
		TaskID:     "t1",
		AgentID:    "main",
		Name:       "browser_scrape",
		Args:       map[string]interface{}{},
	})
	if !resp.IsError {
		t.Error("expected error for missing url")
	}
	if resp.ForLLM != "Missing 'url' argument" {
		t.Errorf("unexpected ForLLM: %s", resp.ForLLM)
	}
}

func TestBrowserScrapeTool_SSRFBlocked(t *testing.T) {
	cfg := config.DefaultConfig()
	policyEval := policy.NewEvaluator(cfg)
	tool := NewBrowserScrapeTool(cfg, policyEval, false)

	resp := tool.Execute(context.Background(), Request{
		ToolCallID: "tc1",
		TaskID:     "t1",
		AgentID:    "main",
		Name:       "browser_scrape",
		Args:       map[string]interface{}{"url": "http://127.0.0.1/page"},
	})
	if !resp.IsError {
		t.Error("expected error for blocked host")
	}
	if resp.ForLLM != "URL host not allowed (internal/private addresses blocked)" {
		t.Errorf("unexpected ForLLM: %s", resp.ForLLM)
	}
}

func TestStripHTML(t *testing.T) {
	tests := []struct {
		html string
		want string
	}{
		{"<p>hello</p>", "hello"},
		{"<div><span>foo</span> bar</div>", "foo bar"},
		{"<a href=\"/x\">link</a>", "link"},
		{"  <b>bold</b>  ", "bold"},
		{"", ""},
		{"no tags", "no tags"},
	}
	for _, tt := range tests {
		got := stripHTML(tt.html)
		if got != tt.want {
			t.Errorf("stripHTML(%q) = %q, want %q", tt.html, got, tt.want)
		}
	}
}

func TestBrowserScrapeTool_NewBrowserScrapeTool_Config(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Tools.BrowserScrape.TimeoutSec = 45
	policyEval := policy.NewEvaluator(cfg)
	tool := NewBrowserScrapeTool(cfg, policyEval, false)

	if tool.timeout != 45*time.Second {
		t.Errorf("expected 45s timeout, got %v", tool.timeout)
	}
}

func TestExtractHost(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://example.com/path", "example.com"},
		{"http://localhost:8080/", "localhost"},
		{"https://sub.example.com:443/x", "sub.example.com"},
	}
	for _, tt := range tests {
		got := extractHost(tt.url)
		if got != tt.want {
			t.Errorf("extractHost(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}

func TestIsBlockedHost(t *testing.T) {
	blocked := []string{"localhost", "127.0.0.1", "10.0.0.1"}
	for _, h := range blocked {
		if !isBlockedHost(h) {
			t.Errorf("expected %q to be blocked", h)
		}
	}
	allowed := []string{"example.com", "google.com"}
	for _, h := range allowed {
		if isBlockedHost(h) {
			t.Errorf("expected %q to be allowed", h)
		}
	}
}
