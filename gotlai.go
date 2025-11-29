// Package gotlai provides an AI-powered HTML translation engine.
//
// Gotlai translates HTML content using AI providers (OpenAI, etc.) with
// intelligent caching, context-aware disambiguation, and support for
// right-to-left languages.
//
// Basic usage:
//
//	import (
//	    "context"
//	    "github.com/ZaguanLabs/gotlai"
//	    "github.com/ZaguanLabs/gotlai/cache"
//	    "github.com/ZaguanLabs/gotlai/processor"
//	    "github.com/ZaguanLabs/gotlai/provider"
//	)
//
//	func main() {
//	    // Create provider
//	    p := provider.NewOpenAIProvider(provider.OpenAIConfig{
//	        APIKey: os.Getenv("OPENAI_API_KEY"),
//	    })
//
//	    // Create translator
//	    t := gotlai.NewTranslator("es_ES", p,
//	        gotlai.WithCache(cache.NewInMemoryCache(3600)),
//	        gotlai.WithProcessor(processor.NewHTMLProcessor()),
//	    )
//
//	    // Translate HTML
//	    result, err := t.ProcessHTML(context.Background(), "<p>Hello World</p>")
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    fmt.Println(result.Content) // <p>Hola Mundo</p>
//	}
package gotlai
