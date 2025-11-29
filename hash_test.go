package gotlai

import "testing"

func TestHashText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple text",
			input:    "Hello World",
			expected: "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e",
		},
		{
			name:     "text with leading whitespace",
			input:    "  Hello World",
			expected: "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e",
		},
		{
			name:     "text with trailing whitespace",
			input:    "Hello World  ",
			expected: "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e",
		},
		{
			name:     "text with both whitespace",
			input:    "  Hello World  ",
			expected: "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e",
		},
		{
			name:  "empty string",
			input: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HashText(tt.input)
			if tt.expected != "" && result != tt.expected {
				t.Errorf("HashText(%q) = %q, want %q", tt.input, result, tt.expected)
			}
			// Verify hash length (SHA-256 = 64 hex chars)
			if len(result) != 64 {
				t.Errorf("HashText(%q) length = %d, want 64", tt.input, len(result))
			}
		})
	}
}

func TestCacheKey(t *testing.T) {
	hash := "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e"
	targetLang := "es_ES"

	result := CacheKey(hash, targetLang)
	expected := "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e:es_ES"

	if result != expected {
		t.Errorf("CacheKey() = %q, want %q", result, expected)
	}
}

func TestCacheKeyExtended(t *testing.T) {
	hash := "abc123"
	sourceLang := "en"
	targetLang := "es_ES"
	model := "gpt-4o-mini"

	result := CacheKeyExtended(hash, sourceLang, targetLang, model)
	expected := "abc123:en:es_ES:gpt-4o-mini"

	if result != expected {
		t.Errorf("CacheKeyExtended() = %q, want %q", result, expected)
	}
}
