package clisession

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPersist_LoadSave(t *testing.T) {
	dir := t.TempDir()
	m := NewManagerWithPersistence(dir)
	if m.persistPath == "" {
		t.Fatal("expected persistPath to be set")
	}

	// Create session and append
	s := m.New("test-tag")
	s.Append("line1\n")
	s.Append("line2\n")
	m.Persist()

	// Create new manager, should load from disk
	m2 := NewManagerWithPersistence(dir)
	list := m2.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 session, got %d", len(list))
	}
	if list[0].Tag != "test-tag" {
		t.Errorf("tag = %q, want test-tag", list[0].Tag)
	}
	s2 := m2.Get(s.ID)
	if s2 == nil {
		t.Fatal("session not found")
	}
	tail := s2.Tail(10)
	if tail != "line1\nline2\n" && tail != "line1\nline2" {
		t.Errorf("tail = %q", tail)
	}
}

func TestPersist_NewManagerNoPersist(t *testing.T) {
	m := NewManager()
	if m.persistPath != "" {
		t.Errorf("NewManager should not have persistPath")
	}
	m.Persist() // no-op, should not panic
}

func TestDefaultSessionsDir(t *testing.T) {
	dir := DefaultSessionsDir()
	if dir == "" {
		t.Fatal("DefaultSessionsDir returned empty")
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("expected absolute path, got %q", dir)
	}
}

func TestPersist_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	m := NewManagerWithPersistence(dir)
	// No sessions - load should not fail
	list := m.List()
	if len(list) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(list))
	}
}

func TestPersist_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sessions.json")
	if err := os.WriteFile(path, []byte("invalid json"), 0644); err != nil {
		t.Fatal(err)
	}
	m := NewManagerWithPersistence(dir)
	// Should not panic, should start fresh
	list := m.List()
	if len(list) != 0 {
		t.Errorf("expected 0 sessions after invalid json, got %d", len(list))
	}
}
