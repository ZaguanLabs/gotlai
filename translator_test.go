package gotlai

import (
	"context"
	"strings"
	"testing"
)

// mockProvider is a simple mock for testing
type mockProvider struct {
	translations map[string]string
	callCount    int
	lastTexts    []string
}

func newMockProvider() *mockProvider {
	return &mockProvider{
		translations: map[string]string{
			"Hello":                "Hola",
			"World":                "Mundo",
			"Hello World":          "Hola Mundo",
			"Welcome to our site.": "Bienvenido a nuestro sitio.",
			"Translate me":         "Tradúceme",
			"Translate this":       "Traduce esto",
		},
	}
}

func (m *mockProvider) Translate(ctx context.Context, req TranslateRequest) ([]string, error) {
	m.callCount++
	m.lastTexts = req.Texts

	results := make([]string, len(req.Texts))
	for i, text := range req.Texts {
		if translation, ok := m.translations[text]; ok {
			results[i] = translation
		} else {
			results[i] = "[" + text + "]"
		}
	}
	return results, nil
}

// mockCache is a simple mock cache for testing
type mockCache struct {
	data map[string]string
}

func newMockCache() *mockCache {
	return &mockCache{data: make(map[string]string)}
}

func (c *mockCache) Get(key string) (string, bool) {
	val, ok := c.data[key]
	return val, ok
}

func (c *mockCache) Set(key string, value string) error {
	c.data[key] = value
	return nil
}

// mockHTMLProcessor is a simple HTML processor for testing
type mockHTMLProcessor struct{}

func (p *mockHTMLProcessor) Extract(content string) (interface{}, []TextNode, error) {
	// Simple extraction: find text between > and <
	var nodes []TextNode
	seenHashes := make(map[string]bool)

	// Very simple parsing for testing
	parts := strings.Split(content, ">")
	for _, part := range parts {
		idx := strings.Index(part, "<")
		if idx > 0 {
			text := strings.TrimSpace(part[:idx])
			if text != "" {
				hash := HashText(text)
				if !seenHashes[hash] {
					seenHashes[hash] = true
					nodes = append(nodes, TextNode{
						ID:       hash[:8],
						Text:     text,
						Hash:     hash,
						NodeType: "html_text",
					})
				}
			}
		}
	}

	return content, nodes, nil
}

func (p *mockHTMLProcessor) Apply(parsed interface{}, nodes []TextNode, translations map[string]string) (string, error) {
	result := parsed.(string)
	for _, node := range nodes {
		if translated, ok := translations[node.Hash]; ok {
			result = strings.ReplaceAll(result, ">"+node.Text+"<", ">"+translated+"<")
		}
	}
	return result, nil
}

func (p *mockHTMLProcessor) ContentType() string {
	return "html"
}

func TestTranslator_BasicTranslation(t *testing.T) {
	provider := newMockProvider()
	processor := &mockHTMLProcessor{}

	translator := NewTranslator("es_ES", provider,
		WithProcessor(processor),
	)

	result, err := translator.Process(context.Background(), "<p>Hello</p>", "html")
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if !strings.Contains(result.Content, "Hola") {
		t.Errorf("Result should contain 'Hola', got: %s", result.Content)
	}

	if result.TranslatedCount != 1 {
		t.Errorf("Expected TranslatedCount 1, got %d", result.TranslatedCount)
	}
}

func TestTranslator_CacheHit(t *testing.T) {
	provider := newMockProvider()
	cache := newMockCache()
	processor := &mockHTMLProcessor{}

	translator := NewTranslator("es_ES", provider,
		WithCache(cache),
		WithProcessor(processor),
	)

	// First call - should translate
	result1, err := translator.Process(context.Background(), "<p>Hello</p>", "html")
	if err != nil {
		t.Fatalf("First Process failed: %v", err)
	}

	if result1.TranslatedCount != 1 {
		t.Errorf("First call: expected TranslatedCount 1, got %d", result1.TranslatedCount)
	}
	if result1.CachedCount != 0 {
		t.Errorf("First call: expected CachedCount 0, got %d", result1.CachedCount)
	}

	// Second call - should use cache
	result2, err := translator.Process(context.Background(), "<p>Hello</p>", "html")
	if err != nil {
		t.Fatalf("Second Process failed: %v", err)
	}

	if result2.CachedCount != 1 {
		t.Errorf("Second call: expected CachedCount 1, got %d", result2.CachedCount)
	}
	if result2.TranslatedCount != 0 {
		t.Errorf("Second call: expected TranslatedCount 0, got %d", result2.TranslatedCount)
	}

	// Provider should only be called once
	if provider.callCount != 1 {
		t.Errorf("Provider should be called once, was called %d times", provider.callCount)
	}
}

func TestTranslator_SourceEqualsTarget(t *testing.T) {
	provider := newMockProvider()
	processor := &mockHTMLProcessor{}

	translator := NewTranslator("en_US", provider,
		WithSourceLang("en"),
		WithProcessor(processor),
	)

	result, err := translator.Process(context.Background(), "<p>Hello</p>", "html")
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should return unchanged content
	if result.TranslatedCount != 0 {
		t.Errorf("Expected TranslatedCount 0 when source==target, got %d", result.TranslatedCount)
	}

	// Provider should not be called
	if provider.callCount != 0 {
		t.Errorf("Provider should not be called when source==target, was called %d times", provider.callCount)
	}
}

func TestTranslator_Deduplication(t *testing.T) {
	provider := newMockProvider()
	processor := &mockHTMLProcessor{}

	translator := NewTranslator("es_ES", provider,
		WithProcessor(processor),
	)

	// Three identical texts
	result, err := translator.Process(context.Background(), "<p>Hello</p><p>Hello</p><p>Hello</p>", "html")
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should only translate once (deduplication)
	if len(provider.lastTexts) != 1 {
		t.Errorf("Provider should receive 1 unique text, got %d", len(provider.lastTexts))
	}

	if result.TranslatedCount != 1 {
		t.Errorf("Expected TranslatedCount 1, got %d", result.TranslatedCount)
	}
}

func TestTranslator_EmptyContent(t *testing.T) {
	provider := newMockProvider()
	processor := &mockHTMLProcessor{}

	translator := NewTranslator("es_ES", provider,
		WithProcessor(processor),
	)

	result, err := translator.Process(context.Background(), "<div></div>", "html")
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if result.TotalNodes != 0 {
		t.Errorf("Expected TotalNodes 0 for empty content, got %d", result.TotalNodes)
	}

	if provider.callCount != 0 {
		t.Errorf("Provider should not be called for empty content")
	}
}

func TestTranslator_NoProcessor(t *testing.T) {
	provider := newMockProvider()

	translator := NewTranslator("es_ES", provider)

	_, err := translator.Process(context.Background(), "<p>Hello</p>", "html")
	if err == nil {
		t.Error("Expected error when no processor registered")
	}

	var procErr *ProcessorError
	if _, ok := err.(*ProcessorError); !ok {
		t.Errorf("Expected ProcessorError, got %T", err)
	}
	_ = procErr
}

func TestTranslator_RTLLanguage(t *testing.T) {
	provider := newMockProvider()
	provider.translations["Hello"] = "مرحبا"

	processor := &mockHTMLProcessor{}

	translator := NewTranslator("ar_SA", provider,
		WithProcessor(processor),
	)

	result, err := translator.Process(context.Background(), "<html><body><p>Hello</p></body></html>", "html")
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if !strings.Contains(result.Content, `dir="rtl"`) {
		t.Errorf("Result should contain dir='rtl' for Arabic, got: %s", result.Content)
	}

	if !strings.Contains(result.Content, `lang="ar-SA"`) {
		t.Errorf("Result should contain lang='ar-SA', got: %s", result.Content)
	}
}

func TestTranslator_Options(t *testing.T) {
	provider := newMockProvider()
	cache := newMockCache()

	translator := NewTranslator("es_ES", provider,
		WithSourceLang("en_US"),
		WithCache(cache),
		WithExcludedTerms([]string{"API", "SDK"}),
		WithContext("Technical documentation"),
	)

	if translator.SourceLang() != "en_US" {
		t.Errorf("Expected source lang 'en_US', got %q", translator.SourceLang())
	}

	if translator.TargetLang() != "es_ES" {
		t.Errorf("Expected target lang 'es_ES', got %q", translator.TargetLang())
	}
}

func TestTranslator_IsSourceLang(t *testing.T) {
	tests := []struct {
		source   string
		target   string
		expected bool
	}{
		{"en", "en_US", true},
		{"en_US", "en_GB", true},
		{"en", "es_ES", false},
		{"en_US", "es_MX", false},
	}

	for _, tt := range tests {
		provider := newMockProvider()
		translator := NewTranslator(tt.target, provider, WithSourceLang(tt.source))

		result := translator.isSourceLang()
		if result != tt.expected {
			t.Errorf("isSourceLang() for source=%q, target=%q: got %v, want %v",
				tt.source, tt.target, result, tt.expected)
		}
	}
}
