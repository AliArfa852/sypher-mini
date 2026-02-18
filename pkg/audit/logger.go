package audit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Logger writes per-task command audit logs.
type Logger struct {
	dir    string
	mu     sync.Mutex
}

// New creates a new audit logger.
func New(dir string) *Logger {
	dir = expandPath(dir)
	_ = os.MkdirAll(dir, 0755)
	return &Logger{dir: dir}
}

// LogCommand logs a command execution for a task.
func (l *Logger) LogCommand(taskID, toolCallID, command, cwd string, exitCode int, outputSummary string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	path := filepath.Join(l.dir, taskID+".log")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	ts := time.Now().Format(time.RFC3339)
	line := fmt.Sprintf("[%s] [%s] %s | exec | cmd=%q cwd=%q exit=%d | %s\n",
		taskID, toolCallID, ts, command, cwd, exitCode, truncate(outputSummary, 200))
	_, err = f.WriteString(line)
	return err
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func expandPath(p string) string {
	if p == "" {
		return p
	}
	if strings.HasPrefix(p, "~") {
		home, _ := os.UserHomeDir()
		if len(p) > 1 && (p[1] == '/' || p[1] == '\\') {
			return filepath.Join(home, p[2:])
		}
		return home
	}
	return p
}
