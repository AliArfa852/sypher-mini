package extensions

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// CheckNodeVersion returns true if the current Node.js version satisfies minMajor (e.g. "20").
func CheckNodeVersion(minMajor string) bool {
	out, err := exec.Command("node", "-v").Output()
	if err != nil {
		return false
	}
	v := strings.TrimSpace(string(out))
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	if len(parts) == 0 {
		return false
	}
	var major, min int
	fmt.Sscanf(parts[0], "%d", &major)
	fmt.Sscanf(minMajor, "%d", &min)
	return major >= min
}

// RunSetup runs the extension setup script. Returns true on success.
func RunSetup(extDir string, m Manifest) bool {
	if m.Setup == "" {
		return false
	}
	setupPath := ResolveScriptPath(extDir, m.Setup)
	if setupPath == "" {
		return false
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", setupPath)
	} else {
		cmd = exec.Command("sh", setupPath)
	}
	cmd.Dir = extDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run() == nil
}

// RunStart returns an exec.Cmd to run the extension start script, or nil if not found.
func RunStart(extDir string, m Manifest) *exec.Cmd {
	if m.Start == "" {
		return nil
	}
	startPath := ResolveScriptPath(extDir, m.Start)
	if startPath == "" {
		return nil
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", startPath)
	} else {
		cmd = exec.Command("sh", startPath)
	}
	return cmd
}

// ResolveScriptPath finds the script file, preferring .cmd/.ps1 on Windows.
func ResolveScriptPath(extDir, rel string) string {
	base := rel
	if ext := filepath.Ext(rel); ext != "" {
		base = strings.TrimSuffix(rel, ext)
	}
	candidates := []string{filepath.Join(extDir, rel)}
	if runtime.GOOS == "windows" {
		candidates = []string{
			filepath.Join(extDir, base+".cmd"),
			filepath.Join(extDir, base+".ps1"),
			filepath.Join(extDir, rel),
		}
	}
	for _, p := range candidates {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}
