package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_InvalidName(t *testing.T) {
	dir := t.TempDir()

	tests := []string{"", ".", "..", "a/b", "a\\b", "..x", "x.."}
	for _, name := range tests {
		_, err := Load(dir, name)
		if err != os.ErrNotExist {
			t.Errorf("Load(%q) err = %v, want ErrNotExist", name, err)
		}
	}
}

func TestLoad_ValidCommand(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mycmd.json")
	if err := os.WriteFile(path, []byte(`{"name":"mycmd","agent_id":"main"}`), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(dir, "mycmd")
	if err != nil {
		t.Fatalf("Load err = %v", err)
	}
	if cfg.Name != "mycmd" {
		t.Errorf("Name = %q, want mycmd", cfg.Name)
	}
	if cfg.AgentID != "main" {
		t.Errorf("AgentID = %q, want main", cfg.AgentID)
	}
}

func TestLoad_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := Load(dir, "nonexistent")
	if err == nil {
		t.Error("Load(nonexistent) should return error")
	}
}

func TestList(t *testing.T) {
	dir := t.TempDir()
	// Empty dir
	names, err := List(dir)
	if err != nil {
		t.Fatalf("List err = %v", err)
	}
	if len(names) != 0 {
		t.Errorf("List empty dir = %v, want []", names)
	}

	// Add JSON files
	for _, f := range []string{"a.json", "b.json", "c.txt"} {
		_ = os.WriteFile(filepath.Join(dir, f), []byte("{}"), 0644)
	}
	names, err = List(dir)
	if err != nil {
		t.Fatalf("List err = %v", err)
	}
	if len(names) != 2 {
		t.Errorf("List = %v, want 2 json names", names)
	}
}
