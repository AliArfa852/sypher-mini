package clisession

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/sypherexx/sypher-mini/pkg/config"
)

// persistedState is the JSON schema for sessions.json.
type persistedState struct {
	Sessions []persistedSession `json:"sessions"`
	NextID   int                `json:"next_id"`
}

type persistedSession struct {
	ID           int       `json:"id"`
	Tag          string    `json:"tag"`
	Created      time.Time `json:"created"`
	LastActivity time.Time `json:"last_activity"`
	OutputTail   string    `json:"output_tail"`
	WorkingDir   string    `json:"working_dir,omitempty"`
}

// DefaultSessionsDir returns ~/.sypher-mini/cli-sessions.
func DefaultSessionsDir() string {
	return filepath.Join(config.ExpandPath("~/.sypher-mini"), "cli-sessions")
}

// NewManagerWithPersistence creates a Manager that persists to dir/sessions.json.
// Loads existing sessions on startup. Pass "" to use DefaultSessionsDir().
func NewManagerWithPersistence(dir string) *Manager {
	if dir == "" {
		dir = DefaultSessionsDir()
	}
	m := &Manager{
		sessions:    make(map[int]*Session),
		nextID:      1,
		persistPath: filepath.Join(dir, "sessions.json"),
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return m
	}
	m.load()
	return m
}

// load reads sessions from disk.
func (m *Manager) load() {
	data, err := os.ReadFile(m.persistPath)
	if err != nil {
		return
	}
	var state persistedState
	if err := json.Unmarshal(data, &state); err != nil {
		return
	}
	if state.NextID < 1 {
		state.NextID = 1
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, ps := range state.Sessions {
		rb := newRingBuffer(MaxTailLines)
		if ps.OutputTail != "" {
			rb.Append(ps.OutputTail)
		}
		s := &Session{
			ID:           ps.ID,
			Tag:          ps.Tag,
			Created:      ps.Created,
			LastActivity: ps.LastActivity,
			Output:       rb,
			WorkingDir:   ps.WorkingDir,
		}
		m.sessions[ps.ID] = s
	}
	m.nextID = state.NextID
}

// Persist writes sessions to disk. No-op if manager was created with NewManager (no persistPath).
func (m *Manager) Persist() {
	if m.persistPath == "" {
		return
	}
	m.mu.RLock()
	sessions := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessions = append(sessions, s)
	}
	nextID := m.nextID
	m.mu.RUnlock()

	state := persistedState{NextID: nextID, Sessions: make([]persistedSession, 0, len(sessions))}
	for _, s := range sessions {
		s.mu.RLock()
		tail := s.Output.Tail(MaxTailLines)
		wd := s.WorkingDir
		s.mu.RUnlock()
		state.Sessions = append(state.Sessions, persistedSession{
			ID:           s.ID,
			Tag:          s.Tag,
			Created:      s.Created,
			LastActivity: s.LastActivity,
			OutputTail:   tail,
			WorkingDir:   wd,
		})
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(m.persistPath, data, 0644)
}
