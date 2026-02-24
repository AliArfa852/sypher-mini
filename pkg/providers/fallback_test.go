package providers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sypherexx/sypher-mini/pkg/config"
)

func TestParseRetryAfter_RetryInFormat(t *testing.T) {
	tests := []struct {
		err  error
		want time.Duration
	}{
		{errors.New("retry in 5s"), 5 * time.Second},
		{errors.New("Retry in 10.5s"), 10500 * time.Millisecond},
		{errors.New("retry in 1s"), time.Second},
		{errors.New("error: retry in 30s later"), 30 * time.Second},
		{errors.New("no match"), 0},
		{nil, 0},
	}
	for _, tt := range tests {
		got := parseRetryAfter(tt.err)
		if got != tt.want {
			t.Errorf("parseRetryAfter(%q) = %v, want %v", tt.err, got, tt.want)
		}
	}
}

func TestParseRetryAfter_JSONFormat(t *testing.T) {
	tests := []struct {
		err  error
		want time.Duration
	}{
		{errors.New(`{"retry_after": 60}`), 60 * time.Second},
		{errors.New(`{"retryAfter": 45}`), 45 * time.Second},
		{errors.New(`"retry_after": 20`), 20 * time.Second},
		{errors.New(`retry_after: 10`), 0}, // not valid JSON
	}
	for _, tt := range tests {
		got := parseRetryAfter(tt.err)
		if got != tt.want {
			t.Errorf("parseRetryAfter(%q) = %v, want %v", tt.err, got, tt.want)
		}
	}
}

func TestParseRetryAfter_Clamping(t *testing.T) {
	// Values below 1s should clamp to 1s
	err := errors.New("retry in 0.5s")
	got := parseRetryAfter(err)
	if got < time.Second {
		t.Errorf("parseRetryAfter(0.5s) should clamp to >= 1s, got %v", got)
	}
	// Values above 90s should clamp to 90s
	err2 := errors.New("retry in 120s")
	got2 := parseRetryAfter(err2)
	if got2 > 90*time.Second {
		t.Errorf("parseRetryAfter(120s) should clamp to <= 90s, got %v", got2)
	}
}

func TestIs429(t *testing.T) {
	tests := []struct {
		err  error
		want bool
	}{
		{errors.New("429 Too Many Requests"), true},
		{errors.New("HTTP 429"), true},
		{errors.New("RESOURCE_EXHAUSTED"), true},
		{errors.New("quota exceeded"), true},
		{errors.New("rate limit"), false},
		{errors.New("500 Internal Server Error"), false},
		{nil, false},
	}
	for _, tt := range tests {
		got := is429(tt.err)
		if got != tt.want {
			t.Errorf("is429(%q) = %v, want %v", tt.err, got, tt.want)
		}
	}
}

func TestNewFallbackProvider_PaidTier(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Providers.PaidTier = true
	fb := NewFallbackProvider(cfg)
	if fb.rateLimit != nil {
		t.Error("paid tier should have nil rate limit")
	}
}

func TestNewFallbackProvider_RateLimitConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Providers.PaidTier = false
	cfg.Providers.LLMRateLimit.MaxPerWindow = 5
	cfg.Providers.LLMRateLimit.WindowSec = 10
	fb := NewFallbackProvider(cfg)
	if fb.rateLimit == nil {
		t.Error("free tier with config should have rate limit")
	}
}

func TestRateLimiter_Wait(t *testing.T) {
	// Use 2 per 1 second window for fast test
	rl := newLLMRateLimiter(2, 1)
	ctx := context.Background()

	// First two should pass immediately
	if err := rl.wait(ctx); err != nil {
		t.Fatalf("first wait: %v", err)
	}
	if err := rl.wait(ctx); err != nil {
		t.Fatalf("second wait: %v", err)
	}

	// Third should block until window slides (~1 sec), then complete
	done := make(chan error, 1)
	go func() {
		done <- rl.wait(ctx)
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("third wait: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Error("third wait took too long (rate limiter may be stuck)")
	}
}

func TestRateLimiter_ContextCancel(t *testing.T) {
	rl := newLLMRateLimiter(1, 60) // 1 per 60 sec - will block
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// First passes
	if err := rl.wait(ctx); err != nil {
		t.Fatal(err)
	}

	// Second will block - cancel context
	done := make(chan error, 1)
	go func() {
		done <- rl.wait(ctx)
	}()
	cancel()
	select {
	case err := <-done:
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("wait did not respect context cancel")
	}
}

func TestCircuitBreaker_RecordAndReset(t *testing.T) {
	cb := &circuitBreaker{}
	cb.record429()
	cb.record429()
	if !cb.isTripped() {
		t.Error("2x 429 should trip circuit breaker")
	}
	cb.recordSuccess()
	if cb.isTripped() {
		t.Error("recordSuccess should reset circuit breaker")
	}
}
