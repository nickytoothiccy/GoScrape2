package loaders

import (
	"testing"

	"stealthfetch/internal/models"
)

func TestNewScreenshotLoaderUsesRod(t *testing.T) {
	loader := NewScreenshotLoader(models.DefaultConfig())
	if loader == nil {
		t.Fatal("expected screenshot loader")
	}
	if loader.Name() != "rod" {
		t.Fatalf("expected rod screenshot loader, got %s", loader.Name())
	}
}
