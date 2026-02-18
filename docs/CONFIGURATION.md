# Sypher-mini Configuration Reference

Complete reference for `~/.sypher-mini/config.json`.

---

## Config file location

| OS | Path |
|----|------|
| Linux/macOS | `~/.sypher-mini/config.json` |
| Windows | `%USERPROFILE%\.sypher-mini\config.json` |

---

## Full schema

```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.sypher-mini/workspace",
      "restrict_to_workspace": true,
      "model": "cerebras/llama-3.1-70b",
      "max_tool_iterations": 20
    },
    "list": [
      {
        "id": "main",
        "default": true,
        "name": "Sypher",
        "workspace": "~/.sypher-mini/workspace",
        "skills": [],
        "allowed_commands": null,
        "command": "",
        "args": [],
        "model": null
      }
    ]
  },
  "bindings": [
    { "agent_id": "main", "match": { "channel": "whatsapp", "account_id": "*" } }
  ],
  "authorized_terminals": ["default"],
  "channels": {
    "whatsapp": {
      "enabled": false,
      "bridge_url": "ws://localhost:3001",
      "use_baileys": false,
      "allow_from": [],
      "operators": [],
      "admins": []
    }
  },
  "providers": {
    "routing_strategy": "cheap_first",
    "cerebras": { "api_key": "", "api_base": "" },
    "openai": { "api_key": "", "api_base": "" },
    "anthropic": { "api_key": "", "api_base": "" },
    "gemini": { "api_key": "", "api_base": "" }
  },
  "task": {
    "timeout_sec": 300,
    "retry_max": 2
  },
  "tools": {
    "exec": {
      "custom_deny_patterns": [],
      "timeout_sec": 60
    },
    "live_monitoring": {
      "allowed_commands": ["npm run", "go run", "tail -f"]
    }
  },
  "audit": {
    "dir": "~/.sypher-mini/audit",
    "retention_days": 30,
    "integrity": "none"
  },
  "policies": {
    "files": [],
    "network": [],
    "rate_limits": []
  },
  "context": {
    "max_tokens": 8192,
    "reserved_for_tools": 2048,
    "summarize_threshold": 6000,
    "cache_tool_outputs": true,
    "cache_max_entries": 10
  },
  "monitors": {
    "http": [],
    "process": []
  },
  "replay": {
    "enabled": false,
    "dir": "~/.sypher-mini/replay"
  },
  "deployment": {
    "mode": "local_dev"
  }
}
```

---

## Sections

### agents

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `defaults.workspace` | string | `~/.sypher-mini/workspace` | Default workspace path |
| `defaults.restrict_to_workspace` | bool | `true` | Limit exec/file access to workspace |
| `defaults.model` | string | `cerebras/llama-3.1-70b` | Default LLM model |
| `defaults.max_tool_iterations` | int | `20` | Max tool calls per task |
| `list[].id` | string | — | Agent ID |
| `list[].default` | bool | `false` | Use as default agent |
| `list[].name` | string | — | Display name |
| `list[].workspace` | string | — | Override workspace |
| `list[].skills` | []string | `[]` | Skill filter (empty = all) |
| `list[].allowed_commands` | []string | `null` | Allowlist for exec |
| `list[].command` | string | — | For CLI agents (e.g. gemini) |
| `list[].args` | []string | — | Args for CLI agents |

### bindings

Maps incoming messages to agents. Priority: peer > account > channel wildcard > default.

```json
{ "agent_id": "main", "match": { "channel": "whatsapp", "account_id": "*" } }
{ "agent_id": "coding", "match": { "channel": "whatsapp", "peer": { "kind": "direct", "id": "+123" } } }
```

### channels.whatsapp

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `false` | Enable WhatsApp |
| `bridge_url` | string | `ws://localhost:3001` | WebSocket bridge URL |
| `use_baileys` | bool | `false` | Use Baileys extension |
| `allow_from` | []string | `[]` | Allowed phone numbers |
| `operators` | []string | `[]` | Operator numbers |
| `admins` | []string | `[]` | Admin numbers |

### providers

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `routing_strategy` | string | `cheap_first` | Provider order |
| `cerebras.api_key` | string | — | Cerebras API key |
| `openai.api_key` | string | — | OpenAI API key |
| `anthropic.api_key` | string | — | Anthropic API key |
| `gemini.api_key` | string | — | Gemini API key |

### task

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `timeout_sec` | int | `300` | Task timeout (seconds) |
| `retry_max` | int | `2` | LLM retry attempts |

### tools.exec

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `custom_deny_patterns` | []string | `[]` | Extra regex patterns to block |
| `timeout_sec` | int | `60` | Exec command timeout |

### audit

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `dir` | string | `~/.sypher-mini/audit` | Audit log directory |
| `retention_days` | int | `30` | Log retention |
| `integrity` | string | `none` | Integrity mode |

### policies

| Field | Description |
|-------|-------------|
| `files` | Per-file access: `{ "path": "~/.sypher-mini/**", "agent_ids": ["*"], "access": "read_write" }` |
| `network` | Network access: `{ "agent_ids": ["*"], "allow_domains": ["*"], "deny_domains": [] }` |
| `rate_limits` | Rate limits: `{ "agent_id": "*", "tool_name": "exec", "requests_per_minute": 30 }` |

### monitors

**HTTP monitors:**

```json
{
  "id": "api-health",
  "url": "https://api.example.com/health",
  "interval_sec": 60,
  "alert_on_status": [400, 500],
  "alert_via_whatsapp": true,
  "cooldown_sec": 300,
  "min_failures": 2
}
```

**Process monitors:**

```json
{
  "id": "dev-server",
  "command": "npm run dev",
  "cwd": "/app",
  "error_pattern": "status (4|5)\\d{2}|Error:",
  "alert_via_whatsapp": true,
  "cooldown_sec": 60
}
```

### deployment

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `mode` | string | `local_dev` | `local_dev` \| `headless_server` \| `container` \| `multi_user` |

---

## CLI config commands

```bash
sypher config get agents.list
sypher config set task.timeout_sec 600
```

---

## Environment overrides

| Variable | Overrides |
|----------|-----------|
| `SYPHER_MINI_MODE` | `deployment.mode` |
| `CEREBRAS_API_KEY` | `providers.cerebras.api_key` |
| `OPENAI_API_KEY` | `providers.openai.api_key` |
| `ANTHROPIC_API_KEY` | `providers.anthropic.api_key` |
| `GEMINI_API_KEY` | `providers.gemini.api_key` |
| `SYPHER_GATEWAY_URL` | Base URL for `sypher cancel` |
