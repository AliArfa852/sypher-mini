package agent

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sypherexx/sypher-mini/pkg/audit"
	"github.com/sypherexx/sypher-mini/pkg/bus"
	"github.com/sypherexx/sypher-mini/pkg/clisession"
	"github.com/sypherexx/sypher-mini/pkg/config"
	"github.com/sypherexx/sypher-mini/pkg/idempotency"
	"github.com/sypherexx/sypher-mini/pkg/intent"
	"github.com/sypherexx/sypher-mini/pkg/observability"
	"github.com/sypherexx/sypher-mini/pkg/process"
	"github.com/sypherexx/sypher-mini/pkg/providers"
	"github.com/sypherexx/sypher-mini/pkg/routing"
	"github.com/sypherexx/sypher-mini/pkg/task"
	"github.com/sypherexx/sypher-mini/pkg/tools"
	"github.com/sypherexx/sypher-mini/pkg/policy"
	"github.com/sypherexx/sypher-mini/pkg/platform"
	"github.com/sypherexx/sypher-mini/pkg/replay"
)

// Loop is the main agent loop that processes inbound messages.
type Loop struct {
	cfg         *config.Config
	msgBus      *bus.MessageBus
	eventBus    *bus.Bus
	taskMgr     *task.Manager
	provider    providers.LLMProvider
	execTool         *tools.ExecTool
	killTool         *tools.KillTool
	webFetch         *tools.WebFetchTool
	messageTool      *tools.MessageTool
	tailOutput       *tools.TailOutputTool
	streamCommand    *tools.StreamCommandTool
	invokeCliAgent   *tools.InvokeCliAgentTool
	cliManager      *clisession.Manager
	metrics         *observability.Metrics
	auditLogger *audit.Logger
	procTracker *process.Tracker
	policyEval  *policy.Evaluator
	replayWriter  *replay.Writer
	idempotency   *idempotency.Cache
	safeMode      bool
	running       atomic.Bool
}

// LoopOptions configures the agent loop.
type LoopOptions struct {
	SafeMode bool
}

// NewLoop creates a new agent loop.
func NewLoop(cfg *config.Config, msgBus *bus.MessageBus, eventBus *bus.Bus, opts *LoopOptions) *Loop {
	if opts == nil {
		opts = &LoopOptions{}
	}
	taskMgr := task.NewManager(cfg.Task.TimeoutSec)
	fb := providers.NewFallbackProvider(cfg)
	var provider providers.LLMProvider = fb
	if len(fb.Entries()) == 0 {
		provider = nil
	}

	auditDir := cfg.Audit.Dir
	if auditDir == "" {
		auditDir = config.ExpandPath("~/.sypher-mini/audit")
	}
	integrity := cfg.Audit.Integrity
	if integrity == "" {
		integrity = "none"
	}
	auditLogger := audit.NewWithIntegrity(auditDir, integrity)
	procTracker := process.New()
	execTool := tools.NewExecTool(cfg, auditLogger, procTracker, opts.SafeMode)
	killTool := tools.NewKillTool(procTracker, opts.SafeMode)
	policyEval := policy.NewEvaluator(cfg)
	webFetch := tools.NewWebFetchTool(cfg, policyEval, opts.SafeMode)
	messageTool := tools.NewMessageTool(msgBus, opts.SafeMode)
	tailOutput := tools.NewTailOutputTool(cfg, opts.SafeMode)
	streamCommand := tools.NewStreamCommandTool(cfg, msgBus, messageTool, opts.SafeMode)
	invokeCliAgent := tools.NewInvokeCliAgentTool(cfg, opts.SafeMode)
	cliManager := clisession.NewManager()
	replayWriter := replay.NewWriter(cfg)
	metrics := observability.NewMetrics()

	var idemCache *idempotency.Cache
	if cfg.Idempotency.Enabled {
		ttl := 60
		if cfg.Idempotency.TTLSec > 0 {
			ttl = cfg.Idempotency.TTLSec
		}
		idemCache = idempotency.New(time.Duration(ttl) * time.Second)
	}

	return &Loop{
		cfg:         cfg,
		msgBus:      msgBus,
		eventBus:    eventBus,
		taskMgr:     taskMgr,
		provider:    provider,
		execTool:       execTool,
		killTool:       killTool,
		webFetch:       webFetch,
		messageTool:    messageTool,
		tailOutput:     tailOutput,
		streamCommand:  streamCommand,
		invokeCliAgent: invokeCliAgent,
		cliManager:     cliManager,
		replayWriter:   replayWriter,
		idempotency:   idemCache,
		metrics:       metrics,
		auditLogger: auditLogger,
		procTracker: procTracker,
		policyEval:  policyEval,
		safeMode:    opts.SafeMode,
	}
}

// Run starts the agent loop. It processes inbound messages until ctx is cancelled.
func (l *Loop) Run(ctx context.Context) error {
	l.running.Store(true)
	defer l.running.Store(false)

	for l.running.Load() {
		select {
		case <-ctx.Done():
			return nil
		default:
			msg, ok := l.msgBus.ConsumeInbound(ctx)
			if !ok {
				continue
			}

			response, err := l.processMessage(ctx, msg)
			if err != nil {
				response = fmt.Sprintf("Error: %v", err)
			}

			if response != "" {
				l.msgBus.PublishOutbound(bus.OutboundMessage{
					Channel: msg.Channel,
					ChatID:  msg.ChatID,
					Content: response,
				})
			}
		}
	}

	return nil
}

// toolDefinitions returns tool definitions for the LLM.
func (l *Loop) toolDefinitions() []providers.ToolDefinition {
	return []providers.ToolDefinition{
		{
			Type: "function",
			Function: providers.ToolFunctionDefinition{
				Name:        "exec",
				Description: "Execute a shell command and return its output. Commands run in the workspace. Use platform-appropriate syntax (Windows: cmd; Linux/macOS: sh). See runtime context in system prompt.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"command":    map[string]interface{}{"type": "string", "description": "The shell command to run"},
						"working_dir": map[string]interface{}{"type": "string", "description": "Working directory (optional)"},
					},
					"required": []interface{}{"command"},
				},
			},
		},
		{
			Type: "function",
			Function: providers.ToolFunctionDefinition{
				Name:        "kill",
				Description: "Kill a process started by Sypher-mini for this task. Only PIDs from exec tool can be killed.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"pid": map[string]interface{}{"type": "integer", "description": "Process ID to kill"},
					},
					"required": []interface{}{"pid"},
				},
			},
		},
		{
			Type: "function",
			Function: providers.ToolFunctionDefinition{
				Name:        "web_fetch",
				Description: "Fetch content from a URL. Use for web search or reading web pages.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"url": map[string]interface{}{"type": "string", "description": "URL to fetch"},
					},
					"required": []interface{}{"url"},
				},
			},
		},
		{
			Type: "function",
			Function: providers.ToolFunctionDefinition{
				Name:        "message",
				Description: "Send a message to the user in the current conversation.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"content": map[string]interface{}{"type": "string", "description": "Message content to send"},
					},
					"required": []interface{}{"content"},
				},
			},
		},
		{
			Type: "function",
			Function: providers.ToolFunctionDefinition{
				Name:        "tail_output",
				Description: "Read the last N lines from a file. Use for live log monitoring.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path":  map[string]interface{}{"type": "string", "description": "File path to read"},
						"lines": map[string]interface{}{"type": "integer", "description": "Number of lines (default 50, max 1000)"},
					},
					"required": []interface{}{"path"},
				},
			},
		},
		{
			Type: "function",
			Function: providers.ToolFunctionDefinition{
				Name:        "stream_command",
				Description: "Run a command and stream output to the user. Only commands in live_monitoring.allowed_commands are permitted (e.g. npm run, go run, gemini).",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"command":     map[string]interface{}{"type": "string", "description": "Command to run"},
						"working_dir": map[string]interface{}{"type": "string", "description": "Working directory (optional)"},
					},
					"required": []interface{}{"command"},
				},
			},
		},
		{
			Type: "function",
			Function: providers.ToolFunctionDefinition{
				Name:        "invoke_cli_agent",
				Description: "Invoke a configured CLI agent (e.g. Gemini CLI) with a task. Use for code generation when an agent with command/args is configured.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"task":        map[string]interface{}{"type": "string", "description": "Task/prompt for the CLI agent"},
						"agent_id":    map[string]interface{}{"type": "string", "description": "Agent ID to use (optional; uses first agent with command/args if omitted)"},
						"working_dir": map[string]interface{}{"type": "string", "description": "Working directory (optional)"},
					},
					"required": []interface{}{"task"},
				},
			},
		},
	}
}

// processMessage handles a single inbound message.
func (l *Loop) processMessage(ctx context.Context, msg bus.InboundMessage) (string, error) {
	// WhatsApp command parsing (config get, agents list, etc.)
	if msg.Channel == "whatsapp" {
		if isCmd, cmd, args, tier := intent.ParseWhatsAppCommand(msg.Content, msg.SenderID, &l.cfg.Channels); isCmd && cmd != "" {
			return l.handleWhatsAppCommand(ctx, cmd, args, tier, msg)
		}
	}

	// Intent parse: fast path for config/command
	parser := intent.New()
	ir := parser.Parse(msg.Content)
	if !ir.NeedsLLM() {
		switch ir.Intent {
		case intent.IntentConfigChange:
			return "Config commands: use 'sypher config get <path>' or 'sypher config set <path> <value>'", nil
		case intent.IntentCommand:
			// For direct commands, we could invoke exec directly; for now route to agent
			break
		case intent.IntentEmergencyAlert:
			return "Alert received. (Notification delivery not yet wired)", nil
		}
	}

	// Route to agent
	route := routing.Resolve(l.cfg, routing.RouteInput{
		Channel:   msg.Channel,
		AccountID: msg.SenderID,
	})
	agentID := route.AgentID
	sessionKey := route.SessionKey
	if sessionKey == "" {
		sessionKey = "agent:" + agentID + ":" + msg.Channel + ":" + msg.ChatID
	}

	// Idempotency: return cached result if same message within TTL
	if l.idempotency != nil {
		if _, result, ok := l.idempotency.Get(sessionKey, msg.Content); ok {
			return result, nil
		}
	}

	// Create task (pending -> authorized)
	t := l.taskMgr.Create(agentID, sessionKey)
	t.Transition(task.StateAuthorized)
	defer func() {
		l.taskMgr.Remove(t.ID)
		l.procTracker.RemoveTask(t.ID)
		l.messageTool.ClearReplyTarget(t.ID)
	}()

	l.messageTool.SetReplyTarget(t.ID, msg.Channel, msg.ChatID)

	// Emit task.started event
	_ = l.eventBus.Publish(ctx, bus.Event{
		Type: "task.started",
		Payload: map[string]interface{}{
			"task_id":     t.ID,
			"agent_id":    agentID,
			"channel":     msg.Channel,
			"chat_id":     msg.ChatID,
			"session_key": sessionKey,
		},
	})

	// Run with timeout
	t.Transition(task.StateExecuting)
	var result string
	err := l.taskMgr.RunWithTimeout(ctx, t, func(ctx context.Context) error {
		if t.IsCancelled() {
			t.Transition(task.StateKilled)
			return context.Canceled
		}

		if l.provider == nil || l.safeMode {
			if l.safeMode {
				result = fmt.Sprintf("Received: %q (LLM disabled in safe mode)", msg.Content)
			} else {
				result = fmt.Sprintf("Received: %q (no LLM provider configured - set CEREBRAS_API_KEY or OPENAI_API_KEY)", msg.Content)
			}
			return nil
		}

		// Build system prompt with bootstrap files (SOUL, AGENT, etc.)
		systemPrompt := l.buildSystemPrompt(agentID)
		messages := []providers.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: msg.Content},
		}
		model := l.cfg.Agents.Defaults.Model
		maxIter := l.cfg.Agents.Defaults.MaxToolIterations
		if maxIter <= 0 {
			maxIter = 20
		}

		for iter := 0; iter < maxIter; iter++ {
			if t.IsCancelled() {
				t.Transition(task.StateKilled)
				return context.Canceled
			}

			// Context summarization: truncate when over threshold (rough: 4 chars = 1 token)
			if thresh := l.cfg.Context.SummarizeThreshold; thresh > 0 {
				messages = truncateMessages(messages, thresh)
			}

			toolsDef := l.toolDefinitions()
			resp, err := l.provider.Chat(ctx, messages, toolsDef, model, map[string]interface{}{
				"max_tokens": 2048,
			})
			if err != nil {
				t.Transition(task.StateFailed)
				result = fmt.Sprintf("LLM error: %v", err)
				return nil
			}

			if len(resp.ToolCalls) == 0 {
				result = resp.Content
				if result == "" {
					result = "(no response)"
				}
				return nil
			}

			// Execute tool calls
			t.Transition(task.StateMonitoring)
			for _, tc := range resp.ToolCalls {
				req := tools.Request{
					ToolCallID: tc.ID,
					TaskID:     t.ID,
					AgentID:    agentID,
					Name:       tc.Name,
					Args:       tc.Arguments,
				}
				if l.policyEval != nil && !l.policyEval.CheckRateLimit(agentID, tc.Name) {
					toolResp := tools.ErrorResponse(tc.ID, "Rate limit exceeded", "Rate limit exceeded.", tools.CodeRateLimited, true)
					messages = append(messages, providers.Message{Role: "assistant", Content: resp.Content, ToolCalls: []providers.ToolCall{tc}, ToolCallID: tc.ID})
					messages = append(messages, providers.Message{Role: "tool", Content: "Error: " + toolResp.ForLLM, ToolCallID: tc.ID})
					continue
				}
				var toolResp tools.Response
				switch tc.Name {
				case "exec":
					toolResp = l.execTool.Execute(ctx, req)
				case "kill":
					toolResp = l.killTool.Execute(ctx, req)
				case "web_fetch":
					toolResp = l.webFetch.Execute(ctx, req)
				case "message":
					toolResp = l.messageTool.Execute(ctx, req)
				case "tail_output":
					toolResp = l.tailOutput.Execute(ctx, req)
				case "stream_command":
					toolResp = l.streamCommand.Execute(ctx, req)
				case "invoke_cli_agent":
					toolResp = l.invokeCliAgent.Execute(ctx, req)
				default:
					toolResp = tools.ErrorResponse(tc.ID, "Unknown tool: "+tc.Name, "Unknown tool.", tools.CodePermissionDenied, false)
				}

				if l.metrics != nil {
					l.metrics.IncToolCall(tc.Name)
					if toolResp.IsError {
						l.metrics.IncToolError(tc.Name)
					}
				}

				// Append assistant message with tool call
				toolContent := toolResp.ForLLM
				if toolResp.IsError {
					toolContent = "Error: " + toolContent
				}
				messages = append(messages, providers.Message{
					Role:       "assistant",
					Content:    resp.Content,
					ToolCalls:  []providers.ToolCall{tc},
					ToolCallID: tc.ID,
				})
				messages = append(messages, providers.Message{
					Role:       "tool",
					Content:    toolContent,
					ToolCallID: tc.ID,
				})
			}
			t.Transition(task.StateExecuting)
		}

		t.Transition(task.StateFailed)
		result = "(max tool iterations reached)"
		return nil
	})

	if err != nil {
		state := t.GetState()
		if state == task.StateTimeout {
			return "Task timed out", nil
		}
		if state == task.StateKilled {
			return "Task cancelled", nil
		}
		if state != task.StateFailed {
			t.Transition(task.StateFailed)
		}
		return fmt.Sprintf("Task failed: %v", err), nil
	}

	state := t.GetState()
	if l.metrics != nil {
		if state == task.StateFailed {
			l.metrics.IncTaskFailed()
		} else {
			l.metrics.IncTaskCompleted()
		}
	}
	if state == task.StateFailed {
		if l.replayWriter != nil {
			_ = l.replayWriter.Write(replay.Record{
				TaskID: t.ID,
				Input:  map[string]string{"content": msg.Content, "channel": msg.Channel},
				Result: result,
				Status: "failed",
			})
		}
		return result, nil
	}
	t.Transition(task.StateCompleted)
	if l.replayWriter != nil {
		_ = l.replayWriter.Write(replay.Record{
			TaskID: t.ID,
			Input:  map[string]string{"content": msg.Content, "channel": msg.Channel},
			Result: result,
			Status: "completed",
		})
	}
	if l.idempotency != nil {
		l.idempotency.Set(sessionKey, msg.Content, t.ID, result)
		l.idempotency.Cleanup()
	}
	return result, nil
}

// handleWhatsAppCommand handles WhatsApp commands (config, agents, monitors, audit, status).
func (l *Loop) handleWhatsAppCommand(ctx context.Context, cmd string, args []string, tier intent.WhatsAppTier, msg bus.InboundMessage) (string, error) {
	switch cmd {
	case "config":
		if len(args) >= 1 {
			if args[0] == "get" && intent.TierLevel(tier) < intent.TierLevel(intent.TierOperator) {
				return "Access denied. Operator tier required.", nil
			}
			if args[0] == "set" && intent.TierLevel(tier) < intent.TierLevel(intent.TierAdmin) {
				return "Access denied. Admin tier required.", nil
			}
		}
		if len(args) >= 1 && args[0] == "get" && len(args) >= 2 {
			key := args[1]
			val := l.cfg.Agents.List
			if key == "agents.list" {
				// Simple response
				var out string
				for i, a := range l.cfg.Agents.List {
					out += fmt.Sprintf("%d: %s\n", i+1, a.ID)
				}
				if out == "" {
					out = "No agents"
				}
				return out, nil
			}
			_ = val
			return "Config get: " + key, nil
		}
	case "agents":
		if intent.TierLevel(tier) < intent.TierLevel(intent.TierOperator) {
			return "Access denied. Operator tier required.", nil
		}
		var out string
		for i, a := range l.cfg.Agents.List {
			out += fmt.Sprintf("%d: %s\n", i+1, a.ID)
		}
		if out == "" {
			out = "No agents"
		}
		return out, nil
	case "monitors":
		if intent.TierLevel(tier) < intent.TierLevel(intent.TierOperator) {
			return "Access denied. Operator tier required.", nil
		}
		out := "HTTP: "
		for _, m := range l.cfg.Monitors.HTTP {
			out += m.ID + " "
		}
		out += "\nProcess: "
		for _, m := range l.cfg.Monitors.Process {
			out += m.ID + " "
		}
		return out, nil
	case "audit":
		if intent.TierLevel(tier) < intent.TierLevel(intent.TierAdmin) {
			return "Access denied. Admin tier required.", nil
		}
		if len(args) >= 1 {
			return "Use: sypher audit show " + args[0], nil
		}
		return "Usage: audit <task_id>", nil
	case "status":
		return fmt.Sprintf("Agents: %d, Timeout: %ds", len(l.cfg.Agents.List), l.cfg.Task.TimeoutSec), nil
	case "cli":
		return l.handleCliCommand(ctx, args, msg)
	}
	return "", nil
}

// handleCliCommand handles sypher cli list|new|N [--tail N].
func (l *Loop) handleCliCommand(ctx context.Context, args []string, msg bus.InboundMessage) (string, error) {
	if len(args) == 0 {
		return "Usage: cli list | cli new -m 'tag' | cli <N> [--tail N]", nil
	}
	switch args[0] {
	case "run":
		if len(args) < 3 {
			return "Usage: cli run <session_id> <command>", nil
		}
		var sid int
		if _, err := fmt.Sscanf(args[1], "%d", &sid); err != nil {
			return "Invalid session ID", nil
		}
		cmdStr := strings.Join(args[2:], " ")
		s := l.cliManager.Get(sid)
		if s == nil {
			return fmt.Sprintf("Session %d not found", sid), nil
		}
		// Run via exec and append output to session
		req := tools.Request{
			ToolCallID: "cli-run",
			TaskID:    "cli-" + args[1],
			AgentID:   "main",
			Name:      "exec",
			Args:      map[string]interface{}{"command": cmdStr},
		}
		resp := l.execTool.Execute(ctx, req)
		output := resp.ForLLM
		if resp.IsError {
			output = "Error: " + output
		}
		s.Append(output)
		return output, nil
	case "list":
		sessions := l.cliManager.List()
		if len(sessions) == 0 {
			return "No active CLI sessions. Use 'cli new -m \"tag\"' to create one.", nil
		}
		var out string
		for _, s := range sessions {
			ago := "just now"
			if d := time.Since(s.LastActivity); d > time.Minute {
				ago = fmt.Sprintf("%.0fm ago", d.Minutes())
			}
			out += fmt.Sprintf("%d: %s (active %s)\n", s.ID, s.Tag, ago)
		}
		return out, nil
	case "new":
		tag := ""
		for i := 1; i < len(args); i++ {
			if args[i] == "-m" && i+1 < len(args) {
				tag = strings.Join(args[i+1:], " ")
				break
			}
		}
		if tag == "" {
			tag = "unnamed"
		}
		s := l.cliManager.New(tag)
		return fmt.Sprintf("Created CLI session %d: %s", s.ID, tag), nil
	default:
		// args[0] is session number, parse --tail N
		var id int
		if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
			return "Usage: cli <session_id> [--tail N]", nil
		}
		tail := clisession.DefaultTailLines
		for i := 1; i < len(args)-1; i++ {
			if args[i] == "--tail" && i+1 < len(args) {
				if n, err := fmt.Sscanf(args[i+1], "%d", &tail); err == nil && n == 1 {
					if tail > clisession.MaxTailLines {
						tail = clisession.MaxTailLines
					}
				}
				break
			}
		}
		s := l.cliManager.Get(id)
		if s == nil {
			return fmt.Sprintf("Session %d not found. Use 'cli list' to see active sessions.", id), nil
		}
		out := s.Tail(tail)
		if out == "" {
			return fmt.Sprintf("Session %d (%s): no output yet", id, s.Tag), nil
		}
		return fmt.Sprintf("Session %d (%s) last %d lines:\n%s", id, s.Tag, tail, out), nil
	}
}

// truncateMessages keeps system + recent messages when total tokens exceed threshold.
// Rough estimate: 4 chars = 1 token.
func truncateMessages(messages []providers.Message, thresholdTokens int) []providers.Message {
	if len(messages) <= 2 {
		return messages
	}
	total := 0
	for _, m := range messages {
		total += len(m.Content) / 4
	}
	if total <= thresholdTokens {
		return messages
	}
	// Keep system (first) + last 6 messages
	keep := 6
	if len(messages) <= keep+1 {
		return messages
	}
	out := make([]providers.Message, 0, keep+1)
	if messages[0].Role == "system" {
		out = append(out, messages[0])
	}
	start := len(messages) - keep
	if start < 1 {
		start = 1
	}
	out = append(out, messages[start:]...)
	return out
}

// buildSystemPrompt builds the system prompt with bootstrap files and hard rules.
func (l *Loop) buildSystemPrompt(agentID string) string {
	workspace := l.cfg.Agents.Defaults.Workspace
	if workspace == "" {
		workspace = config.ExpandPath("~/.sypher-mini/workspace")
	}
	bootstrap := LoadBootstrapFiles(workspace, agentID)

	platformCtx := platform.AgentContext()

	hardRules := `## Hard Rules (non-overridable)
- ALWAYS use tools for actions; never pretend to execute
- Be helpful and accurate
- Use memory file for persistent info
- For messaging channels (WhatsApp, etc.): send ONE consolidated reply per user message; avoid calling the message tool multiple times in one turn`

	parts := []string{}
	if bootstrap != "" {
		parts = append(parts, bootstrap)
	} else {
		parts = append(parts, "You are Sypher, a coding-centric AI assistant.")
	}
	parts = append(parts, platformCtx, hardRules)
	return strings.Join(parts, "\n\n")
}

// Stop stops the agent loop.
func (l *Loop) Stop() {
	l.running.Store(false)
}

// CancelTask cancels a running task by ID.
func (l *Loop) CancelTask(taskID string) bool {
	return l.taskMgr.Cancel(taskID)
}

// Metrics returns the metrics collector for observability.
func (l *Loop) Metrics() *observability.Metrics {
	return l.metrics
}
