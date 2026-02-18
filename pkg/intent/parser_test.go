package intent

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input    string
		want     Intent
		needsLLM bool
	}{
		{"config get foo", IntentConfigChange, false},
		{"config set x y", IntentConfigChange, false},
		{"run ls -la", IntentCommand, false},
		{"/run echo hi", IntentCommand, false},
		{"help", IntentChat, true},
		{"what is 2+2?", IntentChat, true},
		{"hello", IntentChat, true},
		{"schedule backup", IntentAutomationRequest, true},
		{"urgent: server down", IntentEmergencyAlert, false},
	}
	parser := New()
	for _, tt := range tests {
		got := parser.Parse(tt.input)
		if got.Intent != tt.want {
			t.Errorf("Parse(%q).Intent = %v, want %v", tt.input, got.Intent, tt.want)
		}
		if got.NeedsLLM() != tt.needsLLM {
			t.Errorf("Parse(%q).NeedsLLM() = %v, want %v", tt.input, got.NeedsLLM(), tt.needsLLM)
		}
	}
}
