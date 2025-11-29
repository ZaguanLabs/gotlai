package gotlai_test

import (
	"context"
	"testing"

	"github.com/ZaguanLabs/gotlai"
	"github.com/ZaguanLabs/gotlai/cache"
	"github.com/ZaguanLabs/gotlai/processor"
	"github.com/ZaguanLabs/gotlai/provider"
)

// Benchmarks for performance validation

func BenchmarkHashText(b *testing.B) {
	text := "Hello World, this is a sample text for hashing"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gotlai.HashText(text)
	}
}

func BenchmarkCacheKey(b *testing.B) {
	hash := "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e"
	lang := "es_ES"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gotlai.CacheKey(hash, lang)
	}
}

func BenchmarkInMemoryCache_Get(b *testing.B) {
	c := cache.NewInMemoryCache(3600)
	c.Set("test-key", "test-value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get("test-key")
	}
}

func BenchmarkInMemoryCache_Set(b *testing.B) {
	c := cache.NewInMemoryCache(3600)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set("test-key", "test-value")
	}
}

func BenchmarkHTMLProcessor_Extract_Small(b *testing.B) {
	proc := processor.NewHTMLProcessor()
	html := `<div><p>Hello World</p></div>`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		proc.Extract(html)
	}
}

func BenchmarkHTMLProcessor_Extract_Medium(b *testing.B) {
	proc := processor.NewHTMLProcessor()
	html := `<!DOCTYPE html>
<html>
<head><title>Test Page</title></head>
<body>
	<nav><a href="/">Home</a><a href="/about">About</a></nav>
	<main>
		<h1>Welcome to Our Site</h1>
		<p>This is a paragraph with some text.</p>
		<p>Another paragraph here.</p>
		<ul>
			<li>Item one</li>
			<li>Item two</li>
			<li>Item three</li>
		</ul>
	</main>
	<footer><p>Copyright 2024</p></footer>
</body>
</html>`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		proc.Extract(html)
	}
}

func BenchmarkTranslator_Process_Cached(b *testing.B) {
	p := provider.NewMockProvider()
	c := cache.NewInMemoryCache(3600)
	proc := processor.NewHTMLProcessor()

	translator := gotlai.NewTranslator("es_ES", p,
		gotlai.WithCache(c),
		gotlai.WithProcessor(proc),
	)

	html := `<div><p>Hello</p><p>World</p></div>`

	// Prime the cache
	translator.ProcessHTML(context.Background(), html)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		translator.ProcessHTML(context.Background(), html)
	}
}

func BenchmarkTranslator_Process_Uncached(b *testing.B) {
	proc := processor.NewHTMLProcessor()
	html := `<div><p>Hello</p><p>World</p></div>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create fresh translator each time to avoid cache
		p := provider.NewMockProvider()
		translator := gotlai.NewTranslator("es_ES", p,
			gotlai.WithProcessor(proc),
		)
		translator.ProcessHTML(context.Background(), html)
	}
}

func BenchmarkGetDirection(b *testing.B) {
	langs := []string{"en_US", "es_ES", "ar_SA", "ja_JP", "he_IL"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gotlai.GetDirection(langs[i%len(langs)])
	}
}

func BenchmarkGetLanguageName(b *testing.B) {
	langs := []string{"en_US", "es_ES", "ar_SA", "ja_JP", "zh_CN"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gotlai.GetLanguageName(langs[i%len(langs)])
	}
}
