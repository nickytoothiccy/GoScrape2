package scrapegraph

import (
	"context"
	"testing"
)

func TestSearchLinkGraphRun_LocalHTML(t *testing.T) {
	html := `<html><body>
		<a href="https://example.com/a">A</a>
		<a href="/b">B</a>
		<a href="https://example.com/a">dup</a>
	</body></html>`
	g := NewSearchLinkGraph(SearchLinkGraphConfig{
		Source:      html,
		MaxLinks:    10,
		FilterByLLM: false,
	})
	links, err := g.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(links) != 1 || links[0] != "https://example.com/a" {
		t.Fatalf("unexpected links: %#v", links)
	}
}

func TestSearchLinkGraphRun_ResolvesRelativeLinks(t *testing.T) {
	html := `<html><body><a href="/docs/start">Start</a><a href="guide">Guide</a></body></html>`
	g := NewSearchLinkGraph(SearchLinkGraphConfig{
		Source:      html,
		BaseURL:     "https://example.com/base/",
		Prompt:      "docs",
		MaxLinks:    5,
		FilterByLLM: false,
	})
	links, err := g.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(links) != 2 || links[0] != "https://example.com/docs/start" || links[1] != "https://example.com/base/guide" {
		t.Fatalf("unexpected resolved links: %#v", links)
	}
}
