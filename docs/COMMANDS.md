# Sypher-mini Commands Reference

Dedicated reference for all CLI commands.

---

## Build & Clean

### Make (Linux/macOS/Git Bash)

| Command | Description |
|---------|-------------|
| `make build` | Build extensions + sypher binary |
| `make build-go` | Build sypher only (skip extensions) |
| `make extensions` | Install and build Node extensions |
| `make clean` | Remove build/, Go cache, extension node_modules/dist |
| `make rebuild` | Clean then build |
| `make test` | Run tests |
| `make run` | Build and run gateway |
| `make docker` | Build Docker image |
| `make docker-run` | Run with docker-compose |
| `make docker-down` | Stop docker-compose |

### PowerShell (Windows)

```powershell
.\build.ps1 build      # Build extensions + sypher
.\build.ps1 extensions # Extensions only
.\build.ps1 clean      # Remove build/ and extension artifacts
.\build.ps1 rebuild    # Clean then build
.\build.ps1 test       # Run tests
.\build.ps1 run        # Build and run gateway
```

### Go

```bash
go build -o sypher ./cmd/sypher   # Direct build
go clean -cache                   # Clean Go cache
```

---

## CLI Reference Table

| Command | Description |
|---------|-------------|
| `sypher onboard` | Initialize config and workspace |
| `sypher agent -m "msg"` | One-shot message |
| `sypher agent` | Interactive agent loop |
| `sypher gateway` | Start gateway |
| `sypher whatsapp --connect` | Configure WhatsApp (Baileys) and prompt for allow_from |
| `sypher status` | Show config and status |
| `sypher config get <path>` | Read config value |
| `sypher config set <path> <value>` | Write config value |
| `sypher agents list` | List agents |
| `sypher monitors list` | List monitors |
| `sypher audit show <task_id>` | View audit log |
| `sypher replay <task_id>` | Replay stored task |
| `sypher cancel <task_id>` | Cancel running task |
| `sypher extensions` | List extensions |
| `sypher commands list` | List per-command configs |
| `sypher version` | Show version |

---

## Command Details

### onboard

Initialize config and workspace. Creates `~/.sypher-mini/config.json`, workspace directory, subdirs (memory, sessions, state, cron, skills, code-projects), and bootstrap template files (AGENTS.md, SOUL.md, USER.md, etc.).

```bash
sypher onboard
```

---

### agent

Run the agent. Use `-m "message"` for one-shot, or run without `-m` for interactive mode (blocks waiting for inbound; use gateway for channel input).

```bash
sypher agent -m "What is 2+2?"
sypher agent
```

---

### gateway

Start the gateway: HTTP server (health, inbound, cancel, metrics), WhatsApp bridge or Baileys, and monitors.

```bash
sypher gateway
```

---

### whatsapp --connect

Configure WhatsApp (Baileys): enables channel, sets `use_baileys=true`, `baileys_url=http://localhost:3002`. Optionally set allow_from.

```bash
sypher whatsapp --connect
sypher whatsapp --connect --allow-from +1234567890
```

---

### status

Show config path, deployment mode, agent count, task timeout, WhatsApp enabled.

```bash
sypher status
```

---

### config

Read or write config values. Supported paths: `agents`, `agents.list`, `task.timeout_sec`, `channels`.

```bash
sypher config get channels
sypher config set task.timeout_sec 600
```

---

### agents list

List configured agents with default marker.

```bash
sypher agents list
```

---

### monitors list

List HTTP and process monitors.

```bash
sypher monitors list
```

---

### audit show

Display audit log for a task.

```bash
sypher audit show <task_id>
```

---

### replay

Display stored task replay (read-only, no re-execution).

```bash
sypher replay <task_id>
```

---

### cancel

Cancel a running task via gateway API.

```bash
sypher cancel <task_id>
```

---

### extensions

List discovered extensions (from `extensions/` with `sypher.extension.json`).

```bash
sypher extensions
```

---

### commands list

List per-command configs from `~/.sypher-mini/commands/`.

```bash
sypher commands list
```

---

### version

Show version.

```bash
sypher version
```

---

## Global Flags

| Flag | Description |
|------|-------------|
| `--safe` | Disable exec, remote API calls, task killing |

---

## Future Commands (Placeholders)

| Command | Description |
|---------|-------------|
| `sypher cron list` | List scheduled jobs |
| `sypher cron add ...` | Add scheduled job |
| `sypher telegram --connect` | Configure Telegram |
