package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"stealthfetch/internal/models"
	"stealthfetch/pkg/llm"
	"stealthfetch/pkg/scrapegraph"
)

func main() {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENROUTER_API_KEY is required")
	}
	model := os.Getenv("OPENROUTER_MODEL")
	if model == "" {
		model = "openai/gpt-4o-mini"
	}

	cfg := models.DefaultConfig()
	cfg.LLMProvider = llm.ProviderOpenRouter
	cfg.LLMAPIKey = apiKey
	cfg.LLMModel = model
	cfg.LLMHTTPReferer = os.Getenv("OPENROUTER_HTTP_REFERER")
	cfg.LLMAppTitle = os.Getenv("OPENROUTER_APP_TITLE")

	graph := scrapegraph.NewSmartScraperGraph(
		"Extract the page title and description as JSON",
		`<html><body><h1>GoScrape2</h1><p>A Go scraping graph library.</p></body></html>`,
		cfg,
		`{"title":"string","description":"string"}`,
	)
	result, err := graph.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(result))
}
