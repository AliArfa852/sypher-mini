package project

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/sypherexx/sypher-mini/pkg/config"
)

// Project holds metadata for a code project (like Docker Compose).
type Project struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Path              string `json:"path"`               // absolute or relative to workspace
	EnvActivate       string `json:"env_activate"`       // e.g. "source venv/bin/activate", "conda activate myenv"
	RunCommand        string `json:"run_command"`       // e.g. "npm run dev", "go run ."
	BuildCommand      string `json:"build_command"`      // e.g. "npm run build"
	Terminal          string `json:"terminal"`          // e.g. "cmd", "powershell", "bash"
	DockerComposePath string `json:"docker_compose_path,omitempty"`
}

// Load reads a project from a JSON file.
func Load(path string) (*Project, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p Project
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	if p.ID == "" && p.Name != "" {
		p.ID = sanitizeID(p.Name)
	}
	return &p, nil
}

// Save writes a project to a JSON file.
func (p *Project) Save(path string) error {
	if p.ID == "" && p.Name != "" {
		p.ID = sanitizeID(p.Name)
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// AbsPath returns the absolute path of the project, resolving relative paths against workspace.
func (p *Project) AbsPath(workspace string) string {
	ws := config.ExpandPath(workspace)
	if filepath.IsAbs(p.Path) {
		return p.Path
	}
	return filepath.Join(ws, p.Path)
}

func sanitizeID(s string) string {
	var out []rune
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			out = append(out, r)
		} else if r == ' ' {
			out = append(out, '-')
		}
	}
	if len(out) == 0 {
		return "project"
	}
	return string(out)
}
