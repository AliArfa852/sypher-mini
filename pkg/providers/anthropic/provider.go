package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sypherexx/sypher-mini/pkg/providers/types"
)

// Provider implements Anthropic Messages API.
type Provider struct {
	apiKey       string
	httpClient   *http.Client
	defaultModel string
}

// New creates an Anthropic provider.
func New(apiKey, defaultModel string) *Provider {
	if defaultModel == "" {
		defaultModel = "claude-3-5-sonnet-20241022"
	}
	return &Provider{
		apiKey:       apiKey,
		defaultModel: normalizeModel(defaultModel),
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func normalizeModel(model string) string {
	if idx := strings.Index(model, "/"); idx > 0 {
		return model[idx+1:]
	}
	return model
}

// Chat sends a request to Anthropic Messages API.
func (p *Provider) Chat(ctx context.Context, messages []types.Message, tools []types.ToolDefinition, model string, options map[string]interface{}) (*types.LLMResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("anthropic: API key not configured")
	}

	model = normalizeModel(model)
	if model == "" || !strings.HasPrefix(model, "claude-") {
		model = p.defaultModel
	}
	maxTokens := 2048
	if mt, ok := asInt(options["max_tokens"]); ok && mt > 0 {
		maxTokens = mt
	}

	// Convert to Anthropic format: system (from first system message), messages (user/assistant only)
	var system string
	var anthropicMessages []map[string]interface{}
	for _, m := range messages {
		if m.Role == "system" {
			system = m.Content
			continue
		}
		if m.Role == "tool" {
			// Anthropic uses tool_result in assistant turn; skip for simple impl
			continue
		}
		anthropicMessages = append(anthropicMessages, map[string]interface{}{
			"role":    m.Role,
			"content": m.Content,
		})
	}

	reqBody := map[string]interface{}{
		"model":      model,
		"max_tokens": maxTokens,
		"messages":  anthropicMessages,
	}
	if system != "" {
		reqBody["system"] = system
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return parseResponse(body)
}

func parseResponse(body []byte) (*types.LLMResponse, error) {
	var apiResp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		StopReason string `json:"stop_reason"`
		Usage      struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	var content string
	for _, c := range apiResp.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	return &types.LLMResponse{
		Content:      content,
		FinishReason: apiResp.StopReason,
		Usage: &types.UsageInfo{
			PromptTokens:     apiResp.Usage.InputTokens,
			CompletionTokens: apiResp.Usage.OutputTokens,
			TotalTokens:      apiResp.Usage.InputTokens + apiResp.Usage.OutputTokens,
		},
	}, nil
}

func asInt(v interface{}) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case int64:
		return int(val), true
	case float64:
		return int(val), true
	default:
		return 0, false
	}
}

// GetDefaultModel returns the default model.
func (p *Provider) GetDefaultModel() string {
	return p.defaultModel
}
