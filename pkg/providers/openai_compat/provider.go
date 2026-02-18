package openai_compat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sypherexx/sypher-mini/pkg/providers"
)

// Provider is an OpenAI-compatible API provider (OpenAI, Cerebras, etc.).
type Provider struct {
	name       string
	apiKey     string
	apiBase    string
	httpClient *http.Client
	defaultModel string
}

// New creates a new OpenAI-compatible provider.
func New(name, apiKey, apiBase, defaultModel string) *Provider {
	if apiBase == "" {
		switch strings.ToLower(name) {
		case "cerebras":
			apiBase = "https://api.cerebras.ai/v1"
		case "openai":
			apiBase = "https://api.openai.com/v1"
		default:
			apiBase = "https://api.openai.com/v1"
		}
	}
	apiBase = strings.TrimRight(apiBase, "/")
	if defaultModel == "" {
		defaultModel = "gpt-4o-mini"
	}
	return &Provider{
		name:         name,
		apiKey:       apiKey,
		apiBase:      apiBase,
		defaultModel: defaultModel,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// Chat sends a chat completion request.
func (p *Provider) Chat(ctx context.Context, messages []providers.Message, tools []providers.ToolDefinition, model string, options map[string]interface{}) (*providers.LLMResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("%s: API key not configured", p.name)
	}

	model = p.normalizeModel(model)

	requestBody := map[string]interface{}{
		"model":    model,
		"messages": messages,
	}

	if len(tools) > 0 {
		requestBody["tools"] = tools
		requestBody["tool_choice"] = "auto"
	}

	if maxTokens, ok := asInt(options["max_tokens"]); ok && maxTokens > 0 {
		requestBody["max_tokens"] = maxTokens
	}
	if temperature, ok := asFloat(options["temperature"]); ok {
		requestBody["temperature"] = temperature
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.apiBase+"/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

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

func (p *Provider) normalizeModel(model string) string {
	// Strip provider prefix if present (e.g. "cerebras/llama-3.1-70b" -> "llama-3.1-70b")
	if idx := strings.Index(model, "/"); idx > 0 {
		return model[idx+1:]
	}
	return model
}

// GetDefaultModel returns the default model.
func (p *Provider) GetDefaultModel() string {
	return p.defaultModel
}

func parseResponse(body []byte) (*providers.LLMResponse, error) {
	var apiResponse struct {
		Choices []struct {
			Message struct {
				Content   string `json:"content"`
				ToolCalls []struct {
					ID       string `json:"id"`
					Type     string `json:"type"`
					Function *struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage *providers.UsageInfo `json:"usage"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if len(apiResponse.Choices) == 0 {
		return &providers.LLMResponse{
			Content:      "",
			FinishReason: "stop",
		}, nil
	}

	choice := apiResponse.Choices[0]
	toolCalls := make([]providers.ToolCall, 0, len(choice.Message.ToolCalls))
	for _, tc := range choice.Message.ToolCalls {
		arguments := make(map[string]interface{})
		name := ""
		if tc.Function != nil {
			name = tc.Function.Name
			if tc.Function.Arguments != "" {
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &arguments)
			}
		}
		toolCalls = append(toolCalls, providers.ToolCall{
			ID:        tc.ID,
			Name:      name,
			Arguments: arguments,
		})
	}

	return &providers.LLMResponse{
		Content:      choice.Message.Content,
		ToolCalls:    toolCalls,
		FinishReason: choice.FinishReason,
		Usage:        apiResponse.Usage,
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

func asFloat(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	default:
		return 0, false
	}
}
