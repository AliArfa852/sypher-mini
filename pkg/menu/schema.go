package menu

// MenuItem defines a single menu option.
type MenuItem struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Submenu string `json:"submenu,omitempty"`
	Action  string `json:"action,omitempty"`
	Agent   string `json:"agent,omitempty"`
}

// MenuDef defines a menu with title and items.
type MenuDef struct {
	Title string     `json:"title"`
	Items []MenuItem `json:"items"`
}

// MenusConfig holds all menu definitions keyed by menu ID.
type MenusConfig map[string]MenuDef
