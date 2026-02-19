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

// New creates a Gemini provider. Default model is gemini-2.5-flash-lite unless changed.
func New(apiKey, defaultModel string) *Provider {
	if defaultModel == "" {
		defaultModel = "gemini-2.5-flash-lite"
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

// toGeminiTools converts our tool definitions to Gemini functionDeclarations format.
func toGeminiTools(tools []types.ToolDefinition) []map[string]interface{} {
	if len(tools) == 0 {
		return nil
	}
	var decls []map[string]interface{}
	for _, t := range tools {
		if t.Type != "function" || t.Function.Name == "" {
			continue
		}
		params := t.Function.Parameters
		if params == nil {
			params = map[string]interface{}{"type": "object", "properties": map[string]interface{}{}}
		}
		decls = append(decls, map[string]interface{}{
			"name":        t.Function.Name,
			"description": t.Function.Description,
			"parameters":  params,
		})
	}
	if len(decls) == 0 {
		return nil
	}
	return []map[string]interface{}{
		{"functionDeclarations": decls},
	}
}

// isGeminiModel returns true if the model is a Gemini model (e.g. gemini-2.5-flash-lite).
func isGeminiModel(model string) bool {
	return strings.HasPrefix(model, "gemini-") || model == "gemini"
}

// Chat sends a request to Gemini generateContent API with function calling support.
func (p *Provider) Chat(ctx context.Context, messages []types.Message, tools []types.ToolDefinition, model string, options map[string]interface{}) (*types.LLMResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("gemini: API key not configured")
	}

	model = normalizeModel(model)
	if model == "" || !isGeminiModel(model) {
		model = p.defaultModel
	}
	maxTokens := 2048
	if mt, ok := asInt(options["max_tokens"]); ok && mt > 0 {
		maxTokens = mt
	}

	// Convert to Gemini format: contents with role and parts
	contents, err := toGeminiContents(messages)
	if err != nil {
		return nil, fmt.Errorf("convert messages: %w", err)
	}

	reqBody := map[string]interface{}{
		"contents": contents,
		"generationConfig": map[string]interface{}{
			"maxOutputTokens": maxTokens,
		},
	}

	var systemInstruction string
	for _, m := range messages {
		if m.Role == "system" {
			systemInstruction = m.Content
			break
		}
	}
	if systemInstruction != "" {
		reqBody["systemInstruction"] = map[string]interface{}{
			"parts": []map[string]interface{}{{"text": systemInstruction}},
		}
	}

	if geminiTools := toGeminiTools(tools); len(geminiTools) > 0 {
		reqBody["tools"] = geminiTools
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

// toGeminiContents converts our messages to Gemini contents, including tool calls and responses.
func toGeminiContents(messages []types.Message) ([]map[string]interface{}, error) {
	var contents []map[string]interface{}
	var pendingToolResponses []map[string]interface{}
	toolCallIDToName := make(map[string]string)

	for i := 0; i < len(messages); i++ {
		m := messages[i]
		switch m.Role {
		case "system":
			// Handled separately as systemInstruction
			continue
		case "user":
			if len(pendingToolResponses) > 0 {
				contents = append(contents, map[string]interface{}{
					"role":  "user",
					"parts": pendingToolResponses,
				})
				pendingToolResponses = nil
			}
			contents = append(contents, map[string]interface{}{
				"role":  "user",
				"parts": []map[string]interface{}{{"text": m.Content}},
			})
		case "assistant":
			if len(m.ToolCalls) > 0 {
				for _, tc := range m.ToolCalls {
					toolCallIDToName[tc.ID] = tc.Name
				}
				parts := make([]map[string]interface{}, 0, len(m.ToolCalls)+1)
				if m.Content != "" {
					parts = append(parts, map[string]interface{}{"text": m.Content})
				}
				for _, tc := range m.ToolCalls {
					args := tc.Arguments
					if args == nil {
						args = map[string]interface{}{}
					}
					parts = append(parts, map[string]interface{}{
						"functionCall": map[string]interface{}{
							"name": tc.Name,
							"args": args,
						},
					})
				}
				contents = append(contents, map[string]interface{}{
					"role":  "model",
					"parts": parts,
				})
			} else if m.Content != "" {
				contents = append(contents, map[string]interface{}{
					"role":  "model",
					"parts": []map[string]interface{}{{"text": m.Content}},
				})
			}
		case "tool":
			name := toolCallIDToName[m.ToolCallID]
			if name == "" {
				name = "exec" // fallback
			}
			pendingToolResponses = append(pendingToolResponses, map[string]interface{}{
				"functionResponse": map[string]interface{}{
					"name": name,
					"response": map[string]interface{}{
						"result": m.Content,
					},
				},
			})
		}
	}
	// If we ended with tool responses, add a user turn with them
	if len(pendingToolResponses) > 0 {
		contents = append(contents, map[string]interface{}{
			"role":  "user",
			"parts": pendingToolResponses,
		})
	}

	return contents, nil
}

func parseResponse(body []byte) (*types.LLMResponse, error) {
	var apiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text         string                 `json:"text"`
					FunctionCall *struct {
						Name string                 `json:"name"`
						Args map[string]interface{} `json:"args"`
					} `json:"functionCall"`
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
	var toolCalls []types.ToolCall
	toolCallID := 0

	if len(apiResp.Candidates) > 0 {
		for _, p := range apiResp.Candidates[0].Content.Parts {
			if p.Text != "" {
				content += p.Text
			}
			if p.FunctionCall != nil {
				toolCallID++
				toolCalls = append(toolCalls, types.ToolCall{
					ID:        fmt.Sprintf("call_%d", toolCallID),
					Type:      "function",
					Name:      p.FunctionCall.Name,
					Arguments: p.FunctionCall.Args,
				})
			}
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
		ToolCalls:    toolCalls,
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
