package menu

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// DefaultMenus returns the built-in menu definitions.
func DefaultMenus() MenusConfig {
	return MenusConfig{
		"main": {
			Title: "Sypher-mini Control Panel",
			Items: []MenuItem{
				{ID: "1", Label: "Projects", Submenu: "projects"},
				{ID: "2", Label: "Tasks", Submenu: "tasks"},
				{ID: "3", Label: "Logs & Monitoring", Submenu: "logs"},
				{ID: "4", Label: "CLI Sessions", Submenu: "cli"},
				{ID: "5", Label: "Server & Config", Submenu: "server"},
				{ID: "6", Label: "Help / Examples", Submenu: "help"},
			},
		},
		"projects": {
			Title: "Projects",
			Items: []MenuItem{
				{ID: "1", Label: "List active projects", Action: "projects_list"},
				{ID: "2", Label: "Open project workspace", Action: "projects_open"},
				{ID: "3", Label: "Run build/test", Action: "projects_build"},
				{ID: "4", Label: "Pull latest repo changes", Action: "projects_pull"},
				{ID: "0", Label: "Back to main menu", Submenu: "main"},
			},
		},
		"tasks": {
			Title: "Task Manager",
			Items: []MenuItem{
				{ID: "1", Label: "Create new task", Action: "tasks_create"},
				{ID: "2", Label: "View running tasks", Action: "tasks_list"},
				{ID: "3", Label: "Authorize pending task", Action: "tasks_authorize"},
				{ID: "4", Label: "Cancel task", Action: "tasks_cancel"},
				{ID: "0", Label: "Back to main menu", Submenu: "main"},
			},
		},
		"logs": {
			Title: "Logs & Monitoring",
			Items: []MenuItem{
				{ID: "1", Label: "Tail latest logs", Action: "logs_tail"},
				{ID: "2", Label: "Stream command output", Action: "logs_stream"},
				{ID: "3", Label: "Show HTTP monitor status", Action: "monitors_status"},
				{ID: "4", Label: "Check system health", Action: "status"},
				{ID: "0", Label: "Back to main menu", Submenu: "main"},
			},
		},
		"cli": {
			Title: "CLI Sessions",
			Items: []MenuItem{
				{ID: "1", Label: "List sessions", Action: "cli_list"},
				{ID: "2", Label: "Open new session", Action: "cli_new"},
				{ID: "3", Label: "Send command to session", Action: "cli_run"},
				{ID: "4", Label: "Tail session output", Action: "cli_tail"},
				{ID: "0", Label: "Back to main menu", Submenu: "main"},
			},
		},
		"server": {
			Title: "Server & Config",
			Items: []MenuItem{
				{ID: "1", Label: "Status", Action: "status"},
				{ID: "2", Label: "Config summary", Action: "config_status"},
				{ID: "3", Label: "Add API key", Action: "add_api"},
				{ID: "4", Label: "Connect Gemini CLI", Action: "connect_gemini"},
				{ID: "0", Label: "Back to main menu", Submenu: "main"},
			},
		},
		"help": {
			Title: "Help / Examples",
			Items: []MenuItem{
				{ID: "1", Label: "Command formats", Action: "help"},
				{ID: "2", Label: "Slash commands", Action: "help_slash"},
				{ID: "0", Label: "Back to main menu", Submenu: "main"},
			},
		},
	}
}

var (
	menusMu   sync.RWMutex
	menusCfg  MenusConfig
	menusInit bool
)

// LoadMenus loads menus from file if it exists, otherwise returns defaults.
func LoadMenus(menusPath string) MenusConfig {
	if menusPath != "" {
		data, err := os.ReadFile(menusPath)
		if err == nil {
			var cfg MenusConfig
			if json.Unmarshal(data, &cfg) == nil && len(cfg) > 0 {
				return cfg
			}
		}
	}
	// Try ~/.sypher-mini/menus.json
	home, _ := os.UserHomeDir()
	if home != "" {
		p := filepath.Join(home, ".sypher-mini", "menus.json")
		data, err := os.ReadFile(p)
		if err == nil {
			var cfg MenusConfig
			if json.Unmarshal(data, &cfg) == nil && len(cfg) > 0 {
				return cfg
			}
		}
	}
	return DefaultMenus()
}

// GetMenus returns the current menus config, initializing with defaults if needed.
func GetMenus(menusPath string) MenusConfig {
	menusMu.Lock()
	defer menusMu.Unlock()
	if !menusInit {
		menusCfg = LoadMenus(menusPath)
		menusInit = true
	}
	return menusCfg
}

// RenderMenu returns the menu as formatted text for WhatsApp.
func RenderMenu(cfg MenusConfig, menuID string) string {
	m, ok := cfg[menuID]
	if !ok {
		return "Menu not found."
	}
	var b string
	if m.Title != "" {
		b = "*" + m.Title + "*\n\n"
	}
	b += "Reply with a number:\n\n"
	for _, item := range m.Items {
		emoji := ""
		switch item.ID {
		case "1":
			emoji = "1️⃣ "
		case "2":
			emoji = "2️⃣ "
		case "3":
			emoji = "3️⃣ "
		case "4":
			emoji = "4️⃣ "
		case "5":
			emoji = "5️⃣ "
		case "6":
			emoji = "6️⃣ "
		case "0":
			emoji = "0️⃣ "
		default:
			emoji = item.ID + " "
		}
		b += "*" + emoji + item.Label + "*\n"
	}
	return b
}
