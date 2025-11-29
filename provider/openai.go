package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ZaguanLabs/gotlai"
	"github.com/sashabaranov/go-openai"
)

// OpenAIProvider implements AIProvider using OpenAI's API.
type OpenAIProvider struct {
	client      *openai.Client
	model       string
	temperature float32
}

// OpenAIConfig holds configuration for the OpenAI provider.
type OpenAIConfig struct {
	APIKey      string  // OpenAI API key (uses OPENAI_API_KEY env var if empty)
	Model       string  // Model to use (default: "gpt-4o-mini")
	Temperature float32 // Temperature for generation (default: 0.3)
	BaseURL     string  // Custom base URL (optional)
}

// NewOpenAIProvider creates a new OpenAI provider.
func NewOpenAIProvider(cfg OpenAIConfig) *OpenAIProvider {
	config := openai.DefaultConfig(cfg.APIKey)
	if cfg.BaseURL != "" {
		config.BaseURL = cfg.BaseURL
	}

	model := cfg.Model
	if model == "" {
		model = "gpt-4o-mini"
	}

	temperature := cfg.Temperature
	if temperature == 0 {
		temperature = 0.3
	}

	return &OpenAIProvider{
		client:      openai.NewClientWithConfig(config),
		model:       model,
		temperature: temperature,
	}
}

// Translate translates a batch of texts using OpenAI.
func (p *OpenAIProvider) Translate(ctx context.Context, req TranslateRequest) ([]string, error) {
	if len(req.Texts) == 0 {
		return []string{}, nil
	}

	systemPrompt := p.buildSystemPrompt(req)
	userMessage := p.buildUserMessage(req)

	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: p.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userMessage},
		},
		Temperature: p.temperature,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	})
	if err != nil {
		return nil, &gotlai.ProviderError{
			Message:   "OpenAI API call failed",
			Cause:     err,
			Retryable: isRetryableError(err),
		}
	}

	if len(resp.Choices) == 0 {
		return nil, &gotlai.ProviderError{
			Message:   "no response from OpenAI",
			Retryable: true,
		}
	}

	translations, err := p.parseResponse(resp.Choices[0].Message.Content, len(req.Texts))
	if err != nil {
		return nil, err
	}

	return translations, nil
}

func (p *OpenAIProvider) buildSystemPrompt(req TranslateRequest) string {
	sourceLang := req.SourceLang
	if sourceLang == "" {
		sourceLang = "en"
	}

	targetName := gotlai.GetLanguageName(req.TargetLang)
	sourceName := gotlai.GetLanguageName(sourceLang)

	contextText := req.Context
	if contextText == "" {
		contextText = "General web or application content."
	}

	prompt := fmt.Sprintf(`# Role
You are an expert native translator. Translate from %s to %s.

# Context
%s

# Style Guide
- Natural flow: Avoid literal translations
- Vocabulary: Use culturally relevant terminology
- Disambiguation: Use provided context to choose correct translation
  - "Run" in <button> → action verb (execute)
  - "Run" in sports → physical running
- HTML Safety: Do NOT translate tags, classes, IDs
- Variables: Do NOT translate {{name}}, {count}, %%s, etc.

# Input Format
Either:
1. Array: ["text1", "text2"]
2. Object: {"items": [{"text": "Run", "context": "in <button>"}]}

# Output Format
Return ONLY: {"translations": ["translated1", "translated2"]}`, sourceName, targetName, contextText)

	if len(req.ExcludedTerms) > 0 {
		terms := strings.Join(req.ExcludedTerms, "\n- ")
		prompt += fmt.Sprintf("\n\n# Exclusions\nDo NOT translate:\n- %s", terms)
	}

	return prompt
}

func (p *OpenAIProvider) buildUserMessage(req TranslateRequest) string {
	// If we have per-text contexts, use the object format
	hasContexts := false
	for _, ctx := range req.TextContexts {
		if ctx != "" {
			hasContexts = true
			break
		}
	}

	if !hasContexts {
		// Simple array format
		data, _ := json.Marshal(req.Texts)
		return string(data)
	}

	// Object format with contexts
	type item struct {
		Text    string `json:"text"`
		Context string `json:"context,omitempty"`
	}

	items := make([]item, len(req.Texts))
	for i, text := range req.Texts {
		items[i].Text = text
		if i < len(req.TextContexts) {
			items[i].Context = req.TextContexts[i]
		}
	}

	data, _ := json.Marshal(map[string][]item{"items": items})
	return string(data)
}

func (p *OpenAIProvider) parseResponse(content string, expectedCount int) ([]string, error) {
	// Try parsing as object first
	var objResult map[string]interface{}
	if err := json.Unmarshal([]byte(content), &objResult); err == nil {
		// Look for "translations" key
		if translations, ok := objResult["translations"]; ok {
			if arr, ok := translations.([]interface{}); ok {
				return toStringSlice(arr, expectedCount)
			}
		}

		// Fallback: find first array value
		for _, v := range objResult {
			if arr, ok := v.([]interface{}); ok {
				return toStringSlice(arr, expectedCount)
			}
		}
	}

	// Try parsing as direct array
	var arrResult []interface{}
	if err := json.Unmarshal([]byte(content), &arrResult); err == nil {
		return toStringSlice(arrResult, expectedCount)
	}

	return nil, &gotlai.ProviderError{
		Message:   "invalid response format from OpenAI",
		Retryable: false,
	}
}

func toStringSlice(arr []interface{}, expectedCount int) ([]string, error) {
	result := make([]string, len(arr))
	for i, v := range arr {
		if s, ok := v.(string); ok {
			result[i] = s
		} else {
			result[i] = fmt.Sprintf("%v", v)
		}
	}

	if len(result) != expectedCount {
		return nil, &gotlai.CountMismatchError{
			Expected: expectedCount,
			Got:      len(result),
		}
	}

	return result, nil
}

func isRetryableError(err error) bool {
	// Check for common retryable conditions
	errStr := err.Error()
	retryablePatterns := []string{
		"rate limit",
		"timeout",
		"connection refused",
		"temporary",
		"503",
		"502",
		"429",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return true
		}
	}
	return false
}

// Verify OpenAIProvider implements AIProvider
var _ AIProvider = (*OpenAIProvider)(nil)
