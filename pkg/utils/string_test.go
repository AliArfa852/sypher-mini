package utils

import (
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		s      string
		maxLen int
		want   string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello world", 5, "he..."},
		{"hello", 3, "hel"},
		{"hi", 5, "hi"},
		{"日本語", 2, "日本"},
		{"日本語", 3, "日本語"},
		{"a", 1, "a"},
		{"ab", 1, "a"},
	}
	for _, tt := range tests {
		got := Truncate(tt.s, tt.maxLen)
		if got != tt.want {
			t.Errorf("Truncate(%q, %d) = %q, want %q", tt.s, tt.maxLen, got, tt.want)
		}
	}
}
