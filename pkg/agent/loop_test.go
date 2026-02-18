package agent

import (
	"context"
	"testing"

	"github.com/sypherexx/sypher-mini/pkg/bus"
	"github.com/sypherexx/sypher-mini/pkg/config"
)

func TestLoop_ProcessMessage_IntentFastPath(t *testing.T) {
	cfg := config.DefaultConfig()
	msgBus := bus.NewMessageBus(10)
	eventBus := bus.New()
	loop := NewLoop(cfg, msgBus, eventBus, &LoopOptions{SafeMode: true})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go loop.Run(ctx)

	msgBus.PublishInbound(bus.InboundMessage{
		Channel:  "cli",
		ChatID:   "cli",
		Content:  "config get x",
		SenderID: "cli",
	})

	out, ok := msgBus.SubscribeOutbound(ctx)
	if !ok {
		t.Fatal("expected outbound message")
	}
	if out.Content == "" {
		t.Error("expected config response")
	}
}

func TestLoop_ProcessMessage_SafeMode(t *testing.T) {
	cfg := config.DefaultConfig()
	msgBus := bus.NewMessageBus(10)
	eventBus := bus.New()
	loop := NewLoop(cfg, msgBus, eventBus, &LoopOptions{SafeMode: true})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go loop.Run(ctx)

	msgBus.PublishInbound(bus.InboundMessage{
		Channel:  "cli",
		ChatID:   "cli",
		Content:  "hello",
		SenderID: "cli",
	})

	out, ok := msgBus.SubscribeOutbound(ctx)
	if !ok {
		t.Fatal("expected outbound message")
	}
	if out.Content == "" {
		t.Error("expected response")
	}
}
