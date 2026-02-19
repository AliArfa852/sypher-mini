package clisession

import (
	"sync"
	"time"
)

const (
	DefaultTailLines = 10
	MaxTailLines     = 100
)

// Session holds a CLI terminal session with tag and output buffer.
type Session struct {
	ID          int
	Tag         string
	Created     time.Time
	LastActivity time.Time
	Output      *ringBuffer
	mu          sync.RWMutex
}

// Manager stores active CLI sessions.
type Manager struct {
	sessions map[int]*Session
	nextID   int
	mu       sync.RWMutex
}

// NewManager creates a new CLI session manager.
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[int]*Session),
		nextID:   1,
	}
}

// New creates a new session with the given tag.
func (m *Manager) New(tag string) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()
	id := m.nextID
	m.nextID++
	s := &Session{
		ID:          id,
		Tag:         tag,
		Created:     time.Now(),
		LastActivity: time.Now(),
		Output:      newRingBuffer(MaxTailLines),
	}
	m.sessions[id] = s
	return s
}

// Get returns a session by ID.
func (m *Manager) Get(id int) *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[id]
}

// List returns all sessions with ID, tag, last activity.
func (m *Manager) List() []SessionInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]SessionInfo, 0, len(m.sessions))
	for _, s := range m.sessions {
		s.mu.RLock()
		out = append(out, SessionInfo{
			ID:           s.ID,
			Tag:          s.Tag,
			LastActivity: s.LastActivity,
		})
		s.mu.RUnlock()
	}
	return out
}

// SessionInfo is a summary of a session for listing.
type SessionInfo struct {
	ID           int
	Tag          string
	LastActivity time.Time
}

// Append adds lines to the session output buffer.
func (s *Session) Append(lines string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastActivity = time.Now()
	s.Output.Append(lines)
}

// Tail returns the last n lines (capped at MaxTailLines).
func (s *Session) Tail(n int) string {
	if n <= 0 {
		n = DefaultTailLines
	}
	if n > MaxTailLines {
		n = MaxTailLines
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Output.Tail(n)
}

// ringBuffer holds the last N lines of output.
type ringBuffer struct {
	lines []string
	size  int
	head  int
	count int
}

func newRingBuffer(maxLines int) *ringBuffer {
	return &ringBuffer{
		lines: make([]string, maxLines),
		size:  maxLines,
	}
}

func (r *ringBuffer) Append(text string) {
	if text == "" {
		return
	}
	// Split into lines and add each
	lines := splitLines(text)
	for _, line := range lines {
		r.lines[r.head] = line
		r.head = (r.head + 1) % r.size
		if r.count < r.size {
			r.count++
		}
	}
}

func (r *ringBuffer) Tail(n int) string {
	if n <= 0 || r.count == 0 {
		return ""
	}
	if n > r.count {
		n = r.count
	}
	out := make([]string, 0, n)
	start := (r.head - n + r.size) % r.size
	for i := 0; i < n; i++ {
		idx := (start + i) % r.size
		if r.lines[idx] != "" {
			out = append(out, r.lines[idx])
		}
	}
	return joinLines(out)
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i+1])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	result := ""
	for _, l := range lines {
		result += l
		if len(l) > 0 && l[len(l)-1] != '\n' {
			result += "\n"
		}
	}
	return result
}
