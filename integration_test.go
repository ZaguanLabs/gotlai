package gotlai_test

import (
	"context"
	"strings"
	"testing"

	"github.com/ZaguanLabs/gotlai"
	"github.com/ZaguanLabs/gotlai/cache"
	"github.com/ZaguanLabs/gotlai/processor"
	"github.com/ZaguanLabs/gotlai/provider"
)

// Integration tests using all real components

func TestIntegration_BasicTranslation(t *testing.T) {
	p := provider.NewMockProvider()
	c := cache.NewInMemoryCache(3600)
	proc := processor.NewHTMLProcessor()

	translator := gotlai.NewTranslator("es_ES", p,
		gotlai.WithCache(c),
		gotlai.WithProcessor(proc),
	)

	html := `<div><p>Hello</p></div>`
	result, err := translator.ProcessHTML(context.Background(), html)

	if err != nil {
		t.Fatalf("ProcessHTML failed: %v", err)
	}

	if !strings.Contains(result.Content, "Hola") {
		t.Errorf("Expected 'Hola' in result, got: %s", result.Content)
	}

	if result.TranslatedCount != 1 {
		t.Errorf("Expected TranslatedCount 1, got %d", result.TranslatedCount)
	}
}

func TestIntegration_CacheHit(t *testing.T) {
	p := provider.NewMockProvider()
	c := cache.NewInMemoryCache(3600)
	proc := processor.NewHTMLProcessor()

	translator := gotlai.NewTranslator("es_ES", p,
		gotlai.WithCache(c),
		gotlai.WithProcessor(proc),
	)

	html := `<p>Hello</p>`

	// First call
	result1, _ := translator.ProcessHTML(context.Background(), html)
	if result1.TranslatedCount != 1 || result1.CachedCount != 0 {
		t.Errorf("First call: expected 1 translated, 0 cached; got %d, %d",
			result1.TranslatedCount, result1.CachedCount)
	}

	// Second call - should use cache
	result2, _ := translator.ProcessHTML(context.Background(), html)
	if result2.TranslatedCount != 0 || result2.CachedCount != 1 {
		t.Errorf("Second call: expected 0 translated, 1 cached; got %d, %d",
			result2.TranslatedCount, result2.CachedCount)
	}

	// Provider should only be called once
	if p.CallCount != 1 {
		t.Errorf("Provider should be called once, was called %d times", p.CallCount)
	}
}

func TestIntegration_IgnoredTags(t *testing.T) {
	p := provider.NewMockProvider()
	proc := processor.NewHTMLProcessor()

	translator := gotlai.NewTranslator("es_ES", p,
		gotlai.WithProcessor(proc),
	)

	html := `<div>
		<p>Hello</p>
		<script>console.log("Hello");</script>
		<style>.hello { color: red; }</style>
		<code>Hello</code>
	</div>`

	result, err := translator.ProcessHTML(context.Background(), html)
	if err != nil {
		t.Fatalf("ProcessHTML failed: %v", err)
	}

	// Only the <p> content should be translated
	if result.TotalNodes != 1 {
		t.Errorf("Expected 1 translatable node, got %d", result.TotalNodes)
	}

	// Script content should remain unchanged
	if !strings.Contains(result.Content, `console.log("Hello")`) {
		t.Error("Script content should not be translated")
	}
}

func TestIntegration_DataNoTranslate(t *testing.T) {
	p := provider.NewMockProvider()
	proc := processor.NewHTMLProcessor()

	translator := gotlai.NewTranslator("es_ES", p,
		gotlai.WithProcessor(proc),
	)

	html := `<div>
		<p data-no-translate>Hello</p>
		<p>World</p>
	</div>`

	result, err := translator.ProcessHTML(context.Background(), html)
	if err != nil {
		t.Fatalf("ProcessHTML failed: %v", err)
	}

	// Only "World" should be translated
	if result.TotalNodes != 1 {
		t.Errorf("Expected 1 translatable node, got %d", result.TotalNodes)
	}

	// The data-no-translate content should remain
	if !strings.Contains(result.Content, ">Hello<") {
		t.Error("data-no-translate content should not be translated")
	}

	// World should be translated
	if !strings.Contains(result.Content, "Mundo") {
		t.Error("World should be translated to Mundo")
	}
}

func TestIntegration_RTLLanguage(t *testing.T) {
	p := provider.NewMockProvider()
	p.Translations["Hello"] = "مرحبا"
	proc := processor.NewHTMLProcessor()

	translator := gotlai.NewTranslator("ar_SA", p,
		gotlai.WithProcessor(proc),
	)

	html := `<html><body><p>Hello</p></body></html>`
	result, err := translator.ProcessHTML(context.Background(), html)
	if err != nil {
		t.Fatalf("ProcessHTML failed: %v", err)
	}

	if !strings.Contains(result.Content, `dir="rtl"`) {
		t.Errorf("Expected dir='rtl' for Arabic, got: %s", result.Content)
	}

	if !strings.Contains(result.Content, `lang="ar-SA"`) {
		t.Errorf("Expected lang='ar-SA', got: %s", result.Content)
	}
}

func TestIntegration_Deduplication(t *testing.T) {
	p := provider.NewMockProvider()
	proc := processor.NewHTMLProcessor()

	translator := gotlai.NewTranslator("es_ES", p,
		gotlai.WithProcessor(proc),
	)

	// Same text appears 3 times
	html := `<div><p>Hello</p><p>Hello</p><p>Hello</p></div>`
	result, err := translator.ProcessHTML(context.Background(), html)
	if err != nil {
		t.Fatalf("ProcessHTML failed: %v", err)
	}

	// Should only translate once (deduplication)
	if len(p.LastRequest.Texts) != 1 {
		t.Errorf("Expected 1 unique text sent to provider, got %d", len(p.LastRequest.Texts))
	}

	// But all instances should be translated in output
	count := strings.Count(result.Content, "Hola")
	if count != 3 {
		t.Errorf("Expected 3 instances of 'Hola', got %d", count)
	}
}

func TestIntegration_SourceEqualsTarget(t *testing.T) {
	p := provider.NewMockProvider()
	proc := processor.NewHTMLProcessor()

	translator := gotlai.NewTranslator("en_US", p,
		gotlai.WithSourceLang("en"),
		gotlai.WithProcessor(proc),
	)

	html := `<p>Hello</p>`
	result, err := translator.ProcessHTML(context.Background(), html)
	if err != nil {
		t.Fatalf("ProcessHTML failed: %v", err)
	}

	// Should return unchanged
	if result.TranslatedCount != 0 {
		t.Errorf("Expected 0 translations when source==target, got %d", result.TranslatedCount)
	}

	// Provider should not be called
	if p.CallCount != 0 {
		t.Errorf("Provider should not be called when source==target")
	}
}

func TestIntegration_EmptyContent(t *testing.T) {
	p := provider.NewMockProvider()
	proc := processor.NewHTMLProcessor()

	translator := gotlai.NewTranslator("es_ES", p,
		gotlai.WithProcessor(proc),
	)

	html := `<div></div>`
	result, err := translator.ProcessHTML(context.Background(), html)
	if err != nil {
		t.Fatalf("ProcessHTML failed: %v", err)
	}

	if result.TotalNodes != 0 {
		t.Errorf("Expected 0 nodes for empty content, got %d", result.TotalNodes)
	}

	if p.CallCount != 0 {
		t.Error("Provider should not be called for empty content")
	}
}

func TestIntegration_WhitespacePreserved(t *testing.T) {
	p := provider.NewMockProvider()
	proc := processor.NewHTMLProcessor()

	translator := gotlai.NewTranslator("es_ES", p,
		gotlai.WithProcessor(proc),
	)

	html := `<p>  Hello  </p>`
	result, err := translator.ProcessHTML(context.Background(), html)
	if err != nil {
		t.Fatalf("ProcessHTML failed: %v", err)
	}

	// Whitespace should be preserved
	if !strings.Contains(result.Content, "  Hola  ") {
		t.Errorf("Whitespace not preserved, got: %s", result.Content)
	}
}

func TestIntegration_RetryableProvider(t *testing.T) {
	// Create a provider that fails twice then succeeds
	inner := &failingMockProvider{failCount: 2}
	retryable := gotlai.NewRetryableProvider(inner, gotlai.RetryConfig{
		MaxRetries: 3,
		BaseDelay:  1, // 1 nanosecond for fast tests
		MaxDelay:   10,
	})

	proc := processor.NewHTMLProcessor()
	translator := gotlai.NewTranslator("es_ES", retryable,
		gotlai.WithProcessor(proc),
	)

	html := `<p>Hello</p>`
	result, err := translator.ProcessHTML(context.Background(), html)
	if err != nil {
		t.Fatalf("ProcessHTML failed after retries: %v", err)
	}

	if !strings.Contains(result.Content, "translated") {
		t.Errorf("Expected translated content, got: %s", result.Content)
	}

	if inner.callCount != 3 {
		t.Errorf("Expected 3 calls (2 failures + 1 success), got %d", inner.callCount)
	}
}

// Helper: failing provider for retry tests
type failingMockProvider struct {
	failCount int
	callCount int
}

func (p *failingMockProvider) Translate(ctx context.Context, req gotlai.TranslateRequest) ([]string, error) {
	p.callCount++
	if p.callCount <= p.failCount {
		return nil, &gotlai.ProviderError{Message: "temporary failure", Retryable: true}
	}
	results := make([]string, len(req.Texts))
	for i := range req.Texts {
		results[i] = "translated"
	}
	return results, nil
}
