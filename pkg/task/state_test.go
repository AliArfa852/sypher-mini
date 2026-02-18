package task

import (
	"context"
	"testing"
	"time"
)

func TestTask_Transition(t *testing.T) {
	task := New("agent1", "session1")
	if task.State != StatePending {
		t.Errorf("expected pending, got %s", task.State)
	}

	task.Transition(StateAuthorized)
	if task.State != StateAuthorized {
		t.Errorf("expected authorized, got %s", task.State)
	}

	task.Transition(StateExecuting)
	if task.State != StateExecuting {
		t.Errorf("expected executing, got %s", task.State)
	}
}

func TestTask_IsTerminal(t *testing.T) {
	tests := []struct {
		state    State
		terminal bool
	}{
		{StatePending, false},
		{StateExecuting, false},
		{StateCompleted, true},
		{StateFailed, true},
		{StateKilled, true},
		{StateTimeout, true},
	}
	for _, tt := range tests {
		task := New("a", "s")
		task.Transition(tt.state)
		if got := task.IsTerminal(); got != tt.terminal {
			t.Errorf("state %s: IsTerminal() = %v, want %v", tt.state, got, tt.terminal)
		}
	}
}

func TestTask_Cancelled(t *testing.T) {
	task := New("a", "s")
	if task.IsCancelled() {
		t.Error("expected not cancelled")
	}
	task.SetCancelled(true)
	if !task.IsCancelled() {
		t.Error("expected cancelled")
	}
}

func TestTask_RunWithTimeout_Completes(t *testing.T) {
	task := New("a", "s")
	err := task.RunWithTimeout(context.Background(), 5*time.Second, func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if task.GetState() != StateExecuting {
		t.Errorf("state should remain executing after success, got %s", task.GetState())
	}
}

func TestTask_RunWithTimeout_Timeout(t *testing.T) {
	task := New("a", "s")
	err := task.RunWithTimeout(context.Background(), 50*time.Millisecond, func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})
	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
	if task.GetState() != StateTimeout {
		t.Errorf("expected state timeout, got %s", task.GetState())
	}
}
