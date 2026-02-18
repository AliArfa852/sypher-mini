package bus

import (
	"context"
	"testing"
	"time"
)

func TestMessageBus_PublishConsume(t *testing.T) {
	mb := NewMessageBus(10)
	ctx := context.Background()

	mb.PublishInbound(InboundMessage{Channel: "cli", ChatID: "1", Content: "hi", SenderID: "u1"})
	msg, ok := mb.ConsumeInbound(ctx)
	if !ok {
		t.Fatal("expected message")
	}
	if msg.Content != "hi" {
		t.Errorf("got %q", msg.Content)
	}
}

func TestMessageBus_SubscribeOutbound(t *testing.T) {
	mb := NewMessageBus(10)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	mb.PublishOutbound(OutboundMessage{Channel: "cli", ChatID: "1", Content: "reply"})
	msg, ok := mb.SubscribeOutbound(ctx)
	if !ok {
		t.Fatal("expected message")
	}
	if msg.Content != "reply" {
		t.Errorf("got %q", msg.Content)
	}
}
