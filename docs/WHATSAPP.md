# WhatsApp Setup Guide

Connect Sypher-mini to WhatsApp via QR code (Baileys, default) or a WebSocket bridge.

---

## Table of Contents

1. [Overview](#overview)
2. [Option 1: WebSocket Bridge](#option-1-websocket-bridge)
3. [Option 2: Baileys Extension](#option-2-baileys-extension)
4. [Configuration](#configuration)
5. [Menu and Routing](#menu-and-routing)
6. [CLI Session Commands](#cli-session-commands)
7. [Security](#security)
8. [Troubleshooting](#troubleshooting)

---

## Overview

| Method | Pros | Cons |
|--------|------|------|
| **Baileys (default)** | Self-contained, no browser, QR pairing | Node.js 20+ required |
| **Bridge** | Use existing bridge (e.g. whatsapp-web.js) | Requires separate bridge process |

---

## Option 1: WebSocket Bridge

Use any WhatsApp bridge that exposes a WebSocket and sends/receives JSON messages.

### Bridge message format

**Inbound (bridge → Sypher-mini):**

```json
{
  "type": "inbound",
  "from": "+1234567890",
  "content": "Hello",
  "chat_id": "+1234567890"
}
```

**Outbound (Sypher-mini → bridge):**

```json
{
  "type": "outbound",
  "to": "+1234567890",
  "content": "Reply text"
}
```

### Sypher-mini config

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

### Run

1. Start your bridge on port 3001 (or your chosen port)
2. Start Sypher-mini gateway: `sypher gateway`

---

## Option 2: Baileys Extension

The Baileys extension uses [@whiskeysockets/baileys](https://github.com/WhiskeySockets/Baileys) for direct WhatsApp connection (no browser).

### Prerequisites

- Node.js 20+
- npm

### Setup

#### 1. Install extension dependencies

```bash
cd extensions/whatsapp-baileys
npm install
npm run build
```

If you skip this step, the gateway will run `npm install` and `npm run build` automatically when spawning the extension (first run may take longer).

#### 2. Configure environment

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `3002` | Extension HTTP port |
| `SYPHER_CORE_CALLBACK` | `http://localhost:18790/inbound` | Gateway inbound URL |
| `SYPHER_WHATSAPP_AUTH` | `~/.sypher-mini/whatsapp-auth` | Auth storage |

#### 3. Configure and start

**Option A: Integrated (gateway auto-spawns extension)**

```json
{
  "channels": {
    "whatsapp": {
      "enabled": true,
      "use_baileys": true,
      "baileys_url": "http://localhost:3002",
      "allow_from": ["+1234567890"]
    }
  }
}
```

```bash
sypher gateway
```

The gateway will spawn the Baileys extension automatically (from `extensions/whatsapp-baileys`). Run from the project root so the extension can be found.

**Option B: Manual (run extension separately)**

```bash
# Terminal 1: start gateway
sypher gateway

# Terminal 2: start Baileys extension
cd extensions/whatsapp-baileys
npm run build && npm start
```

#### 4. Pair with WhatsApp

- A QR code appears in the terminal
- Scan with WhatsApp (Settings → Linked Devices → Link a Device)
- Auth is saved to `~/.sypher-mini/whatsapp-auth/`

#### 5. Config reference

| Key | Default | Description |
|-----|---------|-------------|
| `use_baileys` | `true` | Use Baileys (QR) extension; set `false` and `bridge_url` for WebSocket bridge |
| `baileys_url` | `http://localhost:3002` | Extension HTTP endpoint |

### Extension API

The extension exposes:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/send` | POST | Send message; body: `{ "to": "+123...", "content": "..." }` |

Inbound messages are sent to the gateway `/inbound` endpoint.

---

## Configuration

### allow_from

Restrict which numbers can interact:

```json
"allow_from": ["+1234567890", "+0987654321"]
```

Empty array = allow all (use with caution).

### operators and admins

For future WhatsApp command tiers:

- **user** — Chat, ask, status (any `allow_from`)
- **operator** — config get, agents list, monitors status
- **admin** — config set, agents add, audit show

```json
"operators": ["+1234567890"],
"admins": ["+1234567890"]
```

---

## Menu and Routing

| You send | What happens |
|----------|--------------|
| `menu` or `/help` | Shows the main menu (Projects, Tasks, Logs, CLI, Server, Help) |
| `1`–`6`, `0`, `back` | Navigate menus (only when already in a menu session) |
| `sypher` + request | Routes to the agent with full tools (e.g. "sypher create a hello world script") |
| `/config`, `/cli`, etc. | Slash commands (see below) |
| `42`, `sudo`, `joke`, `coffee`, `roll dice`, `hello world` | Easter eggs — try them. |
| `7` (from menu) | Roll the dice — 1d6, 2d6, 1d20 |
| Anything else | Agent with full tools |

All features are available via the menu or by talking to the agent. Type `menu` or `/help` to see options.

## CLI Session Commands

Manage persistent CLI terminals from WhatsApp:

| Command | Description |
|---------|-------------|
| `/cli list` or `cli list` | List active CLI sessions (ID, tag, last activity) |
| `/cli new -m 'tag'` | Create new terminal with tag |
| `/cli <N>` | Show last 10 lines of terminal N |
| `/cli <N> --tail 50` | Show last 50 lines (max 100) |
| `/cli run <N> <command>` | Run command in terminal N |

**Examples:** `cli new -m 'starting dev'`, `cli 1 --tail 100`, `cli run 1 npm run dev`

---

## Security

1. **allow_from** — Always restrict in production
2. **Bridge URL** — Use `ws://` only for localhost; use `wss://` for remote
3. **Auth storage** — `~/.sypher-mini/whatsapp-auth/` contains session data; keep private

---

## Troubleshooting

### "Connection refused" to bridge

- Ensure bridge is running before gateway
- Check `bridge_url` port matches bridge

### Baileys: ERR_REQUIRE_ESM or "require() of ES Module not supported"

- Extension uses ESM; ensure `package.json` has `"type": "module"` and `tsconfig` has `"module": "ES2020"`
- Rebuild: `cd extensions/whatsapp-baileys && npm run build`
- Node 18+ required: `node -v`


### Baileys: QR code doesn't appear

- Check Node.js version: `node -v` (need 20+)
- Ensure gateway is running (extension needs `/inbound` to exist)
- Check `SYPHER_CORE_CALLBACK` is correct

### Baileys: "Reconnecting" loop

- Delete `~/.sypher-mini/whatsapp-auth/` and re-pair
- Check WhatsApp account is not logged in elsewhere

### Baileys: "Bad MAC" / "Failed to decrypt message with any known session"

- Session keys are out of sync with the sender's device.
- **Fix:** Delete auth and re-pair: `rm -rf ~/.sypher-mini/whatsapp-auth/` (or `%USERPROFILE%\.sypher-mini\whatsapp-auth` on Windows), then restart and scan a new QR code.
- Ensure only one login per account (unlink other WhatsApp Web/Baileys instances).
- If errors persist, check that you're not in groups/broadcasts with many senders; some sessions may be stale and will recover automatically ("Closing open session in favor of incoming prekey bundle").

### Messages not received

- Verify `allow_from` includes your number (with country code, e.g. `+1234567890`)
- Check gateway logs for errors
