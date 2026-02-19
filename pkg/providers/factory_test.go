package providers

import (
	"errors"
	"strings"
	"testing"

	"github.com/sypherexx/sypher-mini/pkg/config"
)

func TestNewProviderWithFallbacks_EmptyConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	entries := NewProviderWithFallbacks(cfg)
	// With no API keys, entries may be empty or have nil providers
	if len(entries) > 0 {
		for _, e := range entries {
			if e.Provider != nil {
				t.Error("expected nil providers when no API keys set")
			}
		}
	}
}

func TestNewProvider_EmptyConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	p, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("NewProvider err = %v", err)
	}
	if p != nil {
		t.Error("expected nil provider when no API keys set")
	}
}

func TestFailoverError(t *testing.T) {
	wrapped := errors.New("rate limit exceeded")
	err := &FailoverError{
		Reason:   FailoverRateLimit,
		Provider: "openai",
		Model:    "gpt-4",
		Wrapped:  wrapped,
	}
	s := err.Error()
	if s == "" {
		t.Error("FailoverError.Error() returned empty")
	}
	if !strings.Contains(s, "openai") {
		t.Errorf("Error() should contain provider name: %q", s)
	}
	if err.Unwrap() != wrapped {
		t.Error("Unwrap() should return wrapped error")
	}
}
