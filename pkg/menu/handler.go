package menu

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/sypherexx/sypher-mini/pkg/bus"
	"github.com/sypherexx/sypher-mini/pkg/config"
)

// easterEgg returns a fun response for magic phrases, or "" if none matched.
func easterEgg(lower string) string {
	switch lower {
	case "42":
		return "🤖 *42* — The answer to life, the universe, and everything. (Also a valid menu choice if we had 42 options. We don't. Try 1–6.)"
	case "sudo":
		return "🔐 *sudo* — Nice try. I'm already running with maximum enthusiasm. Try `menu` for the real controls."
	case "coffee", "tea":
		return "☕ I'd love some too. While we wait, type `menu` to see what I can actually do for you."
	case "joke", "tell me a joke", "make me laugh":
		return "😄 Why did the developer quit? Because they didn't get arrays. (Type `menu` for the good stuff.)"
	case "hello world":
		return "🌍 Hello, World! Now type `menu` and let's build something real."
	case "beep", "boop":
		return "🤖 Beep boop. I'm fully operational. Type `menu` when you're ready."
	case "help me":
		return "🆘 I've got you! Type `menu` or `/help` for the full control panel."
	case "dice", "roll dice", "roll the dice":
		return fmt.Sprintf("🎲 *Quick roll:* You got *%d* (1d6)", rand.Intn(6)+1)
	default:
		return ""
	}
}

// Handler handles WhatsApp menu workflow messages.
type Handler struct {
	store   *SessionStore
	menus   MenusConfig
	runner  ActionRunner
	cfg     *config.Config
	menusPath string
}

// NewHandler creates a new menu handler.
func NewHandler(cfg *config.Config, runner ActionRunner, menusPath string) *Handler {
	return &Handler{
		store:     NewSessionStore(0),
		menus:     GetMenus(menusPath),
		runner:    runner,
		cfg:       cfg,
		menusPath: menusPath,
	}
}

// Handle processes a WhatsApp message. Returns (handled, response).
// handled=true means the menu workflow handled it; response is the reply.
func (h *Handler) Handle(ctx context.Context, msg bus.InboundMessage) (handled bool, response string) {
	if msg.Channel != "whatsapp" {
		return false, ""
	}
	content := strings.TrimSpace(msg.Content)
	lower := strings.ToLower(content)
	key := Key(msg.Channel, msg.ChatID)

	// 0. Easter eggs — fun responses for magic phrases
	if egg := easterEgg(lower); egg != "" {
		return true, egg
	}

	// 1. Trigger words: menu or /help only (avoid false activation)
	if lower == "menu" || lower == "/help" {
		h.store.ResetToMain(key)
		response = RandomTagline() + "👋 " + RenderMenu(h.menus, "main") + "\n\n" + RandomFooter()
		return true, response
	}

	// 1b. Single digit 1-7 with no session: treat as main menu shortcut (4 -> CLI, 1 -> Projects, etc.)
	if ses, ok := h.store.Get(key); !ok || ses == nil {
		if n, okNum := parseMenuNumber(content); okNum && n >= 1 && n <= 7 {
			mainMenu, hasMain := h.menus["main"]
			if hasMain && n <= len(mainMenu.Items) {
				item := mainMenu.Items[n-1]
				if item.Submenu != "" {
					h.store.Set(key, item.Submenu, "main")
					response = RenderMenu(h.menus, item.Submenu)
					return true, response
				}
				if item.Action != "" {
					res, err := ExecuteAction(ctx, item.Action, h.cfg, h.runner, msg)
					if err != nil {
						response = "Error: " + err.Error()
					} else {
						response = res
						if h.runner != nil {
							if item.Action == "projects_build" || item.Action == "projects_pull" {
								if ids, _ := h.runner.RunProjectsGetIDs(ctx); len(ids) > 0 {
									h.store.SetPendingProjectAction(key, item.Action, ids)
								}
							} else if item.Action == "tasks_cancel" {
								if ids, _ := h.runner.RunTasksGetIDs(ctx); len(ids) > 0 {
									h.store.SetPendingTaskCancel(key, ids)
								}
							}
						}
					}
					h.store.Set(key, "main", "")
					return true, response
				}
			}
		}
	}

	// 2. Numeric input when in menu session
	if ses, ok := h.store.Get(key); ok && ses != nil {
		// Check pending selection first (build/pull/cancel)
		if action, projectIDs, taskIDs, ok := h.store.GetPendingProjectAction(key); ok && h.runner != nil {
			if n, okNum := parseMenuNumber(content); okNum && n >= 1 {
				h.store.ClearPendingProjectAction(key)
				var res string
				var err error
				if action == "projects_build" && n <= len(projectIDs) {
					res, err = h.runner.RunProjectsBuild(ctx, projectIDs[n-1], msg)
				} else if action == "projects_pull" && n <= len(projectIDs) {
					res, err = h.runner.RunProjectsPull(ctx, projectIDs[n-1], msg)
				} else if action == "tasks_cancel" && n <= len(taskIDs) {
					res, err = h.runner.RunTasksCancel(ctx, taskIDs[n-1], msg)
				} else {
					h.store.Set(key, ses.CurrentMenu, ses.ParentMenu)
					return false, ""
				}
				if err != nil {
					response = "Error: " + err.Error()
				} else {
					response = res
				}
				h.store.Set(key, ses.CurrentMenu, ses.ParentMenu)
				return true, response
			}
		}

		// 0 or "back" -> parent or main
		if lower == "0" || lower == "back" {
			h.store.ClearPendingProjectAction(key)
			parent := ses.ParentMenu
			if parent == "" {
				parent = "main"
			}
			h.store.Set(key, parent, "")
			response = RenderMenu(h.menus, parent)
			return true, response
		}

		// 1-9: find matching item
		menuDef, ok := h.menus[ses.CurrentMenu]
		if !ok {
			h.store.Clear(key)
			return false, ""
		}
		for _, item := range menuDef.Items {
			if item.ID == content || (len(content) == 1 && item.ID == content) {
				// Submenu
				if item.Submenu != "" {
					h.store.Set(key, item.Submenu, ses.CurrentMenu)
					response = RenderMenu(h.menus, item.Submenu)
					return true, response
				}
				// Action
				if item.Action != "" {
					res, err := ExecuteAction(ctx, item.Action, h.cfg, h.runner, msg)
					if err != nil {
						response = "Error: " + err.Error()
					} else {
						response = res
						// Set pending state for project build/pull or task cancel (user will reply with number)
						if h.runner != nil {
							if item.Action == "projects_build" || item.Action == "projects_pull" {
								if ids, _ := h.runner.RunProjectsGetIDs(ctx); len(ids) > 0 {
									h.store.SetPendingProjectAction(key, item.Action, ids)
								}
							} else if item.Action == "tasks_cancel" {
								if ids, _ := h.runner.RunTasksGetIDs(ctx); len(ids) > 0 {
									h.store.SetPendingTaskCancel(key, ids)
								}
							}
						}
					}
					// Stay in same menu
					h.store.Set(key, ses.CurrentMenu, ses.ParentMenu)
					return true, response
				}
				break
			}
		}
		// Numeric but no match - maybe user meant something else, fall through to agent
	}

	return false, ""
}

// parseMenuNumber returns (n, true) if content is a positive number.
func parseMenuNumber(content string) (int, bool) {
	n, err := strconv.Atoi(strings.TrimSpace(content))
	if err != nil {
		return 0, false
	}
	return n, n >= 1
}
