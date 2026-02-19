package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Config holds the full Sypher-mini configuration.
type Config struct {
	Agents              AgentsConfig     `json:"agents"`
	Bindings            []AgentBinding   `json:"bindings,omitempty"`
	AuthorizedTerminals []string         `json:"authorized_terminals,omitempty"`
	Channels            ChannelsConfig   `json:"channels"`
	Providers           ProvidersConfig  `json:"providers"`
	Task                TaskConfig       `json:"task"`
	Deployment          DeploymentConfig `json:"deployment"`
	Gateway             GatewayConfig    `json:"gateway,omitempty"`
	Tools               ToolsConfig      `json:"tools,omitempty"`
	Audit               AuditConfig      `json:"audit,omitempty"`
	Policies            PoliciesConfig   `json:"policies,omitempty"`
	Context             ContextConfig    `json:"context,omitempty"`
	Monitors            MonitorsConfig  `json:"monitors,omitempty"`
	Replay              ReplayConfig      `json:"replay,omitempty"`
	Idempotency         IdempotencyConfig `json:"idempotency,omitempty"`
	mu                  sync.RWMutex
}

// IdempotencyConfig holds session dedup config.
type IdempotencyConfig struct {
	Enabled bool `json:"enabled"`
	TTLSec  int  `json:"ttl_sec"`
}

// ReplayConfig holds replay persistence config.
type ReplayConfig struct {
	Enabled bool   `json:"enabled"`
	Dir     string `json:"dir"`
}

// MonitorsConfig holds HTTP and process monitor config.
type MonitorsConfig struct {
	HTTP    []HTTPMonitor    `json:"http,omitempty"`
	Process []ProcessMonitor `json:"process,omitempty"`
}

// HTTPMonitor defines an HTTP health check monitor.
type HTTPMonitor struct {
	ID               string  `json:"id"`
	URL              string  `json:"url"`
	IntervalSec      int     `json:"interval_sec"`
	AlertOnStatus    []int   `json:"alert_on_status,omitempty"`
	AlertViaWhatsApp bool    `json:"alert_via_whatsapp"`
	CooldownSec      int     `json:"cooldown_sec"`
	MinFailures      int     `json:"min_failures"`
}

// ProcessMonitor defines a process output monitor.
type ProcessMonitor struct {
	ID               string `json:"id"`
	Command          string `json:"command"`
	Cwd              string `json:"cwd,omitempty"`
	ErrorPattern     string `json:"error_pattern,omitempty"`
	AlertViaWhatsApp bool   `json:"alert_via_whatsapp"`
	CooldownSec      int    `json:"cooldown_sec"`
}

// ContextConfig holds context window and memory config.
type ContextConfig struct {
	MaxTokens          int  `json:"max_tokens"`
	ReservedForTools    int  `json:"reserved_for_tools"`
	SummarizeThreshold  int  `json:"summarize_threshold"`
	CacheToolOutputs    bool `json:"cache_tool_outputs"`
	CacheMaxEntries     int  `json:"cache_max_entries"`
}

// ToolsConfig holds tool-specific config.
type ToolsConfig struct {
	Exec          ExecToolConfig          `json:"exec,omitempty"`
	LiveMonitoring LiveMonitoringConfig   `json:"live_monitoring,omitempty"`
}

// LiveMonitoringConfig holds config for tail_output and stream_command.
type LiveMonitoringConfig struct {
	AllowedCommands []string `json:"allowed_commands,omitempty"`
}

// ExecToolConfig holds exec tool config.
type ExecToolConfig struct {
	CustomDenyPatterns []string `json:"custom_deny_patterns,omitempty"`
	TimeoutSec         int      `json:"timeout_sec"`
	AllowGitPush       bool     `json:"allow_git_push"`
	AllowDirs          []string `json:"allow_dirs,omitempty"`
}

// AuditConfig holds audit logger config.
type AuditConfig struct {
	Dir           string `json:"dir"`
	RetentionDays int    `json:"retention_days"`
	Integrity     string `json:"integrity"`
}

// PoliciesConfig holds policy config.
type PoliciesConfig struct {
	Files       []FilePolicy   `json:"files,omitempty"`
	Network     []NetPolicy    `json:"network,omitempty"`
	RateLimits  []RateLimit    `json:"rate_limits,omitempty"`
}

// FilePolicy defines per-file access.
type FilePolicy struct {
	Path     string   `json:"path"`
	AgentIDs []string `json:"agent_ids"`
	Access   string   `json:"access"` // read, write, read_write
}

// NetPolicy defines network access.
type NetPolicy struct {
	AgentIDs     []string `json:"agent_ids"`
	AllowDomains []string `json:"allow_domains"`
	DenyDomains  []string `json:"deny_domains"`
	AllowPorts   []int    `json:"allow_ports"`
}

// RateLimit defines per-agent/tool rate limit.
type RateLimit struct {
	AgentID           string `json:"agent_id"`
	ToolName          string `json:"tool_name"`
	RequestsPerMinute int    `json:"requests_per_minute"`
}

// AgentsConfig holds agent defaults and list.
type AgentsConfig struct {
	Defaults AgentDefaults `json:"defaults"`
	List     []AgentConfig `json:"list,omitempty"`
}

// AgentDefaults holds default values for agents.
type AgentDefaults struct {
	Workspace           string  `json:"workspace"`
	RestrictToWorkspace bool    `json:"restrict_to_workspace"`
	Model               string  `json:"model"`
	MaxToolIterations   int     `json:"max_tool_iterations"`
}

// AgentModelConfig supports primary and fallbacks.
type AgentModelConfig struct {
	Primary   string   `json:"primary,omitempty"`
	Fallbacks []string `json:"fallbacks,omitempty"`
}

// AgentConfig defines a single agent.
type AgentConfig struct {
	ID               string            `json:"id"`
	Default          bool              `json:"default,omitempty"`
	Name             string            `json:"name,omitempty"`
	Workspace        string            `json:"workspace,omitempty"`
	Model            *AgentModelConfig `json:"model,omitempty"`
	Skills           []string          `json:"skills,omitempty"`
	Command          string            `json:"command,omitempty"`
	Args             []string          `json:"args,omitempty"`
	AllowedCommands  []string          `json:"allowed_commands,omitempty"`
}

// PeerMatch matches a peer for binding.
type PeerMatch struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

// BindingMatch defines binding criteria.
type BindingMatch struct {
	Channel   string     `json:"channel"`
	AccountID string     `json:"account_id,omitempty"`
	Peer      *PeerMatch `json:"peer,omitempty"`
}

// AgentBinding maps match criteria to agent.
type AgentBinding struct {
	AgentID string       `json:"agent_id"`
	Match   BindingMatch `json:"match"`
}

// ChannelsConfig holds channel configurations.
type ChannelsConfig struct {
	WhatsApp WhatsAppConfig `json:"whatsapp"`
}

// WhatsAppConfig holds WhatsApp channel config.
type WhatsAppConfig struct {
	Enabled    bool     `json:"enabled"`
	BridgeURL  string   `json:"bridge_url"`
	BaileysURL string   `json:"baileys_url"` // Extension HTTP endpoint, e.g. http://localhost:3002
	AllowFrom  []string `json:"allow_from"`
	Operators  []string `json:"operators,omitempty"`
	Admins     []string `json:"admins,omitempty"`
	UseBaileys bool     `json:"use_baileys"`
}

// ProvidersConfig holds LLM provider configs.
type ProvidersConfig struct {
	RoutingStrategy string                 `json:"routing_strategy"`
	LLMRateLimit    LLMRateLimitConfig     `json:"llm_rate_limit,omitempty"`
	Cerebras        ProviderConfig         `json:"cerebras"`
	OpenAI          ProviderConfig         `json:"openai"`
	Anthropic       ProviderConfig         `json:"anthropic"`
	Gemini          ProviderConfig         `json:"gemini"`
}

// LLMRateLimitConfig limits API calls per time window (e.g. 2 per 15 sec).
type LLMRateLimitConfig struct {
	MaxPerWindow int `json:"max_per_window"` // max calls allowed in window (default 2)
	WindowSec    int `json:"window_sec"`     // window duration in seconds (default 15)
}

// ProviderConfig holds a single provider's config.
type ProviderConfig struct {
	APIKey  string `json:"api_key"`
	APIBase string `json:"api_base,omitempty"`
}

// TaskConfig holds task lifecycle config.
type TaskConfig struct {
	TimeoutSec int `json:"timeout_sec"`
	RetryMax   int `json:"retry_max"`
}

// DeploymentConfig holds deployment mode config.
type DeploymentConfig struct {
	Mode string `json:"mode"`
}

// GatewayConfig holds gateway HTTP server config.
type GatewayConfig struct {
	Bind          string `json:"bind,omitempty"`           // e.g. "127.0.0.1:18790" (default) or "0.0.0.0:18790"
	InboundSecret string `json:"inbound_secret,omitempty"` // If set, require X-Sypher-Inbound-Secret header on /inbound and /cancel
}

// GetConfigPath returns the path to the config file.
func GetConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".sypher-mini/config.json"
	}
	return filepath.Join(home, ".sypher-mini", "config.json")
}

// Load loads config from file, with env overrides.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Apply env overrides
	applyEnvOverrides(&cfg)

	// Set defaults for zero values
	if cfg.Task.TimeoutSec == 0 {
		cfg.Task.TimeoutSec = 300
	}
	if cfg.Task.RetryMax == 0 {
		cfg.Task.RetryMax = 2
	}
	if cfg.Providers.RoutingStrategy == "" {
		cfg.Providers.RoutingStrategy = "cheap_first"
	}
	if cfg.Deployment.Mode == "" {
		cfg.Deployment.Mode = "local_dev"
	}

	return &cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	// TODO: use caarlos0/env for full env override support
	if v := os.Getenv("SYPHER_MINI_MODE"); v != "" {
		cfg.Deployment.Mode = v
	}
	if v := os.Getenv("SYPHER_INBOUND_SECRET"); v != "" {
		cfg.Gateway.InboundSecret = v
	}
	if v := os.Getenv("SYPHER_GATEWAY_BIND"); v != "" {
		cfg.Gateway.Bind = v
	}
	// When GEMINI_API_KEY is set, allow GEMINI_MODEL to override agents.defaults.model
	if os.Getenv("GEMINI_API_KEY") != "" || cfg.Providers.Gemini.APIKey != "" {
		if v := os.Getenv("GEMINI_MODEL"); v != "" {
			cfg.Agents.Defaults.Model = "gemini/" + v
		}
	}
}

// DefaultConfig returns a default configuration.
func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	workspace := filepath.Join(home, ".sypher-mini", "workspace")
	return &Config{
		Agents: AgentsConfig{
			Defaults: AgentDefaults{
				Workspace:           workspace,
				RestrictToWorkspace: true,
				Model:               "cerebras/llama-3.1-70b",
				MaxToolIterations:   20,
			},
			List: []AgentConfig{
				{
					ID:        "main",
					Default:   true,
					Name:      "Sypher",
					Workspace: workspace,
				},
			},
		},
		Bindings: []AgentBinding{
			{AgentID: "main", Match: BindingMatch{Channel: "whatsapp", AccountID: "*"}},
		},
		AuthorizedTerminals: []string{"default"},
		Channels: ChannelsConfig{
			WhatsApp: WhatsAppConfig{
				Enabled:    false,
				BridgeURL:  "ws://localhost:3001",
				BaileysURL: "http://localhost:3002",
				UseBaileys: true, // QR connection is default when WhatsApp enabled
				AllowFrom:  []string{},
			},
		},
		Providers: ProvidersConfig{
			RoutingStrategy: "cheap_first",
		},
		Task: TaskConfig{
			TimeoutSec: 300,
			RetryMax:   2,
		},
		Deployment: DeploymentConfig{
			Mode: "local_dev",
		},
		Audit: AuditConfig{
			Dir:           filepath.Join(home, ".sypher-mini", "audit"),
			RetentionDays: 30,
			Integrity:     "none",
		},
		Context: ContextConfig{
			MaxTokens:         8192,
			ReservedForTools:  2048,
			SummarizeThreshold: 6000,
			CacheToolOutputs:   true,
			CacheMaxEntries:    10,
		},
	}
}

// Save writes config to file.
func (c *Config) Save(path string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// ExpandPath expands ~ to home directory.
func ExpandPath(p string) string {
	if p == "" {
		return p
	}
	if strings.HasPrefix(p, "~") {
		home, _ := os.UserHomeDir()
		if len(p) > 1 && (p[1] == '/' || p[1] == '\\') {
			return filepath.Join(home, p[2:])
		}
		return home
	}
	return p
}
