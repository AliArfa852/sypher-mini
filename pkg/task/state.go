package task

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// State represents task lifecycle states.
type State string

const (
	StatePending    State = "pending"
	StateAuthorized State = "authorized"
	StateExecuting  State = "executing"
	StateMonitoring State = "monitoring"
	StateCompleted  State = "completed"
	StateFailed     State = "failed"
	StateKilled     State = "killed"
	StateTimeout    State = "timeout"
)

// Task represents a single task with lifecycle state.
type Task struct {
	ID         string
	State      State
	AgentID    string
	SessionKey string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Cancelled  bool
	mu         sync.RWMutex
}

// New creates a new task in pending state.
func New(agentID, sessionKey string) *Task {
	now := time.Now()
	return &Task{
		ID:         uuid.New().String(),
		State:      StatePending,
		AgentID:    agentID,
		SessionKey: sessionKey,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// Transition transitions to a new state.
func (t *Task) Transition(to State) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.State = to
	t.UpdatedAt = time.Now()
}

// SetCancelled marks the task as cancelled.
func (t *Task) SetCancelled(c bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Cancelled = c
}

// IsCancelled returns whether the task is cancelled.
func (t *Task) IsCancelled() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Cancelled
}

// GetState returns the current state.
func (t *Task) GetState() State {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.State
}

// IsTerminal returns true if the task is in a terminal state.
func (t *Task) IsTerminal() bool {
	s := t.GetState()
	return s == StateCompleted || s == StateFailed || s == StateKilled || s == StateTimeout
}

// RunWithTimeout runs fn with a timeout. On timeout, transitions to StateTimeout.
func (t *Task) RunWithTimeout(ctx context.Context, timeout time.Duration, fn func(context.Context) error) error {
	if timeout <= 0 {
		return fn(ctx)
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- fn(ctx)
	}()

	select {
	case err := <-done:
		if ctx.Err() == context.DeadlineExceeded {
			t.Transition(StateTimeout)
			return context.DeadlineExceeded
		}
		return err
	case <-ctx.Done():
		t.Transition(StateTimeout)
		return ctx.Err()
	}
}
