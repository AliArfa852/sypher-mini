package providers

import (
	"os"
	"strings"

	"github.com/sypherexx/sypher-mini/pkg/config"
	"github.com/sypherexx/sypher-mini/pkg/providers/anthropic"
	"github.com/sypherexx/sypher-mini/pkg/providers/gemini"
	"github.com/sypherexx/sypher-mini/pkg/providers/openai_compat"
)

// RoutingStrategy determines provider fallback order.
type RoutingStrategy string

const (
	RoutingCheapFirst    RoutingStrategy = "cheap_first"
	RoutingFastFirst     RoutingStrategy = "fast_first"
	RoutingPowerfulFirst RoutingStrategy = "powerful_first"
)

// NewProvider creates an LLM provider from config (first available).
func NewProvider(cfg *config.Config) (LLMProvider, error) {
	entries := listProviders(cfg)
	for _, e := range entries {
		if e.Provider != nil {
			return e.Provider, nil
		}
	}
	return nil, nil
}

// NewProviderWithFallbacks returns all configured providers in fallback order.
func NewProviderWithFallbacks(cfg *config.Config) []ProviderEntry {
	return listProviders(cfg)
}

func listProviders(cfg *config.Config) []ProviderEntry {
	strategy := RoutingStrategy(strings.ToLower(cfg.Providers.RoutingStrategy))
	if strategy == "" {
		strategy = RoutingCheapFirst
	}

	var entries []ProviderEntry

	// cheap_first: Cerebras -> OpenAI -> Anthropic (Anthropic/Gemini need separate impl)
	if strategy == RoutingCheapFirst || strategy == RoutingFastFirst {
		if key := getAPIKey("CEREBRAS_API_KEY", cfg.Providers.Cerebras.APIKey); key != "" {
			base := cfg.Providers.Cerebras.APIBase
			if base == "" {
				base = "https://api.cerebras.ai/v1"
			}
		entries = append(entries, ProviderEntry{
			Provider: openai_compat.New("cerebras", key, base, "llama-3.1-70b"),
			Name:     "cerebras",
		})
		}
	}

	if key := getAPIKey("OPENAI_API_KEY", cfg.Providers.OpenAI.APIKey); key != "" {
		base := cfg.Providers.OpenAI.APIBase
		if base == "" {
			base = "https://api.openai.com/v1"
		}
		entries = append(entries, ProviderEntry{
			Provider: openai_compat.New("openai", key, base, "gpt-4o-mini"),
			Name:     "openai",
		})
	}

	if key := getAPIKey("ANTHROPIC_API_KEY", cfg.Providers.Anthropic.APIKey); key != "" {
		entries = append(entries, ProviderEntry{
			Provider: anthropic.New(key, "claude-3-5-sonnet-20241022"),
			Name:     "anthropic",
		})
	}

	if key := getAPIKey("GEMINI_API_KEY", cfg.Providers.Gemini.APIKey); key != "" {
		entries = append(entries, ProviderEntry{
			Provider: gemini.New(key, "gemini-2.5-flash-lite"),
			Name:     "gemini",
		})
	}

	return entries
}

func getAPIKey(envKey, configKey string) string {
	if configKey != "" {
		return configKey
	}
	return os.Getenv(envKey)
}
