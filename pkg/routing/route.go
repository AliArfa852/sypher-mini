package routing

import (
	"strings"

	"github.com/sypherexx/sypher-mini/pkg/config"
)

const DefaultAgentID = "main"

// RouteInput contains the routing context from an inbound message.
type RouteInput struct {
	Channel   string
	AccountID string
	Peer      *PeerMatch
}

// PeerMatch represents a chat peer.
type PeerMatch struct {
	Kind string
	ID   string
}

// ResolvedRoute is the result of agent routing.
type ResolvedRoute struct {
	AgentID   string
	SessionKey string
	MatchedBy string
}

// Resolve determines which agent handles the message.
func Resolve(cfg *config.Config, input RouteInput) ResolvedRoute {
	channel := strings.ToLower(strings.TrimSpace(input.Channel))
	accountID := strings.TrimSpace(input.AccountID)
	if accountID == "" {
		accountID = "default"
	}

	// Priority: peer > account > channel wildcard > default
	bindings := filterBindings(cfg.Bindings, channel)

	if input.Peer != nil && strings.TrimSpace(input.Peer.ID) != "" {
		if m := findPeerMatch(bindings, input.Peer); m != nil {
			return ResolvedRoute{
				AgentID:   pickAgentID(cfg, m.AgentID),
				SessionKey: buildSessionKey(m.AgentID, channel, accountID, input.Peer),
				MatchedBy: "binding.peer",
			}
		}
	}

	if accountID != "" && accountID != "default" {
		if m := findAccountMatch(bindings, accountID); m != nil {
			return ResolvedRoute{
				AgentID:   pickAgentID(cfg, m.AgentID),
				SessionKey: buildSessionKey(m.AgentID, channel, accountID, input.Peer),
				MatchedBy: "binding.account",
			}
		}
	}

	if m := findChannelWildcard(bindings); m != nil {
		return ResolvedRoute{
			AgentID:   pickAgentID(cfg, m.AgentID),
			SessionKey: buildSessionKey(m.AgentID, channel, accountID, input.Peer),
			MatchedBy: "binding.channel",
		}
	}

	agentID := resolveDefaultAgentID(cfg)
	return ResolvedRoute{
		AgentID:   agentID,
		SessionKey: buildSessionKey(agentID, channel, accountID, input.Peer),
		MatchedBy: "default",
	}
}

func filterBindings(bindings []config.AgentBinding, channel string) []config.AgentBinding {
	var out []config.AgentBinding
	for _, b := range bindings {
		mc := strings.ToLower(strings.TrimSpace(b.Match.Channel))
		if mc == "" || mc == channel {
			out = append(out, b)
		}
	}
	return out
}

func findPeerMatch(bindings []config.AgentBinding, peer *PeerMatch) *config.AgentBinding {
	for i := range bindings {
		b := &bindings[i]
		if b.Match.Peer == nil {
			continue
		}
		if strings.EqualFold(b.Match.Peer.Kind, peer.Kind) && b.Match.Peer.ID == peer.ID {
			return b
		}
	}
	return nil
}

func findAccountMatch(bindings []config.AgentBinding, accountID string) *config.AgentBinding {
	for i := range bindings {
		b := &bindings[i]
		if b.Match.AccountID == "*" {
			continue
		}
		if b.Match.Peer != nil {
			continue
		}
		if strings.EqualFold(b.Match.AccountID, accountID) {
			return &bindings[i]
		}
	}
	return nil
}

func findChannelWildcard(bindings []config.AgentBinding) *config.AgentBinding {
	for i := range bindings {
		b := &bindings[i]
		if b.Match.AccountID != "*" {
			continue
		}
		if b.Match.Peer != nil {
			continue
		}
		return &bindings[i]
	}
	return nil
}

func pickAgentID(cfg *config.Config, id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return resolveDefaultAgentID(cfg)
	}
	// Verify agent exists
	for _, a := range cfg.Agents.List {
		if strings.EqualFold(a.ID, id) {
			return a.ID
		}
	}
	return resolveDefaultAgentID(cfg)
}

func resolveDefaultAgentID(cfg *config.Config) string {
	for _, a := range cfg.Agents.List {
		if a.Default {
			if id := strings.TrimSpace(a.ID); id != "" {
				return id
			}
		}
	}
	if len(cfg.Agents.List) > 0 {
		if id := strings.TrimSpace(cfg.Agents.List[0].ID); id != "" {
			return id
		}
	}
	return DefaultAgentID
}

func buildSessionKey(agentID, channel, accountID string, peer *PeerMatch) string {
	parts := []string{"agent", agentID, channel}
	if accountID != "" {
		parts = append(parts, accountID)
	}
	if peer != nil && peer.ID != "" {
		parts = append(parts, peer.Kind, peer.ID)
	}
	return strings.Join(parts, ":")
}
