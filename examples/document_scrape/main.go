package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"stealthfetch/internal/models"
	"stealthfetch/pkg/scrapegraph"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("usage: go run ./examples/document_scrape <file> <prompt>")
	}
	cfg := models.DefaultConfig()
	cfg.LLMAPIKey = os.Getenv("OPENAI_API_KEY")
	if cfg.LLMAPIKey == "" {
		log.Fatal("OPENAI_API_KEY env var required")
	}
	g := scrapegraph.NewDocumentScraperGraph(os.Args[2], os.Args[1], cfg, "")
	result, err := g.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(result))
}
