package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// CommandConfig defines a single command from ~/.sypher-mini/commands/{name}.json.
type CommandConfig struct {
	Name             string   `json:"name"`
	AgentID          string   `json:"agent_id"`
	Args             []string `json:"args,omitempty"`
	AllowedFiles     []string `json:"allowed_files,omitempty"`
	WorkingDirectory string   `json:"working_directory,omitempty"`
}

// Load reads a command config from the commands directory.
func Load(commandsDir, name string) (*CommandConfig, error) {
	if commandsDir == "" {
		home, _ := os.UserHomeDir()
		commandsDir = filepath.Join(home, ".sypher-mini", "commands")
	}
	// Prevent path traversal: reject names with .., /, or \
	if name == "" || name == "." || name == ".." ||
		strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return nil, os.ErrNotExist
	}
	path := filepath.Join(commandsDir, name+".json")
	pathClean := filepath.Clean(path)
	dirAbs, _ := filepath.Abs(commandsDir)
	pathAbs, _ := filepath.Abs(pathClean)
	if dirAbs != "" && pathAbs != "" {
		rel, err := filepath.Rel(dirAbs, pathAbs)
		if err != nil || rel == ".." || (len(rel) >= 2 && rel[:2] == "..") {
			return nil, os.ErrNotExist
		}
	}
	data, err := os.ReadFile(pathClean)
	if err != nil {
		return nil, err
	}
	var cfg CommandConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Name == "" {
		cfg.Name = name
	}
	return &cfg, nil
}

// List returns names of available command configs.
func List(commandsDir string) ([]string, error) {
	if commandsDir == "" {
		home, _ := os.UserHomeDir()
		commandsDir = filepath.Join(home, ".sypher-mini", "commands")
	}
	entries, err := os.ReadDir(commandsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			names = append(names, e.Name()[:len(e.Name())-5])
		}
	}
	return names, nil
}
