package project

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/sypherexx/sypher-mini/pkg/config"
)

// Store manages project metadata in the workspace code-projects directory.
type Store struct {
	mu        sync.RWMutex
	dir       string
	workspace string
}

// NewStore creates a project store for the given workspace.
func NewStore(workspace string) *Store {
	ws := config.ExpandPath(workspace)
	if ws == "" {
		ws, _ = os.Getwd()
	}
	dir := filepath.Join(ws, "code-projects")
	return &Store{
		dir:       dir,
		workspace: ws,
	}
}

// List returns all projects.
func (s *Store) List() ([]*Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := os.MkdirAll(s.dir, 0755); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	var out []*Project
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) != ".json" {
			continue
		}
		path := filepath.Join(s.dir, e.Name())
		p, err := Load(path)
		if err != nil {
			continue
		}
		out = append(out, p)
	}
	return out, nil
}

// Get returns a project by ID.
func (s *Store) Get(id string) (*Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	path := filepath.Join(s.dir, id+".json")
	return Load(path)
}

// Add saves a new project.
func (s *Store) Add(p *Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.MkdirAll(s.dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(s.dir, p.ID+".json")
	return p.Save(path)
}

// Update updates an existing project.
func (s *Store) Update(p *Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	path := filepath.Join(s.dir, p.ID+".json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.ErrNotExist
	}
	return p.Save(path)
}

// Remove deletes a project by ID.
func (s *Store) Remove(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	path := filepath.Join(s.dir, id+".json")
	return os.Remove(path)
}

// Workspace returns the workspace path.
func (s *Store) Workspace() string {
	return s.workspace
}
