package providers

import (
	"context"
	"log"
	"time"

	"github.com/sypherexx/sypher-mini/pkg/config"
)

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

// Chat tries each provider with retries.
func (f *FallbackProvider) Chat(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (*LLMResponse, error) {
	var lastErr error
	for _, e := range f.entries {
		if e.Provider == nil {
			continue
		}
		for attempt := 0; attempt <= f.retryMax; attempt++ {
			if attempt > 0 {
				backoff := f.retryBase * time.Duration(1<<uint(attempt-1))
				if backoff > 30*time.Second {
					backoff = 30 * time.Second
				}
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
