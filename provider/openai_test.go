package provider

import (
	"context"
	"strings"
	"testing"
)

func TestBuildSystemPrompt(t *testing.T) {
	p := NewOpenAIProvider(OpenAIConfig{APIKey: "test"})

	req := TranslateRequest{
		TargetLang:    "es_ES",
		SourceLang:    "en",
		Context:       "E-commerce website",
		ExcludedTerms: []string{"API", "SDK"},
	}

	prompt := p.buildSystemPrompt(req)

	// Check key elements are present
	if !strings.Contains(prompt, "Spanish (Spain)") {
		t.Error("Prompt should contain target language name")
	}
	if !strings.Contains(prompt, "E-commerce website") {
		t.Error("Prompt should contain context")
	}
	if !strings.Contains(prompt, "API") || !strings.Contains(prompt, "SDK") {
		t.Error("Prompt should contain excluded terms")
	}
	if !strings.Contains(prompt, "Castilian Spanish") {
		t.Error("Prompt should contain locale clarification for es_ES")
	}
}

func TestBuildSystemPrompt_WithGlossaryAndStyle(t *testing.T) {
	p := NewOpenAIProvider(OpenAIConfig{APIKey: "test"})

	req := TranslateRequest{
		TargetLang: "nb_NO",
		SourceLang: "en",
		Glossary: map[string]string{
			"on the fly":   "fortløpende",
			"cutting-edge": "banebrytende",
		},
		Style: "marketing",
	}

	prompt := p.buildSystemPrompt(req)

	// Check glossary is included
	if !strings.Contains(prompt, "on the fly") {
		t.Error("Prompt should contain glossary source term")
	}
	if !strings.Contains(prompt, "fortløpende") {
		t.Error("Prompt should contain glossary target term")
	}

	// Check style description is included
	if !strings.Contains(prompt, "persuasive") {
		t.Error("Prompt should contain marketing style description")
	}

	// Check locale clarification for Norwegian
	if !strings.Contains(prompt, "Bokmål") {
		t.Error("Prompt should contain Norwegian locale clarification")
	}
}

func TestBuildUserMessage_SimpleArray(t *testing.T) {
	p := NewOpenAIProvider(OpenAIConfig{APIKey: "test"})

	req := TranslateRequest{
		Texts: []string{"Hello", "World"},
	}

	msg := p.buildUserMessage(req)

	if msg != `["Hello","World"]` {
		t.Errorf("Expected JSON array, got: %s", msg)
	}
}

func TestBuildUserMessage_WithContexts(t *testing.T) {
	p := NewOpenAIProvider(OpenAIConfig{APIKey: "test"})

	req := TranslateRequest{
		Texts:        []string{"Run", "Save"},
		TextContexts: []string{"in <button>", "in file dialog"},
	}

	msg := p.buildUserMessage(req)

	if !strings.Contains(msg, `"text":"Run"`) {
		t.Errorf("Message should contain text field, got: %s", msg)
	}
	if !strings.Contains(msg, `"context":"in \u003cbutton\u003e"`) && !strings.Contains(msg, `"context":"in <button>"`) {
		t.Errorf("Message should contain context field, got: %s", msg)
	}
}

func TestParseResponse_TranslationsKey(t *testing.T) {
	p := NewOpenAIProvider(OpenAIConfig{APIKey: "test"})

	content := `{"translations": ["Hola", "Mundo"]}`
	result, err := p.parseResponse(content, 2)

	if err != nil {
		t.Fatalf("parseResponse failed: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 translations, got %d", len(result))
	}

	if result[0] != "Hola" || result[1] != "Mundo" {
		t.Errorf("Unexpected translations: %v", result)
	}
}

func TestParseResponse_DirectArray(t *testing.T) {
	p := NewOpenAIProvider(OpenAIConfig{APIKey: "test"})

	content := `["Hola", "Mundo"]`
	result, err := p.parseResponse(content, 2)

	if err != nil {
		t.Fatalf("parseResponse failed: %v", err)
	}

	if result[0] != "Hola" || result[1] != "Mundo" {
		t.Errorf("Unexpected translations: %v", result)
	}
}

func TestParseResponse_FallbackArrayKey(t *testing.T) {
	p := NewOpenAIProvider(OpenAIConfig{APIKey: "test"})

	// Some models return with a different key
	content := `{"results": ["Hola", "Mundo"]}`
	result, err := p.parseResponse(content, 2)

	if err != nil {
		t.Fatalf("parseResponse failed: %v", err)
	}

	if result[0] != "Hola" || result[1] != "Mundo" {
		t.Errorf("Unexpected translations: %v", result)
	}
}

func TestParseResponse_CountMismatch(t *testing.T) {
	p := NewOpenAIProvider(OpenAIConfig{APIKey: "test"})

	content := `{"translations": ["Hola"]}`
	_, err := p.parseResponse(content, 2)

	if err == nil {
		t.Error("Expected error for count mismatch")
	}
}

func TestMockProvider(t *testing.T) {
	m := NewMockProvider()

	req := TranslateRequest{
		Texts:      []string{"Hello", "Unknown text"},
		TargetLang: "es_ES",
	}

	result, err := m.Translate(context.Background(), req)
	if err != nil {
		t.Fatalf("MockProvider.Translate failed: %v", err)
	}

	if result[0] != "Hola" {
		t.Errorf("Expected 'Hola', got %q", result[0])
	}

	if result[1] != "[Unknown text]" {
		t.Errorf("Expected '[Unknown text]', got %q", result[1])
	}

	if m.CallCount != 1 {
		t.Errorf("Expected CallCount 1, got %d", m.CallCount)
	}
}
