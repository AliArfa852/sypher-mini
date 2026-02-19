package channels

import (
	"testing"

	"github.com/sypherexx/sypher-mini/pkg/bus"
)

func TestParsePortFromURL(t *testing.T) {
	tests := []struct {
		url    string
		want   string
		hasErr bool
	}{
		{"http://localhost:3002", "3002", false},
		{"https://example.com:443", "443", false},
		{"http://127.0.0.1:18790", "18790", false},
		{"http://localhost", "", true},
		{"localhost:3002", "3002", false},
		{"", "", true},
	}
	for _, tt := range tests {
		got, err := parsePortFromURL(tt.url)
		if (err != nil) != tt.hasErr {
			t.Errorf("parsePortFromURL(%q) err=%v, want hasErr=%v", tt.url, err, tt.hasErr)
		}
		if !tt.hasErr && got != tt.want {
			t.Errorf("parsePortFromURL(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}

func TestNewWhatsAppBaileysClient(t *testing.T) {
	bus := newTestBus()
	client := NewWhatsAppBaileysClient("http://localhost:3002", bus)
	if client == nil {
		t.Fatal("NewWhatsAppBaileysClient returned nil")
	}
	if client.baileysURL != "http://localhost:3002" {
		t.Errorf("baileysURL = %q, want http://localhost:3002", client.baileysURL)
	}

	// Empty URL defaults to localhost:3002
	client2 := NewWhatsAppBaileysClient("", bus)
	if client2.baileysURL != "http://localhost:3002" {
		t.Errorf("empty URL default = %q, want http://localhost:3002", client2.baileysURL)
	}

	// Trailing slash trimmed
	client3 := NewWhatsAppBaileysClient("http://localhost:3002/", bus)
	if client3.baileysURL != "http://localhost:3002" {
		t.Errorf("trailing slash = %q, want http://localhost:3002", client3.baileysURL)
	}
}

func newTestBus() *bus.MessageBus {
	return bus.NewMessageBus(10)
}
