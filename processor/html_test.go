package processor

import (
	"strings"
	"testing"

	"github.com/ZaguanLabs/gotlai"
)

func TestHTMLProcessor_Extract_Basic(t *testing.T) {
	p := NewHTMLProcessor()

	html := `<div><h1>Hello World</h1><p>Welcome to our site.</p></div>`
	parsed, nodes, err := p.Extract(html)

	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if parsed == nil {
		t.Fatal("parsed should not be nil")
	}

	if len(nodes) != 2 {
		t.Fatalf("Expected 2 nodes, got %d", len(nodes))
	}

	// Check first node
	if nodes[0].Text != "Hello World" {
		t.Errorf("Expected 'Hello World', got %q", nodes[0].Text)
	}
	if nodes[0].Hash == "" {
		t.Error("Hash should not be empty")
	}
	if nodes[0].NodeType != "html_text" {
		t.Errorf("Expected node type 'html_text', got %q", nodes[0].NodeType)
	}

	// Check second node
	if nodes[1].Text != "Welcome to our site." {
		t.Errorf("Expected 'Welcome to our site.', got %q", nodes[1].Text)
	}
}

func TestHTMLProcessor_Extract_IgnoredTags(t *testing.T) {
	p := NewHTMLProcessor()

	html := `<div>
		<p>Translate me</p>
		<script>doNotTranslate();</script>
		<style>.class { color: red; }</style>
		<code>const x = 1;</code>
		<pre>preformatted</pre>
		<textarea>form input</textarea>
	</div>`

	_, nodes, err := p.Extract(html)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Only "Translate me" should be extracted
	if len(nodes) != 1 {
		t.Fatalf("Expected 1 node (only 'Translate me'), got %d", len(nodes))
	}

	if nodes[0].Text != "Translate me" {
		t.Errorf("Expected 'Translate me', got %q", nodes[0].Text)
	}
}

func TestHTMLProcessor_Extract_DataNoTranslate(t *testing.T) {
	p := NewHTMLProcessor()

	html := `<div>
		<p data-no-translate>Keep this</p>
		<p>Translate this</p>
	</div>`

	_, nodes, err := p.Extract(html)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("Expected 1 node, got %d", len(nodes))
	}

	if nodes[0].Text != "Translate this" {
		t.Errorf("Expected 'Translate this', got %q", nodes[0].Text)
	}
}

func TestHTMLProcessor_Extract_Deduplication(t *testing.T) {
	p := NewHTMLProcessor()

	html := `<div>
		<p>Hello</p>
		<p>Hello</p>
		<p>Hello</p>
	</div>`

	_, nodes, err := p.Extract(html)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Should only have one unique node
	if len(nodes) != 1 {
		t.Fatalf("Expected 1 unique node, got %d", len(nodes))
	}
}

func TestHTMLProcessor_Extract_Context(t *testing.T) {
	p := NewHTMLProcessor()

	html := `<nav><button class="primary">Run</button></nav>`
	_, nodes, err := p.Extract(html)

	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("Expected 1 node, got %d", len(nodes))
	}

	// Context should mention button and class
	ctx := nodes[0].Context
	if !strings.Contains(ctx, "button") {
		t.Errorf("Context should mention button tag, got: %s", ctx)
	}
	if !strings.Contains(ctx, "primary") {
		t.Errorf("Context should mention class, got: %s", ctx)
	}
}

func TestHTMLProcessor_Apply(t *testing.T) {
	p := NewHTMLProcessor()

	html := `<div><p>Hello</p><p>World</p></div>`
	parsed, nodes, err := p.Extract(html)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	translations := make(map[string]string)
	for _, node := range nodes {
		if node.Text == "Hello" {
			translations[node.Hash] = "Hola"
		} else if node.Text == "World" {
			translations[node.Hash] = "Mundo"
		}
	}

	result, err := p.Apply(parsed, nodes, translations)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if !strings.Contains(result, "Hola") {
		t.Error("Result should contain 'Hola'")
	}
	if !strings.Contains(result, "Mundo") {
		t.Error("Result should contain 'Mundo'")
	}
	if strings.Contains(result, "Hello") {
		t.Error("Result should not contain 'Hello'")
	}
}

func TestHTMLProcessor_Apply_PreservesWhitespace(t *testing.T) {
	p := NewHTMLProcessor()

	html := `<p>  Hello  </p>`
	parsed, nodes, err := p.Extract(html)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	translations := map[string]string{
		nodes[0].Hash: "Hola",
	}

	result, err := p.Apply(parsed, nodes, translations)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Should preserve the whitespace pattern
	if !strings.Contains(result, "  Hola  ") {
		t.Errorf("Result should preserve whitespace, got: %s", result)
	}
}

func TestHTMLProcessor_Apply_DuplicateTexts(t *testing.T) {
	p := NewHTMLProcessor()

	html := `<div><p>Hello</p><p>Hello</p></div>`
	parsed, nodes, err := p.Extract(html)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Only one node due to deduplication
	if len(nodes) != 1 {
		t.Fatalf("Expected 1 node, got %d", len(nodes))
	}

	translations := map[string]string{
		nodes[0].Hash: "Hola",
	}

	result, err := p.Apply(parsed, nodes, translations)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Both instances should be translated
	count := strings.Count(result, "Hola")
	if count != 2 {
		t.Errorf("Expected 2 instances of 'Hola', got %d in: %s", count, result)
	}
}

func TestHTMLProcessor_ContentType(t *testing.T) {
	p := NewHTMLProcessor()
	if p.ContentType() != "html" {
		t.Errorf("Expected 'html', got %q", p.ContentType())
	}
}

func TestPreserveWhitespace(t *testing.T) {
	tests := []struct {
		original   string
		translated string
		expected   string
	}{
		{"Hello", "Hola", "Hola"},
		{"  Hello", "Hola", "  Hola"},
		{"Hello  ", "Hola", "Hola  "},
		{"  Hello  ", "Hola", "  Hola  "},
		{"\n\tHello\n", "Hola", "\n\tHola\n"},
	}

	for _, tt := range tests {
		result := preserveWhitespace(tt.original, tt.translated)
		if result != tt.expected {
			t.Errorf("preserveWhitespace(%q, %q) = %q, want %q",
				tt.original, tt.translated, result, tt.expected)
		}
	}
}

func TestHTMLProcessor_EmptyContent(t *testing.T) {
	p := NewHTMLProcessor()

	html := `<div></div>`
	_, nodes, err := p.Extract(html)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(nodes) != 0 {
		t.Errorf("Expected 0 nodes for empty content, got %d", len(nodes))
	}
}

func TestHTMLProcessor_WhitespaceOnlyContent(t *testing.T) {
	p := NewHTMLProcessor()

	html := `<div>   </div>`
	_, nodes, err := p.Extract(html)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(nodes) != 0 {
		t.Errorf("Expected 0 nodes for whitespace-only content, got %d", len(nodes))
	}
}

// Verify HTMLProcessor implements ContentProcessor
var _ ContentProcessor = (*HTMLProcessor)(nil)

// Verify error types
func TestHTMLProcessor_ExtractError(t *testing.T) {
	p := NewHTMLProcessor()

	// Invalid HTML should still parse (goquery is lenient)
	_, _, err := p.Extract("<div>unclosed")
	if err != nil {
		// goquery is very lenient, so this might not error
		var procErr *gotlai.ProcessorError
		if _, ok := err.(*gotlai.ProcessorError); ok {
			_ = procErr // Just checking type
		}
	}
}
