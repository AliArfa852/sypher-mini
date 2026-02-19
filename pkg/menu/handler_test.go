package menu

import (
	"context"
	"strings"
	"testing"

	"github.com/sypherexx/sypher-mini/pkg/bus"
	"github.com/sypherexx/sypher-mini/pkg/config"
)

func TestHandler_TriggerMenu(t *testing.T) {
	h := NewHandler(config.DefaultConfig(), nil, "")
	ctx := context.Background()
	msg := bus.InboundMessage{Channel: "whatsapp", ChatID: "+123", SenderID: "+123", Content: "menu"}

	handled, resp := h.Handle(ctx, msg)
	if !handled {
		t.Fatal("menu trigger should be handled")
	}
	if !strings.Contains(resp, "Control Panel") || !strings.Contains(resp, "Projects") {
		t.Errorf("response missing expected content: %s", resp)
	}
}

func TestHandler_TriggerHelp(t *testing.T) {
	h := NewHandler(config.DefaultConfig(), nil, "")
	ctx := context.Background()
	msg := bus.InboundMessage{Channel: "whatsapp", ChatID: "+123", SenderID: "+123", Content: "/help"}

	handled, resp := h.Handle(ctx, msg)
	if !handled {
		t.Fatal("/help trigger should be handled")
	}
	if !strings.Contains(resp, "Control Panel") {
		t.Errorf("response missing expected content: %s", resp)
	}
}

func TestHandler_NotTrigger(t *testing.T) {
	h := NewHandler(config.DefaultConfig(), nil, "")
	ctx := context.Background()
	msg := bus.InboundMessage{Channel: "whatsapp", ChatID: "+123", SenderID: "+123", Content: "hello"}

	handled, _ := h.Handle(ctx, msg)
	if handled {
		t.Error("hello should not trigger menu")
	}
}

func TestHandler_NumericInMenu(t *testing.T) {
	h := NewHandler(config.DefaultConfig(), nil, "")
	ctx := context.Background()

	// First trigger menu
	h.Handle(ctx, bus.InboundMessage{Channel: "whatsapp", ChatID: "+123", SenderID: "+123", Content: "menu"})

	// Then send 4 (CLI Sessions)
	msg := bus.InboundMessage{Channel: "whatsapp", ChatID: "+123", SenderID: "+123", Content: "4"}
	handled, resp := h.Handle(ctx, msg)
	if !handled {
		t.Fatal("numeric 4 in menu should be handled")
	}
	if !strings.Contains(resp, "CLI Sessions") {
		t.Errorf("response missing CLI submenu: %s", resp)
	}
}

func TestHandler_NumericNoSession(t *testing.T) {
	h := NewHandler(config.DefaultConfig(), nil, "")
	ctx := context.Background()
	// Send 1 without first triggering menu - should not be handled
	msg := bus.InboundMessage{Channel: "whatsapp", ChatID: "+123", SenderID: "+123", Content: "1"}
	handled, _ := h.Handle(ctx, msg)
	if handled {
		t.Error("numeric without menu session should not be handled")
	}
}

func TestHandler_BackToMain(t *testing.T) {
	h := NewHandler(config.DefaultConfig(), nil, "")
	ctx := context.Background()

	h.Handle(ctx, bus.InboundMessage{Channel: "whatsapp", ChatID: "+123", SenderID: "+123", Content: "menu"})
	h.Handle(ctx, bus.InboundMessage{Channel: "whatsapp", ChatID: "+123", SenderID: "+123", Content: "4"})

	msg := bus.InboundMessage{Channel: "whatsapp", ChatID: "+123", SenderID: "+123", Content: "0"}
	handled, resp := h.Handle(ctx, msg)
	if !handled {
		t.Fatal("0 (back) should be handled")
	}
	if !strings.Contains(resp, "Control Panel") {
		t.Errorf("back should show main menu: %s", resp)
	}
}

func TestHandler_NonWhatsApp(t *testing.T) {
	h := NewHandler(config.DefaultConfig(), nil, "")
	ctx := context.Background()
	msg := bus.InboundMessage{Channel: "cli", ChatID: "1", Content: "menu"}

	handled, _ := h.Handle(ctx, msg)
	if handled {
		t.Error("menu on non-WhatsApp channel should not be handled")
	}
}

func TestHandler_EasterEggs(t *testing.T) {
	h := NewHandler(config.DefaultConfig(), nil, "")
	ctx := context.Background()

	tests := []struct {
		content  string
		wantSub  string
	}{
		{"42", "42"},
		{"sudo", "maximum enthusiasm"},
		{"joke", "arrays"},
		{"hello world", "Hello, World"},
		{"roll dice", "Quick roll"},
	}
	for _, tt := range tests {
		msg := bus.InboundMessage{Channel: "whatsapp", ChatID: "+123", SenderID: "+123", Content: tt.content}
		handled, resp := h.Handle(ctx, msg)
		if !handled {
			t.Errorf("easter egg %q should be handled", tt.content)
		}
		if !strings.Contains(resp, tt.wantSub) {
			t.Errorf("easter egg %q: response missing %q: %s", tt.content, tt.wantSub, resp)
		}
	}
}
