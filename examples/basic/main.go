// Example: Basic HTML translation with gotlai
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
	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Create OpenAI provider
	p := provider.NewOpenAIProvider(provider.OpenAIConfig{
		APIKey: apiKey,
		Model:  "gpt-4o-mini",
	})

	// Wrap with retry logic for resilience
	retryable := gotlai.NewRetryableProvider(p, gotlai.DefaultRetryConfig())

	// Create translator with cache and HTML processor
	t := gotlai.NewTranslator("es_ES", retryable,
		gotlai.WithCache(cache.NewInMemoryCache(3600)),
		gotlai.WithProcessor(processor.NewHTMLProcessor()),
		gotlai.WithContext("E-commerce website"),
		gotlai.WithExcludedTerms([]string{"API", "SDK", "HTML"}),
	)

	// Sample HTML to translate
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Welcome to Our Store</title>
</head>
<body>
    <nav>
        <a href="/">Home</a>
        <a href="/products">Products</a>
        <a href="/about">About Us</a>
    </nav>
    <main>
        <h1>Welcome to Our Store</h1>
        <p>Find the best products at great prices.</p>
        <button>Shop Now</button>
    </main>
    <footer>
        <p>Contact us at support@example.com</p>
    </footer>
    <script>console.log("This should not be translated");</script>
</body>
</html>`

	fmt.Println("=== Original HTML ===")
	fmt.Println(html)
	fmt.Println()

	// Translate
	result, err := t.ProcessHTML(context.Background(), html)
	if err != nil {
		log.Fatalf("Translation failed: %v", err)
	}

	fmt.Println("=== Translated HTML ===")
	fmt.Println(result.Content)
	fmt.Println()

	fmt.Printf("=== Statistics ===\n")
	fmt.Printf("Total nodes found: %d\n", result.TotalNodes)
	fmt.Printf("Newly translated:  %d\n", result.TranslatedCount)
	fmt.Printf("From cache:        %d\n", result.CachedCount)

	// Translate again to demonstrate caching
	fmt.Println("\n=== Second Translation (should use cache) ===")
	result2, err := t.ProcessHTML(context.Background(), html)
	if err != nil {
		log.Fatalf("Second translation failed: %v", err)
	}

	fmt.Printf("Total nodes found: %d\n", result2.TotalNodes)
	fmt.Printf("Newly translated:  %d\n", result2.TranslatedCount)
	fmt.Printf("From cache:        %d\n", result2.CachedCount)
}
