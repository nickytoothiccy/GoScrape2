package scrapegraph

import (
	"context"
	"fmt"

	"stealthfetch/internal/models"
	"stealthfetch/pkg/graph"
	"stealthfetch/pkg/llm"
	"stealthfetch/pkg/loaders"
	"stealthfetch/pkg/nodes"
	"stealthfetch/pkg/telemetry"
)

// SearchLinkGraph fetches a page and returns discovered links.
type SearchLinkGraph struct {
	config      *models.Config
	source      string
	baseURL     string
	prompt      string
	maxLinks    int
	filterByLLM bool
	loader      loaders.Loader
	llmClient   llm.LLM
}

// SearchLinkGraphConfig holds graph construction settings.
type SearchLinkGraphConfig struct {
	Config      *models.Config
	Source      string
	BaseURL     string
	Prompt      string
	MaxLinks    int
	FilterByLLM bool
}

// NewSearchLinkGraph creates a new link-discovery graph.
func NewSearchLinkGraph(cfg SearchLinkGraphConfig) *SearchLinkGraph {
	if cfg.Config == nil {
		cfg.Config = models.DefaultConfig()
	}
	if cfg.MaxLinks <= 0 {
		cfg.MaxLinks = 5
	}
	return &SearchLinkGraph{
		config:      cfg.Config,
		source:      cfg.Source,
		baseURL:     cfg.BaseURL,
		prompt:      cfg.Prompt,
		maxLinks:    cfg.MaxLinks,
		filterByLLM: cfg.FilterByLLM,
		loader:      chooseSearchLinkLoader(cfg.Source, cfg.Config),
		llmClient:   llm.NewOpenAIClient(cfg.Config.LLMAPIKey, cfg.Config),
	}
}

// Run executes the link-discovery workflow.
func (g *SearchLinkGraph) Run(ctx context.Context) ([]string, error) {
	start := telemetry.Start()
	var runErr error
	defer func() { telemetry.LogGraph("search_link", start, runErr) }()
	workflow := graph.NewGraph(g.config)
	workflow.AddNode(nodes.NewFetchNode(g.loader, g.config))
	workflow.AddNode(nodes.NewSearchLinkNode(nodes.SearchLinkConfig{
		LLMClient:   g.llmClient,
		Prompt:      g.prompt,
		MaxLinks:    g.maxLinks,
		FilterByLLM: g.filterByLLM,
		Verbose:     g.config.Verbose,
	}))
	if err := workflow.AddEdge("fetch", "search_link"); err != nil {
		runErr = err
		return nil, err
	}
	state := graph.NewState()
	state.Set("url", g.source)
	state.Set("base_url", g.resolveBaseURL())
	if err := workflow.Execute(ctx, state); err != nil {
		runErr = fmt.Errorf("execute: %w", err)
		return nil, runErr
	}
	links, ok := state.GetStringSlice("links")
	if !ok {
		runErr = fmt.Errorf("no links in state")
		return nil, runErr
	}
	return links, nil
}

func (g *SearchLinkGraph) resolveBaseURL() string {
	if g.baseURL != "" {
		return g.baseURL
	}
	return g.source
}

func chooseSearchLinkLoader(source string, config *models.Config) loaders.Loader {
	return loaders.NewFetchLoader(source, config)
}
