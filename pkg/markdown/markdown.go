// Package markdown handles HTML to text and Markdown conversion
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

// HTMLToMarkdown converts HTML into a lightweight Markdown representation.
func HTMLToMarkdown(rawHTML string, maxChars int) string {
	if maxChars == 0 {
		maxChars = 50000
	}
	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return rawHTML
	}
	var sb strings.Builder
	renderMarkdown(doc, &sb)
	result := normalizeWhitespace(sb.String())
	result = normalizeMarkdownSpacing(result)
	if len(result) > maxChars {
		result = result[:maxChars] + "\n\n[TRUNCATED]"
	}
	return strings.TrimSpace(result)
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

func renderMarkdown(n *html.Node, sb *strings.Builder) {
	if n.Type == html.ElementNode && stripTags[n.Data] {
		return
	}
	if n.Type == html.TextNode {
		text := strings.Join(strings.Fields(n.Data), " ")
		if text != "" {
			sb.WriteString(text)
			sb.WriteString(" ")
		}
		return
	}
	if n.Type == html.ElementNode {
		switch n.Data {
		case "h1":
			sb.WriteString("# ")
		case "h2":
			sb.WriteString("## ")
		case "h3":
			sb.WriteString("### ")
		case "h4":
			sb.WriteString("#### ")
		case "h5":
			sb.WriteString("##### ")
		case "h6":
			sb.WriteString("###### ")
		case "li":
			sb.WriteString("- ")
		case "pre":
			sb.WriteString("```\n")
		case "code":
			if n.Parent == nil || n.Parent.Data != "pre" {
				sb.WriteString("`")
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		renderMarkdown(c, sb)
	}
	if n.Type == html.ElementNode {
		switch n.Data {
		case "p", "div", "section", "article", "header", "footer", "br":
			sb.WriteString("\n\n")
		case "h1", "h2", "h3", "h4", "h5", "h6", "ul", "ol", "li":
			sb.WriteString("\n")
		case "a":
			if href := getAttr(n, "href"); href != "" {
				sb.WriteString(" (")
				sb.WriteString(href)
				sb.WriteString(")")
			}
		case "strong", "b":
			sb.WriteString("**")
		case "em", "i":
			sb.WriteString("*")
		case "pre":
			sb.WriteString("\n```\n")
		case "code":
			if n.Parent == nil || n.Parent.Data != "pre" {
				sb.WriteString("`")
			}
		}
	}
}

func getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func normalizeWhitespace(text string) string {
	// Collapse multiple newlines
	for strings.Contains(text, "\n\n\n") {
		text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	}
	return text
}

func normalizeMarkdownSpacing(text string) string {
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}
	text = strings.ReplaceAll(text, " \n", "\n")
	return text
}
