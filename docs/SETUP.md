# Sypher-mini Setup Guide

Complete step-by-step setup for Sypher-mini on Windows, macOS, and Linux.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Platform setup](#platform-setup)
3. [Installation](#installation)
4. [Build & Clean](#build--clean)
5. [Initial Configuration](#initial-configuration)
6. [API Keys](#api-keys)
7. [First Run](#first-run)
8. [Gateway Mode](#gateway-mode)
9. [WhatsApp Setup](#whatsapp-setup)
10. [Troubleshooting](#troubleshooting)

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
| **Node.js** | 20+ for WhatsApp Baileys extension |
| **npm** | Installing extension dependencies |

### Verify installation

```bash
go version   # Should show go1.22 or higher
git --version
```

### Platform setup

#### macOS (Homebrew)

```bash
# Install Homebrew (if needed): https://brew.sh
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Go (for sypher binary)
brew install go

# Node.js 20+ (for WhatsApp Baileys extension)
brew install node@20
# Link if needed: brew link node@20 --force --overwrite
# Or use latest: brew install node

# Verify
go version   # go1.22+
node -v      # v20.x or higher
npm -v
```

#### Linux (Debian/Ubuntu, Raspberry Pi)

```bash
# Go
sudo apt-get update
sudo apt-get install -y golang-go

# Node.js 20+ (distro default is often 18; use NodeSource)
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# Verify
go version   # go1.22+
node -v     # v20.x or higher
npm -v
```

Alternative (nvm, no sudo):

```bash
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
nvm install 20
nvm use 20
```

#### Windows (cmd)

```cmd
REM Go: https://go.dev/dl/ — download .msi, run installer, add C:\Go\bin to PATH
REM Or via winget:
winget install GoLang.Go

REM Node.js 20+: https://nodejs.org/ — download LTS .msi, run installer
REM Or via winget:
winget install OpenJS.NodeJS.LTS

REM Verify (new cmd window after install)
go version
node -v
npm -v
```

Build from project root:

```cmd
cd path\to\sypher-mini
.\build.ps1 build
REM Or manually: cd extensions\whatsapp-baileys && npm install && npm run build
go build -o sypher.exe .\cmd\sypher
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

### Workspace override (allow_dirs)

To run commands in directories outside the default workspace (e.g. `E:\demo`), add them to `tools.exec.allow_dirs`:

```json
{
  "tools": {
    "exec": {
      "allow_dirs": ["E:\\demo", "D:\\projects"]
    }
  }
}
```

**Windows:** Use double backslashes (`\\`) in JSON. Paths support `~` for home directory.

**Example (E:\\demo setup):** To let the agent create repos and run commands in `E:\demo`, add `"E:\\demo"` to `allow_dirs`. The agent can then `mkdir`, `git init`, and run tools in that directory.

**Git discovery:** The agent can find git repos via exec, e.g. `find . -name .git -type d` (Unix) or `dir /s /b .git` (Windows), or `git rev-parse --show-toplevel` when already inside a repo.

**Platform commands:** The agent receives runtime context (OS, shell) automatically. See [PLATFORMS.md](PLATFORMS.md) for the command compatibility matrix.

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

### Gemini CLI (optional)

To delegate code generation to the Gemini CLI:

1. Install the [Gemini CLI](https://ai.google.dev/gemini-api/docs/cli) and ensure `gemini` is in your PATH.
2. Add an agent with `command` and `args` in `agents.list`:

```json
{
  "agents": {
    "list": [
      { "id": "gemini-cli", "command": "gemini", "args": ["--model", "gemini-2.0"] }
    ]
  }
}
```

3. Add `gemini` to `tools.live_monitoring.allowed_commands` to stream long outputs:

```json
{
  "tools": {
    "live_monitoring": {
      "allowed_commands": ["npm run", "go run", "tail -f", "gemini"]
    }
  }
}
```

The agent can then use the `invoke_cli_agent` tool to run Gemini CLI with a task.

**Tool-capable models:** Use a model that supports tool calls (e.g. `llama-3.1-70b`, `gpt-4o`, `gemini-2.0`). Some models return text only and will not invoke tools.

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

To use a WebSocket bridge instead of QR (Baileys), set `use_baileys: false` and `bridge_url`:

1. Run a WhatsApp bridge (e.g. [whatsapp-web.js](https://github.com/pedroslopez/whatsapp-web.js) or similar) that exposes a WebSocket.
2. Edit config:

```json
{
  "channels": {
    "whatsapp": {
      "enabled": true,
      "use_baileys": false,
      "bridge_url": "ws://localhost:3001",
      "allow_from": ["+1234567890"]
    }
  }
}
```

3. Start gateway: `sypher gateway`

---

## E:\demo Test Scenario Walkthrough

This walkthrough demonstrates creating a repo in `E:\demo`, generating a Python file with animation via Gemini CLI, and committing to git.

### Prerequisites

- Workspace or `tools.exec.allow_dirs` includes `E:\demo`
- [Gemini CLI](https://ai.google.dev/gemini-api/docs/cli) installed and in PATH
- Git installed
- If push needed: enable `tools.exec.allow_git_push` in config

### Step 1: Configure allow_dirs

Add `E:\demo` to `tools.exec.allow_dirs` in `~/.sypher-mini/config.json`:

```json
{
  "tools": {
    "exec": {
      "allow_dirs": ["E:\\demo"]
    }
  }
}
```

### Step 2: Configure Gemini CLI (optional)

Add a Gemini CLI agent and allow streaming:

```json
{
  "agents": {
    "list": [
      { "id": "gemini-cli", "command": "gemini", "args": ["--model", "gemini-2.0"] }
    ]
  },
  "tools": {
    "live_monitoring": {
      "allowed_commands": ["npm run", "go run", "tail -f", "gemini"]
    }
  }
}
```

### Step 3: Run the scenario

Via WhatsApp or `sypher agent -m "..."`:

1. **Create repo:** "Create a repo called test sypher in E:\demo"
   - Agent runs: `mkdir E:\demo\test sypher`, `cd E:\demo\test sypher`, `git init`
2. **Generate code:** "Create hello.py with animation using Gemini"
   - Agent uses `invoke_cli_agent` with task "Create hello.py with animation"
3. **Commit:** "Add and commit the changes"
   - Agent runs: `git add .`, `git commit -m "Add hello.py"`
4. **Push (optional):** Enable `allow_git_push` if you need to push to a remote.

### Troubleshooting

- **"working_dir outside workspace"** — Ensure `E:\demo` is in `allow_dirs`
- **"gemini: command not found"** — Install Gemini CLI and add to PATH
- **"git push blocked"** — Set `tools.exec.allow_git_push: true` for trusted agents

---

## Troubleshooting

### "go: command not found"

- **Windows:** Add Go to PATH: `C:\Program Files\Go\bin` or `C:\Go\bin`
- **Linux/macOS:** Ensure `$GOPATH/bin` or `$HOME/go/bin` is in PATH

### "Config load error"

- Ensure `sypher onboard` was run
- Check file exists: `~/.sypher-mini/config.json` (or `%USERPROFILE%\.sypher-mini\config.json` on Windows)

### "no LLM provider configured"

- Set at least one API key: `GEMINI_API_KEY`, `CEREBRAS_API_KEY`, `OPENAI_API_KEY`, or `ANTHROPIC_API_KEY`
- Copy `.env.example` to `.env`, add your key, then run `sypher agent` again
- Verify env var: `echo $GEMINI_API_KEY` (Linux/macOS) or `echo $env:GEMINI_API_KEY` (PowerShell)

### "models/llama-3.1-70b is not found" (404 from Gemini)

- Config had `cerebras/llama-3.1-70b` but only Gemini was configured. Fixed: providers now use their default model when the config model does not match.
- Optional: set `GEMINI_MODEL=gemini-2.5-flash-lite` to override `agents.defaults.model` when using Gemini

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
