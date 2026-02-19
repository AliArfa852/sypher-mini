// Sypher-mini - Coding-centric agent pipeline
// Combines lightweight efficiency (PicoClaw-style) with flexible API support (OpenClaw-style)

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/sypherexx/sypher-mini/pkg/agent"
	"github.com/sypherexx/sypher-mini/pkg/bus"
	"github.com/sypherexx/sypher-mini/pkg/channels"
	"github.com/sypherexx/sypher-mini/pkg/config"
	"github.com/sypherexx/sypher-mini/pkg/commands"
	"github.com/sypherexx/sypher-mini/pkg/extensions"
	"github.com/sypherexx/sypher-mini/pkg/monitor"
	"github.com/sypherexx/sypher-mini/pkg/observability"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	cmd := os.Args[1]

	// Parse global flags (e.g. --safe)
	safeMode := false
	args := os.Args[2:]
	for i, a := range args {
		if a == "--safe" || a == "-safe" {
			safeMode = true
			args = append(args[:i], args[i+1:]...)
			break
		}
	}

	switch cmd {
	case "agent":
		agentCmd(args, safeMode)
	case "gateway":
		gatewayCmd(args, safeMode)
	case "status":
		statusCmd()
	case "config":
		configCmd(args)
	case "agents":
		agentsCmd(args)
	case "monitors":
		monitorsCmd(args)
	case "audit":
		auditCmd(args)
	case "replay":
		replayCmd(args)
	case "cancel":
		cancelCmd(args)
	case "onboard":
		onboardCmd()
	case "whatsapp":
		whatsappCmd(args)
	case "install-service":
		installServiceCmd()
	case "extensions":
		extensionsCmd()
	case "commands":
		commandsCmd(args)
	case "version", "-v", "--version":
		fmt.Printf("sypher-mini %s\n", version)
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`Sypher-mini - Coding-centric agent pipeline

Usage: sypher <command> [options]

Commands:
  agent      Run agent interactively or with -m "message"
  gateway    Start gateway (channels, monitors)
  status     Show config and status
  config     Get/set config (config get <path> | config set <path> <value>)
  agents     List/add/remove agents (agents list)
  monitors   List/add/remove monitors (monitors list)
  audit      Show audit log (audit show <task_id>)
  replay     Replay stored task (replay <task_id>)
  cancel     Cancel a running task (cancel <task_id>)
  onboard    Initialize config and workspace
  whatsapp   WhatsApp setup (whatsapp --connect)
  install-service  Install auto-start service (systemd/launchd/Task Scheduler)
  extensions  List discovered extensions
  commands   List per-command configs (commands list)
  version    Show version

Global flags:
  --safe     Safe mode: disable exec, remote API calls, task killing

Examples:
  sypher onboard
  sypher agent -m "Hello"
  sypher --safe gateway`)
}

func loadConfig() *config.Config {
	path := config.GetConfigPath()
	cfg, err := config.Load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config load error: %v\n", err)
		os.Exit(1)
	}
	return cfg
}

func agentCmd(args []string, safeMode bool) {
	cfg := loadConfig()

	msgBus := bus.NewMessageBus(100)
	eventBus := bus.New()
	loop := agent.NewLoop(cfg, msgBus, eventBus, &agent.LoopOptions{SafeMode: safeMode})

	// Start async event dispatcher
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go eventBus.RunAsyncDispatcher(ctx)

	// Check for -m "message" flag
	for i, a := range args {
		if a == "-m" && i+1 < len(args) {
			msg := args[i+1]
			go loop.Run(ctx)
			msgBus.PublishInbound(bus.InboundMessage{
				Channel:  "cli",
				ChatID:   "cli",
				Content:  msg,
				SenderID: "cli",
			})
			// Process one message
			out, ok := msgBus.SubscribeOutbound(ctx)
			if ok {
				fmt.Println(out.Content)
			}
			return
		}
	}

	// Interactive mode: run loop with signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		cancel()
		loop.Stop()
	}()

	// For interactive CLI, we'd need to read from stdin and publish to msgBus
	// For now, just run the loop (it will block waiting for inbound)
	fmt.Println("Agent loop running (Ctrl+C to stop). Use gateway for channel input.")
	_ = loop.Run(ctx)
}

func gatewayCmd(args []string, safeMode bool) {
	cfg := loadConfig()

	msgBus := bus.NewMessageBus(100)
	eventBus := bus.New()
	loop := agent.NewLoop(cfg, msgBus, eventBus, &agent.LoopOptions{SafeMode: safeMode})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go eventBus.RunAsyncDispatcher(ctx)

	if safeMode {
		fmt.Println("Safe mode: exec, remote APIs, and task kill disabled")
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		_ = loop.Run(ctx)
	}()

	health := observability.NewHealthChecker()
	health.Set("core", "ok")
	mux := http.NewServeMux()
	mux.Handle("/health", health.Handler())
	if m := loop.Metrics(); m != nil {
		mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			format := r.URL.Query().Get("format")
			if format == "prometheus" {
				w.Header().Set("Content-Type", "text/plain; version=0.0.4")
				w.Write([]byte(m.PrometheusFormat()))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(m.Snapshot())
		})
	}
	mux.HandleFunc("/cancel", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var payload struct {
			TaskID string `json:"task_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		ok := loop.CancelTask(payload.TaskID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"ok": ok})
	})
	mux.HandleFunc("/inbound", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var payload struct {
			Type    string `json:"type"`
			From    string `json:"from"`
			Content string `json:"content"`
			ChatID  string `json:"chat_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		chatID := payload.ChatID
		if chatID == "" {
			chatID = payload.From
		}
		msgBus.PublishInbound(bus.InboundMessage{
			Channel:  "whatsapp",
			ChatID:   chatID,
			Content:  payload.Content,
			SenderID: payload.From,
		})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})
	srv := &http.Server{Addr: ":18790", Handler: mux}
	go func() {
		_ = srv.ListenAndServe()
	}()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	if cfg.Channels.WhatsApp.Enabled {
		health.Set("whatsapp", "ok")

		if cfg.Channels.WhatsApp.UseBaileys {
			// Baileys extension: inbound via /inbound, outbound via HTTP to extension
			baileysURL := cfg.Channels.WhatsApp.BaileysURL
			if baileysURL == "" {
				baileysURL = "http://localhost:3002"
			}
			baileysClient := channels.NewWhatsAppBaileysClient(baileysURL, msgBus)
			go func() {
				_ = baileysClient.Run(ctx)
			}()
			// Optionally spawn extension subprocess
			if extProc := channels.SpawnBaileysExtension(baileysURL, "http://localhost:18790/inbound"); extProc != nil {
				go func() {
					_ = extProc.Wait()
				}()
				fmt.Printf("Gateway running. WhatsApp Baileys: %s (extension spawned)\n", baileysURL)
			} else {
				fmt.Printf("Gateway running. WhatsApp Baileys: %s (run extension separately: cd extensions/whatsapp-baileys && npm start)\n", baileysURL)
			}
		} else if cfg.Channels.WhatsApp.BridgeURL != "" {
			// WebSocket bridge
			bridge := channels.NewWhatsAppBridge(cfg.Channels.WhatsApp.BridgeURL, msgBus, eventBus)
			go func() {
				_ = bridge.Run(ctx)
			}()
			fmt.Printf("Gateway running. WhatsApp bridge: %s\n", cfg.Channels.WhatsApp.BridgeURL)
		} else {
			fmt.Println("Gateway running. WhatsApp enabled but no bridge_url or use_baileys")
		}

		// Start HTTP monitors with WhatsApp alerts
		for _, m := range cfg.Monitors.HTTP {
			if m.AlertViaWhatsApp && m.URL != "" {
				mon := monitor.NewHTTPMonitor(m, func(monitorID, message string) {
					chatID := "broadcast"
					if len(cfg.Channels.WhatsApp.AllowFrom) > 0 {
						chatID = cfg.Channels.WhatsApp.AllowFrom[0]
					}
					msgBus.PublishOutbound(bus.OutboundMessage{
						Channel: "whatsapp",
						ChatID:  chatID,
						Content: "[Monitor " + monitorID + "] " + message,
					})
				})
				go mon.Run(ctx)
			}
		}
	} else {
		health.Set("whatsapp", "disabled")
		fmt.Println("Gateway running. WhatsApp disabled (set channels.whatsapp.enabled)")
	}
	fmt.Println("Health: http://localhost:18790/health")

	<-sigCh
	cancel()
	loop.Stop()
}

func statusCmd() {
	cfg := loadConfig()
	fmt.Println("Sypher-mini status")
	fmt.Println("------------------")
	fmt.Printf("Config: %s\n", config.GetConfigPath())
	fmt.Printf("Deployment mode: %s\n", cfg.Deployment.Mode)
	fmt.Printf("Agents: %d\n", len(cfg.Agents.List))
	fmt.Printf("Task timeout: %ds\n", cfg.Task.TimeoutSec)
	fmt.Printf("WhatsApp enabled: %v\n", cfg.Channels.WhatsApp.Enabled)
}

func configCmd(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: sypher config get <path> | sypher config set <path> <value>")
		return
	}
	sub := args[0]
	cfgPath := config.GetConfigPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config load error: %v\n", err)
		os.Exit(1)
	}

	switch sub {
	case "get":
		key := args[1]
		val := getConfigPath(cfg, strings.Split(key, "."))
		if val != nil {
			fmt.Println(formatConfigValue(val))
		} else {
			fmt.Printf("Key not found: %s\n", key)
		}
	case "set":
		if len(args) < 3 {
			fmt.Println("Usage: sypher config set <path> <value>")
			return
		}
		key := args[1]
		valStr := args[2]
		if err := setConfigPath(cfg, strings.Split(key, "."), valStr); err != nil {
			fmt.Fprintf(os.Stderr, "Config set error: %v\n", err)
			os.Exit(1)
		}
		if err := cfg.Save(cfgPath); err != nil {
			fmt.Fprintf(os.Stderr, "Save error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Config updated")
	default:
		fmt.Printf("Unknown config subcommand: %s\n", sub)
	}
}

func getConfigPath(cfg *config.Config, path []string) interface{} {
	if len(path) == 0 {
		return cfg
	}
	switch path[0] {
	case "agents":
		if len(path) == 1 {
			return cfg.Agents
		}
		if path[1] == "list" && len(path) == 2 {
			return cfg.Agents.List
		}
	case "task", "timeout_sec":
		return cfg.Task.TimeoutSec
	case "channels":
		return cfg.Channels
	}
	return nil
}

func formatConfigValue(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprint(v)
	}
	return string(data)
}

func setConfigPath(cfg *config.Config, path []string, valStr string) error {
	if len(path) < 2 {
		return fmt.Errorf("path too short")
	}
	if path[0] == "task" && path[1] == "timeout_sec" {
		var v int
		if _, err := fmt.Sscanf(valStr, "%d", &v); err != nil {
			return err
		}
		cfg.Task.TimeoutSec = v
		return nil
	}
	return fmt.Errorf("set not supported for path: %s", strings.Join(path, "."))
}

func agentsCmd(args []string) {
	cfg := loadConfig()
	if len(args) > 0 && args[0] == "list" {
		fmt.Println("Agents:")
		for i, a := range cfg.Agents.List {
			def := ""
			if a.Default {
				def = " (default)"
			}
			fmt.Printf("  %d: %s%s\n", i+1, a.ID, def)
			if a.Name != "" {
				fmt.Printf("      name: %s\n", a.Name)
			}
		}
		return
	}
	fmt.Println("Usage: sypher agents list")
}

func monitorsCmd(args []string) {
	cfg := loadConfig()
	if len(cfg.Monitors.HTTP) == 0 && len(cfg.Monitors.Process) == 0 {
		fmt.Println("No monitors configured")
		return
	}
	if len(args) > 0 && args[0] == "list" {
		fmt.Println("HTTP monitors:")
		for _, m := range cfg.Monitors.HTTP {
			fmt.Printf("  - %s: %s (interval %ds)\n", m.ID, m.URL, m.IntervalSec)
		}
		fmt.Println("Process monitors:")
		for _, m := range cfg.Monitors.Process {
			fmt.Printf("  - %s: %s\n", m.ID, m.Command)
		}
		return
	}
	fmt.Println("Usage: sypher monitors list")
}

func auditCmd(args []string) {
	if len(args) < 2 || args[0] != "show" {
		fmt.Println("Usage: sypher audit show <task_id>")
		return
	}
	taskID := args[1]
	cfg := loadConfig()
	auditDir := config.ExpandPath(cfg.Audit.Dir)
	if auditDir == "" {
		auditDir = config.ExpandPath("~/.sypher-mini/audit")
	}
	path := filepath.Join(auditDir, taskID+".log")
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Audit log not found: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(string(data))
}

func replayCmd(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: sypher replay <task_id>")
		return
	}
	taskID := args[0]
	cfg := loadConfig()
	replayDir := config.ExpandPath("~/.sypher-mini/replay")
	if cfg.Replay.Dir != "" {
		replayDir = config.ExpandPath(cfg.Replay.Dir)
	}
	path := filepath.Join(replayDir, taskID+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Replay file not found: %v (replay persistence may be disabled)\n", err)
		os.Exit(1)
	}
	var replay struct {
		Input         interface{} `json:"input"`
		ToolCalls     interface{} `json:"tool_calls"`
		ToolResults   interface{} `json:"tool_results"`
		LLMResponses  interface{} `json:"llm_responses"`
	}
	if err := json.Unmarshal(data, &replay); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid replay file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Replay (read-only, no re-execution):")
	fmt.Println(formatConfigValue(replay))
}

func cancelCmd(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: sypher cancel <task_id>")
		fmt.Println("Note: Cancel via gateway API: POST http://localhost:18790/cancel with {\"task_id\": \"<id>\"}")
		return
	}
	taskID := args[0]
	url := "http://localhost:18790/cancel"
	if v := os.Getenv("SYPHER_GATEWAY_URL"); v != "" {
		url = v + "/cancel"
	}
	body := fmt.Sprintf(`{"task_id":"%s"}`, taskID)
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Request error: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cancel failed (is gateway running?): %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	var result struct {
		OK bool `json:"ok"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&result)
	if result.OK {
		fmt.Printf("Task %s cancelled\n", taskID)
	} else {
		fmt.Printf("Task %s not found or already completed\n", taskID)
	}
}

func whatsappCmd(args []string) {
	connect := false
	allowFrom := ""
	for i, a := range args {
		if a == "--connect" || a == "-connect" {
			connect = true
		}
		if (a == "--allow-from" || a == "-allow-from") && i+1 < len(args) {
			allowFrom = args[i+1]
		}
	}
	if !connect {
		fmt.Println("Usage: sypher whatsapp --connect [--allow-from +1234567890]")
		fmt.Println("Configures WhatsApp (Baileys) and enables the channel.")
		return
	}

	cfgPath := config.GetConfigPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config load error: %v\n", err)
		os.Exit(1)
	}

	cfg.Channels.WhatsApp.Enabled = true
	cfg.Channels.WhatsApp.UseBaileys = true
	if cfg.Channels.WhatsApp.BaileysURL == "" {
		cfg.Channels.WhatsApp.BaileysURL = "http://localhost:3002"
	}
	if allowFrom != "" {
		cfg.Channels.WhatsApp.AllowFrom = []string{allowFrom}
	}

	if err := cfg.Save(cfgPath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to save config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("WhatsApp configured (Baileys).")
	if len(cfg.Channels.WhatsApp.AllowFrom) > 0 {
		fmt.Printf("Allow from: %v\n", cfg.Channels.WhatsApp.AllowFrom)
	} else {
		fmt.Println("Allow from: (empty = allow all — restrict in production)")
	}
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Run: sypher gateway")
	fmt.Println("  2. Scan the QR code in the terminal with WhatsApp (Settings → Linked Devices)")
	fmt.Println("  3. Auth saved to ~/.sypher-mini/whatsapp-auth/")
}

func onboardCmd() {
	path := config.GetConfigPath()
	cfg := config.DefaultConfig()

	dir := config.ExpandPath(cfg.Agents.Defaults.Workspace)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create workspace: %v\n", err)
		os.Exit(1)
	}

	// Create workspace subdirectories (PicoClaw/OpenClaw alignment)
	for _, sub := range []string{"memory", "sessions", "state", "cron", "skills", "code-projects"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create workspace subdir %s: %v\n", sub, err)
			os.Exit(1)
		}
	}

	// Create bootstrap template files if missing
	bootstrapFiles := map[string]string{
		"AGENTS.md":    "# Agent behavior guide\n\nEdit this file to define the agent's role, instructions, and how it should behave.\n",
		"AGENT.md":     "# Agent role (alias for AGENTS.md)\n\nSame content as AGENTS.md. Edit to define the agent's role and instructions.\n",
		"SOUL.md":      "# Agent soul\n\nPersonality, tone, values, and boundaries. Loaded every session.\n",
		"USER.md":      "# User context\n\nWho the user is and how to address them. Loaded every session.\n",
		"IDENTITY.md":  "# Agent identity\n\nThe agent's name, vibe, and identity. Optional override.\n",
		"HEARTBEAT.md": "# Periodic tasks\n\nOptional checklist for heartbeat runs (when implemented). Keep short.\n",
		"TOOLS.md":     "# Tool descriptions\n\nNotes about local tools and conventions. Guidance only; does not control tool availability.\n",
	}
	for name, content := range bootstrapFiles {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			if err := os.WriteFile(p, []byte(content), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create %s: %v\n", name, err)
				os.Exit(1)
			}
		}
	}

	configDir := ""
	if len(path) > 0 {
		for i := len(path) - 1; i >= 0; i-- {
			if path[i] == '/' || path[i] == '\\' {
				configDir = path[:i]
				break
			}
		}
	}
	if configDir != "" {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create config dir: %v\n", err)
			os.Exit(1)
		}
	}

	if err := cfg.Save(path); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to save config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Onboard complete. Config: %s\n", path)
	fmt.Printf("Workspace: %s\n", dir)
	fmt.Println("Edit config and run: sypher agent")
}

func installServiceCmd() {
	binPath, err := os.Executable()
	if err != nil {
		binPath = "sypher"
	}

	fmt.Println("Sypher-mini install-service")
	fmt.Println("----------------------------")
	fmt.Printf("Binary: %s\n", binPath)
	fmt.Printf("Config: %s\n", config.GetConfigPath())
	fmt.Println()
	fmt.Println("Auto-start options:")
	fmt.Println("  Linux (systemd): Create ~/.config/systemd/user/sypher-mini.service")
	fmt.Println("  macOS (launchd): Create ~/Library/LaunchAgents/com.sypher-mini.plist")
	fmt.Println("  Windows: Use Task Scheduler to run at logon")
	fmt.Println()
	fmt.Println("Example systemd unit (~/.config/systemd/user/sypher-mini.service):")
	fmt.Printf(`[Unit]
Description=Sypher-mini Gateway
After=network.target

[Service]
Type=simple
ExecStart=%s gateway
Restart=on-failure
RestartSec=10

[Install]
WantedBy=default.target
`, binPath)
	fmt.Println()
	fmt.Println("Then: systemctl --user enable sypher-mini && systemctl --user start sypher-mini")
}

func extensionsCmd() {
	wd, _ := os.Getwd()
	exts, err := extensions.DiscoverFromWorkspace(wd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Extensions discovery error: %v\n", err)
		os.Exit(1)
	}
	if len(exts) == 0 {
		fmt.Println("No extensions found (scan extensions/ for sypher.extension.json)")
		return
	}
	fmt.Println("Discovered extensions:")
	for _, e := range exts {
		caps := ""
		if len(e.Manifest.Capabilities) > 0 {
			caps = " [" + strings.Join(e.Manifest.Capabilities, ", ") + "]"
		}
		fmt.Printf("  %s v%s%s\n", e.Manifest.ID, e.Manifest.Version, caps)
	}
}

func commandsCmd(args []string) {
	commandsDir := ""
	if home, err := os.UserHomeDir(); err == nil {
		commandsDir = filepath.Join(home, ".sypher-mini", "commands")
	}
	if len(args) >= 1 && args[0] == "list" {
		names, err := commands.List(commandsDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Commands list error: %v\n", err)
			os.Exit(1)
		}
		if len(names) == 0 {
			fmt.Println("No command configs (create ~/.sypher-mini/commands/{name}.json)")
			return
		}
		fmt.Println("Available commands:")
		for _, n := range names {
			fmt.Printf("  %s\n", n)
		}
		return
	}
	fmt.Println("Usage: sypher commands list")
}
