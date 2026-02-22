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

// Default models per provider (cost-effective, sufficient quality for most tasks).
const (
	defaultCerebrasModel  = "llama-3.1-70b"
	defaultOpenAIModel    = "gpt-4o-mini"
	defaultAnthropicModel = "claude-3-5-sonnet-20241022"
	defaultGeminiModel    = "gemini-2.5-flash-lite"
	defaultDeepSeekModel  = "deepseek-chat"
)

func defaultModel(cfg *config.ProviderConfig, fallback string) string {
	if cfg.DefaultModel != "" {
		return cfg.DefaultModel
	}
	return fallback
}

func listProviders(cfg *config.Config) []ProviderEntry {
	strategy := RoutingStrategy(strings.ToLower(cfg.Providers.RoutingStrategy))
	if strategy == "" {
		strategy = RoutingCheapFirst
	}

	var entries []ProviderEntry

	// cheap_first: Cerebras -> DeepSeek -> OpenAI -> Anthropic -> Gemini
	if strategy == RoutingCheapFirst || strategy == RoutingFastFirst {
		if key := getAPIKey("CEREBRAS_API_KEY", cfg.Providers.Cerebras.APIKey); key != "" {
			base := cfg.Providers.Cerebras.APIBase
			if base == "" {
				base = "https://api.cerebras.ai/v1"
			}
			model := defaultModel(&cfg.Providers.Cerebras, defaultCerebrasModel)
			entries = append(entries, ProviderEntry{
				Provider: openai_compat.New("cerebras", key, base, model),
				Name:     "cerebras",
			})
		}

		if key := getAPIKey("DEEPSEEK_API_KEY", cfg.Providers.DeepSeek.APIKey); key != "" {
			base := cfg.Providers.DeepSeek.APIBase
			if base == "" {
				base = "https://api.deepseek.com/v1"
			}
			model := defaultModel(&cfg.Providers.DeepSeek, defaultDeepSeekModel)
			entries = append(entries, ProviderEntry{
				Provider: openai_compat.New("deepseek", key, base, model),
				Name:     "deepseek",
			})
		}
	}

	if key := getAPIKey("OPENAI_API_KEY", cfg.Providers.OpenAI.APIKey); key != "" {
		base := cfg.Providers.OpenAI.APIBase
		if base == "" {
			base = "https://api.openai.com/v1"
		}
		model := defaultModel(&cfg.Providers.OpenAI, defaultOpenAIModel)
		entries = append(entries, ProviderEntry{
			Provider: openai_compat.New("openai", key, base, model),
			Name:     "openai",
		})
	}

	if key := getAPIKey("ANTHROPIC_API_KEY", cfg.Providers.Anthropic.APIKey); key != "" {
		model := defaultModel(&cfg.Providers.Anthropic, defaultAnthropicModel)
		entries = append(entries, ProviderEntry{
			Provider: anthropic.New(key, model),
			Name:     "anthropic",
		})
	}

	if key := getAPIKey("GEMINI_API_KEY", cfg.Providers.Gemini.APIKey); key != "" {
		model := defaultModel(&cfg.Providers.Gemini, defaultGeminiModel)
		entries = append(entries, ProviderEntry{
			Provider: gemini.New(key, model),
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
