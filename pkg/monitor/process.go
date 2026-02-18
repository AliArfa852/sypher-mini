package monitor

import (
	"context"
	"regexp"
	"sync"
	"time"

	"github.com/sypherexx/sypher-mini/pkg/config"
)

// ProcessMonitor attaches to a process and watches for error patterns.
type ProcessMonitor struct {
	cfg       config.ProcessMonitor
	pattern   *regexp.Regexp
	lastAlert time.Time
	mu        sync.Mutex
	onAlert   func(monitorID, message string)
}

// NewProcessMonitor creates a process monitor.
func NewProcessMonitor(cfg config.ProcessMonitor, onAlert func(monitorID, message string)) *ProcessMonitor {
	var re *regexp.Regexp
	if cfg.ErrorPattern != "" {
		re, _ = regexp.Compile(cfg.ErrorPattern)
	}
	return &ProcessMonitor{
		cfg:     cfg,
		pattern: re,
		onAlert: onAlert,
	}
}

// Run starts monitoring. For process monitors, we would typically spawn the
// command and attach to stdout/stderr. This is a placeholder - full implementation
// would use exec.Command and scan output.
func (m *ProcessMonitor) Run(ctx context.Context) {
	// Placeholder: process monitoring requires spawning the command and
	// streaming output. Defer to extension or external runner.
	ticker := time.NewTicker(time.Duration(m.cfg.CooldownSec) * time.Second)
	if m.cfg.CooldownSec <= 0 {
		ticker = time.NewTicker(time.Minute)
	}
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// No-op: actual implementation would check process output
			_ = m.pattern
		}
	}
}

// MatchError checks if output matches the error pattern.
func (m *ProcessMonitor) MatchError(output string) bool {
	if m.pattern == nil {
		return false
	}
	return m.pattern.MatchString(output)
}
