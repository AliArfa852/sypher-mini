# Sypher-mini Breakdown

Complete reference for every command, service connection, and workspace layout.

---

## Table of Contents

1. [Build & Clean](#build--clean)
2. [Commands](#commands)
3. [Connecting Services](#connecting-services)
4. [Config and Environment](#config-and-environment)
5. [Gateway Endpoints](#gateway-endpoints)
6. [Workspace Layout](#workspace-layout)
7. [Provider Architecture](#provider-architecture)
8. [Security Sandbox](#security-sandbox)

---

## Build & Clean

### Make (Linux/macOS/Git Bash)

| Command | Description |
|---------|-------------|
| `make build` | Build extensions (npm install + build) and sypher binary |
| `make build-go` | Build sypher binary only (skip extensions) |
| `make extensions` | Install and build all Node extensions |
| `make clean` | Remove build/, Go cache, and extension node_modules/dist |
| `make rebuild` | Clean then build |
| `make test` | Run tests |
| `make run` | Build and run gateway |
| `make docker` | Build Docker image |
| `make docker-run` | Run with docker-compose |
| `make docker-down` | Stop docker-compose |

### Go (any platform)

```bash
# Build
go build -o sypher ./cmd/sypher

# Build to build/
go build -o build/sypher ./cmd/sypher

# Clean Go cache
go clean -cache

# Clean build artifacts (PowerShell)
Remove-Item -Recurse -Force build -ErrorAction SilentlyContinue

# Clean build artifacts (Bash)
rm -rf build
```

### Docker

```bash
# Build image
docker build -t sypher-mini:latest .

# Run with compose
docker-compose up -d

# Stop
docker-compose down

# Rebuild (no cache)
docker-compose build --no-cache
docker-compose up -d
```

### PowerShell (Windows)

```powershell
.\build.ps1 build      # Build extensions + sypher binary
.\build.ps1 extensions # Install and build extensions only
.\build.ps1 clean      # Remove build/, Go cache, extension node_modules/dist
.\build.ps1 rebuild    # Clean then build
.\build.ps1 test       # Run tests
.\build.ps1 run        # Build and run gateway
```

### Full rebuild (clean slate)

```bash
make clean
make build
# or (PowerShell)
.\build.ps1 rebuild
# or (Go only)
go clean -cache
rm -rf build   # or: Remove-Item -Recurse -Force build
go build -o build/sypher ./cmd/sypher
```

---

## Commands

| Command | Description |
|---------|-------------|
| `sypher onboard` | Initialize config and workspace |
| `sypher agent -m "msg"` | One-shot message |
| `sypher agent` | Interactive agent loop |
| `sypher gateway` | Start gateway (channels, monitors) |
| `sypher whatsapp --connect` | Configure WhatsApp (Baileys) |
| `sypher status` | Show config and status |
| `sypher config get <path>` | Read config value |
| `sypher config set <path> <value>` | Write config value |
| `sypher agents list` | List agents |
| `sypher monitors list` | List monitors |
| `sypher audit show <task_id>` | View audit log |
| `sypher replay <task_id>` | Replay stored task |
| `sypher cancel <task_id>` | Cancel running task |
| `sypher extensions` | List discovered extensions |
| `sypher commands list` | List per-command configs |
| `sypher version` | Show version |

**Global flags:** `--safe` — Disable exec, remote API calls, task killing.

**Future:** `sypher cron list`, `sypher cron add`, `sypher telegram --connect`.

---

## Connecting Services

### Channels (Chat Apps)

| Channel | Setup | Notes |
|---------|-------|-------|
| **WhatsApp (Baileys)** | Easy | `sypher whatsapp --connect`; Node.js 18+ required |
| **WhatsApp (Bridge)** | Medium | Requires separate WebSocket bridge process |
| **Telegram** | — | Coming soon |

---

### WhatsApp (Baileys) — Recommended

**1. Prerequisites**

- Node.js 18+
- npm

**2. Configure**

```bash
sypher whatsapp --connect --allow-from +1234567890
```

Or without allowlist (allow all; restrict in production):

```bash
sypher whatsapp --connect
```

**3. Install extension** (if not using auto-spawn)

```bash
cd extensions/whatsapp-baileys
npm install
npm run build
```

**4. Run**

```bash
sypher gateway
```

**5. Pair**

- QR code appears in the terminal
- Scan with WhatsApp: Settings → Linked Devices → Link a Device
- Auth saved to `~/.sypher-mini/whatsapp-auth/`

---

### WhatsApp (WebSocket Bridge)

Use any bridge that exposes a WebSocket and sends/receives JSON.

**1. Bridge message format**

Inbound: `{"type":"inbound","from":"+123","content":"Hello","chat_id":"+123"}`  
Outbound: `{"type":"outbound","to":"+123","content":"Reply"}`

**2. Configure** (`~/.sypher-mini/config.json`)

```json
{
  "channels": {
    "whatsapp": {
      "enabled": true,
      "bridge_url": "ws://localhost:3001",
      "allow_from": ["+1234567890"]
    }
  }
}
```

**3. Run**

1. Start your bridge on port 3001
2. `sypher gateway`

---

### LLM Providers

| Provider | Env var | Config path |
|----------|---------|-------------|
| Cerebras | `CEREBRAS_API_KEY` | `providers.cerebras.api_key` |
| OpenAI | `OPENAI_API_KEY` | `providers.openai.api_key` |
| Anthropic | `ANTHROPIC_API_KEY` | `providers.anthropic.api_key` |
| Gemini | `GEMINI_API_KEY` | `providers.gemini.api_key` |

Routing: `cheap_first` (default) — tries providers in order until one succeeds.

---

## Config and Environment

### Config paths

| OS | Path |
|----|------|
| Linux/macOS | `~/.sypher-mini/config.json` |
| Windows | `%USERPROFILE%\.sypher-mini\config.json` |

### Environment variables

| Variable | Purpose |
|----------|---------|
| `CEREBRAS_API_KEY` | Cerebras API |
| `OPENAI_API_KEY` | OpenAI API |
| `ANTHROPIC_API_KEY` | Anthropic API |
| `GEMINI_API_KEY` | Gemini API |
| `SYPHER_MINI_MODE` | deployment.mode |
| `SYPHER_GATEWAY_URL` | Base URL for cancel |
| `SYPHER_CORE_CALLBACK` | Baileys callback (default: http://localhost:18790/inbound) |

---

## Gateway Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `http://localhost:18790/health` | GET | Health check |
| `http://localhost:18790/inbound` | POST | Inbound messages |
| `http://localhost:18790/cancel` | POST | Cancel task (body: `{"task_id":"<id>"}`) |
| `http://localhost:18790/metrics` | GET | Metrics (JSON or `?format=prometheus`) |

---

## Workspace Layout

```
~/.sypher-mini/
├── config.json           # Main config
├── workspace/            # Default agent workspace
│   ├── AGENTS.md         # Role/instructions
│   ├── AGENT.md          # Alias for AGENTS.md
│   ├── SOUL.md           # Personality, values
│   ├── USER.md           # User context
│   ├── IDENTITY.md       # Override identity
│   ├── HEARTBEAT.md      # Optional heartbeat checklist
│   ├── TOOLS.md          # Tool descriptions
│   ├── memory/           # Daily memory (YYYY-MM-DD.md)
│   ├── sessions/         # Session metadata/transcripts
│   ├── state/            # Persistent state
│   ├── cron/             # Scheduled jobs
│   ├── skills/           # Workspace skills
│   └── code-projects/    # Project context
├── workspace-{agent_id}/ # Per-agent workspace (optional)
├── audit/                # Per-task command logs
├── replay/               # Stored task replays (if enabled)
└── whatsapp-auth/        # Baileys auth (if using extension)
```

### Bootstrap files

| File | Purpose |
|------|---------|
| AGENTS.md / AGENT.md | Role and instructions |
| SOUL.md | Personality, values, identity |
| USER.md | User context |
| IDENTITY.md | Override identity |
| HEARTBEAT.md | Optional heartbeat checklist |
| TOOLS.md | Tool guidance (does not control availability) |

---

## Provider Architecture

- **Routing:** `cheap_first` — tries Cerebras, then OpenAI, Anthropic, Gemini.
- **Env overrides:** API keys from env take precedence over config.
- **Fallback:** Retries with backoff, then next provider.

---

## Security Sandbox

- **Workspace restriction:** `restrict_to_workspace: true` (default) limits exec and file access to workspace.
- **Deny patterns:** Blocks `rm -rf`, `sudo`, `chmod`, etc.
- **Kill scope:** Only PIDs started by Sypher-mini for current task.
- **Audit:** All commands logged to `~/.sypher-mini/audit/`.
- **Safe mode:** `--safe` disables exec, remote APIs, task killing.

See [docs/SECURITY.md](docs/SECURITY.md) for details.

---

## Quick Start

```bash
sypher onboard
# Edit ~/.sypher-mini/config.json — set API key (or use env)
sypher whatsapp --connect --allow-from +1234567890
sypher gateway
# Scan QR code in terminal
```
