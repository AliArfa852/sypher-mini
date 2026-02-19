# Adding and Removing Extensions

How to add, modify, and remove Sypher-mini extensions.

---

## Extension structure

Each extension lives in `extensions/<name>/` with:

| File | Purpose |
|------|---------|
| `sypher.extension.json` | Manifest (id, entry, runtime, setup, start) |
| `package.json` | Node deps (for Node extensions) |
| `scripts/setup` | Unix setup script |
| `scripts/setup.cmd` | Windows setup script |
| `scripts/start` | Unix start script |
| `scripts/start.cmd` | Windows start script |

---

## Adding an extension

1. Copy `extensions/_template/` to `extensions/<your-extension>/`.

2. Edit `sypher.extension.json`:
   - `id`: unique extension ID
   - `entry`: path to main entry (e.g. `dist/index.js`)
   - `runtime`: `"node"` (default)
   - `node_min`: minimum Node version (e.g. `"20"`)
   - `setup`: path to setup script (e.g. `scripts/setup`)
   - `start`: path to start script (e.g. `scripts/start`)

3. Add `package.json`, source files, and build config (e.g. `tsconfig.json` for TypeScript).

4. Update `scripts/setup` and `scripts/setup.cmd` to install deps and build.

5. Update `scripts/start` and `scripts/start.cmd` to run your entry.

6. Run `sypher extensions` to verify discovery.

---

## Modifying an extension

- Edit manifest, source, or scripts as needed.
- Rebuild: `cd extensions/<name> && npm run build` (or run setup).
- No config changes needed; extensions are discovered by directory.

---

## Removing an extension

Delete the extension directory:

```bash
rm -rf extensions/<name>
```

No config changes needed. The gateway will no longer discover or spawn it.

---

## Manifest schema

```json
{
  "id": "my-extension",
  "version": "1.0.0",
  "sypher_mini_version": ">=0.1.0",
  "capabilities": ["channel"],
  "entry": "dist/index.js",
  "runtime": "node",
  "node_min": "20",
  "setup": "scripts/setup",
  "start": "scripts/start"
}
```

- `id`: Required. Unique identifier.
- `entry`: Required. Path to executable entry (relative to extension dir).
- `runtime`: Optional. `"node"` (default).
- `node_min`: Optional. Minimum Node.js major version for Node extensions.
- `setup`: Optional. Path to setup script (runs before first spawn).
- `start`: Optional. Path to start script. If omitted, gateway runs `node <entry>`.
