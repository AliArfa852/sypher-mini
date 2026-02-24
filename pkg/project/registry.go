package project

import (
	"os"
	"path/filepath"
)

// Known project file markers for auto-discovery.
var projectMarkers = []string{
	"package.json",
	"go.mod",
	"Cargo.toml",
	"pyproject.toml",
	"requirements.txt",
	"docker-compose.yml",
	"docker-compose.yaml",
	".git",
}

// ScanPath scans a directory for project markers (same logic as ScanWorkspace).
func ScanPath(dir string) []string {
	d := filepath.Clean(dir)
	return scanDir(d)
}

// ScanWorkspace finds directories that look like projects (contain package.json, go.mod, etc.).
func ScanWorkspace(workspace string) []string {
	ws := filepath.Clean(workspace)
	return scanDir(ws)
}

func scanDir(ws string) []string {
	var found []string
	_ = filepath.Walk(ws, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}
		// Skip hidden dirs (except .git) and common non-project dirs
		base := filepath.Base(path)
		if len(base) > 0 && base[0] == '.' && base != ".git" {
			return filepath.SkipDir
		}
		if base == ".git" {
			// Include parent of .git as project
			parent := filepath.Dir(path)
			if !contains(found, parent) {
				found = append(found, parent)
			}
			return filepath.SkipDir
		}
		if base == "node_modules" || base == "vendor" || base == "__pycache__" {
			return filepath.SkipDir
		}
		for _, marker := range projectMarkers {
			if _, err := os.Stat(filepath.Join(path, marker)); err == nil {
				if !contains(found, path) {
					found = append(found, path)
				}
				break
			}
		}
		return nil
	})
	return found
}

// SuggestProject creates a Project from a directory path (for auto-registration).
func SuggestProject(dir, workspace string) *Project {
	abs, _ := filepath.Abs(dir)
	ws, _ := filepath.Abs(workspace)
	rel, _ := filepath.Rel(ws, abs)
	if rel == ".." || len(rel) >= 2 && rel[:2] == ".." {
		rel = abs
	}
	name := filepath.Base(dir)
	return &Project{
		ID:   sanitizeID(name),
		Name: name,
		Path: rel,
	}
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
