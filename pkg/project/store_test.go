package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStore_Add_List_Get(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	p := &Project{ID: "proj1", Name: "Project 1", Path: "p1", BuildCommand: "go build"}
	if err := store.Add(p); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	list, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("List: got %d projects, want 1", len(list))
	}
	if list[0].ID != "proj1" {
		t.Errorf("List[0].ID = %q, want proj1", list[0].ID)
	}

	got, err := store.Get("proj1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Name != "Project 1" {
		t.Errorf("Get: Name = %q, want Project 1", got.Name)
	}
}

func TestStore_Get_NotFound(t *testing.T) {
	store := NewStore(t.TempDir())
	_, err := store.Get("nonexistent")
	if err == nil {
		t.Error("expected error for missing project")
	}
}

func TestStore_Update(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	p := &Project{ID: "upd", Name: "Original", Path: ".", BuildCommand: "build"}
	if err := store.Add(p); err != nil {
		t.Fatal(err)
	}

	p.BuildCommand = "npm run build"
	if err := store.Update(p); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, _ := store.Get("upd")
	if got.BuildCommand != "npm run build" {
		t.Errorf("Update: BuildCommand = %q, want npm run build", got.BuildCommand)
	}
}

func TestStore_Update_NotFound(t *testing.T) {
	store := NewStore(t.TempDir())
	p := &Project{ID: "missing", Name: "X", Path: "."}
	err := store.Update(p)
	if err == nil {
		t.Error("expected error when updating non-existent project")
	}
}

func TestStore_Remove(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	p := &Project{ID: "del", Name: "ToDelete", Path: "."}
	if err := store.Add(p); err != nil {
		t.Fatal(err)
	}
	if err := store.Remove("del"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	_, err := store.Get("del")
	if err == nil {
		t.Error("expected project to be removed")
	}
}

func TestStore_List_Empty(t *testing.T) {
	store := NewStore(t.TempDir())
	list, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("List: got %d projects, want 0", len(list))
	}
}

func TestStore_List_IgnoresNonJSON(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	// Create code-projects dir and add a .txt file
	cp := filepath.Join(dir, "code-projects")
	if err := os.MkdirAll(cp, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cp, "readme.txt"), []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}

	list, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("List should ignore .txt, got %d", len(list))
	}
}

func TestStore_Workspace(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	if store.Workspace() != dir {
		t.Errorf("Workspace = %q, want %q", store.Workspace(), dir)
	}
}

