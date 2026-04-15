// Package scrapegraph provides the DepthSearchGraph for recursive crawling
package scrapegraph

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"stealthfetch/internal/models"
	"stealthfetch/pkg/graph"
	"stealthfetch/pkg/llm"
	"stealthfetch/pkg/loaders"
	"stealthfetch/pkg/nodes"
)

// DepthSearchGraph recursively follows links and extracts data.
// Workflow per depth level: Fetch → ExtractLinks → [SmartScraper per link] → collect
// After all levels: MergeAnswers
type DepthSearchGraph struct {
	config          *models.Config
	prompt          string
	schemaHint      string
	source          string // seed URL
	maxDepth        int
	maxLinksPerPage int
	filterByLLM     bool
	perURLTimeout   time.Duration
	timeout         time.Duration
	verbose         bool
	visitedURLs     map[string]bool
	allResults      []json.RawMessage
	allURLs         []string
	allowedDomains  []string
	restrictHost    bool
	pathPrefixes    []string
	includePatterns []string
	excludePatterns []string
	maxPages        int
}

// DepthSearchConfig holds configuration for creating a DepthSearchGraph
type DepthSearchConfig struct {
	Config          *models.Config
	Prompt          string
	SchemaHint      string
	Source          string
	MaxDepth        int           // default 2
	MaxLinksPerPage int           // default 3
	FilterByLLM     bool          // use LLM to filter links, default true
	PerURLTimeout   time.Duration // default 60s
	Timeout         time.Duration // overall timeout, 0 = no limit
	AllowedDomains  []string
	RestrictToHost  bool
	PathPrefixes    []string
	IncludePatterns []string
	ExcludePatterns []string
	MaxPages        int
}

// NewDepthSearchGraph creates a new depth search graph
func NewDepthSearchGraph(cfg DepthSearchConfig) *DepthSearchGraph {
	if cfg.Config == nil {
		cfg.Config = models.DefaultConfig()
	}
	if cfg.MaxDepth <= 0 {
		cfg.MaxDepth = 2
	}
	if cfg.MaxLinksPerPage <= 0 {
		cfg.MaxLinksPerPage = 3
	}
	if cfg.PerURLTimeout == 0 {
		cfg.PerURLTimeout = 60 * time.Second
	}

	return &DepthSearchGraph{
		config:          cfg.Config,
		prompt:          cfg.Prompt,
		schemaHint:      cfg.SchemaHint,
		source:          cfg.Source,
		maxDepth:        cfg.MaxDepth,
		maxLinksPerPage: cfg.MaxLinksPerPage,
		filterByLLM:     cfg.FilterByLLM,
		perURLTimeout:   cfg.PerURLTimeout,
		timeout:         cfg.Timeout,
		verbose:         cfg.Config.Verbose,
		visitedURLs:     make(map[string]bool),
		allowedDomains:  cfg.AllowedDomains,
		restrictHost:    cfg.RestrictToHost,
		pathPrefixes:    cfg.PathPrefixes,
		includePatterns: cfg.IncludePatterns,
		excludePatterns: cfg.ExcludePatterns,
		maxPages:        cfg.MaxPages,
	}
}

// Run executes the depth search workflow
func (d *DepthSearchGraph) Run(ctx context.Context) (json.RawMessage, error) {
	if d.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, d.timeout)
		defer cancel()
	}

	if d.verbose {
		log.Printf("[depth_search] starting from %s, max depth %d", d.source, d.maxDepth)
	}

	// Crawl recursively starting from seed URL
	d.crawl(ctx, d.source, 0)

	if len(d.allResults) == 0 {
		return nil, fmt.Errorf("depth_search: no results extracted from any page")
	}

	if d.verbose {
		log.Printf("[depth_search] collected %d results from %d pages", len(d.allResults), len(d.allURLs))
	}

	// Merge all results
	return d.mergeResults(ctx)
}

// GetVisitedURLs returns all URLs that were visited
func (d *DepthSearchGraph) GetVisitedURLs() []string {
	return d.allURLs
}

// crawl recursively fetches and extracts from a URL, then follows links
func (d *DepthSearchGraph) crawl(ctx context.Context, pageURL string, depth int) {
	if d.maxPages > 0 && len(d.allURLs) >= d.maxPages {
		return
	}

	// Check context
	select {
	case <-ctx.Done():
		return
	default:
	}

	normalized := normalizeURL(pageURL)
	if !d.shouldVisitURL(normalized) {
		return
	}

	// Check if already visited
	if d.visitedURLs[normalized] {
		return
	}
	d.visitedURLs[normalized] = true

	if d.verbose {
		log.Printf("[depth_search] depth %d: scraping %s", depth, normalized)
	}

	// Scrape this page
	urlCtx, cancel := context.WithTimeout(ctx, d.perURLTimeout)
	scraper := NewSmartScraperGraph(d.prompt, normalized, d.config, d.schemaHint)
	result, err := scraper.Run(urlCtx)
	cancel()

	if err != nil {
		if d.verbose {
			log.Printf("[depth_search] depth %d: scrape failed for %s: %v", depth, normalized, err)
		}
	} else {
		d.allResults = append(d.allResults, result)
		d.allURLs = append(d.allURLs, normalized)
	}

	// Stop if we've reached max depth
	if depth >= d.maxDepth {
		return
	}

	// Find links on this page
	links, err := d.discoverLinks(ctx, normalized)
	if err != nil {
		if d.verbose {
			log.Printf("[depth_search] depth %d: link discovery failed for %s: %v", depth, normalized, err)
		}
		return
	}

	if d.verbose {
		log.Printf("[depth_search] depth %d: found %d links to follow from %s", depth, len(links), normalized)
	}

	// Recursively crawl each discovered link
	for _, link := range links {
		d.crawl(ctx, link, depth+1)
	}
}

// discoverLinks fetches a page and extracts links using SearchLinkNode
func (d *DepthSearchGraph) discoverLinks(ctx context.Context, pageURL string) ([]string, error) {
	loader := loaders.NewFetchLoader(pageURL, d.config)

	// Fetch the page
	fetchResult, err := loader.Load(ctx, pageURL)
	if err != nil {
		return nil, fmt.Errorf("fetch for links: %w", err)
	}

	// Use SearchLinkNode to extract links
	llmClient := llm.NewOpenAIClient(d.config.LLMAPIKey, d.config)
	linkNode := nodes.NewSearchLinkNode(nodes.SearchLinkConfig{
		LLMClient:   llmClient,
		Prompt:      d.prompt,
		MaxLinks:    d.maxLinksPerPage,
		FilterByLLM: d.filterByLLM,
		Verbose:     d.verbose,
	})

	state := graph.NewState()
	state.Set("html", fetchResult.HTML)
	state.Set("url", pageURL)

	if err := linkNode.Execute(ctx, state); err != nil {
		return nil, err
	}

	links, ok := state.GetStringSlice("links")
	if !ok {
		return nil, nil
	}

	return links, nil
}

// mergeResults combines all extraction results
func (d *DepthSearchGraph) mergeResults(ctx context.Context) (json.RawMessage, error) {
	if len(d.allResults) == 1 {
		return d.attachMeta(d.allResults[0]), nil
	}

	llmClient := llm.NewOpenAIClient(d.config.LLMAPIKey, d.config)
	mergeNode := nodes.NewMergeAnswersNode(llmClient, d.prompt, d.verbose)

	state := graph.NewState()
	state.Set("answers", d.allResults)
	state.Set("urls", d.allURLs)

	if err := mergeNode.Execute(ctx, state); err != nil {
		return nil, fmt.Errorf("merge: %w", err)
	}

	result, ok := state.GetJSON("extracted_data")
	if !ok {
		return nil, fmt.Errorf("no merged data in state")
	}

	return d.attachMeta(result), nil
}

// attachMeta adds crawl metadata to the result
func (d *DepthSearchGraph) attachMeta(data json.RawMessage) json.RawMessage {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return data
	}

	result["sources"] = d.allURLs
	result["pages_crawled"] = len(d.allURLs)
	result["max_depth"] = d.maxDepth
	result["max_pages"] = d.maxPages

	enriched, err := json.Marshal(result)
	if err != nil {
		return data
	}
	return json.RawMessage(enriched)
}
