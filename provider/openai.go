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
	localeHint := gotlai.GetLocaleClarification(req.TargetLang)

	// Get style description (default to neutral)
	styleDesc := gotlai.GetStyleDescription(req.Style)

	// Build context section
	contextText := "The content is general web content."
	if req.Context != "" {
		contextText = fmt.Sprintf("The content is for: %s. Adapt the tone to be appropriate for this context.", req.Context)
	}

	prompt := fmt.Sprintf(`# Role
You are an expert native translator. You translate content to %s with the fluency and nuance of a highly educated native speaker.

# Context
%s

# Register
%s

# Task
Translate the provided texts into idiomatic %s.

# Style Guide
- **Natural Flow**: Avoid literal translations. Rephrase sentences to sound completely natural to a native speaker.
- **Vocabulary**: Use precise, culturally relevant terminology. Avoid awkward "translationese" or robotic phrasing.
- **Tone**: Maintain the original intent but adapt the wording to fit the target culture's expectations.
- **Idioms**: Never translate idioms literally. Replace English idioms with natural %s equivalents.
- **HTML/Code Safety**: Do NOT translate HTML tags, class names, IDs, attributes, URLs, email addresses, or content inside backticks or <code> blocks.
- **Interpolation**: Do NOT translate variables or placeholders (e.g., {{name}}, {count}, %%s, $1).
- **Formatting**: Preserve meaningful whitespace (leading/trailing spaces, multiple spaces, newlines). Use idiomatic punctuation for the target language.
- **Context Hints**: If you see {{__ctx__:...}}, use that hint to disambiguate the translation, then REMOVE the hint from your output.`, targetName, contextText, styleDesc, targetName, targetName)

	// Add locale clarification if available
	if localeHint != "" {
		prompt += fmt.Sprintf("\n- **Locale**: %s", localeHint)
	}

	// Add user-provided glossary if available
	if len(req.Glossary) > 0 {
		prompt += "\n\n# Glossary\nWhen you encounter these phrases, prefer these translations (unless context demands otherwise):"
		for source, target := range req.Glossary {
			prompt += fmt.Sprintf("\n- \"%s\" → %s", source, target)
		}
	}

	// Add quality check instruction
	prompt += fmt.Sprintf("\n\n# Quality Check\nAfter translating each string, verify it sounds like native %s and not a calque. If any phrase sounds like a literal translation, rewrite it naturally.", targetName)

	// Add format requirements
	prompt += `

# Format
Return a valid JSON object with a single key "translations" containing an array of strings in the exact same order as the input.
Example: { "translations": ["translated string 1", "translated string 2"] }
- Do NOT wrap in Markdown code blocks.
- Do NOT include any {{__ctx__:...}} markers in your output.`

	// Add exclusions if provided
	if len(req.ExcludedTerms) > 0 {
		terms := strings.Join(req.ExcludedTerms, "\n- ")
		prompt += fmt.Sprintf("\n\n# Exclusions\nDo NOT translate the following terms. Keep them exactly as they appear in the source:\n- %s", terms)
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
