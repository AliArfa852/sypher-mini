package platform

import "runtime"

// Info holds runtime platform context for the agent.
type Info struct {
	OS      string // "windows", "linux", "darwin"
	Shell   string // "cmd" or "sh"
	PathSep string // "\" or "/"
}

// Current returns the current platform info.
func Current() Info {
	os := runtime.GOOS
	shell := "sh"
	pathSep := "/"
	if os == "windows" {
		shell = "cmd"
		pathSep = "\\"
	}
	return Info{OS: os, Shell: shell, PathSep: pathSep}
}

// AgentContext returns a short string to inject into the system prompt
// so the agent uses platform-appropriate commands.
func AgentContext() string {
	p := Current()
	switch p.OS {
	case "windows":
		return `## Runtime (exec tool)
- OS: Windows
- Shell: cmd.exe (/c)
- Path separator: backslash (\)
- Create dir: mkdir E:\path\to\dir (parent must exist; use multiple mkdir if needed)
- Chain commands: use && (cmd supports it)
- List dir: dir
- Find files: dir /s /b .git
- Use exec tool for file ops (mkdir, git init, etc.); invoke_cli_agent does NOT run commands on this machine`
	case "darwin":
		return `## Runtime (exec tool)
- OS: macOS
- Shell: sh
- Path separator: /
- Create dir: mkdir -p /path/to/dir
- Chain commands: && or ;
- List dir: ls
- Find files: find . -name .git -type d`
	default:
		return `## Runtime (exec tool)
- OS: Linux
- Shell: sh
- Path separator: /
- Create dir: mkdir -p /path/to/dir
- Chain commands: && or ;
- List dir: ls
- Find files: find . -name .git -type d`
	}
}
