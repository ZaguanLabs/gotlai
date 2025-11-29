# Gotlai

Go Translation AI - An AI-powered HTML translation engine for Go.

[![Go Reference](https://pkg.go.dev/badge/github.com/ZaguanLabs/gotlai.svg)](https://pkg.go.dev/github.com/ZaguanLabs/gotlai)

## Features

- **AI-Powered Translation**: Uses OpenAI (default to GPT-4o-mini) for high-quality translations
- **Intelligent Caching**: SHA-256 based caching with TTL support (in-memory or Redis)
- **Context-Aware**: Disambiguates words based on surrounding HTML context
- **RTL Support**: Automatic `dir="rtl"` for Arabic, Hebrew, Persian, etc.
- **Whitespace Preservation**: Maintains original formatting
- **Batch Processing**: Efficient API usage by batching translations
- **Retry Logic**: Exponential backoff for transient failures

## Installation

```bash
go get github.com/ZaguanLabs/gotlai
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/ZaguanLabs/gotlai"
    "github.com/ZaguanLabs/gotlai/cache"
    "github.com/ZaguanLabs/gotlai/processor"
    "github.com/ZaguanLabs/gotlai/provider"
)

func main() {
    // Create OpenAI provider
    p := provider.NewOpenAIProvider(provider.OpenAIConfig{
        APIKey: os.Getenv("OPENAI_API_KEY"),
        Model:  "gpt-4o-mini",
    })

    // Create translator with cache and HTML processor
    t := gotlai.NewTranslator("es_ES", p,
        gotlai.WithCache(cache.NewInMemoryCache(3600)),
        gotlai.WithProcessor(processor.NewHTMLProcessor()),
        gotlai.WithContext("E-commerce website"),
        gotlai.WithExcludedTerms([]string{"API", "SDK"}),
    )

    // Translate HTML
    html := `<html><body><h1>Welcome</h1><p>Hello World</p></body></html>`
    result, err := t.ProcessHTML(context.Background(), html)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(result.Content)
    fmt.Printf("Translated: %d, Cached: %d\n", result.TranslatedCount, result.CachedCount)
}
```

## Configuration Options

```go
gotlai.NewTranslator("es_ES", provider,
    gotlai.WithSourceLang("en"),                    // Source language (default: "en")
    gotlai.WithCache(cache),                        // Translation cache
    gotlai.WithProcessor(processor),                // Content processor
    gotlai.WithExcludedTerms([]string{"API"}),      // Terms to never translate
    gotlai.WithContext("Technical documentation"),  // Global context for AI
)
```

## Caching

### In-Memory Cache

```go
c := cache.NewInMemoryCache(3600) // TTL in seconds
```

### Redis Cache

```go
c, err := cache.NewRedisCache(cache.RedisConfig{
    URL:       "redis://localhost:6379",
    TTL:       3600,
    KeyPrefix: "gotlai:",
})
```

## Retry Logic

Wrap your provider with retry logic for resilience:

```go
p := provider.NewOpenAIProvider(cfg)
retryable := gotlai.NewRetryableProvider(p, gotlai.RetryConfig{
    MaxRetries: 3,
    BaseDelay:  1 * time.Second,
    MaxDelay:   30 * time.Second,
})
```

## Supported Languages

### Tier 1 (High Quality)
`en_US`, `en_GB`, `de_DE`, `es_ES`, `es_MX`, `fr_FR`, `it_IT`, `ja_JP`, `pt_BR`, `pt_PT`, `zh_CN`, `zh_TW`

### Tier 2 (Good Quality)
`ar_SA`, `ko_KR`, `ru_RU`, `nl_NL`, `pl_PL`, `tr_TR`, `vi_VN`, and more...

### RTL Languages
Arabic (`ar`), Hebrew (`he`), Persian (`fa`), Urdu (`ur`), Pashto (`ps`), Sindhi (`sd`), Uyghur (`ug`)

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                           Gotlai Core                               │
├─────────────────────────────────────────────────────────────────────┤
│  Content Input → Text Extraction → Cache Check → AI Translation    │
│                                         ↓                           │
│                                   Cache Store                       │
│                                         ↓                           │
│  Content Output ← Text Replacement ← Translations                   │
└─────────────────────────────────────────────────────────────────────┘
```

## CLI

### Installation

```bash
go install github.com/ZaguanLabs/gotlai/cmd/gotlai@latest
```

### Usage

```bash
# Translate a file
gotlai --lang es_ES input.html > output.html

# Translate with context
gotlai --lang ja_JP --context "E-commerce website" input.html -o output.html

# Exclude terms from translation
gotlai --lang de_DE --exclude "API,SDK,HTML" input.html

# Read from stdin
cat input.html | gotlai --lang fr_FR > output.html

# Use a specific model
gotlai --lang zh_CN --model gpt-4o input.html
```

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `--lang` | Target language code (required) | - |
| `--source` | Source language code | `en` |
| `--output`, `-o` | Output file | stdout |
| `--api-key` | OpenAI API key | `$OPENAI_API_KEY` |
| `--model` | OpenAI model | `gpt-4o-mini` |
| `--context` | Translation context | - |
| `--exclude` | Comma-separated terms to skip | - |
| `--cache-ttl` | Cache TTL in seconds | `3600` |
| `--quiet` | Suppress progress output | `false` |
| `--dry-run` | Show what would be translated | `false` |
| `--json` | Output result as JSON | `false` |
| `--diff` | Compare with previous version | - |
| `--update` | Only translate new/changed content | `false` |

### Diff Mode (Incremental Updates)

Compare a new version with a previous version to see what changed:

```bash
# See what strings changed between versions
gotlai --lang es_ES --diff old.html new.html

# Output as JSON for CI/CD integration
gotlai --lang es_ES --diff old.html --json new.html
```

This is useful when you update your source content and want to:
- See exactly what strings need re-translation
- Avoid re-translating unchanged content
- Track content changes over time

### Dry Run

Preview what would be translated without calling the API:

```bash
gotlai --lang es_ES --dry-run input.html
```

### JSON Output

Get structured output for programmatic use:

```bash
gotlai --lang es_ES --json input.html
```

## Advanced Features

### Cache Export/Import

Save translations for offline use or sharing:

```go
exporter := cache.NewExporter(myCache)
exporter.ExportToFile("translations.json", map[string]string{"lang": "es_ES"})

importer := cache.NewImporter(newCache)
importer.ImportFromFile("translations.json")
```

### Go Source Translation

Translate Go source code strings and comments:

```go
proc := processor.NewGoProcessor()
// Or customize:
proc := processor.NewGoProcessor(
    processor.WithComments(true),
    processor.WithStrings(true),
)
```

### Rate Limiting

Control API request rate:

```go
provider := gotlai.NewRateLimitedProvider(openaiProvider, gotlai.RateLimitConfig{
    RequestsPerMinute: 60,
    BurstSize:         10,
})
```

### Parallel Cache Lookups

For high-performance scenarios:

```go
translations, misses := gotlai.ParallelCacheLookup(cache, nodes, "es_ES")
```

## Testing

```bash
go test -v ./...
```

## Benchmarks

```bash
go test -bench=. -benchmem ./...
```

## License

MIT
