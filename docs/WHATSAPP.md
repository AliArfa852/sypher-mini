# WhatsApp Setup Guide

Connect Sypher-mini to WhatsApp via the WebSocket bridge or the Baileys extension.

---

## Table of Contents

1. [Overview](#overview)
2. [Option 1: WebSocket Bridge](#option-1-websocket-bridge)
3. [Option 2: Baileys Extension](#option-2-baileys-extension)
4. [Configuration](#configuration)
5. [Security](#security)
6. [Troubleshooting](#troubleshooting)

---

## Overview

| Method | Pros | Cons |
|--------|------|------|
| **Bridge** | Use existing bridge (e.g. whatsapp-web.js) | Requires separate bridge process |
| **Baileys** | Self-contained, no browser | Node.js required, QR pairing |

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

- Node.js 18+
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
| `use_baileys` | `false` | Use Baileys extension instead of WebSocket bridge |
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

## Security

1. **allow_from** — Always restrict in production
2. **Bridge URL** — Use `ws://` only for localhost; use `wss://` for remote
3. **Auth storage** — `~/.sypher-mini/whatsapp-auth/` contains session data; keep private

---

## Troubleshooting

### "Connection refused" to bridge

- Ensure bridge is running before gateway
- Check `bridge_url` port matches bridge

### Baileys: QR code doesn't appear

- Check Node.js version: `node -v` (need 18+)
- Ensure gateway is running (extension needs `/inbound` to exist)
- Check `SYPHER_CORE_CALLBACK` is correct

### Baileys: "Reconnecting" loop

- Delete `~/.sypher-mini/whatsapp-auth/` and re-pair
- Check WhatsApp account is not logged in elsewhere

### Messages not received

- Verify `allow_from` includes your number (with country code, e.g. `+1234567890`)
- Check gateway logs for errors
