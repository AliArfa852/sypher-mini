package bus

import (
	"context"
	"sync"
)

// Event represents a structured event on the bus.
type Event struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// EventHandler processes an event.
type EventHandler func(ctx context.Context, ev Event) error

// Bus is the event bus for Sypher-mini.
type Bus struct {
	syncHandlers   map[string][]EventHandler
	asyncHandlers  map[string][]EventHandler
	asyncBuffer    chan Event
	asyncBufferLen int
	onFull         string // "drop" or "block"
	closed         bool
	mu             sync.RWMutex
}

// New creates a new event bus.
func New(opts ...Option) *Bus {
	b := &Bus{
		syncHandlers:   make(map[string][]EventHandler),
		asyncHandlers:  make(map[string][]EventHandler),
		asyncBufferLen: 100,
		onFull:         "drop",
	}
	for _, opt := range opts {
		opt(b)
	}
	b.asyncBuffer = make(chan Event, b.asyncBufferLen)
	return b
}

// Option configures the bus.
type Option func(*Bus)

// WithAsyncBufferSize sets the async buffer size.
func WithAsyncBufferSize(n int) Option {
	return func(b *Bus) {
		b.asyncBufferLen = n
	}
}

// Publish publishes an event. Sync handlers run before return; async handlers are queued.
func (b *Bus) Publish(ctx context.Context, ev Event) error {
	b.mu.RLock()
	syncHandlers := b.syncHandlers[ev.Type]
	asyncHandlers := b.asyncHandlers[ev.Type]
	b.mu.RUnlock()

	for _, h := range syncHandlers {
		if err := h(ctx, ev); err != nil {
			return err
		}
	}

	if len(asyncHandlers) > 0 {
		select {
		case b.asyncBuffer <- ev:
		default:
			if b.onFull == "drop" {
				// Drop oldest would require a different buffer; for now just drop new
				return nil
			}
			// block
			b.asyncBuffer <- ev
		}
	}

	return nil
}

// SubscribeSync registers a sync handler for event type.
func (b *Bus) SubscribeSync(eventType string, h EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.syncHandlers[eventType] = append(b.syncHandlers[eventType], h)
}

// SubscribeAsync registers an async handler for event type.
func (b *Bus) SubscribeAsync(eventType string, h EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.asyncHandlers[eventType] = append(b.asyncHandlers[eventType], h)
}

// RunAsyncDispatcher starts the async handler loop. Call in a goroutine.
func (b *Bus) RunAsyncDispatcher(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-b.asyncBuffer:
			if !ok {
				return
			}
			b.mu.RLock()
			handlers := b.asyncHandlers[ev.Type]
			b.mu.RUnlock()
			for _, h := range handlers {
				_ = h(ctx, ev)
			}
		}
	}
}

// Close closes the bus.
func (b *Bus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	b.closed = true
	close(b.asyncBuffer)
}
