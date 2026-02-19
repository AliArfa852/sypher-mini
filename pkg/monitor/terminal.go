package monitor

// TerminalMonitor optionally tracks active PTYs/shells from authorized list.
// Stub: full implementation would attach to terminal output and log commands.
type TerminalMonitor struct {
	AuthorizedTerminals []string
}

// NewTerminalMonitor creates a terminal monitor. Stub implementation.
func NewTerminalMonitor(authorized []string) *TerminalMonitor {
	return &TerminalMonitor{AuthorizedTerminals: authorized}
}

// Run starts monitoring. No-op in stub.
func (t *TerminalMonitor) Run() {
	// TODO: attach to PTYs, log commands to audit
}
