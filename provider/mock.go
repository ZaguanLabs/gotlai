package provider

import (
	"context"
	"fmt"
)

// MockProvider is a mock AI provider for testing.
type MockProvider struct {
	Translations map[string]string // Map of source text to translation
	CallCount    int               // Number of times Translate was called
	LastRequest  *TranslateRequest // Last request received
}

// NewMockProvider creates a new mock provider with default translations.
func NewMockProvider() *MockProvider {
	return &MockProvider{
		Translations: map[string]string{
			"Hello":                "Hola",
			"World":                "Mundo",
			"Hello World":          "Hola Mundo",
			"Welcome to our site.": "Bienvenido a nuestro sitio.",
		},
	}
}

// Translate returns mock translations.
func (m *MockProvider) Translate(ctx context.Context, req TranslateRequest) ([]string, error) {
	m.CallCount++
	m.LastRequest = &req

	results := make([]string, len(req.Texts))
	for i, text := range req.Texts {
		if translation, ok := m.Translations[text]; ok {
			results[i] = translation
		} else {
			// Return bracketed text for unknown translations
			results[i] = fmt.Sprintf("[%s]", text)
		}
	}

	return results, nil
}

// Reset resets the call count and last request.
func (m *MockProvider) Reset() {
	m.CallCount = 0
	m.LastRequest = nil
}

// Verify MockProvider implements AIProvider
var _ AIProvider = (*MockProvider)(nil)
