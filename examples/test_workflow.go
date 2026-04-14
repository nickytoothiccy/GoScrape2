// Test the full SmartScraperGraph workflow
package main

import (
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

	// Test 1: Local HTML
	fmt.Println("=== Test 1: Local HTML ===")
	testLocalHTML(apiKey)

	// Test 2: Live URL
	fmt.Println("\n=== Test 2: Live URL ===")
	testLiveURL(apiKey)
}

func testLocalHTML(apiKey string) {
	html := `
	<html>
		<body>
			<h1>Product Catalog</h1>
			<div class="product">
				<h2>Widget Pro</h2>
				<span class="price">$99.99</span>
			</div>
			<div class="product">
				<h2>Gadget Ultra</h2>
				<span class="price">$149.99</span>
			</div>
		</body>
	</html>
	`

	config := models.DefaultConfig()
	config.LLMAPIKey = apiKey
	config.LLMModel = "gpt-4o-mini"
	config.Verbose = true

	scraper := scrapegraph.NewSmartScraperGraph(
		"Extract all products with their names and prices as a JSON array",
		html,
		config,
		`{"products": [{"name": "string", "price": "string"}]}`,
	)

	result, err := scraper.Run(context.Background())
	if err != nil {
		log.Fatalf("scrape failed: %v", err)
	}

	var pretty map[string]interface{}
	json.Unmarshal(result, &pretty)
	formatted, _ := json.MarshalIndent(pretty, "", "  ")
	fmt.Println(string(formatted))
}

func testLiveURL(apiKey string) {
	config := models.DefaultConfig()
	config.LLMAPIKey = apiKey
	config.LLMModel = "gpt-4o-mini"
	config.Verbose = true

	scraper := scrapegraph.NewSmartScraperGraph(
		"Extract the main title of the page",
		"https://example.com",
		config,
		"",
	)

	result, err := scraper.Run(context.Background())
	if err != nil {
		log.Fatalf("scrape failed: %v", err)
	}

	var pretty map[string]interface{}
	json.Unmarshal(result, &pretty)
	formatted, _ := json.MarshalIndent(pretty, "", "  ")
	fmt.Println(string(formatted))
}
