package extensions

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveScriptPath(t *testing.T) {
	dir := t.TempDir()

	// Create scripts/setup (no extension)
	setupDir := filepath.Join(dir, "scripts")
	if err := os.MkdirAll(setupDir, 0755); err != nil {
		t.Fatal(err)
	}
	setupPath := filepath.Join(setupDir, "setup")
	if err := os.WriteFile(setupPath, []byte("#!/bin/sh\nexit 0"), 0644); err != nil {
		t.Fatal(err)
	}

	// Resolve scripts/setup
	got := ResolveScriptPath(dir, "scripts/setup")
	if got != setupPath {
		t.Errorf("ResolveScriptPath(scripts/setup) = %q, want %q", got, setupPath)
	}

	// Non-existent returns empty
	got2 := ResolveScriptPath(dir, "scripts/nonexistent")
	if got2 != "" {
		t.Errorf("ResolveScriptPath(nonexistent) = %q, want empty", got2)
	}
}

func TestResolveScriptPathWindowsPreference(t *testing.T) {
	dir := t.TempDir()
	setupDir := filepath.Join(dir, "scripts")
	if err := os.MkdirAll(setupDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create both setup and setup.cmd
	setupPath := filepath.Join(setupDir, "setup")
	setupCmdPath := filepath.Join(setupDir, "setup.cmd")
	if err := os.WriteFile(setupPath, []byte("#!/bin/sh"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(setupCmdPath, []byte("@echo off"), 0644); err != nil {
		t.Fatal(err)
	}

	got := ResolveScriptPath(dir, "scripts/setup")
	// On Windows, .cmd is preferred; on Unix, scripts/setup is used
	if got == "" {
		t.Error("ResolveScriptPath returned empty")
	}
	// Should return one of the existing files
	if got != setupPath && got != setupCmdPath {
		t.Errorf("ResolveScriptPath = %q, expected scripts/setup or scripts/setup.cmd", got)
	}
}
