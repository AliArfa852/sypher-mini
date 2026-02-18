package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}
	if len(cfg.Agents.List) == 0 {
		t.Error("expected agents")
	}
	if cfg.Task.TimeoutSec == 0 {
		t.Error("expected timeout")
	}
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()
	if got := ExpandPath("~"); got != home {
		t.Errorf("ExpandPath(~) = %q, want %q", got, home)
	}
	if got := ExpandPath("~/foo"); got != filepath.Join(home, "foo") {
		t.Errorf("ExpandPath(~/foo) = %q", got)
	}
	// Absolute path should be returned as-is
	absPath := filepath.Join(os.TempDir(), "foo")
	if got := ExpandPath(absPath); got != absPath {
		t.Errorf("ExpandPath(abs) = %q, want %q", got, absPath)
	}
}
