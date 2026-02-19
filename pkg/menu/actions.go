package menu

import (
	"context"
	"fmt"

	"github.com/sypherexx/sypher-mini/pkg/bus"
	"github.com/sypherexx/sypher-mini/pkg/config"
)

// ActionRunner runs menu actions (CLI, status, etc.). Implemented by agent.Loop.
type ActionRunner interface {
	RunCliList(ctx context.Context, msg bus.InboundMessage) (string, error)
	RunCliNew(ctx context.Context, tag string, msg bus.InboundMessage) (string, error)
	RunStatus(ctx context.Context, msg bus.InboundMessage) (string, error)
	RunConfigStatus(ctx context.Context, msg bus.InboundMessage) (string, error)
}

// ExecuteAction runs the given action and returns the response.
func ExecuteAction(ctx context.Context, actionID string, cfg *config.Config, runner ActionRunner, msg bus.InboundMessage) (string, error) {
	switch actionID {
	case "cli_run", "cli_tail":
		return "Reply with: /cli list to see sessions, then /cli run <N> <command> or /cli <N>", nil
	case "add_api":
		return `*Add API key*

Reply with: /config set providers.X.api_key YOUR_KEY (admin only)

Replace X with: cerebras, openai, anthropic, or gemini.`, nil
	case "connect_gemini":
		return `*Connect Gemini CLI*

1. Install Gemini CLI: https://ai.google.dev/gemini-api/docs/cli
2. Add to config (~/.sypher-mini/config.json):

{
  "agents": {
    "list": [
      { "id": "gemini-cli", "command": "gemini", "args": ["--model", "gemini-2.0"] }
    ]
  }
}

3. Add gemini to tools.live_monitoring.allowed_commands`, nil
	case "help", "help_slash":
		return helpText(), nil
	case "projects_list", "projects_open", "projects_build", "projects_pull":
		return "Say 'sypher' + your request to use the agent for project tasks.", nil
	case "tasks_create", "tasks_list", "tasks_authorize", "tasks_cancel":
		return "Say 'sypher' + your request to use the agent for task management.", nil
	case "logs_tail", "logs_stream":
		return "Say 'sypher tail <file>' or 'sypher stream <command>' to use the agent.", nil
	case "cli_list":
		if runner != nil {
			return runner.RunCliList(ctx, msg)
		}
		return "Use /cli list to see sessions.", nil
	case "cli_new":
		if runner != nil {
			return runner.RunCliNew(ctx, "unnamed", msg)
		}
		return "Use /cli new -m \"tag\" to create a session.", nil
	case "status":
		if runner != nil {
			return runner.RunStatus(ctx, msg)
		}
		return "Use /status or say 'sypher status'.", nil
	case "config_status", "monitors_status":
		if runner != nil {
			return runner.RunConfigStatus(ctx, msg)
		}
		return "Use /status for status. Use /agents for agents list (operator).", nil
	default:
		return preDesignedAction(actionID, cfg), nil
	}
}

func preDesignedAction(actionID string, cfg *config.Config) string {
	switch actionID {
	case "add_api":
		return "Add API: use /config set providers.X.api_key YOUR_KEY (admin only)"
	case "connect_gemini":
		return "Add agents.list with command/args for gemini. See docs."
	default:
		return fmt.Sprintf("Action %q not implemented. Say 'sypher' + your request.", actionID)
	}
}

func helpText() string {
	return `*Help*

*Menu:* Type menu or /help for options.

*Agent:* Say sypher + your request (e.g. "sypher create a hello world script").

*Slash commands:*
• /status - Status
• /cli list - List CLI sessions
• /cli new -m "tag" - New session
• /config get <path> - Get config (operator)
• /agents - List agents (operator)`
}
