package scrapegraph

import (
	"context"
	"strings"
	"testing"
)

func TestMarkdownifyGraphRun_LocalHTML(t *testing.T) {
	g := NewMarkdownifyGraph(`<html><body><h2>Docs</h2><p>Some text</p></body></html>`, nil)
	out, err := g.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !strings.Contains(out, "## Docs") || !strings.Contains(out, "Some text") {
		t.Fatalf("unexpected markdown output: %q", out)
	}
}
