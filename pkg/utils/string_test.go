package utils

import (
	"testing"
)

func TestNormalizeWhatsAppID(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"1234567890@s.whatsapp.net", "1234567890"},
		{"+1234567890", "1234567890"},
		{"1234567890", "1234567890"},
		{"+1 234 567 890", "1234567890"},
		{"60838547296357@lid", "60838547296357"},
		{"202383321759875@lid", "202383321759875"},
		{"", ""},
	}
	for _, tt := range tests {
		got := NormalizeWhatsAppID(tt.in)
		if got != tt.want {
			t.Errorf("NormalizeWhatsAppID(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestToWhatsAppJID(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"+1234567890", "1234567890@s.whatsapp.net"},
		{"1234567890", "1234567890@s.whatsapp.net"},
		{"1234567890@s.whatsapp.net", "1234567890@s.whatsapp.net"},
		{"", ""},
		{"broadcast", ""},
	}
	for _, tt := range tests {
		got := ToWhatsAppJID(tt.in)
		if got != tt.want {
			t.Errorf("ToWhatsAppJID(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

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
