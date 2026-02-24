package utils

import (
	"strings"
	"unicode"
)

// NormalizeWhatsAppID extracts digits from a WhatsApp ID for comparison.
// Handles: "1234567890@s.whatsapp.net", "+1234567890", "1234567890"
func NormalizeWhatsAppID(s string) string {
	var out []rune
	for _, r := range s {
		if unicode.IsDigit(r) {
			out = append(out, r)
		}
	}
	return string(out)
}

// ToWhatsAppJID converts a phone/ID to Baileys JID format (number@s.whatsapp.net).
// Handles: "+1234567890", "1234567890", "1234567890@s.whatsapp.net"
func ToWhatsAppJID(id string) string {
	if id == "" {
		return ""
	}
	if strings.Contains(id, "@s.whatsapp.net") {
		return id
	}
	digits := NormalizeWhatsAppID(id)
	if digits == "" {
		return ""
	}
	return digits + "@s.whatsapp.net"
}

// Truncate returns a truncated version of s with at most maxLen runes.
// Handles multi-byte Unicode characters properly.
// If the string is truncated, "..." is appended to indicate truncation.
func Truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	// Reserve 3 chars for "..."
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}
