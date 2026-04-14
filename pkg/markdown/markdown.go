// Package markdown handles HTML to text conversion
package markdown

import (
	"strings"

	"golang.org/x/net/html"
)

var stripTags = map[string]bool{
	"script": true, "style": true, "noscript": true, "svg": true,
	"iframe": true, "link": true, "meta": true, "head": true,
}

// HTMLToText converts HTML to clean text, stripping scripts and styles
func HTMLToText(rawHTML string, maxChars int) string {
	if maxChars == 0 {
		maxChars = 50000
	}

	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return rawHTML
	}

	var sb strings.Builder
	extractText(doc, &sb)

	result := sb.String()
	result = normalizeWhitespace(result)

	if len(result) > maxChars {
		result = result[:maxChars] + "\n\n[TRUNCATED]"
	}

	return result
}

func extractText(n *html.Node, sb *strings.Builder) {
	if n.Type == html.ElementNode && stripTags[n.Data] {
		return
	}

	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			sb.WriteString(text)
			sb.WriteString("\n")
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractText(c, sb)
	}
}

func normalizeWhitespace(text string) string {
	// Collapse multiple newlines
	for strings.Contains(text, "\n\n\n") {
		text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	}
	return text
}
