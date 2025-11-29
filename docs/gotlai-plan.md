# Gotlai Implementation Plan

Go Translation AI - A Go implementation of the translation engine based on Tstlai (TypeScript) and lessons learned from Pytlai (Python).

---

## Project Overview

| Aspect | Value |
|--------|-------|
| Language | Go 1.21+ |
| HTML Parser | `golang.org/x/net/html` + `goquery` |
| HTTP Client | `net/http` / OpenAI Go SDK |
| Cache | `sync.Map` / `go-redis/redis` |
| Test Framework | `testing` + `testify` |
| Package Manager | Go modules |

**Target**: Full feature parity with Pytlai, leveraging Go's concurrency strengths.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                           Gotlai Core                               │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────────────┐ │
│  │   Content    │────▶│    Text      │────▶│   Translation        │ │
│  │   Input      │     │  Extraction  │     │   Pipeline           │ │
│  └──────────────┘     └──────────────┘     └──────────────────────┘ │
│                              │                       │              │
│                              ▼                       ▼              │
│                       ┌──────────────┐     ┌──────────────────────┐ │
│                       │   TextNode   │     │   Cache Layer        │ │
│                       │   + Context  │     │   (goroutine-safe)   │ │
│                       └──────────────┘     └──────────────────────┘ │
│                                             ┌────────┴────────┐     │
│                                        cache hit         cache miss │
│                                             │                 │     │
│                                             ▼                 ▼     │
│                                    ┌──────────────┐  ┌─────────────┐│
│                                    │   Return     │  │ AI Provider ││
│                                    │   Cached     │  │ (batch call)││
│                                    └──────────────┘  └─────────────┘│
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Key Decisions (Based on Porting Report)

### 1. Address Known Pitfalls Upfront

| Issue from Pytlai | Gotlai Solution |
|-------------------|-----------------|
| Response format ambiguity | Standardize on `{"translations": [...]}`, robust fallback parsing |
| Whitespace preservation | Store leading/trailing whitespace, reapply after translation |
| Error handling gaps | Define custom error types, implement retry with backoff |
| Cache key collisions | Support extended key format: `{hash}:{source}:{target}:{model}` |
| Context disambiguation | Include `Context` field in `TextNode` from day one |
| Empty cache truthiness | Use explicit nil checks (Go's zero values handle this naturally) |

### 2. Go-Specific Advantages

- **Concurrency**: Use goroutines for parallel cache lookups
- **Type safety**: Strong typing catches errors at compile time
- **Performance**: Native compilation, no runtime overhead
- **Deployment**: Single binary, easy distribution

### 3. Go-Specific Gotchas (from improved guide)

- HTML parser (`golang.org/x/net/html`) doesn't preserve whitespace well → use `goquery` for better handling
- Need explicit mutex/sync for concurrent cache access
- Error handling via return values, not exceptions

---

## Implementation Phases

### Phase 1: Core Data Structures

**Files**: `types.go`, `errors.go`

```go
// TextNode - fundamental unit of translatable content
type TextNode struct {
    ID       string            // UUID
    Text     string            // Original text (trimmed)
    Hash     string            // SHA-256 of Text
    NodeType string            // "html_text", "go_comment", etc.
    Context  string            // Disambiguation context
    Metadata map[string]string // Additional info
}

// TranslationConfig
type TranslationConfig struct {
    TargetLang    string
    SourceLang    string   // Default: "en"
    ExcludedTerms []string
    Context       string   // Global context
}

// ProcessedContent - return value
type ProcessedContent struct {
    Content         string
    TranslatedCount int
    CachedCount     int
    TotalNodes      int
}

// Custom errors
type TranslationError struct { ... }
type ProviderError struct { ... }
type CacheError struct { ... }
type ProcessorError struct { ... }
```

**Checklist**:
- [ ] Define all core types
- [ ] Implement custom error types with `Error()` method
- [ ] Add helper functions (hash generation, cache key formatting)

---

### Phase 2: Cache Layer

**Files**: `cache/cache.go`, `cache/memory.go`, `cache/redis.go`

```go
// TranslationCache interface
type TranslationCache interface {
    Get(key string) (string, bool)
    Set(key string, value string) error
}

// InMemoryCache with TTL and mutex
type InMemoryCache struct {
    cache      map[string]cacheEntry
    mu         sync.RWMutex
    ttl        time.Duration
}

// RedisCache
type RedisCache struct {
    client    *redis.Client
    ttl       time.Duration
    keyPrefix string
}
```

**Cache Key Format**:
```go
func CacheKey(hash, targetLang string) string {
    return hash + ":" + targetLang
}

// Extended format (optional)
func CacheKeyExtended(hash, sourceLang, targetLang, model string) string {
    return fmt.Sprintf("%s:%s:%s:%s", hash, sourceLang, targetLang, model)
}
```

**Checklist**:
- [ ] Define `TranslationCache` interface
- [ ] Implement `InMemoryCache` with TTL and thread-safety
- [ ] Implement `RedisCache` with connection pooling
- [ ] Add cache key helper functions
- [ ] Write unit tests for both implementations

---

### Phase 3: AI Provider

**Files**: `provider/provider.go`, `provider/openai.go`

```go
// AIProvider interface
type AIProvider interface {
    Translate(ctx context.Context, req TranslateRequest) ([]string, error)
}

type TranslateRequest struct {
    Texts         []string
    TargetLang    string
    SourceLang    string
    ExcludedTerms []string
    Context       string
    TextContexts  []string // Per-text disambiguation
}

// OpenAIProvider
type OpenAIProvider struct {
    client      *openai.Client
    model       string
    temperature float32
}
```

**System Prompt** (from improved guide):
```
# Role
You are an expert native translator. Translate from {source} to {target}.

# Context
{context or "General web or application content."}

# Style Guide
- Natural flow: Avoid literal translations
- Vocabulary: Use culturally relevant terminology
- Disambiguation: Use provided context to choose correct translation
- HTML Safety: Do NOT translate tags, classes, IDs
- Variables: Do NOT translate {{name}}, {count}, %s, etc.

# Output Format
Return ONLY: {"translations": ["translated1", "translated2"]}
```

**Response Parsing** (handle both formats):
```go
func parseResponse(content string, expectedCount int) ([]string, error) {
    var result map[string]interface{}
    if err := json.Unmarshal([]byte(content), &result); err != nil {
        // Try direct array
        var arr []string
        if err := json.Unmarshal([]byte(content), &arr); err != nil {
            return nil, fmt.Errorf("invalid response format")
        }
        return arr, nil
    }
    
    // Look for "translations" key first
    if translations, ok := result["translations"].([]interface{}); ok {
        return toStringSlice(translations), nil
    }
    
    // Fallback: first array value
    for _, v := range result {
        if arr, ok := v.([]interface{}); ok {
            return toStringSlice(arr), nil
        }
    }
    
    return nil, fmt.Errorf("no translations array in response")
}
```

**Checklist**:
- [ ] Define `AIProvider` interface
- [ ] Implement `OpenAIProvider` with configurable model/temperature
- [ ] Build system prompt with exclusions and context
- [ ] Implement robust response parsing
- [ ] Add retry logic with exponential backoff
- [ ] Write unit tests with mock provider

---

### Phase 4: Content Processor

**Files**: `processor/processor.go`, `processor/html.go`

```go
// ContentProcessor interface
type ContentProcessor interface {
    Extract(content string) (interface{}, []TextNode, error)
    Apply(parsed interface{}, nodes []TextNode, translations map[string]string) (string, error)
    ContentType() string
}

// HTMLProcessor
type HTMLProcessor struct {
    ignoredTags map[string]bool
}

var defaultIgnoredTags = map[string]bool{
    "script": true, "style": true, "code": true,
    "pre": true, "textarea": true, "noscript": true,
}
```

**Context Building** (for disambiguation):
```go
func (p *HTMLProcessor) buildContext(node *goquery.Selection) string {
    var parts []string
    
    parent := node.Parent()
    if parent.Length() > 0 {
        tag := goquery.NodeName(parent)
        if class, exists := parent.Attr("class"); exists {
            parts = append(parts, fmt.Sprintf("in <%s class=\"%s\">", tag, class))
        } else {
            parts = append(parts, fmt.Sprintf("in <%s>", tag))
        }
    }
    
    // Add sibling context, ancestor path...
    return strings.Join(parts, " | ")
}
```

**Whitespace Preservation** (lesson from Pytlai):
```go
func preserveWhitespace(original, translated string) string {
    leading := len(original) - len(strings.TrimLeft(original, " \t\n"))
    trailing := len(original) - len(strings.TrimRight(original, " \t\n"))
    
    result := original[:leading] + translated
    if trailing > 0 {
        result += original[len(original)-trailing:]
    }
    return result
}
```

**Checklist**:
- [ ] Define `ContentProcessor` interface
- [ ] Implement `HTMLProcessor` with goquery
- [ ] Handle ignored tags (`script`, `style`, `code`, `pre`, `textarea`)
- [ ] Handle `data-no-translate` attribute
- [ ] Implement context building for disambiguation
- [ ] Preserve whitespace in translations
- [ ] Set `lang` and `dir` attributes on `<html>` tag
- [ ] Write comprehensive tests

---

### Phase 5: Main Translator

**Files**: `translator.go`

```go
type Translator struct {
    targetLang    string
    sourceLang    string
    provider      AIProvider
    cache         TranslationCache
    excludedTerms []string
    context       string
    processors    map[string]ContentProcessor
}

var rtlLanguages = map[string]bool{
    "ar": true, "he": true, "fa": true,
    "ur": true, "ps": true, "sd": true, "ug": true,
}

func (t *Translator) Process(content string, contentType string) (*ProcessedContent, error) {
    // 1. Skip if source == target
    if t.isSourceLang() {
        return &ProcessedContent{Content: content}, nil
    }
    
    // 2. Get processor
    processor := t.processors[contentType]
    
    // 3. Extract text nodes
    parsed, nodes, err := processor.Extract(content)
    
    // 4. Translate batch (with cache)
    translations, cached, translated := t.translateBatch(nodes)
    
    // 5. Apply translations
    result, err := processor.Apply(parsed, nodes, translations)
    
    // 6. Set HTML attributes if applicable
    if contentType == "html" {
        result = t.setHTMLAttributes(result)
    }
    
    return &ProcessedContent{
        Content:         result,
        TranslatedCount: translated,
        CachedCount:     cached,
        TotalNodes:      len(nodes),
    }, nil
}
```

**Batch Translation with Deduplication**:
```go
func (t *Translator) translateBatch(nodes []TextNode) (map[string]string, int, int) {
    translations := make(map[string]string)
    var cacheMisses []TextNode
    seenHashes := make(map[string]bool)
    cachedCount := 0
    
    // Check cache (can parallelize with goroutines)
    for _, node := range nodes {
        key := CacheKey(node.Hash, t.targetLang)
        if cached, ok := t.cache.Get(key); ok {
            translations[node.Hash] = cached
            cachedCount++
        } else if !seenHashes[node.Hash] {
            cacheMisses = append(cacheMisses, node)
            seenHashes[node.Hash] = true
        }
    }
    
    // Translate misses via AI
    if len(cacheMisses) > 0 && t.provider != nil {
        texts := make([]string, len(cacheMisses))
        contexts := make([]string, len(cacheMisses))
        for i, node := range cacheMisses {
            texts[i] = node.Text
            contexts[i] = node.Context
        }
        
        results, _ := t.provider.Translate(context.Background(), TranslateRequest{
            Texts:         texts,
            TargetLang:    t.targetLang,
            SourceLang:    t.sourceLang,
            ExcludedTerms: t.excludedTerms,
            Context:       t.context,
            TextContexts:  contexts,
        })
        
        for i, node := range cacheMisses {
            translations[node.Hash] = results[i]
            key := CacheKey(node.Hash, t.targetLang)
            t.cache.Set(key, results[i])
        }
    }
    
    return translations, cachedCount, len(cacheMisses)
}
```

**Checklist**:
- [ ] Implement `Translator` struct with all dependencies
- [ ] Implement `Process()` method
- [ ] Implement `translateBatch()` with deduplication
- [ ] Add source/target language check bypass
- [ ] Implement RTL detection and HTML attribute setting
- [ ] Add content type auto-detection
- [ ] Write integration tests

---

### Phase 6: Error Handling & Retry

**Files**: `retry.go`

```go
type RetryConfig struct {
    MaxRetries int
    BaseDelay  time.Duration
    MaxDelay   time.Duration
}

func WithRetry[T any](cfg RetryConfig, fn func() (T, error)) (T, error) {
    var lastErr error
    var zero T
    
    for attempt := 0; attempt < cfg.MaxRetries; attempt++ {
        result, err := fn()
        if err == nil {
            return result, nil
        }
        
        lastErr = err
        
        // Check if retryable
        if !isRetryable(err) {
            return zero, err
        }
        
        delay := cfg.BaseDelay * time.Duration(1<<attempt)
        if delay > cfg.MaxDelay {
            delay = cfg.MaxDelay
        }
        time.Sleep(delay)
    }
    
    return zero, lastErr
}

func isRetryable(err error) bool {
    // Rate limits, timeouts, temporary network errors
    var providerErr *ProviderError
    if errors.As(err, &providerErr) {
        return providerErr.Retryable
    }
    return false
}
```

**Checklist**:
- [ ] Implement retry with exponential backoff
- [ ] Define retryable vs non-retryable errors
- [ ] Add timeout handling
- [ ] Log retry attempts

---

### Phase 7: Testing

**Files**: `*_test.go`

**Required Test Cases** (from improved guide):

| Test | Description |
|------|-------------|
| `TestBasicTranslation` | Single text node is translated |
| `TestCacheHit` | Second call uses cache |
| `TestIgnoredTags` | script/style/code not translated |
| `TestDataNoTranslate` | Elements with attribute skipped |
| `TestRTLLanguage` | Arabic gets `dir="rtl"` |
| `TestWhitespacePreserved` | Leading/trailing spaces kept |
| `TestDeduplication` | Identical texts translated once |
| `TestEmptyContent` | Empty returns unchanged |
| `TestSourceEqualsTarget` | No translation when same |

**Mock Provider**:
```go
type MockProvider struct {
    Translations map[string]string
    CallCount    int
    LastTexts    []string
}

func (m *MockProvider) Translate(ctx context.Context, req TranslateRequest) ([]string, error) {
    m.CallCount++
    m.LastTexts = req.Texts
    
    results := make([]string, len(req.Texts))
    for i, text := range req.Texts {
        if t, ok := m.Translations[text]; ok {
            results[i] = t
        } else {
            results[i] = "[" + text + "]"
        }
    }
    return results, nil
}
```

**Checklist**:
- [ ] Implement `MockProvider` for testing
- [ ] Write all required test cases
- [ ] Add benchmarks for performance targets
- [ ] Achieve >80% code coverage

---

### Phase 8: Extensions (Optional)

**Files**: `export/`, `cli/`

| Feature | Priority | Description |
|---------|----------|-------------|
| Export to JSON | Medium | Save translations for offline use |
| Export to PO | Medium | Standard gettext format |
| Import from JSON/PO | Medium | Load pre-translated content |
| CLI interface | Low | Command-line tool |
| Go source translation | Low | Translate comments/strings in Go files |
| File-based cache | Low | Persist cache to disk |

---

## Project Structure

```
gotlai/
├── go.mod
├── go.sum
├── README.md
├── gotlai.go           # Main package exports
├── translator.go       # Translator struct and Process()
├── types.go            # Core data structures
├── errors.go           # Custom error types
├── retry.go            # Retry logic
├── hash.go             # SHA-256 hashing utilities
├── languages.go        # Language codes, RTL detection
├── cache/
│   ├── cache.go        # TranslationCache interface
│   ├── memory.go       # InMemoryCache
│   ├── memory_test.go
│   ├── redis.go        # RedisCache
│   └── redis_test.go
├── provider/
│   ├── provider.go     # AIProvider interface
│   ├── openai.go       # OpenAI implementation
│   ├── openai_test.go
│   └── mock.go         # MockProvider for testing
├── processor/
│   ├── processor.go    # ContentProcessor interface
│   ├── html.go         # HTML processor
│   └── html_test.go
├── export/             # Optional
│   ├── json.go
│   └── po.go
├── cli/                # Optional
│   └── main.go
└── docs/
    ├── porting-guide.md
    ├── improved-porting-guide.md
    ├── porting-report.md
    └── gotlai-plan.md
```

---

## Dependencies

**Required**:
```go
require (
    github.com/PuerkitoBio/goquery v1.8.1
    github.com/sashabaranov/go-openai v1.17.9
    golang.org/x/net v0.19.0
)
```

**Optional**:
```go
require (
    github.com/redis/go-redis/v9 v9.3.0
    github.com/stretchr/testify v1.8.4
)
```

---

## Performance Targets

| Operation | Target |
|-----------|--------|
| Cache lookup (memory) | < 1ms |
| Cache lookup (Redis) | < 5ms |
| HTML parsing (1KB) | < 10ms |
| API call (10 texts) | < 2s |
| Full page (50 nodes) | < 5s |

**Go-specific optimizations**:
- Use goroutines for parallel cache lookups
- Connection pooling for Redis and HTTP
- Reuse `goquery.Document` where possible
- Pre-allocate slices when size is known

---

## Implementation Order

1. **Phase 1**: Core types and errors ✅
2. **Phase 2**: Cache layer with tests ✅
3. **Phase 3**: AI provider with tests ✅
4. **Phase 4**: HTML processor with tests ✅
5. **Phase 5**: Main translator with integration tests ✅
6. **Phase 6**: Error handling and retry ✅
7. **Phase 7**: Full test suite and benchmarks ✅
8. **Phase 8**: CLI ✅

---

## Success Criteria

- [x] All 9 required test cases passing (10 integration tests + 91 unit tests = 101 total)
- [x] >80% code coverage on core packages (90.6% on main, 83.2% on processor)
- [x] Performance within target benchmarks (cache <1ms, HTML parse <10ms)
- [x] Clean `go vet` and `golint` output
- [x] Comprehensive README with examples
- [x] Working example in `examples/` directory

---

## References

- Original porting guide: `docs/porting-guide.md`
- Improved guide (with Pytlai lessons): `docs/improved-porting-guide.md`
- Pytlai porting report: `docs/porting-report.md`
