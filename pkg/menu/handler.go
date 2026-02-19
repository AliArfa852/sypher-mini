package menu

import (
	"context"
	"strings"

	"github.com/sypherexx/sypher-mini/pkg/bus"
	"github.com/sypherexx/sypher-mini/pkg/config"
)

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

	// 1. Trigger words: menu or /help only (avoid false activation)
	if lower == "menu" || lower == "/help" {
		h.store.ResetToMain(key)
		response = RenderMenu(h.menus, "main")
		response = "👋 " + response + "\n\n_Type a number or say 'sypher' + your request._"
		return true, response
	}

	// 2. Numeric input when in menu session
	if ses, ok := h.store.Get(key); ok && ses != nil {
		// 0 or "back" -> parent or main
		if lower == "0" || lower == "back" {
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
