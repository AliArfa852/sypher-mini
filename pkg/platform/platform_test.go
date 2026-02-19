package platform

import (
	"runtime"
	"strings"
	"testing"
)

func TestCurrent(t *testing.T) {
	p := Current()
	if p.OS == "" {
		t.Error("Current() returned empty OS")
	}
	if p.Shell == "" {
		t.Error("Current() returned empty Shell")
	}
	if p.PathSep == "" {
		t.Error("Current() returned empty PathSep")
	}
	if runtime.GOOS == "windows" {
		if p.Shell != "cmd" {
			t.Errorf("windows: Shell = %q, want cmd", p.Shell)
		}
		if p.PathSep != "\\" {
			t.Errorf("windows: PathSep = %q, want \\", p.PathSep)
		}
	} else {
		if p.Shell != "sh" {
			t.Errorf("unix: Shell = %q, want sh", p.Shell)
		}
		if p.PathSep != "/" {
			t.Errorf("unix: PathSep = %q, want /", p.PathSep)
		}
	}
}

func TestAgentContext(t *testing.T) {
	s := AgentContext()
	if s == "" {
		t.Fatal("AgentContext() returned empty string")
	}
	if !strings.Contains(s, "Runtime") {
		t.Error("AgentContext() should contain 'Runtime'")
	}
	if !strings.Contains(s, "OS:") {
		t.Error("AgentContext() should contain 'OS:'")
	}
}
