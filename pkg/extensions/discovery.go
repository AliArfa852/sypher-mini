package extensions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Manifest is the sypher.extension.json schema.
type Manifest struct {
	ID                 string   `json:"id"`
	Version            string   `json:"version"`
	SypherMiniVersion  string   `json:"sypher_mini_version"`
	Capabilities       []string `json:"capabilities"`
	Entry              string   `json:"entry"`
	Runtime            string   `json:"runtime"`   // e.g. "node"
	NodeMin            string   `json:"node_min"`  // e.g. "20"
	Setup              string   `json:"setup"`     // e.g. "scripts/setup"
	Start              string   `json:"start"`     // e.g. "scripts/start"
}

// DiscoveredExtension holds a discovered extension with its manifest.
type DiscoveredExtension struct {
	Dir      string
	Manifest Manifest
}

// Discover scans the extensions directory for sypher.extension.json manifests.
// extensionsDir is typically "./extensions" or an absolute path.
func Discover(extensionsDir string) ([]DiscoveredExtension, error) {
	extDir := filepath.Clean(extensionsDir)
	entries, err := os.ReadDir(extDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read extensions dir: %w", err)
	}

	var result []DiscoveredExtension
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), "_") {
			continue
		}
		manifestPath := filepath.Join(extDir, e.Name(), "sypher.extension.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			continue // skip invalid entries
		}

		var m Manifest
		if err := json.Unmarshal(data, &m); err != nil {
			continue
		}
		if m.ID == "" {
			continue
		}

		result = append(result, DiscoveredExtension{
			Dir:      filepath.Join(extDir, e.Name()),
			Manifest: m,
		})
	}
	return result, nil
}

// DiscoverFromWorkspace discovers extensions relative to the workspace root.
// It looks for extensions/ in the same directory as the binary or in cwd.
func DiscoverFromWorkspace(workspaceRoot string) ([]DiscoveredExtension, error) {
	candidates := []string{
		filepath.Join(workspaceRoot, "extensions"),
		"extensions",
		"./extensions",
	}
	for _, d := range candidates {
		abs, _ := filepath.Abs(d)
		exts, err := Discover(abs)
		if err != nil {
			continue
		}
		if len(exts) > 0 {
			return exts, nil
		}
	}
	return nil, nil
}

// VersionSatisfies checks if coreVersion satisfies the constraint (e.g. ">=0.1.0").
// Simplified: only supports ">=X.Y.Z" for now.
func VersionSatisfies(coreVersion, constraint string) bool {
	if constraint == "" {
		return true
	}
	constraint = strings.TrimSpace(constraint)
	if strings.HasPrefix(constraint, ">=") {
		minVer := strings.TrimSpace(constraint[2:])
		return compareVersions(coreVersion, minVer) >= 0
	}
	return coreVersion == constraint
}

func compareVersions(a, b string) int {
	parse := func(s string) (major, minor, patch int) {
		parts := strings.Split(strings.TrimPrefix(s, "v"), ".")
		if len(parts) > 0 {
			fmt.Sscanf(parts[0], "%d", &major)
		}
		if len(parts) > 1 {
			fmt.Sscanf(parts[1], "%d", &minor)
		}
		if len(parts) > 2 {
			fmt.Sscanf(parts[2], "%d", &patch)
		}
		return
	}
	ma, mia, pa := parse(a)
	mb, mib, pb := parse(b)
	if ma != mb {
		return ma - mb
	}
	if mia != mib {
		return mia - mib
	}
	return pa - pb
}
