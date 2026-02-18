package agent

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sypherexx/sypher-mini/pkg/config"
)

// BootstrapFiles are loaded in order and injected into the system prompt.
const (
	BootstrapAGENT    = "AGENT.md"
	BootstrapAGENTS   = "AGENTS.md"
	BootstrapSOUL     = "SOUL.md"
	BootstrapUSER     = "USER.md"
	BootstrapIDENTITY = "IDENTITY.md"
)

// LoadBootstrapFiles loads workspace bootstrap files and returns concatenated content.
func LoadBootstrapFiles(workspace string, agentID string) string {
	base := config.ExpandPath(workspace)
	if base == "" {
		return ""
	}
	// Per-agent workspace: ~/.sypher-mini/workspace-{id}/
	agentWorkspace := filepath.Join(filepath.Dir(base), "workspace-"+agentID)
	if _, err := os.Stat(agentWorkspace); os.IsNotExist(err) {
		agentWorkspace = base
	}

	var out []string
	files := []string{BootstrapAGENTS, BootstrapAGENT, BootstrapSOUL, BootstrapUSER, BootstrapIDENTITY}
	for _, name := range files {
		path := filepath.Join(agentWorkspace, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		out = append(out, string(data))
	}
	return strings.TrimSpace(strings.Join(out, "\n\n"))
}
