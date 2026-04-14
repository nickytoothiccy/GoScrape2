package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"stealthfetch/internal/models"
	"stealthfetch/pkg/scrapegraph"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY env var required")
	}

	config := models.DefaultConfig()
	config.LLMAPIKey = apiKey
	config.LLMModel = "gpt-4o-mini"

	multi := scrapegraph.NewSmartScraperMultiGraph(scrapegraph.SmartScraperMultiConfig{
		Config:        config,
		Prompt:        "Extract the page title and a short summary as JSON",
		URLs:          []string{"https://example.com", "https://example.org"},
		ConcatResults: false,
	})

	result, err := multi.Run(context.Background())
	if err != nil {
		log.Fatalf("multi scrape failed: %v", err)
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, result, "", "  "); err != nil {
		log.Fatalf("format result: %v", err)
	}

	fmt.Println(pretty.String())
	if failed := multi.GetFailedURLs(); len(failed) > 0 {
		fmt.Printf("failed URLs: %v\n", failed)
	}
}
