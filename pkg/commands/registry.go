package commands

import (
	"sort"
	"strings"
)

// CommandCategory groups commands by type.
type CommandCategory struct {
	Name     string
	Commands []string
}

// Registry provides a summary of available commands for the agent.
type Registry struct {
	slashCommands   []string
	systemCommands  []string
	projectCommands []string
}

// NewRegistry creates a registry. Call AddSlash, AddSystem, AddProject to populate.
func NewRegistry() *Registry {
	return &Registry{
		slashCommands: []string{
			"/status - Status",
			"/cli list - List CLI sessions",
			"/cli new -m \"tag\" - New session",
			"/cli run <N> <command> - Run command in session",
			"/cli <N> - Tail session output",
			"/config get <path> - Get config (operator)",
			"/agents - List agents (operator)",
			"/monitors - List monitors (operator)",
			"/cancel <task_id> - Cancel task",
		},
		systemCommands: []string{
			"dir, ls, ls -la - List directory",
			"top, ps, tasklist - Process list",
			"exec <command> - Run shell command",
		},
		projectCommands: []string{},
	}
}

// AddProjectCommands adds project IDs from the project store.
func (r *Registry) AddProjectCommands(projectIDs []string) {
	r.projectCommands = projectIDs
}

// BuildSummary returns a formatted string for the agent's system prompt.
func (r *Registry) BuildSummary(commandsDir string) string {
	var out []string
	out = append(out, "## Available Commands")
	out = append(out, "")
	out = append(out, "### WhatsApp Slash Commands")
	for _, c := range r.slashCommands {
		out = append(out, "- "+c)
	}
	out = append(out, "")
	out = append(out, "### System Commands (via exec)")
	for _, c := range r.systemCommands {
		out = append(out, "- "+c)
	}
	if len(r.projectCommands) > 0 {
		out = append(out, "")
		out = append(out, "### Registered Projects")
		for _, id := range r.projectCommands {
			out = append(out, "- "+id+" (build, pull, deploy)")
		}
	}
	if commandsDir != "" {
		names, _ := List(commandsDir)
		if len(names) > 0 {
			sort.Strings(names)
			out = append(out, "")
			out = append(out, "### Custom Commands")
			for _, n := range names {
				out = append(out, "- "+n)
			}
		}
	}
	return strings.Join(out, "\n")
}

// BuildForAgent returns a command summary for the agent's system prompt.
func BuildForAgent(commandsDir string, projectIDs []string) string {
	r := NewRegistry()
	r.AddProjectCommands(projectIDs)
	return r.BuildSummary(commandsDir)
}
