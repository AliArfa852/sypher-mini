# Platform and Command Compatibility

Sypher-mini runs on Windows, Linux, and macOS. The **exec** tool uses different shells per platform. Agents receive runtime context automatically; this doc is the canonical reference for command compatibility.

## Exec Tool Shell

| Platform | Shell | Invocation |
|----------|-------|------------|
| Windows | cmd.exe | `cmd /c "<command>"` |
| Linux | sh | `sh -c "<command>"` |
| macOS | sh | `sh -c "<command>"` |

## Command Compatibility Matrix

### Directory operations

| Action | Windows (cmd) | Linux/macOS (sh) |
|--------|---------------|------------------|
| Create dir (single) | `mkdir E:\demo\foo` | `mkdir -p /path/to/foo` |
| Create dir (parents) | `mkdir E:\demo\foo` (creates parents on Win 10+) | `mkdir -p /path/to/foo` |
| List directory | `dir` | `ls` or `ls -la` |
| Change directory | `cd E:\demo` | `cd /path/to/dir` |
| Current directory | `cd` | `pwd` |

### File operations

| Action | Windows (cmd) | Linux/macOS (sh) |
|--------|---------------|------------------|
| Remove file | `del file` | `rm file` |
| Remove dir (empty) | `rmdir dir` | `rmdir dir` |
| Remove dir (recursive) | `rmdir /s /q dir` | `rm -rf dir` (blocked by deny patterns) |
| Copy file | `copy src dst` | `cp src dst` |
| Move/rename | `move src dst` | `mv src dst` |

### Command chaining

| Syntax | Windows (cmd) | Linux/macOS (sh) |
|--------|---------------|------------------|
| And-then | `cmd1 && cmd2` | `cmd1 && cmd2` |
| Or | `cmd1 \|\| cmd2` | `cmd1 \|\| cmd2` |
| Sequential | `cmd1 & cmd2` | `cmd1 ; cmd2` |

### Paths

| Platform | Separator | Example |
|----------|------------|---------|
| Windows | `\` | `E:\demo\test-sypher` |
| Linux/macOS | `/` | `/home/user/demo/test-sypher` |

**Windows in JSON config:** Use double backslashes: `"E:\\demo"`.

### Git discovery

| Platform | Command |
|----------|---------|
| Windows | `dir /s /b .git` |
| Linux/macOS | `find . -name .git -type d` |
| Any | `git rev-parse --show-toplevel` (when inside repo) |

### Incompatible or blocked

- **Unix-only on Windows:** `mkdir -p`, `ls`, `find`, `rm -rf`, `pwd` (use Windows equivalents)
- **Windows-only on Unix:** `dir`, `del`, `rmdir /s`, `copy`, `move` (use Unix equivalents)
- **Blocked by safety:** `rm -rf`, `sudo`, `curl | sh`, `git push` (unless allow_git_push), etc. See [SECURITY.md](SECURITY.md)

## invoke_cli_agent vs exec

| Tool | What it does |
|------|---------------|
| **exec** | Runs commands on the host. Use for file ops (mkdir, git init, npm install), listing dirs, etc. |
| **invoke_cli_agent** | Runs a CLI (e.g. Gemini) with a task. The CLI returns text; it does **not** execute commands on the host. |

For "create a repo and run git init", use **exec** with platform-appropriate commands. Do not use invoke_cli_agent for file operations.

## Config Override (future)

A `platform` or `runtime` config section may be added to let users override detected OS (e.g. WSL, remote). For now, runtime is auto-detected from `runtime.GOOS`.
