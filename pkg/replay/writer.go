package replay

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/sypherexx/sypher-mini/pkg/config"
)

// Record holds task replay data.
type Record struct {
	TaskID    string      `json:"task_id"`
	Input     interface{} `json:"input"`
	Result    string      `json:"result"`
	ToolCalls interface{} `json:"tool_calls,omitempty"`
	Status    string      `json:"status"`
}

// Writer writes replay records when enabled.
type Writer struct {
	dir     string
	enabled bool
}

// NewWriter creates a replay writer.
func NewWriter(cfg *config.Config) *Writer {
	enabled := cfg.Replay.Enabled
	dir := config.ExpandPath(cfg.Replay.Dir)
	if dir == "" {
		dir = config.ExpandPath("~/.sypher-mini/replay")
	}
	return &Writer{dir: dir, enabled: enabled}
}

// Write records a task completion.
func (w *Writer) Write(r Record) error {
	if !w.enabled {
		return nil
	}
	if err := os.MkdirAll(w.dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(w.dir, r.TaskID+".json")
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
