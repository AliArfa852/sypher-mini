package menu

import (
	"sync"
	"time"
)

const defaultSessionTTL = 10 * time.Minute

// Session holds menu state for a chat.
type Session struct {
	CurrentMenu       string
	ParentMenu        string
	LastActive        time.Time
	PendingAction     string   // e.g. "projects_build", "projects_pull", "tasks_cancel"
	PendingProjectIDs []string // project IDs for selection (1-indexed)
	PendingTaskIDs    []string // task IDs for cancel selection (1-indexed)
}

// SessionStore holds per-chat menu sessions.
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	ttl      time.Duration
}

// NewSessionStore creates a new session store.
func NewSessionStore(ttl time.Duration) *SessionStore {
	if ttl <= 0 {
		ttl = defaultSessionTTL
	}
	return &SessionStore{
		sessions: make(map[string]*Session),
		ttl:      ttl,
	}
}

// Key returns the session key for channel and chatID.
func Key(channel, chatID string) string {
	return channel + ":" + chatID
}

// Get returns the session if it exists and is not expired.
func (s *SessionStore) Get(key string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ses, ok := s.sessions[key]
	if !ok || ses == nil {
		return nil, false
	}
	if time.Since(ses.LastActive) > s.ttl {
		return nil, false
	}
	return ses, true
}

// Set creates or updates a session.
func (s *SessionStore) Set(key string, currentMenu, parentMenu string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	if ses, ok := s.sessions[key]; ok && ses != nil {
		ses.CurrentMenu = currentMenu
		ses.ParentMenu = parentMenu
		ses.LastActive = now
		return
	}
	s.sessions[key] = &Session{
		CurrentMenu: currentMenu,
		ParentMenu:  parentMenu,
		LastActive:  now,
	}
}

// Clear removes the session (e.g. on "back" to main).
func (s *SessionStore) Clear(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, key)
}

// ResetToMain sets session to main menu.
func (s *SessionStore) ResetToMain(key string) {
	s.Set(key, "main", "")
}

// SetPendingProjectAction sets state for project selection (build/pull).
// Creates session if it does not exist.
func (s *SessionStore) SetPendingProjectAction(key, action string, projectIDs []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	if ses, ok := s.sessions[key]; ok && ses != nil {
		ses.PendingAction = action
		ses.PendingProjectIDs = projectIDs
		ses.PendingTaskIDs = nil
		ses.LastActive = now
	} else {
		s.sessions[key] = &Session{
			CurrentMenu:       "projects",
			ParentMenu:        "main",
			LastActive:        now,
			PendingAction:     action,
			PendingProjectIDs: projectIDs,
		}
	}
}

// SetPendingTaskCancel sets state for task cancel selection.
// Creates session if it does not exist.
func (s *SessionStore) SetPendingTaskCancel(key string, taskIDs []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	if ses, ok := s.sessions[key]; ok && ses != nil {
		ses.PendingAction = "tasks_cancel"
		ses.PendingProjectIDs = nil
		ses.PendingTaskIDs = taskIDs
		ses.LastActive = now
	} else {
		s.sessions[key] = &Session{
			CurrentMenu:    "tasks",
			ParentMenu:     "main",
			LastActive:     now,
			PendingAction:  "tasks_cancel",
			PendingTaskIDs: taskIDs,
		}
	}
}

// GetPendingProjectAction returns (action, projectIDs, taskIDs, ok).
func (s *SessionStore) GetPendingProjectAction(key string) (string, []string, []string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ses, ok := s.sessions[key]
	if !ok || ses == nil || ses.PendingAction == "" {
		return "", nil, nil, false
	}
	if time.Since(ses.LastActive) > s.ttl {
		return "", nil, nil, false
	}
	return ses.PendingAction, ses.PendingProjectIDs, ses.PendingTaskIDs, true
}

// ClearPendingProjectAction clears the pending state.
func (s *SessionStore) ClearPendingProjectAction(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ses, ok := s.sessions[key]; ok && ses != nil {
		ses.PendingAction = ""
		ses.PendingProjectIDs = nil
		ses.PendingTaskIDs = nil
	}
}
