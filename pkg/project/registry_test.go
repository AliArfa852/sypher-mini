package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanWorkspace_Empty(t *testing.T) {
	dir := t.TempDir()
	found := ScanWorkspace(dir)
	if len(found) != 0 {
		t.Errorf("ScanWorkspace(empty) = %v, want []", found)
	}
}

func TestScanWorkspace_FindsPackageJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	found := ScanWorkspace(dir)
	if len(found) != 1 {
		t.Fatalf("ScanWorkspace: got %d, want 1", len(found))
	}
	if found[0] != dir {
		t.Errorf("found[0] = %q, want %q", found[0], dir)
	}
}

func TestScanWorkspace_FindsGoMod(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x"), 0600); err != nil {
		t.Fatal(err)
	}
	found := ScanWorkspace(dir)
	if len(found) != 1 {
		t.Fatalf("ScanWorkspace(go.mod): got %d, want 1", len(found))
	}
}

func TestScanWorkspace_FindsGit(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatal(err)
	}
	found := ScanWorkspace(dir)
	if len(found) != 1 {
		t.Fatalf("ScanWorkspace(.git): got %d, want 1", len(found))
	}
	if found[0] != dir {
		t.Errorf(".git parent = %q, want %q", found[0], dir)
	}
}

func TestScanWorkspace_SkipsNodeModules(t *testing.T) {
	dir := t.TempDir()
	nm := filepath.Join(dir, "node_modules")
	if err := os.MkdirAll(nm, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nm, "package.json"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	found := ScanWorkspace(dir)
	if len(found) != 0 {
		t.Errorf("ScanWorkspace should skip node_modules, got %v", found)
	}
}

func TestScanWorkspace_NestedProjects(t *testing.T) {
	dir := t.TempDir()
	sub1 := filepath.Join(dir, "app1")
	sub2 := filepath.Join(dir, "app2")
	if err := os.MkdirAll(sub1, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(sub2, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub1, "package.json"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub2, "go.mod"), []byte("module x"), 0600); err != nil {
		t.Fatal(err)
	}
	found := ScanWorkspace(dir)
	if len(found) != 2 {
		t.Fatalf("ScanWorkspace(nested): got %d, want 2", len(found))
	}
}

func TestSuggestProject(t *testing.T) {
	dir := t.TempDir()
	ws := filepath.Dir(dir)
	p := SuggestProject(dir, ws)
	if p == nil {
		t.Fatal("SuggestProject returned nil")
	}
	if p.Name != filepath.Base(dir) {
		t.Errorf("Name = %q, want %q", p.Name, filepath.Base(dir))
	}
	if p.ID == "" {
		t.Error("ID should be set")
	}
}
