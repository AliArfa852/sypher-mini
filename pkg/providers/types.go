package providers

import (
	"fmt"

	"github.com/sypherexx/sypher-mini/pkg/providers/types"
)

// Re-export types for backward compatibility.
type (
	Message             = types.Message
	ToolCall            = types.ToolCall
	LLMResponse         = types.LLMResponse
	ToolDefinition      = types.ToolDefinition
	ToolFunctionDefinition = types.ToolFunctionDefinition
	UsageInfo           = types.UsageInfo
	LLMProvider         = types.LLMProvider
)

// ProviderEntry holds a provider and its priority.
type ProviderEntry struct {
	Provider LLMProvider
	Name     string
}

// FailoverReason classifies why an LLM request failed.
type FailoverReason string

const (
	FailoverAuth      FailoverReason = "auth"
	FailoverRateLimit FailoverReason = "rate_limit"
	FailoverTimeout   FailoverReason = "timeout"
	FailoverFormat    FailoverReason = "format"
	FailoverUnknown   FailoverReason = "unknown"
)

// FailoverError wraps an LLM provider error.
type FailoverError struct {
	Reason   FailoverReason
	Provider string
	Model    string
	Status   int
	Wrapped  error
}

func (e *FailoverError) Error() string {
	return fmt.Sprintf("failover(%s): provider=%s model=%s: %v", e.Reason, e.Provider, e.Model, e.Wrapped)
}

func (e *FailoverError) Unwrap() error {
	return e.Wrapped
}
