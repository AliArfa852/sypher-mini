package bus

import (
	"context"
	"sync"
)

// MessageBus handles inbound/outbound message flow for the agent loop.
type MessageBus struct {
	inbound  chan InboundMessage
	outbound chan OutboundMessage
	closed   bool
	mu       sync.RWMutex
}

// NewMessageBus creates a new message bus.
func NewMessageBus(bufferSize int) *MessageBus {
	if bufferSize <= 0 {
		bufferSize = 100
	}
	return &MessageBus{
		inbound:  make(chan InboundMessage, bufferSize),
		outbound: make(chan OutboundMessage, bufferSize),
	}
}

// PublishInbound publishes an inbound message.
func (mb *MessageBus) PublishInbound(msg InboundMessage) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	if mb.closed {
		return
	}
	select {
	case mb.inbound <- msg:
	default:
		// Buffer full, drop
	}
}

// ConsumeInbound consumes the next inbound message.
func (mb *MessageBus) ConsumeInbound(ctx context.Context) (InboundMessage, bool) {
	select {
	case msg := <-mb.inbound:
		return msg, true
	case <-ctx.Done():
		return InboundMessage{}, false
	}
}

// PublishOutbound publishes an outbound message.
func (mb *MessageBus) PublishOutbound(msg OutboundMessage) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	if mb.closed {
		return
	}
	select {
	case mb.outbound <- msg:
	default:
		// Buffer full, drop
	}
}

// SubscribeOutbound consumes the next outbound message.
func (mb *MessageBus) SubscribeOutbound(ctx context.Context) (OutboundMessage, bool) {
	select {
	case msg := <-mb.outbound:
		return msg, true
	case <-ctx.Done():
		return OutboundMessage{}, false
	}
}

// Close closes the message bus.
func (mb *MessageBus) Close() {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	if mb.closed {
		return
	}
	mb.closed = true
	close(mb.inbound)
	close(mb.outbound)
}
