package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry returned nil")
	}
	if len(r.slashCommands) == 0 {
		t.Error("slash commands should be populated")
	}
	if len(r.systemCommands) == 0 {
		t.Error("system commands should be populated")
	}
}

func TestRegistry_AddProjectCommands(t *testing.T) {
	r := NewRegistry()
	r.AddProjectCommands([]string{"proj1", "proj2"})
	if len(r.projectCommands) != 2 {
		t.Errorf("projectCommands len = %d, want 2", len(r.projectCommands))
	}
}

func TestRegistry_BuildSummary_Base(t *testing.T) {
	r := NewRegistry()
	summary := r.BuildSummary("")
	if !strings.Contains(summary, "## Available Commands") {
		t.Error("summary should contain header")
	}
	if !strings.Contains(summary, "/status") {
		t.Error("summary should contain slash commands")
	}
	if !strings.Contains(summary, "dir, ls") {
		t.Error("summary should contain system commands")
	}
}

func TestRegistry_BuildSummary_WithProjects(t *testing.T) {
	r := NewRegistry()
	r.AddProjectCommands([]string{"myapp", "api"})
	summary := r.BuildSummary("")
	if !strings.Contains(summary, "Registered Projects") {
		t.Error("summary should contain projects section")
	}
	if !strings.Contains(summary, "myapp") || !strings.Contains(summary, "api") {
		t.Errorf("summary should list projects: %s", summary)
	}
}

func TestRegistry_BuildSummary_WithCustomCommands(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "deploy.json"), []byte(`{"name":"deploy"}`), 0644); err != nil {
		t.Fatal(err)
	}
	r := NewRegistry()
	summary := r.BuildSummary(dir)
	if !strings.Contains(summary, "Custom Commands") {
		t.Error("summary should contain custom commands section")
	}
	if !strings.Contains(summary, "deploy") {
		t.Errorf("summary should list deploy: %s", summary)
	}
}

func TestBuildForAgent(t *testing.T) {
	summary := BuildForAgent("", []string{"proj1"})
	if summary == "" {
		t.Error("BuildForAgent should return non-empty")
	}
	if !strings.Contains(summary, "proj1") {
		t.Errorf("BuildForAgent should include projects: %s", summary)
	}
}

func TestBuildForAgent_EmptyProjects(t *testing.T) {
	summary := BuildForAgent("", nil)
	if !strings.Contains(summary, "Slash Commands") {
		t.Error("BuildForAgent should include slash commands even with no projects")
	}
}
