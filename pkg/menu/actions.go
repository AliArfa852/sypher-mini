package menu

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/sypherexx/sypher-mini/pkg/bus"
	"github.com/sypherexx/sypher-mini/pkg/config"
)

// ActionRunner runs menu actions (CLI, status, projects, etc.). Implemented by agent.Loop.
type ActionRunner interface {
	RunCliList(ctx context.Context, msg bus.InboundMessage) (string, error)
	RunCliNew(ctx context.Context, tag string, msg bus.InboundMessage) (string, error)
	RunStatus(ctx context.Context, msg bus.InboundMessage) (string, error)
	RunConfigStatus(ctx context.Context, msg bus.InboundMessage) (string, error)
	RunProjectsList(ctx context.Context, msg bus.InboundMessage) (string, error)
	RunProjectsBuild(ctx context.Context, projectID string, msg bus.InboundMessage) (string, error)
	RunProjectsPull(ctx context.Context, projectID string, msg bus.InboundMessage) (string, error)
	RunProjectsGetIDs(ctx context.Context) ([]string, error)
	RunTasksList(ctx context.Context, msg bus.InboundMessage) (string, error)
	RunTasksCancel(ctx context.Context, taskID string, msg bus.InboundMessage) (string, error)
	RunTasksGetIDs(ctx context.Context) ([]string, error)
}

// ExecuteAction runs the given action and returns the response.
func ExecuteAction(ctx context.Context, actionID string, cfg *config.Config, runner ActionRunner, msg bus.InboundMessage) (string, error) {
	switch actionID {
	case "cli_run":
		if runner != nil {
			list, _ := runner.RunCliList(ctx, msg)
			if strings.Contains(list, "No active") {
				return list + "\n\n_Create a session first (option 2)._", nil
			}
			return "*Send command to session:*\n\n" + list + "\n\n_Reply: N <command> (e.g. 1 ls -a) or use /cli run N <cmd>_", nil
		}
		return "Reply with: /cli run <N> <command>", nil
	case "cli_tail":
		if runner != nil {
			list, _ := runner.RunCliList(ctx, msg)
			if strings.Contains(list, "No active") {
				return list + "\n\n_Create a session first (option 2)._", nil
			}
			return "*Tail session output:*\n\n" + list + "\n\n_Reply with number or use /cli N [--tail 20]_", nil
		}
		return "Reply with: /cli <N> [--tail 20]", nil
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
	case "projects_list":
		if runner != nil {
			return runner.RunProjectsList(ctx, msg)
		}
		return "No project runner. Say 'sypher list projects'.", nil
	case "projects_add":
		return "*Add project*\n\nCreate a JSON file in:\n~/.sypher-mini/workspace/code-projects/<id>.json\n\nExample:\n" +
			"{\"id\":\"myapp\",\"name\":\"My App\",\"path\":\"myapp\",\"build_command\":\"npm run build\"}\n\n" +
			"Path is relative to workspace. Then use \"List active projects\" to see it.", nil
	case "projects_open":
		if runner != nil {
			list, _ := runner.RunProjectsList(ctx, msg)
			return list + "\n\n_To open: say 'sypher open project-name' or use /cli._", nil
		}
		return "Say 'sypher' + your request.", nil
	case "projects_build":
		// Empty projectID is intentional: shows list and sets pending for numeric reply
		if runner != nil {
			return runner.RunProjectsBuild(ctx, "", msg)
		}
		return "Say 'sypher build <project>' for agent.", nil
	case "projects_pull":
		// Empty projectID is intentional: shows list and sets pending for numeric reply
		if runner != nil {
			return runner.RunProjectsPull(ctx, "", msg)
		}
		return "Say 'sypher pull <project>' for agent.", nil
	case "tasks_list":
		if runner != nil {
			return runner.RunTasksList(ctx, msg)
		}
		return "Say 'sypher status' for task info.", nil
	case "tasks_cancel":
		if runner != nil {
			return runner.RunTasksCancel(ctx, "", msg)
		}
		return "Use /cancel <task_id> or say 'sypher cancel <task_id>'.", nil
	case "tasks_create", "tasks_authorize":
		return "Say 'sypher' + your task to create, or 'sypher authorize' for pending tasks.", nil
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
	case "roll_1d6":
		return fmt.Sprintf("🎲 *1d6:* You rolled *%d*", rand.Intn(6)+1), nil
	case "roll_2d6":
		a, b := rand.Intn(6)+1, rand.Intn(6)+1
		return fmt.Sprintf("🎲 *2d6:* You rolled *%d* + *%d* = *%d*", a, b, a+b), nil
	case "roll_1d20":
		return fmt.Sprintf("🎲 *1d20:* You rolled *%d*", rand.Intn(20)+1), nil
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
• /cli run <N> <command> - Run in session (e.g. /cli run 1 ls -a)
• /cli <N> [--tail 20] - Tail session output
• /projects list - List projects
• /projects build <id> - Build project
• /projects pull <id> - Git pull project
• /projects run <id> - Run/deploy project
• /projects add <path> - Register project
• /projects scan - Detect projects in workspace
• /config get <path> - Get config (operator)
• /agents - List agents (operator)
• /cancel <task_id> - Cancel task`
}
