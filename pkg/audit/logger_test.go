package audit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLogger_LogCommand(t *testing.T) {
	dir := t.TempDir()
	l := New(dir)

	err := l.LogCommand("task1", "tc1", "echo hi", "/tmp", 0, "hi")
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(dir, "task1.log")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("expected log content")
	}
}
