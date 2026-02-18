# Sypher-mini Quick Reference

One-page cheat sheet.

---

## Commands

```bash
sypher onboard
sypher agent -m "message"
sypher agent
sypher gateway
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
