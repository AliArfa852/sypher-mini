package replay

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sypherexx/sypher-mini/pkg/config"
)

func TestNewWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Replay.Enabled = false
	w := NewWriter(cfg)
	if w == nil {
		t.Fatal("NewWriter returned nil")
	}
	if w.enabled {
		t.Error("expected disabled when cfg.Replay.Enabled=false")
	}
}

func TestWriter_Write_Disabled(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Replay.Enabled = false
	w := NewWriter(cfg)
	err := w.Write(Record{TaskID: "t1", Status: "ok"})
	if err != nil {
		t.Errorf("Write when disabled err = %v", err)
	}
}

func TestWriter_Write_Enabled(t *testing.T) {
	dir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.Replay.Enabled = true
	cfg.Replay.Dir = dir
	w := NewWriter(cfg)

	err := w.Write(Record{TaskID: "task-123", Input: "hello", Result: "hi", Status: "completed"})
	if err != nil {
		t.Fatalf("Write err = %v", err)
	}
	path := filepath.Join(dir, "task-123.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("Write should create task file")
	}
}
