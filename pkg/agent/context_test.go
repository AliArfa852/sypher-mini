package agent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sypherexx/sypher-mini/pkg/config"
)

func TestLoadBootstrapFiles_EmptyWorkspace(t *testing.T) {
	content := LoadBootstrapFiles("", "main")
	if content != "" {
		t.Errorf("expected empty for empty workspace, got %q", content)
	}
}

func TestLoadBootstrapFiles_Nonexistent(t *testing.T) {
	content := LoadBootstrapFiles("/nonexistent/path", "main")
	if content != "" {
		t.Errorf("expected empty for nonexistent path, got %q", content)
	}
}

func TestLoadBootstrapFiles_WithSOUL(t *testing.T) {
	dir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.Agents.Defaults.Workspace = dir

	soulPath := filepath.Join(dir, "SOUL.md")
	if err := os.WriteFile(soulPath, []byte("# Soul\nI am helpful."), 0644); err != nil {
		t.Fatal(err)
	}

	content := LoadBootstrapFiles(dir, "main")
	if content == "" {
		t.Error("expected content from SOUL.md")
	}
	if content != "# Soul\nI am helpful." {
		t.Errorf("unexpected content: %q", content)
	}
}
