# Sypher-mini Quick Reference

One-page cheat sheet.

---

## Build & Clean

```bash
make build      # Build extensions + sypher (Linux/macOS/Git Bash)
make extensions # Install and build Node extensions
make clean      # Remove build/, Go cache, extension node_modules/dist
make rebuild    # Clean then build
make test       # Run tests
make run        # Build and run gateway
```

```powershell
.\build.ps1 build      # Windows PowerShell (extensions + binary)
.\build.ps1 extensions # Extensions only
.\build.ps1 rebuild
```

```bash
go build -o sypher ./cmd/sypher    # Direct build (no extensions)
go clean -cache                    # Clean Go cache
```

---

## Commands

```bash
sypher onboard
sypher agent -m "message"
sypher agent
sypher gateway
sypher whatsapp --connect [--allow-from +1234567890]
sypher status
sypher config get <path>
sypher config set <path> <value>
sypher agents list
sypher monitors list
sypher audit show <task_id>
sypher replay <task_id>
sypher cancel <task_id>
sypher version
sypher --safe gateway
```

---

## Config paths

| Path | Default |
|------|---------|
| `~/.sypher-mini/config.json` | Linux/macOS |
| `%USERPROFILE%\.sypher-mini\config.json` | Windows |

---

## Env vars

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

## Gateway endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `http://localhost:18790/health` | GET | Health check |
| `http://localhost:18790/inbound` | POST | Inbound messages |
| `http://localhost:18790/cancel` | POST | Cancel task |

---

## Workspace

```
~/.sypher-mini/
├── config.json
├── workspace/
│   ├── AGENTS.md, AGENT.md, SOUL.md, USER.md, IDENTITY.md
│   ├── HEARTBEAT.md, TOOLS.md
│   ├── memory/, sessions/, state/, cron/, skills/, code-projects/
├── audit/
├── replay/
└── whatsapp-auth/   (Baileys)
```

---

## Bootstrap files

| File | Purpose |
|------|---------|
| AGENTS.md / AGENT.md | Role |
| SOUL.md | Personality |
| USER.md | User context |
| IDENTITY.md | Override |
| HEARTBEAT.md | Optional heartbeat |
| TOOLS.md | Tool guidance |
