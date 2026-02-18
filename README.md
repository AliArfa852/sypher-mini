# Sypher-mini

**Coding-centric AI agent pipeline** — A lightweight Go core with optional Node extensions, supporting multiple LLM providers, WhatsApp connectivity, per-task audit logging, and server monitoring.

<p>
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/license-MIT-green" alt="License">
</p>

---

## Overview

Sypher-mini combines **PicoClaw-style** efficiency with **OpenClaw-style** flexibility:

- **Go core** — Fast, single binary, minimal footprint
- **Multi-provider** — Cerebras, OpenAI, Anthropic, Gemini with cheap-first routing
- **WhatsApp** — Bridge (WebSocket) or Baileys extension
- **Audit & security** — Per-task command logging, process tracking, deny patterns
- **Extensible** — Node.js extensions (e.g. WhatsApp Baileys)

## Features

| Feature | Description |
|---------|-------------|
| **Agent loop** | LLM-driven tool use (exec, kill) with intent parsing |
| **Task lifecycle** | States: pending → authorized → executing → monitoring → completed/failed/timeout/killed |
| **Intent parser** | Fast path for config/command/alert before LLM |
| **Capability registry** | Map tools and agents to capabilities |
| **Policy layer** | File access, network, rate limits |
| **Audit logging** | Per-task command logs in `~/.sypher-mini/audit/` |
| **Process tracker** | Kill only PIDs started by Sypher-mini |
| **HTTP monitors** | Poll URLs, alert on 4xx/5xx |
| **Health endpoint** | `GET /health` when gateway runs |
| **Metrics endpoint** | `GET /metrics` for tool/task counters |
| **Live streaming** | `tail_output`, `stream_command` tools |
| **Extension discovery** | `sypher extensions` lists extensions |
| **Safe mode** | `--safe` disables exec, LLM, kill |

## Quick Start

### 1. Prerequisites

- **Go 1.22+** ([install](https://go.dev/doc/install))
- **API key** — Cerebras, OpenAI, or Anthropic

### 2. Build

```bash
git clone https://github.com/sypherexx/sypher-mini.git
cd sypher-mini
go build -o sypher ./cmd/sypher
```

### 3. Initialize

```bash
./sypher onboard
```

Creates `~/.sypher-mini/config.json` and workspace.

### 4. Configure

Edit `~/.sypher-mini/config.json` and set your API key:

```json
{
  "providers": {
    "cerebras": { "api_key": "YOUR_CEREBRAS_API_KEY" },
    "openai": { "api_key": "YOUR_OPENAI_API_KEY" }
  }
}
```

Or use environment variables:

```bash
export CEREBRAS_API_KEY="your-key"
# or
export OPENAI_API_KEY="your-key"
```

### 5. Chat

```bash
./sypher agent -m "What is 2+2?"
```

## Commands

| Command | Description |
|---------|-------------|
| `sypher onboard` | Initialize config and workspace |
| `sypher agent -m "msg"` | One-shot message |
| `sypher agent` | Interactive agent loop |
| `sypher gateway` | Start gateway (HTTP, WhatsApp bridge) |
| `sypher status` | Show config and status |
| `sypher config get <path>` | Read config value |
| `sypher config set <path> <value>` | Write config value |
| `sypher agents list` | List agents |
| `sypher monitors list` | List monitors |
| `sypher audit show <task_id>` | View audit log |
| `sypher replay <task_id>` | Replay stored task |
| `sypher cancel <task_id>` | Cancel running task (via gateway) |
| `sypher extensions` | List discovered extensions |
| `sypher version` | Show version |

### Global flags

- `--safe` — Disable exec, remote API calls, and task killing

## Project Structure

```
sypher-mini/
├── cmd/sypher/           # CLI entry point
├── pkg/
│   ├── agent/            # Agent loop, context, bootstrap files
│   ├── config/           # Config load, validation
│   ├── routing/          # Agent bindings, route resolution
│   ├── providers/        # Cerebras, OpenAI, Anthropic, Gemini
│   ├── channels/         # WhatsApp bridge
│   ├── tools/            # exec, kill, contract
│   ├── audit/            # Per-task command logging
│   ├── process/          # PID tracking
│   ├── monitor/          # HTTP + process monitors
│   ├── bus/              # Event bus
│   ├── task/             # Task state machine
│   ├── policy/           # Permission evaluation
│   ├── capabilities/     # Capability registry
│   ├── intent/           # Intent parser
│   └── observability/    # Health, metrics
├── extensions/
│   └── whatsapp-baileys/ # Baileys Node extension
├── config/
│   └── config.example.json
└── docs/                 # Setup and configuration guides
```

## Module Status

See [MODULES.md](MODULES.md) for a complete list of implemented modules and their status (completed, testing, pending, not implemented).

## Documentation

| Document | Description |
|----------|-------------|
| [SETUP.md](docs/SETUP.md) | Full setup guide (install, configure, run) |
| [CONFIGURATION.md](docs/CONFIGURATION.md) | Complete config reference |
| [ARCHITECTURE.md](docs/ARCHITECTURE.md) | Architecture and design |
| [WHATSAPP.md](docs/WHATSAPP.md) | WhatsApp bridge and Baileys setup |
| [SECURITY.md](docs/SECURITY.md) | Security model and best practices |
| [QUICK_REFERENCE.md](docs/QUICK_REFERENCE.md) | One-page cheat sheet |

## Workspace Layout

```
~/.sypher-mini/
├── config.json           # Main config
├── workspace/            # Default agent workspace
├── workspace-{agent_id}/ # Per-agent workspace (optional)
├── audit/                # Per-task command logs
├── replay/               # Stored task replays (if enabled)
└── whatsapp-auth/        # Baileys auth (if using extension)
```

### Bootstrap files (in workspace)

| File | Purpose |
|------|---------|
| `AGENTS.md` / `AGENT.md` | Role and instructions |
| `SOUL.md` | Personality, values, identity |
| `USER.md` | User context |
| `IDENTITY.md` | Override identity |

## Security

- **Workspace restriction** — Exec limited to workspace by default
- **Deny patterns** — Blocks `rm -rf`, `sudo`, `chmod`, etc.
- **Kill scope** — Only PIDs started by Sypher-mini for current task
- **Audit** — All commands logged
- **Safe mode** — `--safe` for recovery/debugging

## License

MIT
