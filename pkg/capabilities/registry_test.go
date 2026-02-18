package capabilities

import (
	"testing"
)

func TestDefaultRegistry(t *testing.T) {
	r := DefaultRegistry()
	if len(r.Tools) == 0 {
		t.Error("default registry should have tools")
	}
	if len(r.Agents) == 0 {
		t.Error("default registry should have agents")
	}
}

func TestResolveTools(t *testing.T) {
	r := DefaultRegistry()
	tools := r.ResolveTools("code_generation")
	if len(tools) == 0 {
		t.Error("expected tools for code_generation")
	}
	found := false
	for _, t := range tools {
		if t == "exec" {
			found = true
			break
		}
	}
	if !found {
		t.Error("exec should satisfy code_generation")
	}
}

func TestResolveAgents(t *testing.T) {
	r := DefaultRegistry()
	agents := r.ResolveAgents("code_generation")
	if len(agents) == 0 {
		t.Error("expected agents for code_generation")
	}
}
