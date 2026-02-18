package intent

import (
	"strings"

	"github.com/sypherexx/sypher-mini/pkg/config"
)

// WhatsAppTier is the authorization tier for WhatsApp commands.
type WhatsAppTier string

const (
	TierUser     WhatsAppTier = "user"
	TierOperator WhatsAppTier = "operator"
	TierAdmin    WhatsAppTier = "admin"
)

// TierLevel returns numeric level for comparison.
func TierLevel(t WhatsAppTier) int {
	switch t {
	case TierAdmin:
		return 3
	case TierOperator:
		return 2
	case TierUser:
		return 1
	default:
		return 0
	}
}

// ParseWhatsAppCommand parses inbound WhatsApp message for commands.
// Returns (isCommand, command, args, tier).
func ParseWhatsAppCommand(content string, from string, cfg *config.ChannelsConfig) (bool, string, []string, WhatsAppTier) {
	content = strings.TrimSpace(content)
	if content == "" {
		return false, "", nil, TierUser
	}

	wc := cfg.WhatsApp
	tier := resolveTier(from, wc.AllowFrom, wc.Operators, wc.Admins)
	if tier == "" {
		return false, "", nil, TierUser
	}

	// Command prefixes: /config, /agents, /monitors, /audit, /status
	lower := strings.ToLower(content)
	var cmd string
	var args []string

	if strings.HasPrefix(lower, "/config ") || strings.HasPrefix(lower, "config ") {
		parts := strings.Fields(content)
		if len(parts) >= 2 {
			cmd = "config"
			args = parts[1:]
		}
	} else if strings.HasPrefix(lower, "/agents") || lower == "agents" {
		cmd = "agents"
		args = strings.Fields(content)[1:]
	} else if strings.HasPrefix(lower, "/monitors") || lower == "monitors" {
		cmd = "monitors"
		args = strings.Fields(content)[1:]
	} else if strings.HasPrefix(lower, "/audit ") || strings.HasPrefix(lower, "audit ") {
		parts := strings.Fields(content)
		if len(parts) >= 2 {
			cmd = "audit"
			args = parts[1:]
		}
	} else if strings.HasPrefix(lower, "/status") || lower == "status" {
		cmd = "status"
	}

	if cmd == "" {
		return false, "", nil, tier
	}

	// Check tier for command - if insufficient, don't treat as command (let agent handle)
	switch cmd {
	case "config":
		if len(args) >= 1 {
			if args[0] == "get" && TierLevel(tier) >= TierLevel(TierOperator) {
				return true, cmd, args, tier
			}
			if args[0] == "set" && TierLevel(tier) >= TierLevel(TierAdmin) {
				return true, cmd, args, tier
			}
		}
	case "agents", "monitors":
		if TierLevel(tier) >= TierLevel(TierOperator) {
			return true, cmd, args, tier
		}
	case "audit":
		if TierLevel(tier) >= TierLevel(TierAdmin) {
			return true, cmd, args, tier
		}
	case "status":
		if TierLevel(tier) >= TierLevel(TierUser) {
			return true, cmd, args, tier
		}
	}

	// Command recognized but tier insufficient - still return as command so we can say "access denied"
	return true, cmd, args, tier
}

func resolveTier(from string, allowFrom, operators, admins []string) WhatsAppTier {
	if !contains(allowFrom, from) && len(allowFrom) > 0 {
		return ""
	}
	if len(allowFrom) == 0 {
		// No allow list = allow all as user
	}
	if contains(admins, from) {
		return TierAdmin
	}
	if contains(operators, from) {
		return TierOperator
	}
	return TierUser
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
