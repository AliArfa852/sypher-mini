# Sypher-mini Security Guide

Security model, best practices, vulnerabilities, and safe mode.

---

## Overview

Sypher-mini executes shell commands and connects to external services. This document describes built-in protections, known vulnerabilities and mitigations, and recommended practices.

---

## Vulnerability and Weak Point Reference

### Critical Vulnerabilities (Mitigated)

| ID | Severity | Description | Mitigation |
|----|----------|-------------|------------|
| V1 | High | Deny patterns can be disabled (Picoclaw) | Picoclaw now requires `PICOCLAW_DANGER_DISABLE_DENY=1` to honor config disable |
| V2 | High | Command path traversal | Command path validation added; paths in command string checked against workspace |
| V3 | Medium | Web fetch prompt injection | Prefix guard in place; consider stronger fencing for untrusted URLs |
| V4 | Medium | Windows path HasPrefix flaw | Workspace root rejection added; `filepath.Rel` used for path comparison |

### Weak Points (Feature Gaps)

| ID | Description |
|----|-------------|
| W1 | No MCP tools – limited extensibility |
| W2 | Gemini CLI not wired – cannot delegate to Gemini CLI |
| W3 | Git push blocked – no opt-in for trusted agents |
| W4 | Sypher-mini lacks file tools – relies on exec for file ops |
| W5 | No git discovery tool |
| W6 | Stream_command allowlist – gemini not in default list |
| W7 | Workspace restriction – E:\demo requires allow_dirs |

### Exploit Scenarios and Mitigations

| Exploit | Vector | Mitigation |
|---------|--------|------------|
| E1 | Config sets `EnableDenyPatterns: false` | Env var required (Picoclaw) |
| E2 | Malicious URL in web_fetch injects instructions | Prefix guard; avoid fetching untrusted URLs |
| E3 | Command `cat ../../../etc/passwd` with valid working_dir | Path extraction and validation in exec |
| E4 | Workspace set to `E:\` | Root path rejection; workspace cannot be filesystem root |

---

## Built-in protections

### 1. Workspace restriction

**Default:** `restrict_to_workspace: true`

- Exec commands run with working directory inside workspace
- File access (read/write) limited to workspace
- Prevents access to system paths outside `~/.sypher-mini/workspace`
- **Hardened:** Workspace cannot be a filesystem root (e.g. `E:\`, `/`); `filepath.Rel` used for path comparison
- **Command path validation:** Paths inside the command string (e.g. `cat /etc/passwd`) are validated; paths outside workspace are blocked

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

Every exec command (including `cli run`) is logged to `~/.sypher-mini/audit/{task_id}.log`:

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

---

## Config security notes

- **Deny patterns:** Never disable via config without understanding the risk. Picoclaw requires `PICOCLAW_DANGER_DISABLE_DENY=1` to honor disable.
- **Workspace:** Do not set workspace to a drive root (`E:\`, `C:\`) or `/`; these are rejected.
- **allow_dirs:** When added, only list directories you explicitly trust for `working_dir` override.

---

## Safe deployment checklist

- [ ] API keys in environment variables, not config
- [ ] `restrict_to_workspace: true` (default)
- [ ] `allow_from` set for WhatsApp
- [ ] Audit logging enabled
- [ ] Workspace is not a filesystem root
- [ ] Deny patterns enabled (do not disable without env var)
