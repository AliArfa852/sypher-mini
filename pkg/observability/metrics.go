package observability

import (
	"sync"
)

// Metrics holds simple counters and histograms for observability.
type Metrics struct {
	mu sync.RWMutex

	ToolCallsTotal   map[string]int
	ToolErrorsTotal  map[string]int
	LLMRequestsTotal map[string]int
	TaskCompleted    int
	TaskFailed       int
}

// NewMetrics creates a new metrics collector.
func NewMetrics() *Metrics {
	return &Metrics{
		ToolCallsTotal:   make(map[string]int),
		ToolErrorsTotal:  make(map[string]int),
		LLMRequestsTotal: make(map[string]int),
	}
}

// IncToolCall increments tool call count.
func (m *Metrics) IncToolCall(tool string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ToolCallsTotal[tool]++
}

// IncToolError increments tool error count.
func (m *Metrics) IncToolError(tool string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ToolErrorsTotal[tool]++
}

// IncLLMRequest increments LLM request count by provider.
func (m *Metrics) IncLLMRequest(provider string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.LLMRequestsTotal[provider]++
}

// IncTaskCompleted increments completed task count.
func (m *Metrics) IncTaskCompleted() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TaskCompleted++
}

// IncTaskFailed increments failed task count.
func (m *Metrics) IncTaskFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TaskFailed++
}

// Snapshot returns a copy of current metrics.
func (m *Metrics) Snapshot() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	toolCalls := make(map[string]int)
	for k, v := range m.ToolCallsTotal {
		toolCalls[k] = v
	}
	toolErrors := make(map[string]int)
	for k, v := range m.ToolErrorsTotal {
		toolErrors[k] = v
	}
	llmReqs := make(map[string]int)
	for k, v := range m.LLMRequestsTotal {
		llmReqs[k] = v
	}
	return map[string]interface{}{
		"tool_calls_total":   toolCalls,
		"tool_errors_total":  toolErrors,
		"llm_requests_total": llmReqs,
		"task_completed":     m.TaskCompleted,
		"task_failed":        m.TaskFailed,
	}
}
