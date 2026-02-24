# Sypher-mini Architecture

Technical overview of the Sypher-mini agent pipeline.

---

## High-level flow

```
Inbound message (CLI / WhatsApp / HTTP)
    ↓
Intent parser (fast path: config, command, alert)
    ↓
Agent routing (bindings → agent_id)
    ↓
Task created (pending → authorized)
    ↓
Agent loop (LLM + tools)
    ↓
Outbound message
```

---

## Components

### 1. CLI (`cmd/sypher`)

Entry point. Subcommands:

- `agent` — Run agent loop (one-shot or interactive)
- `gateway` — HTTP server + channels + agent loop
- `config` — Get/set config
- `agents` — List agents
- `monitors` — List monitors
- `audit` — Show task audit log
- `replay` — Replay stored task
- `cancel` — Cancel task via gateway API
- `onboard` — Initialize config and workspace

### 2. Message bus (`pkg/bus`)

- **Inbound** — Channels publish; agent loop consumes
- **Outbound** — Agent loop publishes; channels/CLI consume
- **Event bus** — Internal events (task.started, tool_request, etc.)

### 3. Intent parser (`pkg/intent`)

Rule-based classifier. Runs before the agent loop.

| Intent | Fast path | LLM |
|--------|-----------|-----|
| `config_change` | Config get/set | No |
| `command` | Direct exec (or route to agent) | No |
| `emergency_alert` | Notify | No |
| `question`, `chat`, `automation_request` | Agent loop | Yes |

### 4. Routing (`pkg/routing`)

Resolves `(channel, account_id, peer)` → `agent_id`.

Priority: peer match > account match > channel wildcard > default agent.

### 5. Task lifecycle (`pkg/task`)

States:

```
pending → authorized → executing ⇄ monitoring → completed | failed | timeout | killed
```

- **pending** — Message received, not yet routed
- **authorized** — Routed, policy OK
- **executing** — Agent loop active
- **monitoring** — Waiting on tool call
- **completed** — Success
- **failed** — Error, max iterations, or tool failure
- **timeout** — Task timeout exceeded
- **killed** — User cancelled

### 6. Agent loop (`pkg/agent`)

1. Consume inbound message
2. Parse intent (fast path if applicable)
3. Resolve agent
4. Create task
5. Call LLM with tools
6. Execute tool calls (exec, kill)
7. Loop until no more tool calls or max iterations
8. Publish outbound

### 7. Providers (`pkg/providers`)

- **OpenAI-compatible** — Cerebras, OpenAI (shared implementation)
- **Factory** — Selects provider by `routing_strategy` (e.g. cheap_first)
- **Model format** — `provider/model` (e.g. `cerebras/llama-3.1-70b`)

### 8. Tools (`pkg/tools`)

| Tool | Description |
|------|-------------|
| `exec` | Run shell command; deny patterns, workspace check |
| `kill` | Kill PID (only if owned by current task) |
| `invoke_cli_agent` | Run configured CLI agent (e.g. Gemini) with task |

### 8b. CLI session manager (`pkg/clisession`)

- **SessionStore** — Active terminals with ID, tag, output buffer
- **Per-session buffer** — Last N lines (default 10, max 100 via `--tail`)
- **WhatsApp commands** — `cli list`, `cli new -m 'tag'`, `cli <N> [--tail N]`, `cli run <N> <cmd>`

### 9. Audit (`pkg/audit`)

Per-task log: `{task_id}.log` with timestamp, command, cwd, exit code, output summary.

### 10. Process tracker (`pkg/process`)

Maps `task_id` → PIDs. Kill tool only allows PIDs in this map.

### 11. Policy (`pkg/policy`)

- **File access** — Path globs, agent_ids, read/write
- **Network** — allow_domains, deny_domains
- **Rate limits** — requests_per_minute per agent/tool

### 12. Capabilities (`pkg/capabilities`)

Maps tools and agents to capabilities (e.g. `code_generation`, `notify_user`). Used for capability-based routing.

### 13. Channels (`pkg/channels`)

- **WhatsApp bridge** — WebSocket client; relays inbound/outbound

### 14. Monitor (`pkg/monitor`)

- **HTTP** — Poll URL; alert on 4xx/5xx
- **Process** — (Scaffold) Watch process output for error patterns

### 15. Observability (`pkg/observability`)

- **Health** — `GET /health` with status and checks

---

## Data flow

### CLI one-shot

```
User: sypher agent -m "hello"
    → msgBus.PublishInbound
    → loop.Run (goroutine)
    → loop consumes, processes
    → msgBus.PublishOutbound
    → main subscribes, prints
```

### Gateway + WhatsApp

```
WhatsApp bridge (WebSocket)
    → inbound message
    → msgBus.PublishInbound
    → loop consumes, processes
    → msgBus.PublishOutbound
    → bridge forwards to WhatsApp
```

### Baileys extension

```
Baileys (Node) receives WhatsApp message
    → POST /inbound to gateway
    → msgBus.PublishInbound
    → (same as above)
```

---

## Extension contract

Extensions live in `extensions/{name}/` with `sypher.extension.json`:

```json
{
  "id": "whatsapp-baileys",
  "version": "1.0.0",
  "sypher_mini_version": ">=0.1.0",
  "capabilities": ["channel"],
  "entry": "dist/index.js",
  "node_min_version": "18",
  "setup_script": "npm run setup",
  "start_script": "npm run start"
}
```

### Extension structure (add, modify, remove)

| File | Purpose |
|------|---------|
| `sypher.extension.json` | Manifest: id, entry, node_min_version, setup_script, start_script |
| `scripts/setup.cjs` | Setup: check Node version, `npm install`, `npm run build` |
| `package.json` | `"type": "module"` for ESM; `engines.node`: `>=18` |
| `src/` | TypeScript source; build to `dist/` |

**Adding an extension:** Create `extensions/{name}/`, add manifest, setup script, and wire into gateway if needed.

**Modifying:** Edit manifest or scripts; rebuild with `npm run build` in extension dir.

**Removing:** Delete `extensions/{name}/`; gateway skips missing extensions.

**Gateway auto-setup:** When spawning, gateway checks Node version, runs `setup_script` if dist missing, then `start_script` or `node entry`.

Protocol: HTTP callback for inbound, HTTP POST for outbound.
