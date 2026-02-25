package channels

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sypherexx/sypher-mini/pkg/bus"
)

func TestNewTelegramClient(t *testing.T) {
	msgBus := bus.NewMessageBus(10)
	client := NewTelegramClient("http://localhost:3003", msgBus, 0)
	if client == nil {
		t.Fatal("NewTelegramClient returned nil")
	}
	if client.botURL != "http://localhost:3003" {
		t.Errorf("botURL = %q, want http://localhost:3003", client.botURL)
	}
	if client.minInterval != defaultTelegramMinInterval {
		t.Errorf("minInterval = %v, want %v", client.minInterval, defaultTelegramMinInterval)
	}

	// Empty URL defaults to localhost:3003
	client2 := NewTelegramClient("", msgBus, 0)
	if client2.botURL != "http://localhost:3003" {
		t.Errorf("empty URL default = %q, want http://localhost:3003", client2.botURL)
	}

	// Trailing slash trimmed
	client3 := NewTelegramClient("http://localhost:3003/", msgBus, 0)
	if client3.botURL != "http://localhost:3003" {
		t.Errorf("trailing slash = %q, want http://localhost:3003", client3.botURL)
	}

	// Custom min interval
	client4 := NewTelegramClient("http://localhost:3003", msgBus, 5)
	if client4.minInterval != 5*time.Second {
		t.Errorf("minInterval = %v, want 5s", client4.minInterval)
	}
}

func TestTelegramClient_Send(t *testing.T) {
	received := make(chan []byte, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/send" || r.Method != http.MethodPost {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		body := make([]byte, 1024)
		n, _ := r.Body.Read(body)
		received <- body[:n]
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	msgBus := bus.NewMessageBus(10)
	client := NewTelegramClient(server.URL, msgBus, 0)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go client.Run(ctx)

	msgBus.PublishOutbound(bus.OutboundMessage{
		Channel: "telegram",
		ChatID:  "12345",
		Content: "test message",
	})

	select {
	case body := <-received:
		if len(body) == 0 {
			t.Error("expected request body")
		}
		bodyStr := string(body)
		if !strings.Contains(bodyStr, "12345") || !strings.Contains(bodyStr, "test message") {
			t.Errorf("unexpected body: %s", bodyStr)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for send")
	}
}

