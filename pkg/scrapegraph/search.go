// Package scrapegraph provides the SearchGraph workflow
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
	"stealthfetch/pkg/nodes"
	"stealthfetch/pkg/utils"
)

// SearchGraph searches the internet for information and extracts data
// Workflow: SearchInternet → [SmartScraper per URL] → MergeAnswers
type SearchGraph struct {
	config        *models.Config
	prompt        string
	schemaHint    string
	maxResults    int
	searchFunc    utils.SearchFunc
	timeout       time.Duration
	perURLTimeout time.Duration
	urls          []string // discovered URLs after execution
}

// SearchGraphConfig holds configuration for creating a SearchGraph
type SearchGraphConfig struct {
	Config        *models.Config
	Prompt        string
	SchemaHint    string
	MaxResults    int              // default 3
	SearchFunc    utils.SearchFunc // nil defaults to DuckDuckGo
	Timeout       time.Duration    // overall timeout, 0 = no limit
	PerURLTimeout time.Duration    // per-URL scrape timeout, default 60s
}

// NewSearchGraph creates a new search graph
func NewSearchGraph(cfg SearchGraphConfig) *SearchGraph {
	if cfg.Config == nil {
		cfg.Config = models.DefaultConfig()
	}
	if cfg.MaxResults <= 0 {
		cfg.MaxResults = 3
	}
	if cfg.PerURLTimeout == 0 {
		cfg.PerURLTimeout = 60 * time.Second
	}

	return &SearchGraph{
		config:        cfg.Config,
		prompt:        cfg.Prompt,
		schemaHint:    cfg.SchemaHint,
		maxResults:    cfg.MaxResults,
		searchFunc:    cfg.SearchFunc,
		timeout:       cfg.Timeout,
		perURLTimeout: cfg.PerURLTimeout,
	}
}

// Run executes the search graph workflow
func (s *SearchGraph) Run(ctx context.Context) (json.RawMessage, error) {
	// Apply overall timeout if configured
	if s.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.timeout)
		defer cancel()
	}

	llmClient := llm.NewOpenAIClient(s.config.LLMAPIKey, s.config)

	// Phase 1: Search the internet for relevant URLs
	urls, err := s.searchPhase(ctx, llmClient)
	if err != nil {
		return nil, fmt.Errorf("search phase: %w", err)
	}
	s.urls = urls

	if s.config.Verbose {
		log.Printf("[search_graph] found %d URLs to scrape", len(urls))
	}

	// Phase 2: Scrape each URL using SmartScraperGraph
	answers, err := s.scrapePhase(ctx, urls)
	if err != nil {
		return nil, fmt.Errorf("scrape phase: %w", err)
	}

	// Phase 3: Merge all answers
	result, err := s.mergePhase(ctx, llmClient, answers, urls)
	if err != nil {
		return nil, fmt.Errorf("merge phase: %w", err)
	}

	return result, nil
}

// GetConsideredURLs returns URLs discovered during the search
func (s *SearchGraph) GetConsideredURLs() []string {
	return s.urls
}

// searchPhase uses SearchInternetNode to find relevant URLs
func (s *SearchGraph) searchPhase(ctx context.Context, llmClient llm.LLM) ([]string, error) {
	searchNode := nodes.NewSearchInternetNode(nodes.SearchInternetConfig{
		LLMClient:  llmClient,
		SearchFunc: s.searchFunc,
		MaxResults: s.maxResults,
		Verbose:    s.config.Verbose,
	})

	state := graph.NewState()
	state.Set("user_prompt", s.prompt)

	if err := searchNode.Execute(ctx, state); err != nil {
		return nil, err
	}

	urls, ok := state.GetStringSlice("urls")
	if !ok || len(urls) == 0 {
		return nil, fmt.Errorf("no URLs found from search")
	}

	return urls, nil
}

// scrapePhase runs SmartScraperGraph for each URL with per-URL timeouts
func (s *SearchGraph) scrapePhase(ctx context.Context, urls []string) ([]json.RawMessage, error) {
	var answers []json.RawMessage

	for i, url := range urls {
		// Check if overall context is already done
		select {
		case <-ctx.Done():
			if s.config.Verbose {
				log.Printf("[search_graph] context cancelled, stopping at URL %d/%d", i+1, len(urls))
			}
			if len(answers) > 0 {
				return answers, nil // return what we have
			}
			return nil, ctx.Err()
		default:
		}

		if s.config.Verbose {
			log.Printf("[search_graph] scraping URL %d/%d: %s", i+1, len(urls), url)
		}

		// Per-URL timeout
		urlCtx, urlCancel := context.WithTimeout(ctx, s.perURLTimeout)
		scraper := NewSmartScraperGraph(s.prompt, url, s.config, s.schemaHint)

		result, err := scraper.Run(urlCtx)
		urlCancel()

		if err != nil {
			if s.config.Verbose {
				log.Printf("[search_graph] URL %d failed: %v", i+1, err)
			}
			continue // skip failed URLs
		}

		answers = append(answers, result)
	}

	if len(answers) == 0 {
		return nil, fmt.Errorf("all %d URLs failed to scrape", len(urls))
	}

	return answers, nil
}

// mergePhase combines results from multiple sources
func (s *SearchGraph) mergePhase(ctx context.Context, llmClient llm.LLM, answers []json.RawMessage, urls []string) (json.RawMessage, error) {
	// Single result: return directly with sources
	if len(answers) == 1 {
		return s.attachSources(answers[0], urls), nil
	}

	// Multiple results: use MergeAnswersNode
	mergeNode := nodes.NewMergeAnswersNode(llmClient, s.prompt, s.config.Verbose)

	state := graph.NewState()
	state.Set("answers", answers)
	state.Set("urls", urls)

	if err := mergeNode.Execute(ctx, state); err != nil {
		return nil, err
	}

	result, ok := state.GetJSON("extracted_data")
	if !ok {
		return nil, fmt.Errorf("no merged data in state")
	}

	return result, nil
}

// attachSources adds source URLs to a result
func (s *SearchGraph) attachSources(data json.RawMessage, urls []string) json.RawMessage {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return data
	}

	result["sources"] = urls
	enriched, err := json.Marshal(result)
	if err != nil {
		return data
	}
	return json.RawMessage(enriched)
}
