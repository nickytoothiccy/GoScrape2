// Package main shows basic usage of the scrapegraph library
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

	// Configure
	config := models.DefaultConfig()
	config.LLMAPIKey = apiKey
	config.LLMModel = "gpt-4o-mini"
	config.Verbose = true

	// Create scraper
	prompt := "Extract the main article title and first paragraph"
	url := "https://news.ycombinator.com"

	scraper := scrapegraph.NewSmartScraperGraph(prompt, url, config, "")

	// Run
	result, err := scraper.Run(context.Background())
	if err != nil {
		log.Fatalf("scrape failed: %v", err)
	}

	// Pretty print
	var pretty bytes.Buffer
	json.Indent(&pretty, result, "", "  ")
	fmt.Println(pretty.String())
}
