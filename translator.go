package gotlai

import (
	"context"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Translator is the main translation engine.
type Translator struct {
	targetLang    string
	sourceLang    string
	provider      AIProvider
	cache         TranslationCache
	excludedTerms []string
	context       string
	glossary      map[string]string
	style         TranslationStyle
	processors    map[string]ContentProcessor
}

// AIProvider is the interface for AI translation backends.
type AIProvider interface {
	Translate(ctx context.Context, req TranslateRequest) ([]string, error)
}

// TranslateRequest contains the parameters for a translation request.
type TranslateRequest struct {
	Texts         []string
	TargetLang    string
	SourceLang    string
	ExcludedTerms []string
	Context       string
	TextContexts  []string
	Glossary      map[string]string
	Style         TranslationStyle
}

// TranslationCache is the interface for translation caching.
type TranslationCache interface {
	Get(key string) (string, bool)
	Set(key string, value string) error
}

// ContentProcessor is the interface for content processing.
type ContentProcessor interface {
	Extract(content string) (interface{}, []TextNode, error)
	Apply(parsed interface{}, nodes []TextNode, translations map[string]string) (string, error)
	ContentType() string
}

// TranslatorOption is a functional option for configuring the Translator.
type TranslatorOption func(*Translator)

// WithSourceLang sets the source language.
func WithSourceLang(lang string) TranslatorOption {
	return func(t *Translator) {
		t.sourceLang = lang
	}
}

// WithCache sets the translation cache.
func WithCache(cache TranslationCache) TranslatorOption {
	return func(t *Translator) {
		t.cache = cache
	}
}

// WithExcludedTerms sets terms that should not be translated.
func WithExcludedTerms(terms []string) TranslatorOption {
	return func(t *Translator) {
		t.excludedTerms = terms
	}
}

// WithContext sets the global translation context.
func WithContext(ctx string) TranslatorOption {
	return func(t *Translator) {
		t.context = ctx
	}
}

// WithGlossary sets preferred translations for specific phrases.
func WithGlossary(glossary map[string]string) TranslatorOption {
	return func(t *Translator) {
		t.glossary = glossary
	}
}

// WithStyle sets the translation style/register.
func WithStyle(style TranslationStyle) TranslatorOption {
	return func(t *Translator) {
		t.style = style
	}
}

// WithProcessor registers a content processor.
func WithProcessor(processor ContentProcessor) TranslatorOption {
	return func(t *Translator) {
		t.processors[processor.ContentType()] = processor
	}
}

// NewTranslator creates a new Translator with the given target language and provider.
func NewTranslator(targetLang string, provider AIProvider, opts ...TranslatorOption) *Translator {
	t := &Translator{
		targetLang: targetLang,
		sourceLang: "en",
		provider:   provider,
		style:      StyleNeutral,
		processors: make(map[string]ContentProcessor),
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// Process translates content of the specified type.
func (t *Translator) Process(ctx context.Context, content string, contentType string) (*ProcessedContent, error) {
	// Skip if source == target
	if t.isSourceLang() {
		return &ProcessedContent{
			Content:         content,
			TranslatedCount: 0,
			CachedCount:     0,
			TotalNodes:      0,
		}, nil
	}

	// Get processor
	processor, ok := t.processors[contentType]
	if !ok {
		return nil, &ProcessorError{
			Message:     "no processor registered for content type",
			ContentType: contentType,
		}
	}

	// Extract text nodes
	parsed, nodes, err := processor.Extract(content)
	if err != nil {
		return nil, err
	}

	if len(nodes) == 0 {
		return &ProcessedContent{
			Content:         content,
			TranslatedCount: 0,
			CachedCount:     0,
			TotalNodes:      0,
		}, nil
	}

	// Translate batch
	translations, cachedCount, translatedCount, err := t.translateBatch(ctx, nodes)
	if err != nil {
		return nil, err
	}

	// Apply translations
	result, err := processor.Apply(parsed, nodes, translations)
	if err != nil {
		return nil, err
	}

	// Set HTML attributes if applicable
	if contentType == "html" {
		result = t.setHTMLAttributes(result)
	}

	return &ProcessedContent{
		Content:         result,
		TranslatedCount: translatedCount,
		CachedCount:     cachedCount,
		TotalNodes:      len(nodes),
	}, nil
}

// ProcessHTML is a convenience method for processing HTML content.
func (t *Translator) ProcessHTML(ctx context.Context, html string) (*ProcessedContent, error) {
	return t.Process(ctx, html, "html")
}

// translateBatch translates nodes, using cache where possible.
func (t *Translator) translateBatch(ctx context.Context, nodes []TextNode) (map[string]string, int, int, error) {
	translations := make(map[string]string)
	var cacheMisses []TextNode
	seenHashes := make(map[string]bool)
	cachedCount := 0

	// Check cache for each node
	for _, node := range nodes {
		cacheKey := CacheKey(node.Hash, t.targetLang)

		if t.cache != nil {
			if cached, ok := t.cache.Get(cacheKey); ok {
				translations[node.Hash] = cached
				cachedCount++
				continue
			}
		}

		// Deduplicate cache misses
		if !seenHashes[node.Hash] {
			cacheMisses = append(cacheMisses, node)
			seenHashes[node.Hash] = true
		}
	}

	// Translate cache misses via AI
	translatedCount := 0
	if len(cacheMisses) > 0 && t.provider != nil {
		texts := make([]string, len(cacheMisses))
		textContexts := make([]string, len(cacheMisses))
		for i, node := range cacheMisses {
			texts[i] = node.Text
			textContexts[i] = node.Context
		}

		results, err := t.provider.Translate(ctx, TranslateRequest{
			Texts:         texts,
			TargetLang:    t.targetLang,
			SourceLang:    t.sourceLang,
			ExcludedTerms: t.excludedTerms,
			Context:       t.context,
			TextContexts:  textContexts,
			Glossary:      t.glossary,
			Style:         t.style,
		})
		if err != nil {
			return nil, 0, 0, err
		}

		// Cache and store results
		for i, node := range cacheMisses {
			translations[node.Hash] = results[i]
			if t.cache != nil {
				cacheKey := CacheKey(node.Hash, t.targetLang)
				_ = t.cache.Set(cacheKey, results[i]) // Ignore cache set errors
			}
			translatedCount++
		}
	}

	return translations, cachedCount, translatedCount, nil
}

// isSourceLang checks if target matches source (no translation needed).
func (t *Translator) isSourceLang() bool {
	target := strings.Split(t.targetLang, "_")[0]
	target = strings.ToLower(target)

	source := strings.Split(t.sourceLang, "_")[0]
	source = strings.ToLower(source)

	return target == source
}

// setHTMLAttributes sets lang and dir attributes on the <html> tag.
func (t *Translator) setHTMLAttributes(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return html
	}

	htmlTag := doc.Find("html")
	if htmlTag.Length() > 0 {
		htmlTag.SetAttr("lang", ToHTMLLang(t.targetLang))
		htmlTag.SetAttr("dir", GetDirection(t.targetLang))
	}

	result, err := doc.Html()
	if err != nil {
		return html
	}

	return result
}

// TargetLang returns the target language.
func (t *Translator) TargetLang() string {
	return t.targetLang
}

// SourceLang returns the source language.
func (t *Translator) SourceLang() string {
	return t.sourceLang
}

// IsSourceLang checks if the target language matches the source language.
// When true, translation can be bypassed.
func (t *Translator) IsSourceLang(targetLangOverride ...string) bool {
	targetLang := t.targetLang
	if len(targetLangOverride) > 0 && targetLangOverride[0] != "" {
		targetLang = targetLangOverride[0]
	}
	return t.isSourceLang() || normalizeBaseLang(targetLang) == normalizeBaseLang(t.sourceLang)
}

// IsRTL returns true if the target language uses right-to-left text direction.
func (t *Translator) IsRTL(targetLangOverride ...string) bool {
	targetLang := t.targetLang
	if len(targetLangOverride) > 0 && targetLangOverride[0] != "" {
		targetLang = targetLangOverride[0]
	}
	return IsRTL(targetLang)
}

// GetDir returns the text direction for the target language ("ltr" or "rtl").
func (t *Translator) GetDir(targetLangOverride ...string) string {
	targetLang := t.targetLang
	if len(targetLangOverride) > 0 && targetLangOverride[0] != "" {
		targetLang = targetLangOverride[0]
	}
	return GetDirection(targetLang)
}

// Glossary returns the glossary of preferred translations.
func (t *Translator) Glossary() map[string]string {
	return t.glossary
}

// Style returns the translation style.
func (t *Translator) Style() TranslationStyle {
	return t.style
}

// Context returns the global translation context.
func (t *Translator) Context() string {
	return t.context
}

// ExcludedTerms returns the list of excluded terms.
func (t *Translator) ExcludedTerms() []string {
	return t.excludedTerms
}

// normalizeBaseLang extracts the base language code (e.g., "en" from "en_US").
func normalizeBaseLang(lang string) string {
	parts := strings.Split(lang, "_")
	if len(parts) > 0 {
		return strings.ToLower(parts[0])
	}
	return strings.ToLower(lang)
}
