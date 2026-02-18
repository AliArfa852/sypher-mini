package task

import (
	"context"
	"testing"
)

func TestManager_CreateGetRemove(t *testing.T) {
	m := NewManager(300)
	task := m.Create("agent1", "s1")
	if task == nil {
		t.Fatal("Create returned nil")
	}
	if task.ID == "" {
		t.Error("task ID empty")
	}

	got, ok := m.Get(task.ID)
	if !ok || got != task {
		t.Errorf("Get: ok=%v, task mismatch", ok)
	}

	m.Remove(task.ID)
	_, ok = m.Get(task.ID)
	if ok {
		t.Error("Remove did not remove task")
	}
}

func TestManager_Cancel(t *testing.T) {
	m := NewManager(300)
	task := m.Create("a", "s")
	if !m.Cancel(task.ID) {
		t.Error("Cancel returned false")
	}
	if !task.IsCancelled() {
		t.Error("task not cancelled")
	}
	if m.Cancel("nonexistent") {
		t.Error("Cancel nonexistent should return false")
	}
}

func TestManager_RunWithTimeout(t *testing.T) {
	m := NewManager(1) // 1 second
	task := m.Create("a", "s")
	err := m.RunWithTimeout(context.Background(), task, func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
