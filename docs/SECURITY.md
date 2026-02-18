# Sypher-mini Security Guide

Security model, best practices, and safe mode.

---

## Overview

Sypher-mini executes shell commands and connects to external services. This document describes built-in protections and recommended practices.

---

## Built-in protections

### 1. Workspace restriction

**Default:** `restrict_to_workspace: true`

- Exec commands run with working directory inside workspace
- File access (read/write) limited to workspace
- Prevents access to system paths outside `~/.sypher-mini/workspace`

### 2. Deny patterns (exec tool)

The exec tool blocks commands matching these patterns (and custom ones):

| Pattern | Blocks |
|---------|--------|
| `rm -rf`, `del /f`, `rmdir /s` | Bulk deletion |
| `format`, `mkfs`, `diskpart` | Disk formatting |
| `dd if=` | Disk imaging |
| `> /dev/sd[a-z]` | Direct disk writes |
| `shutdown`, `reboot`, `poweroff` | System shutdown |
| `sudo` | Privilege escalation |
| `chmod`, `chown` | Permission changes |
| `curl \| sh`, `wget \| bash` | Pipe-to-shell |
| `eval`, `source *.sh` | Code injection |
| Fork bombs | `:(){ :|:& };:` |

Add custom patterns in config:

```json
{
  "tools": {
    "exec": {
      "custom_deny_patterns": ["\\b dangerous_cmd \\b"]
    }
  }
}
```

### 3. Kill scope

The `kill` tool only kills processes that:

- Were started by Sypher-mini for the **current task**
- Are recorded in the process tracker

You cannot kill arbitrary system processes.

### 4. Audit logging

Every exec command is logged to `~/.sypher-mini/audit/{task_id}.log`:

```
[task_id] [tool_call_id] timestamp | exec | cmd="..." cwd="..." exit=0 | output...
```

### 5. Rate limits

Configure per-agent, per-tool rate limits:

```json
{
  "policies": {
    "rate_limits": [
      { "agent_id": "*", "tool_name": "exec", "requests_per_minute": 30 }
    ]
  }
}
```

---

## Safe mode

**Flag:** `--safe`

Disables:

- Exec tool (no command execution)
- LLM API calls
- Kill tool

Use for:

- Recovering from misconfiguration
- Inspecting state without risk
- Debugging

```bash
sypher --safe gateway
sypher --safe agent -m "test"
```

---

## API keys

**Never** commit API keys to config in version control.

**Recommended:** Environment variables

```bash
export CEREBRAS_API_KEY="..."
export OPENAI_API_KEY="..."
```

**Alternative:** Config file with restricted permissions

```bash
chmod 600 ~/.sypher-mini/config.json
```

---

## WhatsApp

- Set `allow_from` to restrict who can interact
- Use `operators` and `admins` for privileged commands (when implemented)
- Keep `~/.sypher-mini/whatsapp-auth/` private (Baileys session)

---

## Disabling restrictions (risk)

Only in controlled environments:

```json
{
  "agents": {
    "defaults": {
      "restrict_to_workspace": false
    }
  }
}
```

⚠️ **Warning:** This allows the agent to access any path and run commands anywhere. Use with extreme caution.
