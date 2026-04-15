package scrapegraph

import (
	"testing"

	"stealthfetch/internal/models"
	"stealthfetch/pkg/loaders"
)

func TestChooseSmartScraperLoader_UsesAutoStrategy(t *testing.T) {
	cfg := models.DefaultConfig()
	cfg.FetchStrategy = "auto"
	loader := chooseSmartScraperLoader("https://example.com", cfg)
	if loader == nil || loader.Name() != "auto" {
		t.Fatalf("expected auto loader, got %#v", loader)
	}
}

func TestChooseSearchLinkLoader_LocalHTML(t *testing.T) {
	loader := chooseSearchLinkLoader("<html></html>", models.DefaultConfig())
	if _, ok := loader.(*loaders.LocalLoader); !ok {
		t.Fatalf("expected local loader, got %T", loader)
	}
}
