package capabilities

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Registry maps tools and agents to capabilities.
type Registry struct {
	Tools  map[string][]string `json:"tools"`
	Agents map[string][]string `json:"agents"`
}

// Load loads the registry from path.
func Load(path string) (*Registry, error) {
	path = expandPath(path)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultRegistry(), nil
		}
		return nil, err
	}
	var r Registry
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if r.Tools == nil {
		r.Tools = make(map[string][]string)
	}
	if r.Agents == nil {
		r.Agents = make(map[string][]string)
	}
	return &r, nil
}

// DefaultRegistry returns built-in capability mappings.
func DefaultRegistry() *Registry {
	return &Registry{
		Tools: map[string][]string{
			"exec":       {"code_generation", "log_analysis", "deploy_service"},
			"edit_file":  {"code_generation", "file_edit"},
			"read_file":  {"log_analysis", "file_edit"},
			"message":    {"notify_user"},
			"web_fetch":  {"web_search"},
		},
		Agents: map[string][]string{
			"cursor":      {"code_generation"},
			"gemini-cli":  {"code_generation"},
			"main":        {"code_generation", "log_analysis", "deploy_service", "notify_user", "web_search", "file_edit"},
		},
	}
}

// ResolveTools returns tool names that satisfy the given capability.
func (r *Registry) ResolveTools(capability string) []string {
	capability = strings.ToLower(strings.TrimSpace(capability))
	var out []string
	for tool, caps := range r.Tools {
		for _, c := range caps {
			if strings.EqualFold(c, capability) {
				out = append(out, tool)
				break
			}
		}
	}
	return out
}

// ResolveAgents returns agent IDs that satisfy the given capability.
func (r *Registry) ResolveAgents(capability string) []string {
	capability = strings.ToLower(strings.TrimSpace(capability))
	var out []string
	for agent, caps := range r.Agents {
		for _, c := range caps {
			if strings.EqualFold(c, capability) {
				out = append(out, agent)
				break
			}
		}
	}
	return out
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
