package observability

import (
	"fmt"
	"sort"
	"strings"
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

// PrometheusFormat returns metrics in Prometheus text exposition format.
func (m *Metrics) PrometheusFormat() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var b strings.Builder
	b.WriteString("# HELP sypher_task_completed Total completed tasks\n")
	b.WriteString("# TYPE sypher_task_completed counter\n")
	b.WriteString(fmt.Sprintf("sypher_task_completed %d\n", m.TaskCompleted))
	b.WriteString("# HELP sypher_task_failed Total failed tasks\n")
	b.WriteString("# TYPE sypher_task_failed counter\n")
	b.WriteString(fmt.Sprintf("sypher_task_failed %d\n", m.TaskFailed))

	// tool_calls_total
	b.WriteString("# HELP sypher_tool_calls_total Total tool calls by tool\n")
	b.WriteString("# TYPE sypher_tool_calls_total counter\n")
	var toolNames []string
	for k := range m.ToolCallsTotal {
		toolNames = append(toolNames, k)
	}
	sort.Strings(toolNames)
	for _, tool := range toolNames {
		b.WriteString(fmt.Sprintf("sypher_tool_calls_total{tool=%q} %d\n", tool, m.ToolCallsTotal[tool]))
	}

	// tool_errors_total
	b.WriteString("# HELP sypher_tool_errors_total Total tool errors by tool\n")
	b.WriteString("# TYPE sypher_tool_errors_total counter\n")
	toolNames = nil
	for k := range m.ToolErrorsTotal {
		toolNames = append(toolNames, k)
	}
	sort.Strings(toolNames)
	for _, tool := range toolNames {
		b.WriteString(fmt.Sprintf("sypher_tool_errors_total{tool=%q} %d\n", tool, m.ToolErrorsTotal[tool]))
	}

	return b.String()
}
