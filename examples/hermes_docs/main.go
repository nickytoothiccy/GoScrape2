package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"stealthfetch/internal/envutil"
	"stealthfetch/internal/models"
	"stealthfetch/pkg/scrapegraph"
)

type crawlProfile struct {
	name            string
	maxDepth        int
	maxPages        int
	maxLinksPerPage int
}

func main() {
	_ = envutil.LoadDotEnv(".env")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY env var required")
	}

	config := models.DefaultConfig()
	config.LLMAPIKey = apiKey
	config.LLMModel = "gpt-5.4-mini"
	config.Verbose = true
	config.Headless = true

	base := "https://hermes-agent.nousresearch.com/docs"
	profile := balancedHermesProfile()
	researchPrompt := "Browse this documentation site and aggregate all the information needed for setting up and developing with Hermes Agent. Return JSON with setup steps, installation methods, quickstart workflow, development concepts, important features, and source pages."
	urls := []string{
		base + "/",
		base + "/getting-started/installation",
		base + "/getting-started/quickstart",
		base + "/user-guide/features/memory",
		base + "/user-guide/features/skills",
	}

	// Single-page scrape: prove the docs landing page extracts cleanly.
	single := scrapegraph.NewSmartScraperGraph(
		"Extract the product name, short description, and the quick links as JSON.",
		urls[0],
		config,
		`{"product":"string","summary":"string","quick_links":[{"title":"string","url":"string","description":"string"}]}`,
	)

	singleResult, err := single.Run(context.Background())
	if err != nil {
		log.Fatalf("single page scrape failed: %v", err)
	}

	fmt.Println("=== Single page result ===")
	printJSON(singleResult)

	// Multi-page scrape: summarize key documentation areas in one run.
	multi := scrapegraph.NewSmartScraperMultiGraph(scrapegraph.SmartScraperMultiConfig{
		Config:        config,
		Prompt:        "Extract the page title, the main purpose of the page, and 3-5 key takeaways as JSON.",
		SchemaHint:    `{"title":"string","purpose":"string","takeaways":["string"]}`,
		URLs:          urls,
		ConcatResults: true,
	})

	multiResult, err := multi.Run(context.Background())
	if err != nil {
		log.Fatalf("multi page scrape failed: %v", err)
	}

	fmt.Println("\n=== Multi page result ===")
	printJSON(multiResult)

	if failed := multi.GetFailedURLs(); len(failed) > 0 {
		fmt.Printf("\nfailed URLs: %v\n", failed)
	}

	// Site-wide research scrape: use the new high-level orchestration API.
	fmt.Printf("\n=== ResearchGraph profile: %s (depth=%d, max_pages=%d, max_links_per_page=%d) ===\n", profile.name, profile.maxDepth, profile.maxPages, profile.maxLinksPerPage)
	research := scrapegraph.NewResearchGraph(config)
	researchResult, err := research.Run(context.Background(), scrapegraph.ResearchRequest{
		Prompt:             researchPrompt,
		SeedURL:            base,
		SchemaHint:         `{"setup":{"installation_methods":["string"],"prerequisites":["string"],"steps":["string"]},"quickstart":{"summary":"string","steps":["string"]},"development":{"core_concepts":["string"],"features":["string"],"workflows":["string"]},"recommended_pages":[{"title":"string","url":"string","reason":"string"}]}`,
		FollowSubpages:     true,
		MaxDepth:           profile.maxDepth,
		MaxPages:           profile.maxPages,
		MaxLinksPerPage:    profile.maxLinksPerPage,
		RestrictToSeedHost: true,
		PathPrefixes:       []string{"/docs"},
		ExcludePatterns:    []string{"changelog", "blog", "github"},
	})
	if err != nil {
		log.Fatalf("research scrape failed: %v", err)
	}

	fmt.Println("\n=== ResearchGraph site-wide result ===")
	printJSON(researchResult.Result)
	fmt.Printf("mode: %s\n", researchResult.Mode)
	fmt.Printf("pages used: %d\n", researchResult.PagesUsed)
	if len(researchResult.Sources) > 0 {
		fmt.Printf("sources: %v\n", researchResult.Sources)
	}
	if len(researchResult.FailedURLs) > 0 {
		fmt.Printf("failed URLs: %v\n", researchResult.FailedURLs)
	}

	outputPath := filepath.Join("examples", "hermes_docs", "output.json")
	if err := writeJSONFile(outputPath, researchResult.Result); err != nil {
		log.Fatalf("failed to write output file: %v", err)
	}
	fmt.Printf("saved ResearchGraph result to %s\n", outputPath)
}

func printJSON(raw json.RawMessage) {
	var out bytes.Buffer
	if err := json.Indent(&out, raw, "", "  "); err != nil {
		fmt.Println(string(raw))
		return
	}
	fmt.Println(out.String())
}

func writeJSONFile(path string, raw json.RawMessage) error {
	var out bytes.Buffer
	if err := json.Indent(&out, raw, "", "  "); err != nil {
		out.Write(raw)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, out.Bytes(), 0o644)
}

func balancedHermesProfile() crawlProfile {
	return crawlProfile{
		name:            "balanced",
		maxDepth:        2,
		maxPages:        20,
		maxLinksPerPage: 10,
	}
}
