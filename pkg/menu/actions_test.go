package menu

import (
	"context"
	"strings"
	"testing"

	"github.com/sypherexx/sypher-mini/pkg/bus"
	"github.com/sypherexx/sypher-mini/pkg/config"
)

func TestExecuteAction_Help(t *testing.T) {
	resp, err := ExecuteAction(context.Background(), "help", config.DefaultConfig(), nil, bus.InboundMessage{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp, "Menu") || !strings.Contains(resp, "sypher") {
		t.Errorf("help response: %s", resp)
	}
}

func TestExecuteAction_AddAPI(t *testing.T) {
	resp, err := ExecuteAction(context.Background(), "add_api", config.DefaultConfig(), nil, bus.InboundMessage{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp, "API") || !strings.Contains(resp, "config") {
		t.Errorf("add_api response: %s", resp)
	}
}

func TestExecuteAction_ConnectGemini(t *testing.T) {
	resp, err := ExecuteAction(context.Background(), "connect_gemini", config.DefaultConfig(), nil, bus.InboundMessage{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp, "Gemini") {
		t.Errorf("connect_gemini response: %s", resp)
	}
}

func TestRenderMenu(t *testing.T) {
	cfg := DefaultMenus()
	resp := RenderMenu(cfg, "main")
	if !strings.Contains(resp, "Control Panel") || !strings.Contains(resp, "Projects") {
		t.Errorf("RenderMenu main: %s", resp)
	}
}
