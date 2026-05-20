package main

import (
	"context"
	"log"
	"os"

	"stealthfetch/internal/models"
	"stealthfetch/pkg/loaders"
)

func main() {
	config := models.DefaultConfig()
	config.ScreenshotWaitSecs = 2

	loader := loaders.NewScreenshotLoader(config)
	result, err := loader.Capture(context.Background(), "https://example.com")
	if err != nil {
		log.Fatalf("capture failed: %v", err)
	}
	if err := os.WriteFile("screenshot.png", result.PNG, 0o644); err != nil {
		log.Fatalf("write screenshot: %v", err)
	}
	log.Printf("saved screenshot.png (%d bytes)", len(result.PNG))
}
