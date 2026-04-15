// Package scrapegraph provides high-level scraping workflows
package scrapegraph

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"stealthfetch/internal/models"
	"stealthfetch/pkg/graph"
	"stealthfetch/pkg/llm"
	"stealthfetch/pkg/loaders"
	"stealthfetch/pkg/nodes"
	"stealthfetch/pkg/telemetry"
)

// SmartScraperGraph implements the ScrapeGraphAI SmartScraperGraph workflow
type SmartScraperGraph struct {
	config     *models.Config
	prompt     string
	schemaHint string
	source     string
	maxRetries int
}

// NewSmartScraperGraph creates a new smart scraper
func NewSmartScraperGraph(prompt, source string, config *models.Config, schemaHint string) *SmartScraperGraph {
	if config == nil {
		config = models.DefaultConfig()
	}

	return &SmartScraperGraph{
		config:     config,
		prompt:     prompt,
		schemaHint: schemaHint,
		source:     source,
		maxRetries: 2,
	}
}

// Run executes the scraping workflow with retry logic
func (s *SmartScraperGraph) Run(ctx context.Context) (json.RawMessage, error) {
	start := telemetry.Start()
	var runErr error
	defer func() { telemetry.LogGraph("smart_scraper", start, runErr) }()
	var lastErr error

	for attempt := 0; attempt <= s.maxRetries; attempt++ {
		if attempt > 0 && s.config.Verbose {
			log.Printf("[smart_scraper] retry attempt %d/%d", attempt, s.maxRetries)
		}

		result, err := s.executeOnce(ctx)
		if err != nil {
			lastErr = err
			runErr = err
			continue
		}

		// Validate the result
		if s.isValidResult(result) {
			runErr = nil
			return result, nil
		}

		if s.config.Verbose {
			log.Printf("[smart_scraper] invalid result on attempt %d, retrying", attempt+1)
		}
		lastErr = fmt.Errorf("extraction returned invalid/empty result")
		runErr = lastErr
	}

	runErr = fmt.Errorf("all attempts failed: %w", lastErr)
	return nil, fmt.Errorf("all attempts failed: %w", lastErr)
}

// executeOnce runs a single extraction attempt
func (s *SmartScraperGraph) executeOnce(ctx context.Context) (json.RawMessage, error) {
	g := graph.NewGraph(s.config)

	// Create LLM client
	llmClient := llm.NewOpenAIClient(s.config.LLMAPIKey, s.config)

	loader := loaders.NewFetchLoader(s.source, s.config)

	// Create nodes
	fetchNode := nodes.NewFetchNode(loader, s.config)
	parseNode := nodes.NewParseNode(s.config)
	generateNode := nodes.NewGenerateAnswerNode(llmClient, s.prompt, s.schemaHint)

	// Add nodes
	g.AddNode(fetchNode)
	g.AddNode(parseNode)
	g.AddNode(generateNode)

	// Add edges
	if err := g.AddEdge("fetch", "parse"); err != nil {
		return nil, err
	}
	if err := g.AddEdge("parse", "generate_answer"); err != nil {
		return nil, err
	}

	// Create initial state
	state := graph.NewState()
	state.Set("url", s.source)

	// Execute
	if err := g.Execute(ctx, state); err != nil {
		return nil, fmt.Errorf("execute: %w", err)
	}

	// Retrieve result
	result, ok := state.GetJSON("extracted_data")
	if !ok {
		return nil, fmt.Errorf("no extracted data in state")
	}

	return result, nil
}

// isValidResult checks if the extraction result is meaningful
func (s *SmartScraperGraph) isValidResult(data json.RawMessage) bool {
	if len(data) == 0 {
		return false
	}

	str := string(data)

	// Check for error responses from the LLM
	if strings.Contains(str, `"error"`) && strings.Contains(str, `"not_found"`) {
		return false
	}

	// Check for empty objects/arrays
	trimmed := strings.TrimSpace(str)
	if trimmed == "{}" || trimmed == "[]" || trimmed == "null" {
		return false
	}

	return true
}
