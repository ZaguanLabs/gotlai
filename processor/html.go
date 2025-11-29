package processor

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ZaguanLabs/gotlai"
	"golang.org/x/net/html"
)

// HTMLProcessor extracts and applies translations to HTML content.
type HTMLProcessor struct {
	ignoredTags map[string]bool
}

// NewHTMLProcessor creates a new HTML processor with default ignored tags.
func NewHTMLProcessor() *HTMLProcessor {
	return &HTMLProcessor{
		ignoredTags: gotlai.IgnoredTags,
	}
}

// NewHTMLProcessorWithIgnoredTags creates a new HTML processor with custom ignored tags.
func NewHTMLProcessorWithIgnoredTags(tags []string) *HTMLProcessor {
	ignored := make(map[string]bool)
	for _, tag := range tags {
		ignored[strings.ToLower(tag)] = true
	}
	return &HTMLProcessor{
		ignoredTags: ignored,
	}
}

// parsedHTML holds the parsed document and node mappings.
type parsedHTML struct {
	doc     *goquery.Document
	nodeMap map[string]*html.Node // Maps node ID to HTML node for mutation
}

// Extract parses HTML and extracts translatable text nodes.
func (p *HTMLProcessor) Extract(content string) (interface{}, []gotlai.TextNode, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return nil, nil, &gotlai.ProcessorError{
			Message:     "failed to parse HTML",
			Cause:       err,
			ContentType: "html",
		}
	}

	var nodes []gotlai.TextNode
	nodeMap := make(map[string]*html.Node)
	seenHashes := make(map[string]bool)

	// Walk the DOM tree
	var walk func(*html.Node, *goquery.Selection)
	walk = func(n *html.Node, parentSel *goquery.Selection) {
		if n.Type == html.ElementNode {
			// Skip ignored tags
			if p.ignoredTags[strings.ToLower(n.Data)] {
				return
			}

			// Skip elements with data-no-translate attribute
			for _, attr := range n.Attr {
				if attr.Key == "data-no-translate" {
					return
				}
			}
		}

		if n.Type == html.TextNode {
			text := n.Data
			trimmed := strings.TrimSpace(text)

			if trimmed != "" {
				hash := gotlai.HashText(trimmed)

				// Deduplicate by hash
				if !seenHashes[hash] {
					seenHashes[hash] = true

					nodeID := fmt.Sprintf("node-%d", len(nodes))
					context := p.buildContext(n, parentSel)

					node := gotlai.TextNode{
						ID:       nodeID,
						Text:     trimmed,
						Hash:     hash,
						NodeType: "html_text",
						Context:  context,
						Metadata: map[string]string{},
					}

					if n.Parent != nil {
						node.Metadata["parent_tag"] = n.Parent.Data
					}

					nodes = append(nodes, node)
				}

				// Always map this node for later mutation (even if duplicate hash)
				nodeID := fmt.Sprintf("node-%d-%d", len(nodes)-1, len(nodeMap))
				nodeMap[nodeID] = n
			}
		}

		// Recurse into children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			childSel := parentSel
			if c.Type == html.ElementNode {
				childSel = parentSel.Find(c.Data).First()
			}
			walk(c, childSel)
		}
	}

	// Start walking from the root
	doc.Each(func(i int, s *goquery.Selection) {
		for _, n := range s.Nodes {
			walk(n, s)
		}
	})

	return &parsedHTML{doc: doc, nodeMap: nodeMap}, nodes, nil
}

// Apply applies translations back to the HTML document.
func (p *HTMLProcessor) Apply(parsed interface{}, nodes []gotlai.TextNode, translations map[string]string) (string, error) {
	ph, ok := parsed.(*parsedHTML)
	if !ok {
		return "", &gotlai.ProcessorError{
			Message:     "invalid parsed content type",
			ContentType: "html",
		}
	}

	// Build a map of hash to translation
	hashToTranslation := translations

	// Walk the DOM and apply translations
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// Skip ignored tags
			if p.ignoredTags[strings.ToLower(n.Data)] {
				return
			}

			// Skip elements with data-no-translate attribute
			for _, attr := range n.Attr {
				if attr.Key == "data-no-translate" {
					return
				}
			}
		}

		if n.Type == html.TextNode {
			text := n.Data
			trimmed := strings.TrimSpace(text)

			if trimmed != "" {
				hash := gotlai.HashText(trimmed)
				if translated, ok := hashToTranslation[hash]; ok {
					// Preserve original whitespace
					n.Data = preserveWhitespace(text, translated)
				}
			}
		}

		// Recurse into children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	ph.doc.Each(func(i int, s *goquery.Selection) {
		for _, n := range s.Nodes {
			walk(n)
		}
	})

	html, err := ph.doc.Html()
	if err != nil {
		return "", &gotlai.ProcessorError{
			Message:     "failed to serialize HTML",
			Cause:       err,
			ContentType: "html",
		}
	}

	return html, nil
}

// ContentType returns "html".
func (p *HTMLProcessor) ContentType() string {
	return "html"
}

// buildContext creates a disambiguation context string for a text node.
func (p *HTMLProcessor) buildContext(n *html.Node, parentSel *goquery.Selection) string {
	var parts []string

	if n.Parent != nil {
		parent := n.Parent
		tag := parent.Data

		// Get class or id if available
		var classAttr, idAttr string
		for _, attr := range parent.Attr {
			if attr.Key == "class" {
				classAttr = attr.Val
			} else if attr.Key == "id" {
				idAttr = attr.Val
			}
		}

		if classAttr != "" {
			parts = append(parts, fmt.Sprintf("in <%s class=\"%s\">", tag, classAttr))
		} else if idAttr != "" {
			parts = append(parts, fmt.Sprintf("in <%s id=\"%s\">", tag, idAttr))
		} else {
			parts = append(parts, fmt.Sprintf("in <%s>", tag))
		}

		// Get sibling text (up to 3 items)
		var siblings []string
		for sib := parent.FirstChild; sib != nil; sib = sib.NextSibling {
			if sib == n {
				continue
			}
			if sib.Type == html.TextNode {
				sibText := strings.TrimSpace(sib.Data)
				if sibText != "" && len(sibText) < 100 {
					siblings = append(siblings, sibText)
				}
			}
		}
		if len(siblings) > 3 {
			siblings = siblings[:3]
		}
		if len(siblings) > 0 {
			parts = append(parts, fmt.Sprintf("with: %s", strings.Join(siblings, ", ")))
		}

		// Get ancestor path (up to 3 levels)
		var ancestors []string
		ancestor := parent.Parent
		for i := 0; i < 3 && ancestor != nil; i++ {
			if ancestor.Type == html.ElementNode {
				name := ancestor.Data
				if name != "html" && name != "body" {
					ancestors = append(ancestors, name)
				}
			}
			ancestor = ancestor.Parent
		}
		if len(ancestors) > 0 {
			// Reverse to show outer to inner
			for i, j := 0, len(ancestors)-1; i < j; i, j = i+1, j-1 {
				ancestors[i], ancestors[j] = ancestors[j], ancestors[i]
			}
			parts = append(parts, fmt.Sprintf("inside: %s", strings.Join(ancestors, " > ")))
		}
	}

	return strings.Join(parts, " | ")
}

// preserveWhitespace preserves the original leading/trailing whitespace.
func preserveWhitespace(original, translated string) string {
	// Find leading whitespace
	leadingLen := len(original) - len(strings.TrimLeft(original, " \t\n\r"))
	leading := original[:leadingLen]

	// Find trailing whitespace
	trailingLen := len(original) - len(strings.TrimRight(original, " \t\n\r"))
	trailing := ""
	if trailingLen > 0 {
		trailing = original[len(original)-trailingLen:]
	}

	return leading + translated + trailing
}

// Verify HTMLProcessor implements ContentProcessor
var _ ContentProcessor = (*HTMLProcessor)(nil)
