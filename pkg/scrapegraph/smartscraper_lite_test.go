package scrapegraph

import (
	"context"
	"testing"

	"stealthfetch/internal/models"
	"stealthfetch/pkg/loaders"
)

func TestNewSmartScraperLiteGraph_AppliesLiteDefaults(t *testing.T) {
	g := NewSmartScraperLiteGraph("extract title", "<html></html>", nil, "")
	if g == nil || g.config == nil || g.loader == nil {
		t.Fatalf("expected initialized lite graph, got %#v", g)
	}
	if g.config.HTMLMaxChars != defaultLiteHTMLMaxChars {
		t.Fatalf("expected HTMLMaxChars=%d, got %d", defaultLiteHTMLMaxChars, g.config.HTMLMaxChars)
	}
	if g.config.ChunkSize != defaultLiteChunkSize {
		t.Fatalf("expected ChunkSize=%d, got %d", defaultLiteChunkSize, g.config.ChunkSize)
	}
}

func TestSmartScraperLiteGraphRun_LocalHTML(t *testing.T) {
	g := NewSmartScraperLiteGraph("extract title", `<html><body><h1>Lite Docs</h1></body></html>`, models.DefaultConfig(), "")
	g.llmClient = stubLLM{}
	g.loader = loaders.NewLocalLoader()
	result, err := g.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if string(result) != `{"ok":true}` {
		t.Fatalf("unexpected result: %s", string(result))
	}
}
