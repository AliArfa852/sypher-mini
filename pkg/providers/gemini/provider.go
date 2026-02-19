package gemini

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

const baseURL = "https://generativelanguage.googleapis.com/v1beta"

// Provider implements Google AI Studio (Gemini) generateContent API.
type Provider struct {
	apiKey       string
	httpClient   *http.Client
	defaultModel string
}

// New creates a Gemini provider.
func New(apiKey, defaultModel string) *Provider {
	if defaultModel == "" {
		defaultModel = "gemini-1.5-flash"
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

// Chat sends a request to Gemini generateContent API.
func (p *Provider) Chat(ctx context.Context, messages []types.Message, tools []types.ToolDefinition, model string, options map[string]interface{}) (*types.LLMResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("gemini: API key not configured")
	}

	model = normalizeModel(model)
	maxTokens := 2048
	if mt, ok := asInt(options["max_tokens"]); ok && mt > 0 {
		maxTokens = mt
	}

	// Convert to Gemini format: contents with role and parts
	var contents []map[string]interface{}
	var systemInstruction string
	for _, m := range messages {
		if m.Role == "system" {
			systemInstruction = m.Content
			continue
		}
		if m.Role == "tool" {
			continue
		}
		role := "user"
		if m.Role == "assistant" {
			role = "model"
		}
		contents = append(contents, map[string]interface{}{
			"role": role,
			"parts": []map[string]interface{}{{"text": m.Content}},
		})
	}

	reqBody := map[string]interface{}{
		"contents": contents,
		"generationConfig": map[string]interface{}{
			"maxOutputTokens": maxTokens,
		},
	}
	if systemInstruction != "" {
		reqBody["systemInstruction"] = map[string]interface{}{
			"parts": []map[string]interface{}{{"text": systemInstruction}},
		}
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", baseURL, model, p.apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
			FinishReason string `json:"finishReason"`
		} `json:"candidates"`
		UsageMetadata *struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	var content, finishReason string
	if len(apiResp.Candidates) > 0 {
		for _, p := range apiResp.Candidates[0].Content.Parts {
			content += p.Text
		}
		finishReason = apiResp.Candidates[0].FinishReason
	}

	usage := (*types.UsageInfo)(nil)
	if apiResp.UsageMetadata != nil {
		usage = &types.UsageInfo{
			PromptTokens:     apiResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: apiResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      apiResp.UsageMetadata.TotalTokenCount,
		}
	}

	return &types.LLMResponse{
		Content:      content,
		FinishReason: finishReason,
		Usage:        usage,
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
