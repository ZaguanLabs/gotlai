// Package gotlai provides an AI-powered HTML translation engine.
package gotlai

// TextNode represents a translatable unit of content.
type TextNode struct {
	ID       string            // Unique identifier (UUID)
	Text     string            // Original text content (trimmed)
	Hash     string            // SHA-256 hash of Text
	NodeType string            // Content type: "html_text", "go_comment", etc.
	Context  string            // Disambiguation context for AI
	Metadata map[string]string // Additional info (parent tag, line number, etc.)
}

// TranslationConfig holds configuration for the translator.
type TranslationConfig struct {
	TargetLang    string   // Target language code (e.g., "es_ES", "ja_JP")
	SourceLang    string   // Source language code (default: "en")
	ExcludedTerms []string // Terms to never translate (e.g., ["API", "SDK"])
	Context       string   // Global context for all translations
}

// ProcessedContent is the result of a translation operation.
type ProcessedContent struct {
	Content         string // Translated content
	TranslatedCount int    // Number of newly translated items
	CachedCount     int    // Number of cache hits
	TotalNodes      int    // Total translatable nodes found
}

// RTLLanguages contains language codes that use right-to-left text direction.
var RTLLanguages = map[string]bool{
	"ar": true, // Arabic
	"he": true, // Hebrew
	"fa": true, // Persian/Farsi
	"ur": true, // Urdu
	"ps": true, // Pashto
	"sd": true, // Sindhi
	"ug": true, // Uyghur
}

// IgnoredTags contains HTML tags whose content should not be translated.
var IgnoredTags = map[string]bool{
	"script":   true,
	"style":    true,
	"code":     true,
	"pre":      true,
	"textarea": true,
	"noscript": true,
}
