package processor

import (
	"strings"
	"testing"
)

func TestGoProcessor_Extract_Strings(t *testing.T) {
	p := NewGoProcessor()

	src := `package main

func main() {
	msg := "Hello World"
	fmt.Println(msg)
}
`
	_, nodes, err := p.Extract(src)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Should find "Hello World"
	found := false
	for _, n := range nodes {
		if n.Text == "Hello World" {
			found = true
			if n.NodeType != "go_string" {
				t.Errorf("Expected node type 'go_string', got %q", n.NodeType)
			}
		}
	}

	if !found {
		t.Error("Expected to find 'Hello World' string")
	}
}

func TestGoProcessor_Extract_Comments(t *testing.T) {
	p := NewGoProcessor()

	src := `package main

// This is a comment
func main() {
	/* Another comment */
}
`
	_, nodes, err := p.Extract(src)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Should find comments
	foundLine := false
	foundBlock := false
	for _, n := range nodes {
		if n.Text == "This is a comment" {
			foundLine = true
		}
		if n.Text == "Another comment" {
			foundBlock = true
		}
	}

	if !foundLine {
		t.Error("Expected to find line comment")
	}
	if !foundBlock {
		t.Error("Expected to find block comment")
	}
}

func TestGoProcessor_Extract_SkipsNonTranslatable(t *testing.T) {
	p := NewGoProcessor()

	src := `package main

func main() {
	path := "/api/v1/users"
	format := "%s"
	constant := "API_KEY"
}
`
	_, nodes, err := p.Extract(src)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Should skip paths, format specifiers, and constants
	for _, n := range nodes {
		if n.Text == "/api/v1/users" {
			t.Error("Should skip path-like strings")
		}
		if n.Text == "%s" {
			t.Error("Should skip format specifiers")
		}
		if n.Text == "API_KEY" {
			t.Error("Should skip all-caps constants")
		}
	}
}

func TestGoProcessor_Apply(t *testing.T) {
	p := NewGoProcessor()

	src := `package main

func main() {
	msg := "Hello"
}
`
	parsed, nodes, err := p.Extract(src)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	translations := make(map[string]string)
	for _, n := range nodes {
		if n.Text == "Hello" {
			translations[n.Hash] = "Hola"
		}
	}

	result, err := p.Apply(parsed, nodes, translations)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if !strings.Contains(result, `"Hola"`) {
		t.Errorf("Expected translated string 'Hola', got:\n%s", result)
	}
}

func TestGoProcessor_Apply_Comments(t *testing.T) {
	p := NewGoProcessor()

	src := `package main

// Hello
func main() {}
`
	parsed, nodes, err := p.Extract(src)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	translations := make(map[string]string)
	for _, n := range nodes {
		if n.Text == "Hello" {
			translations[n.Hash] = "Hola"
		}
	}

	result, err := p.Apply(parsed, nodes, translations)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if !strings.Contains(result, "// Hola") {
		t.Errorf("Expected translated comment '// Hola', got:\n%s", result)
	}
}

func TestGoProcessor_WithOptions(t *testing.T) {
	// Comments only
	p1 := NewGoProcessor(WithStrings(false))
	src := `package main

// Comment
func main() {
	msg := "String"
}
`
	_, nodes1, _ := p1.Extract(src)
	for _, n := range nodes1 {
		if n.NodeType == "go_string" {
			t.Error("Should not extract strings when disabled")
		}
	}

	// Strings only
	p2 := NewGoProcessor(WithComments(false))
	_, nodes2, _ := p2.Extract(src)
	for _, n := range nodes2 {
		if n.NodeType == "go_comment" {
			t.Error("Should not extract comments when disabled")
		}
	}
}

func TestGoProcessor_ContentType(t *testing.T) {
	p := NewGoProcessor()
	if p.ContentType() != "go" {
		t.Errorf("Expected 'go', got %q", p.ContentType())
	}
}

func TestGoProcessor_InvalidSource(t *testing.T) {
	p := NewGoProcessor()

	_, _, err := p.Extract("this is not valid go code {{{")
	if err == nil {
		t.Error("Expected error for invalid Go source")
	}
}

func TestGoProcessor_Deduplication(t *testing.T) {
	p := NewGoProcessor()

	src := `package main

func main() {
	a := "Hello"
	b := "Hello"
	c := "Hello"
}
`
	_, nodes, err := p.Extract(src)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Should only have one node for "Hello" due to deduplication
	count := 0
	for _, n := range nodes {
		if n.Text == "Hello" {
			count++
		}
	}

	if count != 1 {
		t.Errorf("Expected 1 unique 'Hello' node, got %d", count)
	}
}

func TestGoProcessor_BacktickStrings(t *testing.T) {
	p := NewGoProcessor()

	src := "package main\n\nfunc main() {\n\tmsg := `Hello World`\n}\n"

	parsed, nodes, err := p.Extract(src)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	translations := make(map[string]string)
	for _, n := range nodes {
		if n.Text == "Hello World" {
			translations[n.Hash] = "Hola Mundo"
		}
	}

	result, err := p.Apply(parsed, nodes, translations)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if !strings.Contains(result, "`Hola Mundo`") {
		t.Errorf("Expected backtick string, got:\n%s", result)
	}
}
