package constants

import (
	"testing"
)

func TestIsInternalChannel(t *testing.T) {
	tests := []struct {
		channel string
		want    bool
	}{
		{"cli", true},
		{"system", true},
		{"whatsapp", false},
		{"", false},
		{"CLI", false},
	}
	for _, tt := range tests {
		got := IsInternalChannel(tt.channel)
		if got != tt.want {
			t.Errorf("IsInternalChannel(%q) = %v, want %v", tt.channel, got, tt.want)
		}
	}
}
