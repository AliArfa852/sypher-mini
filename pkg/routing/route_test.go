package routing

import (
	"testing"

	"github.com/sypherexx/sypher-mini/pkg/config"
)

func TestResolve_Default(t *testing.T) {
	cfg := config.DefaultConfig()
	route := Resolve(cfg, RouteInput{Channel: "whatsapp", AccountID: "+123"})
	if route.AgentID != "main" {
		t.Errorf("expected main, got %s", route.AgentID)
	}
}

func TestResolve_PeerMatch(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Bindings = []config.AgentBinding{
		{AgentID: "coding", Match: config.BindingMatch{
			Channel: "whatsapp",
			Peer:   &config.PeerMatch{Kind: "direct", ID: "+999"},
		}},
		{AgentID: "main", Match: config.BindingMatch{Channel: "whatsapp", AccountID: "*"}},
	}
	cfg.Agents.List = append(cfg.Agents.List, config.AgentConfig{ID: "coding"})
	route := Resolve(cfg, RouteInput{
		Channel:   "whatsapp",
		AccountID: "+999",
		Peer:      &PeerMatch{Kind: "direct", ID: "+999"},
	})
	if route.AgentID != "coding" {
		t.Errorf("expected coding for peer match, got %s", route.AgentID)
	}
}
