package task

import (
	"context"
	"sync"
	"time"
)

// Manager tracks active tasks.
type Manager struct {
	tasks   map[string]*Task
	timeout time.Duration
	mu      sync.RWMutex
}

// NewManager creates a new task manager.
func NewManager(timeoutSec int) *Manager {
	timeout := 300 * time.Second
	if timeoutSec > 0 {
		timeout = time.Duration(timeoutSec) * time.Second
	}
	return &Manager{
		tasks:   make(map[string]*Task),
		timeout: timeout,
	}
}

// Create creates a new task and registers it.
func (m *Manager) Create(agentID, sessionKey string) *Task {
	m.mu.Lock()
	defer m.mu.Unlock()
	t := New(agentID, sessionKey)
	m.tasks[t.ID] = t
	return t
}

// Get returns a task by ID.
func (m *Manager) Get(id string) (*Task, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.tasks[id]
	return t, ok
}

// Remove removes a task (e.g. after completion).
func (m *Manager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tasks, id)
}

// Cancel marks a task as cancelled.
func (m *Manager) Cancel(id string) bool {
	m.mu.RLock()
	t, ok := m.tasks[id]
	m.mu.RUnlock()
	if !ok {
		return false
	}
	t.SetCancelled(true)
	return true
}

// Timeout returns the default task timeout.
func (m *Manager) Timeout() time.Duration {
	return m.timeout
}

// List returns all active (non-terminal) tasks.
func (m *Manager) List() []*Task {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []*Task
	for _, t := range m.tasks {
		if !t.IsTerminal() {
			out = append(out, t)
		}
	}
	return out
}

// RunWithTimeout runs fn with the manager's timeout.
func (m *Manager) RunWithTimeout(ctx context.Context, t *Task, fn func(context.Context) error) error {
	return t.RunWithTimeout(ctx, m.timeout, fn)
}
