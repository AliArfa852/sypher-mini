package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProject_Load_Save(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	p := &Project{
		ID:           "myapp",
		Name:         "My App",
		Path:         "../myapp",
		BuildCommand: "npm run build",
		EnvActivate:  "source venv/bin/activate",
	}
	if err := p.Save(path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("Save did not create file")
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.ID != p.ID || loaded.Name != p.Name || loaded.BuildCommand != p.BuildCommand {
		t.Errorf("Load mismatch: got %+v", loaded)
	}
}

func TestProject_Load_AutoIDFromName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	content := `{"name":"My Cool App","path":"/tmp/app"}`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.ID == "" {
		t.Error("expected ID to be derived from name")
	}
	if loaded.ID != "My-Cool-App" {
		t.Errorf("expected sanitized ID, got %q", loaded.ID)
	}
}

func TestProject_Save_CreatesDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "nested", "proj.json")

	p := &Project{ID: "x", Name: "X", Path: "."}
	if err := p.Save(path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("Save did not create nested dirs")
	}
}

func TestProject_AbsPath_Relative(t *testing.T) {
	p := &Project{Path: "subdir/app"}
	// Use temp dir as workspace to avoid platform-specific ~ expansion
	ws := t.TempDir()
	got := p.AbsPath(ws)
	want := filepath.Join(ws, "subdir", "app")
	if got != want {
		t.Errorf("AbsPath(relative) = %q, want %q", got, want)
	}
}

func TestProject_AbsPath_Absolute(t *testing.T) {
	absPath := filepath.Join(t.TempDir(), "absolute", "path")
	p := &Project{Path: absPath}
	got := p.AbsPath("/any/workspace")
	if got != absPath {
		t.Errorf("AbsPath(absolute) = %q, want %q", got, absPath)
	}
}

func TestProject_Load_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("not json"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestProject_Load_NotFound(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "nonexistent.json"))
	if err == nil {
		t.Error("expected error for missing file")
	}
}
