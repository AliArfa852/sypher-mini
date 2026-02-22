# Sypher-mini Issues Audit

Code review and workflow trace (2025-02-22). Issues found in CLI registration, project management, WhatsApp flows, and pipelines.

---

## Critical Issues

### 1. **WhatsApp `allow_from` / Sender ID Mismatch**

**Problem:** Baileys sends `from` as `1234567890@s.whatsapp.net`, but config uses `+1234567890`. Exact string match fails, so messages are silently dropped when `allow_from` is set.

**Location:** `cmd/sypher/main.go` (isAllowedSender), `pkg/agent/loop.go` (allow_from check), `pkg/intent/whatsapp_commands.go` (resolveTier)

**Fix:** Normalize both values before comparison (strip `@s.whatsapp.net`, normalize `+` prefix).

---

### 2. **CLI Sessions Are In-Memory Only**

**Problem:** `clisession.Manager` stores sessions in memory. Gateway restart loses all sessions. User creates `/cli new -m "tag"`, then after restart `/cli list` shows nothing.

**Location:** `pkg/clisession/manager.go`, `pkg/agent/loop.go` (cliManager)

**Fix:** Persist sessions to `~/.sypher-mini/cli-sessions/` or similar, load on startup.

---

### 3. **No Way to Add Projects via WhatsApp**

**Problem:** Projects are loaded from `workspace/code-projects/*.json` only. There is no menu action or slash command to add/register a project from WhatsApp.

**Location:** `pkg/project/store.go`, `pkg/menu/actions.go`, `config/menus.json`

**Fix:** Add `projects_add` action and/or `/project add <name> <path>` command. Or document that users must create JSON files manually.

---

### 4. **Menu Session Requires "menu" First**

**Problem:** If user sends `4` (expecting CLI submenu) without first sending `menu`, there is no session. The handler falls through and `4` goes to the agent instead of the CLI menu.

**Location:** `pkg/menu/handler.go` – session exists only after `menu` or `\help`

**Fix:** Either auto-create session on first numeric input, or make the main menu show on any of `menu`, `help`, `1`–`7` when no session exists.

---

### 5. **Outbound ChatID Format Mismatch**

**Problem:** Outbound uses `msg.ChatID` (from inbound). If inbound `from` is `123@s.whatsapp.net`, ChatID is the same. Baileys `sendMessage(to, ...)` expects JID format. Need to verify extension handles both.

**Location:** `pkg/channels/whatsapp_baileys.go` (send), `extensions/whatsapp-baileys/src/index.ts`

**Status:** Extension uses `to` directly; if ChatID is JID format it should work. But if config `allow_from` has `+123` and we use that for fallback ChatID, format may differ.

---

## Medium Issues

### 6. **Commands Registry vs CLI Sessions Confusion**

**Problem:** Two different concepts:
- **Commands** (`~/.sypher-mini/commands/*.json`): Custom command configs for invoke_cli_agent
- **CLI Sessions** (in-memory): Terminal sessions created via `/cli new`

User may expect "registered CLI" to mean commands, but "CLI list" shows sessions. Naming is confusing.

**Fix:** Clarify in docs and help text.

---

### 7. **Project Build/Pull: Pending State Uses Session TTL**

**Problem:** After selecting "build" or "pull", user must reply with a number. Session TTL is 10 minutes. If user waits too long, pending state expires and numeric reply is ignored.

**Location:** `pkg/menu/session.go` (defaultSessionTTL, GetPendingProjectAction)

---

### 8. **RunProjectsBuild/RunProjectsPull Called with Empty projectID from Menu**

**Problem:** In `actions.go`, `projects_build` calls `runner.RunProjectsBuild(ctx, "", msg)`. The handler then sets pending and shows the list. But the initial call passes `""` – RunProjectsBuild handles that by showing the list. However, the handler overwrites the response with the list from RunProjectsBuild. Flow is correct but the action could be clearer.

---

### 9. **Menus.json Schema: Title vs title**

**Problem:** `MenuDef` uses `json:"title"` (lowercase). `config/menus.json` has `"title"`. Go struct uses `Title`. Unmarshaling works. But if user creates `~/.sypher-mini/menus.json` with `"Title"`, it would not populate.

**Status:** Current config is correct. Document that keys must be lowercase.

---

### 10. **resolveTier Returns Empty for Non-Allowed Senders**

**Problem:** When `allowFrom` has entries and sender is not in the list, `resolveTier` returns `""`. `ParseWhatsAppCommand` still returns `isCmd=true` for the command. But the message is dropped earlier in `processMessage` by the allow_from check. So we never reach ParseWhatsAppCommand for disallowed senders. Flow is OK.

---

## Low / Documentation

### 11. **Agent System Prompt: Command Summary**

**Problem:** `BuildForAgent` includes project IDs and custom commands. But it does NOT include the current CLI session list. The agent cannot see "you have CLI session 1 (tag: x)" in the system prompt.

**Fix:** Optionally append CLI session summary to command summary when sessions exist.

---

### 12. **Stream Command: allowed_commands**

**Problem:** `stream_command` only allows commands in `tools.live_monitoring.allowed_commands`. If not configured, streaming may fail. Document required config.

---

### 13. **Invoke CLI Agent: Requires agents.list with command/args**

**Problem:** `invoke_cli_agent` needs an agent with `command` and `args` in config. If only default agent exists without command, tool returns error. Document clearly.

---

## Summary of Fixes to Apply

| Priority | Issue | Fix |
|----------|-------|-----|
| P0 | allow_from mismatch | Normalize WhatsApp JID for comparison |
| P0 | CLI sessions ephemeral | Persist sessions (or document limitation) |
| P1 | No add project via WhatsApp | Add projects_add action or document |
| P1 | Menu requires "menu" first | Auto-init session on 1–7 when no session |
| P2 | Commands vs CLI confusion | Improve help text and docs |

---

## Workflow Verification

### CLI Flow (Expected)
1. User: `menu` → Main menu
2. User: `4` → CLI submenu
3. User: `2` → "Open new session" → RunCliNew("unnamed") → "Created CLI session 1: unnamed"
4. User: `1` → "List sessions" → RunCliList → Shows "1: unnamed (active just now)"
5. User: `/cli run 1 dir` → Runs dir in session 1, appends output

**Potential failure:** If gateway restarts between 3 and 4, session is gone.

### Projects Flow (Expected)
1. User: `menu` → `1` (Projects) → `3` (Build)
2. If projects exist: "Select project to build" + numbered list, pending set
3. User: `1` → Build runs for projectIDs[0]

**Potential failure:** No projects in `code-projects/` → "No projects. Add projects to workspace/code-projects/". No way to add from WhatsApp.

### allow_from Flow
1. Config has `allow_from: ["+1234567890"]`
2. WhatsApp message from `1234567890@s.whatsapp.net`
3. isAllowedSender("1234567890@s.whatsapp.net", ["+1234567890"]) → false (no match)
4. Message silently dropped

**Fix:** Normalize before compare.
