package extensions

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscover(t *testing.T) {
	dir := t.TempDir()
	extDir := filepath.Join(dir, "extensions")
	if err := os.MkdirAll(extDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a valid extension
	ext1 := filepath.Join(extDir, "test-ext")
	if err := os.MkdirAll(ext1, 0755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"id":"test-ext","version":"1.0.0","sypher_mini_version":">=0.1.0","capabilities":["channel"],"entry":"dist/index.js"}`
	if err := os.WriteFile(filepath.Join(ext1, "sypher.extension.json"), []byte(manifest), 0644); err != nil {
		t.Fatal(err)
	}

	exts, err := Discover(extDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(exts) != 1 {
		t.Fatalf("expected 1 extension, got %d", len(exts))
	}
	if exts[0].Manifest.ID != "test-ext" {
		t.Errorf("expected id test-ext, got %s", exts[0].Manifest.ID)
	}
	if exts[0].Manifest.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", exts[0].Manifest.Version)
	}
}

func TestVersionSatisfies(t *testing.T) {
	tests := []struct {
		core   string
		constraint string
		want   bool
	}{
		{"0.1.0", ">=0.1.0", true},
		{"0.2.0", ">=0.1.0", true},
		{"0.0.9", ">=0.1.0", false},
		{"1.0.0", ">=0.1.0", true},
		{"0.1.0", "", true},
	}
	for _, tt := range tests {
		got := VersionSatisfies(tt.core, tt.constraint)
		if got != tt.want {
			t.Errorf("VersionSatisfies(%q, %q) = %v, want %v", tt.core, tt.constraint, got, tt.want)
		}
	}
}
