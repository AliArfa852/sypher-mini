package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Config holds logging configuration.
type Config struct {
	Level  string `json:"level"`  // debug, info, warn, error
	Output string `json:"output"` // stdout, stderr, or file path
	JSON   bool   `json:"json"`   // emit JSON lines
}

// Logger writes structured JSON log lines.
type Logger struct {
	cfg  Config
	file *os.File
	mu   sync.Mutex
}

// New creates a logger from config.
func New(cfg Config) (*Logger, error) {
	l := &Logger{cfg: cfg}
	if cfg.Output != "" && cfg.Output != "stdout" && cfg.Output != "stderr" {
		f, err := os.OpenFile(cfg.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return nil, err
		}
		l.file = f
	}
	return l, nil
}

// Log writes a JSON log line.
func (l *Logger) Log(level, msg string, fields map[string]interface{}) {
	if !l.cfg.JSON {
		return
	}
	entry := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"level":    level,
		"msg":      msg,
	}
	for k, v := range fields {
		entry[k] = v
	}
	data, _ := json.Marshal(entry)
	line := string(data) + "\n"

	l.mu.Lock()
	defer l.mu.Unlock()
	w := os.Stdout
	if l.cfg.Output == "stderr" {
		w = os.Stderr
	}
	if l.file != nil {
		w = l.file
	}
	fmt.Fprint(w, line)
}

// Info logs at info level.
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	l.Log("info", msg, fields)
}

// Error logs at error level.
func (l *Logger) Error(msg string, fields map[string]interface{}) {
	l.Log("error", msg, fields)
}

// Close closes the log file if open.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		err := l.file.Close()
		l.file = nil
		return err
	}
	return nil
}
