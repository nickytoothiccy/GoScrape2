// Package scrapegraph provides multi-URL scraping workflows.
package scrapegraph

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"stealthfetch/internal/models"
	"stealthfetch/pkg/graph"
	"stealthfetch/pkg/nodes"
)

// SmartScraperMultiGraph runs SmartScraperGraph across multiple URLs.
// It can return separate per-URL results or one concatenated JSON array.
type SmartScraperMultiGraph struct {
	config        *models.Config
	prompt        string
	schemaHint    string
	urls          []string
	perURLTimeout time.Duration
	concatResults bool
	failedURLs    []string
}

// SmartScraperMultiConfig configures SmartScraperMultiGraph.
type SmartScraperMultiConfig struct {
	Config        *models.Config
	Prompt        string
	SchemaHint    string
	URLs          []string
	PerURLTimeout time.Duration
	ConcatResults bool
}

// NewSmartScraperMultiGraph creates a multi-URL scraping workflow.
func NewSmartScraperMultiGraph(cfg SmartScraperMultiConfig) *SmartScraperMultiGraph {
	if cfg.Config == nil {
		cfg.Config = models.DefaultConfig()
	}
	if cfg.PerURLTimeout == 0 {
		cfg.PerURLTimeout = 60 * time.Second
	}

	return &SmartScraperMultiGraph{
		config:        cfg.Config,
		prompt:        cfg.Prompt,
		schemaHint:    cfg.SchemaHint,
		urls:          cfg.URLs,
		perURLTimeout: cfg.PerURLTimeout,
		concatResults: cfg.ConcatResults,
	}
}

// Run executes the multi-URL scraping workflow.
func (s *SmartScraperMultiGraph) Run(ctx context.Context) (json.RawMessage, error) {
	state := graph.NewState()
	state.Set("urls", s.urls)

	iterator := nodes.NewGraphIteratorNode(nodes.GraphIteratorConfig{
		InputKey:       "urls",
		OutputKey:      "answers",
		PerItemTimeout: s.perURLTimeout,
		Verbose:        s.config.Verbose,
		Factory: func(ctx context.Context, url string) (json.RawMessage, error) {
			scraper := NewSmartScraperGraph(s.prompt, url, s.config, s.schemaHint)
			return scraper.Run(ctx)
		},
	})

	if err := iterator.Execute(ctx, state); err != nil {
		return nil, fmt.Errorf("smart_scraper_multi: iterate urls: %w", err)
	}

	if failed, ok := state.GetStringSlice("failed_items"); ok {
		s.failedURLs = failed
	}

	if s.concatResults {
		concat := nodes.NewConcatAnswersNode(nodes.ConcatAnswersConfig{})
		if err := concat.Execute(ctx, state); err != nil {
			return nil, fmt.Errorf("smart_scraper_multi: concat results: %w", err)
		}
		result, ok := state.GetJSON("extracted_data")
		if !ok {
			return nil, fmt.Errorf("smart_scraper_multi: missing concatenated output")
		}
		return result, nil
	}

	answers, ok := state.Get("answers")
	if !ok {
		return nil, fmt.Errorf("smart_scraper_multi: missing answers in state")
	}

	raw, err := json.Marshal(answers)
	if err != nil {
		return nil, fmt.Errorf("smart_scraper_multi: marshal answers: %w", err)
	}
	return json.RawMessage(raw), nil
}

// GetURLs returns the configured URLs.
func (s *SmartScraperMultiGraph) GetURLs() []string { return s.urls }

// GetFailedURLs returns any URLs that failed during execution.
func (s *SmartScraperMultiGraph) GetFailedURLs() []string { return s.failedURLs }
