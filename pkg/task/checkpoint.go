package task

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Checkpoint holds task state for crash recovery.
type Checkpoint struct {
	TaskID     string `json:"task_id"`
	State      string `json:"state"`
	HistoryHash string `json:"history_hash,omitempty"`
}

// WriteCheckpoint writes task state to disk for recovery.
func WriteCheckpoint(dir, taskID, state, historyHash string) error {
	if dir == "" {
		return nil
	}
	_ = os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, taskID+".checkpoint.json")
	data, _ := json.Marshal(Checkpoint{
		TaskID:      taskID,
		State:       state,
		HistoryHash: historyHash,
	})
	return os.WriteFile(path, data, 0600)
}

// RemoveCheckpoint removes a checkpoint file.
func RemoveCheckpoint(dir, taskID string) error {
	if dir == "" {
		return nil
	}
	return os.Remove(filepath.Join(dir, taskID+".checkpoint.json"))
}
