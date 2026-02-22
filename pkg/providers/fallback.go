package providers

import (
	"context"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sypherexx/sypher-mini/pkg/config"
)

// retryAfterRegex parses "retry in X.XXXs" or "retry in Xs" from API error bodies.
var retryAfterRegex = regexp.MustCompile(`[Rr]etry in (\d+(?:\.\d+)?)s`)

// retryAfterJSONRegex parses "retry_after": N or "retryAfter": N from JSON error bodies.
var retryAfterJSONRegex = regexp.MustCompile(`"(?:retry_after|retryAfter)"\s*:\s*(\d+(?:\.\d+)?)`)

// llmRateLimiter limits API calls per sliding window (e.g. 2 per 15 sec).
type llmRateLimiter struct {
	mu     sync.Mutex
	times  []time.Time
	max    int
	window time.Duration
}

func newLLMRateLimiter(maxPerWindow, windowSec int) *llmRateLimiter {
	if maxPerWindow <= 0 {
		maxPerWindow = 2
	}
	if windowSec <= 0 {
		windowSec = 15
	}
	return &llmRateLimiter{
		max:    maxPerWindow,
		window: time.Duration(windowSec) * time.Second,
	}
}

func (r *llmRateLimiter) wait(ctx context.Context) error {
	for {
		r.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-r.window)
		var valid []time.Time
		for _, t := range r.times {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		if len(valid) < r.max {
			r.times = append(valid, now)
			r.mu.Unlock()
			return nil
		}
		oldest := valid[0]
		waitUntil := oldest.Add(r.window)
		waitDur := time.Until(waitUntil)
		r.mu.Unlock()
		if waitDur <= 0 {
			continue
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitDur):
		}
	}
}

// circuitBreaker backs off after consecutive 429 errors.
type circuitBreaker struct {
	mu             sync.Mutex
	consecutive429 int
	trippedUntil   time.Time
}

func (c *circuitBreaker) record429() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.consecutive429++
	if c.consecutive429 >= 2 && c.trippedUntil.IsZero() {
		backoff := 60 * time.Second
		if c.consecutive429 >= 3 {
			backoff = 120 * time.Second
		}
		c.trippedUntil = time.Now().Add(backoff)
	}
}

func (c *circuitBreaker) recordSuccess() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.consecutive429 = 0
	c.trippedUntil = time.Time{}
}

func (c *circuitBreaker) wait(ctx context.Context) error {
	c.mu.Lock()
	until := c.trippedUntil
	c.mu.Unlock()
	if until.IsZero() {
		return nil
	}
	waitDur := time.Until(until)
	if waitDur <= 0 {
		c.mu.Lock()
		c.trippedUntil = time.Time{}
		c.mu.Unlock()
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(waitDur):
	}
	c.mu.Lock()
	c.trippedUntil = time.Time{}
	c.mu.Unlock()
	return nil
}

func (c *circuitBreaker) isTripped() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return !c.trippedUntil.IsZero() && time.Now().Before(c.trippedUntil)
}

// FallbackProvider tries providers in order with retries.
type FallbackProvider struct {
	entries       []ProviderEntry
	retryMax      int
	retryBase     time.Duration
	rateLimit     *llmRateLimiter
	circuitBreaker *circuitBreaker
}

// NewFallbackProvider creates a provider that falls back on failure.
func NewFallbackProvider(cfg *config.Config) *FallbackProvider {
	retryMax := cfg.Task.RetryMax
	if retryMax <= 0 {
		retryMax = 2
	}
	rl := (*llmRateLimiter)(nil)
	if !cfg.Providers.PaidTier {
		if cfg.Providers.LLMRateLimit.MaxPerWindow > 0 || cfg.Providers.LLMRateLimit.WindowSec > 0 {
			rl = newLLMRateLimiter(cfg.Providers.LLMRateLimit.MaxPerWindow, cfg.Providers.LLMRateLimit.WindowSec)
		} else {
			rl = newLLMRateLimiter(2, 15) // default: 2 per 15 sec (free tier)
		}
	}
	return &FallbackProvider{
		entries:        NewProviderWithFallbacks(cfg),
		retryMax:       retryMax,
		retryBase:      time.Second,
		rateLimit:      rl,
		circuitBreaker: &circuitBreaker{},
	}
}

// Entries returns the configured provider entries.
func (f *FallbackProvider) Entries() []ProviderEntry {
	return f.entries
}

// is429 returns true if the error indicates a rate limit (429).
func is429(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "429") || strings.Contains(s, "RESOURCE_EXHAUSTED") || strings.Contains(s, "quota")
}

// parseRetryAfter extracts suggested wait time in seconds from error body.
// Handles "retry in Xs", JSON "retry_after": N, and "retryAfter": N.
func parseRetryAfter(err error) time.Duration {
	if err == nil {
		return 0
	}
	s := err.Error()
	// Try "retry in Xs" format first
	if m := retryAfterRegex.FindStringSubmatch(s); len(m) >= 2 {
		if sec, _ := strconv.ParseFloat(m[1], 64); sec > 0 {
			return clampRetryDuration(sec)
		}
	}
	// Try JSON retry_after / retryAfter
	if m := retryAfterJSONRegex.FindStringSubmatch(s); len(m) >= 2 {
		if sec, _ := strconv.ParseFloat(m[1], 64); sec > 0 {
			return clampRetryDuration(sec)
		}
	}
	return 0
}

func clampRetryDuration(sec float64) time.Duration {
	d := time.Duration(sec * float64(time.Second))
	if d < time.Second {
		d = time.Second
	}
	if d > 90*time.Second {
		d = 90 * time.Second
	}
	return d
}

// Chat tries each provider with retries. Respects LLM rate limit (default 2 per 15 sec).
func (f *FallbackProvider) Chat(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (*LLMResponse, error) {
	var lastErr error
	for _, e := range f.entries {
		if e.Provider == nil {
			continue
		}
		// Circuit breaker: wait if we've had 2+ consecutive 429s
		if f.circuitBreaker != nil && f.circuitBreaker.isTripped() {
			if err := f.circuitBreaker.wait(ctx); err != nil {
				return nil, err
			}
		}
		maxAttempts := f.retryMax + 1
		for attempt := 0; attempt < maxAttempts; attempt++ {
			if f.rateLimit != nil {
				if err := f.rateLimit.wait(ctx); err != nil {
					return nil, err
				}
			}
			if attempt > 0 {
				backoff := f.retryBase * time.Duration(1<<uint(attempt-1))
				if backoff > 30*time.Second {
					backoff = 30 * time.Second
				}
				if is429(lastErr) {
					f.circuitBreaker.record429()
					if parsed := parseRetryAfter(lastErr); parsed > 0 {
						backoff = parsed
					} else {
						backoff = 60 * time.Second
					}
					if attempt >= 2 {
						break
					}
				}
				log.Printf("LLM %s rate limited, waiting %v before retry", e.Name, backoff)
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(backoff):
				}
			}
			resp, err := e.Provider.Chat(ctx, messages, tools, model, options)
			if err == nil {
				f.circuitBreaker.recordSuccess()
				return resp, nil
			}
			lastErr = err
			if !is429(err) {
				f.circuitBreaker.recordSuccess()
			}
			log.Printf("LLM %s attempt %d failed: %v", e.Name, attempt+1, err)
		}
	}
	return nil, lastErr
}

// GetDefaultModel returns the first provider's default model.
func (f *FallbackProvider) GetDefaultModel() string {
	for _, e := range f.entries {
		if e.Provider != nil {
			return e.Provider.GetDefaultModel()
		}
	}
	return "gpt-4o-mini"
}
