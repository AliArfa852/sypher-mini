package process

import (
	"testing"
)

func TestTracker_RecordAndCanKill(t *testing.T) {
	tr := New()
	tr.Record("task1", 123)
	tr.Record("task1", 456)

	if !tr.CanKill("task1", 123) {
		t.Error("expected CanKill task1/123")
	}
	if !tr.CanKill("task1", 456) {
		t.Error("expected CanKill task1/456")
	}
	if tr.CanKill("task1", 999) {
		t.Error("should not allow kill of unrecorded PID")
	}
	if tr.CanKill("task2", 123) {
		t.Error("should not allow kill from different task")
	}
}

func TestTracker_RemoveTask(t *testing.T) {
	tr := New()
	tr.Record("task1", 123)
	tr.RemoveTask("task1")
	if tr.CanKill("task1", 123) {
		t.Error("RemoveTask should clear PIDs")
	}
}
