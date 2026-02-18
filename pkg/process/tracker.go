package process

import (
	"sync"
)

// Tracker tracks PIDs started by Sypher-mini per task.
type Tracker struct {
	taskPIDs map[string]map[int]bool
	mu       sync.RWMutex
}

// New creates a new process tracker.
func New() *Tracker {
	return &Tracker{
		taskPIDs: make(map[string]map[int]bool),
	}
}

// Record records a PID for a task.
func (t *Tracker) Record(taskID string, pid int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.taskPIDs[taskID] == nil {
		t.taskPIDs[taskID] = make(map[int]bool)
	}
	t.taskPIDs[taskID][pid] = true
}

// CanKill returns true if the PID belongs to the task and can be killed.
func (t *Tracker) CanKill(taskID string, pid int) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	pids, ok := t.taskPIDs[taskID]
	if !ok {
		return false
	}
	return pids[pid]
}

// RemoveTask removes all PIDs for a task (e.g. on completion).
func (t *Tracker) RemoveTask(taskID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.taskPIDs, taskID)
}
