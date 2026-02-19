package providers

import (
	"context"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sypherexx/sypher-mini/pkg/config"
)

// retryAfterRegex parses "retry in X.XXXs" or "retry in Xs" from API error bodies.
var retryAfterRegex = regexp.MustCompile(`[Rr]etry in (\d+(?:\.\d+)?)s`)

// FallbackProvider tries providers in order with retries.
type FallbackProvider struct {
	entries   []ProviderEntry
	retryMax  int
	retryBase time.Duration
}

// NewFallbackProvider creates a provider that falls back on failure.
func NewFallbackProvider(cfg *config.Config) *FallbackProvider {
	retryMax := cfg.Task.RetryMax
	if retryMax <= 0 {
		retryMax = 2
	}
	return &FallbackProvider{
		entries:   NewProviderWithFallbacks(cfg),
		retryMax:  retryMax,
		retryBase: time.Second,
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
func parseRetryAfter(err error) time.Duration {
	if err == nil {
		return 0
	}
	m := retryAfterRegex.FindStringSubmatch(err.Error())
	if len(m) < 2 {
		return 0
	}
	sec, _ := strconv.ParseFloat(m[1], 64)
	if sec <= 0 {
		return 0
	}
	d := time.Duration(sec * float64(time.Second))
	if d < time.Second {
		d = time.Second
	}
	if d > 90*time.Second {
		d = 90 * time.Second
	}
	return d
}

// Chat tries each provider with retries.
func (f *FallbackProvider) Chat(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (*LLMResponse, error) {
	var lastErr error
	for _, e := range f.entries {
		if e.Provider == nil {
			continue
		}
		maxAttempts := f.retryMax + 1
		for attempt := 0; attempt < maxAttempts; attempt++ {
			if attempt > 0 {
				backoff := f.retryBase * time.Duration(1<<uint(attempt-1))
				if backoff > 30*time.Second {
					backoff = 30 * time.Second
				}
				if is429(lastErr) {
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
				return resp, nil
			}
			lastErr = err
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
