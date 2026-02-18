package channels

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sypherexx/sypher-mini/pkg/bus"
)

// WhatsAppBridge connects to a WhatsApp bridge WebSocket and relays messages.
type WhatsAppBridge struct {
	bridgeURL string
	msgBus    *bus.MessageBus
	eventBus  *bus.Bus
	conn      *websocket.Conn
	mu        sync.Mutex
	running   bool
}

// BridgeMessage is the JSON format for bridge messages.
type BridgeMessage struct {
	Type    string `json:"type"`
	From    string `json:"from,omitempty"`
	To      string `json:"to,omitempty"`
	Content string `json:"content,omitempty"`
	ChatID  string `json:"chat_id,omitempty"`
}

// NewWhatsAppBridge creates a WhatsApp bridge channel.
func NewWhatsAppBridge(bridgeURL string, msgBus *bus.MessageBus, eventBus *bus.Bus) *WhatsAppBridge {
	return &WhatsAppBridge{
		bridgeURL: bridgeURL,
		msgBus:    msgBus,
		eventBus:  eventBus,
	}
}

// Run connects to the bridge and relays messages until ctx is cancelled.
func (w *WhatsAppBridge) Run(ctx context.Context) error {
	u, err := url.Parse(w.bridgeURL)
	if err != nil {
		return fmt.Errorf("invalid bridge URL: %w", err)
	}
	if u.Scheme == "ws" || u.Scheme == "wss" {
		// already websocket
	} else if u.Scheme == "http" {
		u.Scheme = "ws"
	} else if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}

	w.mu.Lock()
	w.running = true
	w.mu.Unlock()

	backoff := time.Second
	maxBackoff := 60 * time.Second
	for ctx.Err() == nil {
		conn, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
		if err != nil {
			log.Printf("WhatsApp bridge connect failed: %v (retry in %v)", err, backoff)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
				if backoff < maxBackoff {
					backoff *= 2
					if backoff > maxBackoff {
						backoff = maxBackoff
					}
				}
			}
			continue
		}
		backoff = time.Second // reset on successful connect

		w.mu.Lock()
		w.conn = conn
		w.mu.Unlock()

		// Forward outbound to bridge (runs until ctx done)
		go w.forwardOutbound(ctx, conn)

		// Read inbound from bridge
		for ctx.Err() == nil {
			_, data, err := conn.ReadMessage()
			if err != nil {
				log.Printf("WhatsApp bridge read error: %v", err)
				break
			}
			var msg BridgeMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				continue
			}
			if msg.Type == "inbound" || msg.Type == "message" {
				w.msgBus.PublishInbound(bus.InboundMessage{
					Channel:  "whatsapp",
					ChatID:   msg.ChatID,
					Content:  msg.Content,
					SenderID: msg.From,
				})
			}
		}

		conn.Close()
		w.mu.Lock()
		w.conn = nil
		w.mu.Unlock()
	}

	w.mu.Lock()
	w.running = false
	w.mu.Unlock()
	return ctx.Err()
}

// forwardOutbound sends outbound messages to the bridge.
func (w *WhatsAppBridge) forwardOutbound(ctx context.Context, conn *websocket.Conn) {
	for ctx.Err() == nil {
		out, ok := w.msgBus.SubscribeOutbound(ctx)
		if !ok {
			return
		}
		if out.Channel != "whatsapp" {
			continue
		}
		payload := BridgeMessage{
			Type:    "outbound",
			To:      out.ChatID,
			Content: out.Content,
		}
		data, _ := json.Marshal(payload)
		w.mu.Lock()
		err := conn.WriteMessage(websocket.TextMessage, data)
		w.mu.Unlock()
		if err != nil {
			log.Printf("WhatsApp bridge write error: %v", err)
			return
		}
	}
}
