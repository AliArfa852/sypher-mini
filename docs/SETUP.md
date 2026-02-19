# Sypher-mini Setup Guide

Complete step-by-step setup for Sypher-mini on Windows, macOS, and Linux.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Installation](#installation)
3. [Build & Clean](#build--clean)
4. [Initial Configuration](#initial-configuration)
5. [API Keys](#api-keys)
6. [First Run](#first-run)
7. [Gateway Mode](#gateway-mode)
8. [WhatsApp Setup](#whatsapp-setup)
9. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Required

| Requirement | Version | Notes |
|-------------|---------|-------|
| **Go** | 1.22+ | [Download](https://go.dev/dl/) |
| **Git** | Any | For cloning |

### Optional (for extensions)

| Requirement | Purpose |
|-------------|---------|
| **Node.js** | WhatsApp Baileys extension |
| **npm** | Installing extension dependencies |

### Verify installation

```bash
go version   # Should show go1.22 or higher
git --version
```

---

## Installation

### Option A: Build from source (recommended)

```bash
# Clone the repository
git clone https://github.com/sypherexx/sypher-mini.git
cd sypher-mini

# Build
go build -o sypher ./cmd/sypher

# Optional: install to PATH
# Linux/macOS:
sudo mv sypher /usr/local/bin/

# Windows (PowerShell):
# Copy sypher.exe to a directory in your PATH, e.g. C:\Go\bin
```

### Option B: Using go install

```bash
go install github.com/sypherexx/sypher-mini/cmd/sypher@latest
# Binary will be in $GOPATH/bin or $HOME/go/bin
```

### Option C: Pre-built binaries (when available)

Download from [Releases](https://github.com/sypherexx/sypher-mini/releases) for your platform:

- `sypher-windows-amd64.exe`
- `sypher-linux-amd64`
- `sypher-linux-arm64`
- `sypher-darwin-amd64`
- `sypher-darwin-arm64`

### Build & clean

Build automatically installs and builds Node extensions (e.g. WhatsApp Baileys).

| Command | Description |
|---------|-------------|
| `make build` | Build extensions + sypher (Linux/macOS/Git Bash) |
| `make extensions` | Install and build Node extensions only |
| `make clean` | Remove build artifacts, Go cache, extension node_modules |
| `make rebuild` | Clean then build |
| `make test` | Run tests |
| `.\build.ps1 build` | Build extensions + sypher (Windows PowerShell) |
| `.\build.ps1 rebuild` | Clean then build |
| `go build -o sypher ./cmd/sypher` | Direct build (no extensions) |
| `go clean -cache` | Clean Go cache |

---

## Initial Configuration

### Step 1: Run onboard

```bash
sypher onboard
```

This creates:

- `~/.sypher-mini/config.json` — Main configuration
- `~/.sypher-mini/workspace/` — Default workspace directory

**Windows:** Config path is `%USERPROFILE%\.sypher-mini\config.json`

### Step 2: Locate your config file

| OS | Path |
|----|------|
| Linux/macOS | `~/.sypher-mini/config.json` |
| Windows | `C:\Users\<You>\.sypher-mini\config.json` |

### Step 3: Edit config

Open the config file in your editor. The minimal required change is adding an API key.

---

## API Keys

Sypher-mini supports multiple LLM providers. You need **at least one** API key.

### Provider options

| Provider | Config key | Env variable | Get key |
|----------|------------|--------------|---------|
| **Cerebras** | `providers.cerebras.api_key` | `CEREBRAS_API_KEY` | [Cerebras Cloud](https://cloud.cerebras.ai/) |
| **OpenAI** | `providers.openai.api_key` | `OPENAI_API_KEY` | [OpenAI Platform](https://platform.openai.com/api-keys) |
| **Anthropic** | `providers.anthropic.api_key` | `ANTHROPIC_API_KEY` | [Anthropic Console](https://console.anthropic.com/) |
| **Gemini** | `providers.gemini.api_key` | `GEMINI_API_KEY` | [Google AI Studio](https://aistudio.google.com/) |

### Method 1: Config file

Edit `~/.sypher-mini/config.json`:

```json
{
  "providers": {
    "cerebras": { "api_key": "your-cerebras-key-here" },
    "openai": { "api_key": "" },
    "anthropic": { "api_key": "" },
    "gemini": { "api_key": "" }
  }
}
```

### Method 2: Environment variables (recommended for security)

```bash
# Linux/macOS - add to ~/.bashrc or ~/.zshrc
export CEREBRAS_API_KEY="your-key"
# or
export OPENAI_API_KEY="your-key"

# Windows PowerShell (current session)
$env:CEREBRAS_API_KEY = "your-key"

# Windows (permanent - System Properties > Environment Variables)
# Add CEREBRAS_API_KEY to User variables
```

### Routing strategy

By default, Sypher-mini uses `cheap_first` routing: it tries cheaper providers first (e.g. Cerebras) before falling back to more expensive ones (OpenAI, Anthropic).

---

## First Run

### One-shot message

```bash
sypher agent -m "Hello, what can you do?"
```

Expected: A response from the LLM (or a placeholder if no API key is set).

### Interactive agent

```bash
sypher agent
```

Starts the agent loop. Use Ctrl+C to stop. For interactive input, use the gateway with a channel (e.g. WhatsApp).

### Safe mode (no exec, no LLM)

```bash
sypher --safe agent -m "test"
```

Useful for testing config without executing commands or calling APIs.

### Check status

```bash
sypher status
```

Shows config path, deployment mode, agents, task timeout, and WhatsApp status.

---

## Gateway Mode

The gateway runs the agent loop plus HTTP server and optional channels (WhatsApp).

### Start gateway

```bash
sypher gateway
```

This:

- Starts HTTP server on port **18790**
- Exposes `/health` and `/inbound` endpoints
- Connects to WhatsApp bridge if configured
- Runs the agent loop to process messages

### Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check; returns `{ "status": "ok", "checks": {...} }` |
| `/inbound` | POST | Receive inbound messages (e.g. from Baileys extension) |
| `/cancel` | POST | Cancel a task; body: `{ "task_id": "..." }` |

### Health check

```bash
curl http://localhost:18790/health
```

---

## WhatsApp Setup

### Option 1: Baileys (recommended) — `sypher whatsapp --connect`

The simplest path: run the connect command, then start the gateway.

```bash
sypher whatsapp --connect --allow-from +1234567890
sypher gateway
```

Scan the QR code in the terminal with WhatsApp (Settings → Linked Devices). Auth is saved to `~/.sypher-mini/whatsapp-auth/`.

See [docs/WHATSAPP.md](WHATSAPP.md) for details.

### Option 2: WebSocket bridge

1. Run a WhatsApp bridge (e.g. [whatsapp-web.js](https://github.com/pedroslopez/whatsapp-web.js) or similar) that exposes a WebSocket.
2. Edit config:

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

3. Start gateway: `sypher gateway`

---

## Troubleshooting

### "go: command not found"

- **Windows:** Add Go to PATH: `C:\Program Files\Go\bin` or `C:\Go\bin`
- **Linux/macOS:** Ensure `$GOPATH/bin` or `$HOME/go/bin` is in PATH

### "Config load error"

- Ensure `sypher onboard` was run
- Check file exists: `~/.sypher-mini/config.json` (or `%USERPROFILE%\.sypher-mini\config.json` on Windows)

### "no LLM provider configured"

- Set at least one API key in config or environment
- Verify env var is exported: `echo $CEREBRAS_API_KEY` (Linux/macOS) or `echo $env:CEREBRAS_API_KEY` (PowerShell)

### "Command blocked by safety guard"

- The exec tool blocks dangerous commands (e.g. `rm -rf`, `sudo`)
- Use `restrict_to_workspace: false` only in controlled environments (security risk)

### Gateway won't start

- Check port 18790 is free: `netstat -an | findstr 18790` (Windows) or `lsof -i :18790` (Linux/macOS)
- Run with `--safe` to disable external connections during debug

### Cancel command fails

- `sypher cancel <task_id>` requires the gateway to be running
- It sends a POST to `http://localhost:18790/cancel`
- Set `SYPHER_GATEWAY_URL` if gateway runs elsewhere
