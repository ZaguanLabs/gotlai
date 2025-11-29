package processor

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"

	"github.com/ZaguanLabs/gotlai"
)

// GoProcessor extracts and applies translations to Go source code.
// It translates string literals and comments.
type GoProcessor struct {
	translateComments bool
	translateStrings  bool
}

// GoProcessorOption configures the Go processor.
type GoProcessorOption func(*GoProcessor)

// WithComments enables/disables comment translation.
func WithComments(enabled bool) GoProcessorOption {
	return func(p *GoProcessor) {
		p.translateComments = enabled
	}
}

// WithStrings enables/disables string literal translation.
func WithStrings(enabled bool) GoProcessorOption {
	return func(p *GoProcessor) {
		p.translateStrings = enabled
	}
}

// NewGoProcessor creates a new Go source processor.
func NewGoProcessor(opts ...GoProcessorOption) *GoProcessor {
	p := &GoProcessor{
		translateComments: true,
		translateStrings:  true,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// parsedGo holds the parsed Go AST and file set.
type parsedGo struct {
	fset    *token.FileSet
	file    *ast.File
	content string
}

// Extract parses Go source and extracts translatable text nodes.
func (p *GoProcessor) Extract(content string) (interface{}, []gotlai.TextNode, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "source.go", content, parser.ParseComments)
	if err != nil {
		return nil, nil, &gotlai.ProcessorError{
			Message:     "failed to parse Go source",
			Cause:       err,
			ContentType: "go",
		}
	}

	var nodes []gotlai.TextNode
	seenHashes := make(map[string]bool)

	// Extract comments
	if p.translateComments {
		for _, cg := range file.Comments {
			for _, c := range cg.List {
				text := extractCommentText(c.Text)
				if text == "" {
					continue
				}

				hash := gotlai.HashText(text)
				if seenHashes[hash] {
					continue
				}
				seenHashes[hash] = true

				nodes = append(nodes, gotlai.TextNode{
					ID:       fmt.Sprintf("comment-%d", c.Pos()),
					Text:     text,
					Hash:     hash,
					NodeType: "go_comment",
					Context:  "Go source comment",
					Metadata: map[string]string{
						"pos": fmt.Sprintf("%d", c.Pos()),
					},
				})
			}
		}
	}

	// Extract string literals
	if p.translateStrings {
		ast.Inspect(file, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}

			// Get the actual string value (remove quotes)
			text := strings.Trim(lit.Value, "`\"")
			if text == "" || !isTranslatableString(text) {
				return true
			}

			hash := gotlai.HashText(text)
			if seenHashes[hash] {
				return true
			}
			seenHashes[hash] = true

			// Build context from parent
			ctx := "Go string literal"

			nodes = append(nodes, gotlai.TextNode{
				ID:       fmt.Sprintf("string-%d", lit.Pos()),
				Text:     text,
				Hash:     hash,
				NodeType: "go_string",
				Context:  ctx,
				Metadata: map[string]string{
					"pos":   fmt.Sprintf("%d", lit.Pos()),
					"quote": string(lit.Value[0]),
				},
			})

			return true
		})
	}

	return &parsedGo{fset: fset, file: file, content: content}, nodes, nil
}

// Apply applies translations back to the Go source.
func (p *GoProcessor) Apply(parsed interface{}, nodes []gotlai.TextNode, translations map[string]string) (string, error) {
	pg, ok := parsed.(*parsedGo)
	if !ok {
		return "", &gotlai.ProcessorError{
			Message:     "invalid parsed content type",
			ContentType: "go",
		}
	}

	// Build position to translation map
	posToTranslation := make(map[token.Pos]string)
	for _, node := range nodes {
		if translated, ok := translations[node.Hash]; ok {
			if posStr, ok := node.Metadata["pos"]; ok {
				var pos token.Pos
				if _, err := fmt.Sscanf(posStr, "%d", &pos); err == nil {
					posToTranslation[pos] = translated
				}
			}
		}
	}

	// Apply translations to comments
	if p.translateComments {
		for _, cg := range pg.file.Comments {
			for _, c := range cg.List {
				if translated, ok := posToTranslation[c.Pos()]; ok {
					if strings.HasPrefix(c.Text, "//") {
						c.Text = "// " + translated
					} else if strings.HasPrefix(c.Text, "/*") {
						c.Text = "/* " + translated + " */"
					}
				}
			}
		}
	}

	// Apply translations to string literals
	if p.translateStrings {
		ast.Inspect(pg.file, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}

			if translated, ok := posToTranslation[lit.Pos()]; ok {
				quote := string(lit.Value[0])
				if quote == "`" {
					lit.Value = "`" + translated + "`"
				} else {
					// Escape the translated string for double quotes
					lit.Value = `"` + escapeString(translated) + `"`
				}
			}

			return true
		})
	}

	// Print the modified AST
	var buf strings.Builder
	if err := printer.Fprint(&buf, pg.fset, pg.file); err != nil {
		return "", &gotlai.ProcessorError{
			Message:     "failed to print Go source",
			Cause:       err,
			ContentType: "go",
		}
	}

	return buf.String(), nil
}

// ContentType returns "go".
func (p *GoProcessor) ContentType() string {
	return "go"
}

// extractCommentText extracts the text content from a comment.
func extractCommentText(comment string) string {
	if strings.HasPrefix(comment, "//") {
		return strings.TrimSpace(comment[2:])
	}
	if strings.HasPrefix(comment, "/*") && strings.HasSuffix(comment, "*/") {
		return strings.TrimSpace(comment[2 : len(comment)-2])
	}
	return ""
}

// isTranslatableString checks if a string should be translated.
func isTranslatableString(s string) bool {
	// Skip empty or very short strings
	if len(s) < 2 {
		return false
	}

	// Skip strings that look like identifiers or paths
	if strings.Contains(s, "/") && !strings.Contains(s, " ") {
		return false
	}

	// Skip strings that look like format specifiers
	if strings.HasPrefix(s, "%") && len(s) < 5 {
		return false
	}

	// Skip strings that are all uppercase (likely constants)
	if s == strings.ToUpper(s) && !strings.Contains(s, " ") {
		return false
	}

	// Must contain at least one letter
	hasLetter := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			hasLetter = true
			break
		}
	}

	return hasLetter
}

// escapeString escapes a string for use in a Go string literal.
func escapeString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	return s
}

// Verify GoProcessor implements ContentProcessor
var _ ContentProcessor = (*GoProcessor)(nil)
