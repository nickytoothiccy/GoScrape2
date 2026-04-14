package scrapegraph

import (
	"context"
	"encoding/json"
	"time"

	"stealthfetch/internal/models"
	"stealthfetch/pkg/utils"
)

type ResearchRequest struct {
	Prompt             string
	SeedURL            string
	SchemaHint         string
	SearchFirst        bool
	FollowSubpages     bool
	MaxDepth           int
	MaxPages           int
	MaxResults         int
	MaxLinksPerPage    int
	PerURLTimeout      time.Duration
	AllowedDomains     []string
	RestrictToSeedHost bool
	PathPrefixes       []string
	IncludePatterns    []string
	ExcludePatterns    []string
}

type ResearchResult struct {
	Mode        string          `json:"mode"`
	Result      json.RawMessage `json:"result"`
	Sources     []string        `json:"sources,omitempty"`
	FailedURLs  []string        `json:"failed_urls,omitempty"`
	SearchQuery string          `json:"search_query,omitempty"`
	PagesUsed   int             `json:"pages_used"`
}

type ResearchGraph struct {
	config     *models.Config
	searchFunc utils.SearchFunc
	runSmart   func(context.Context, *models.Config, ResearchRequest) (json.RawMessage, []string, []string, error)
	runDepth   func(context.Context, *models.Config, ResearchRequest) (json.RawMessage, []string, []string, error)
	runSearch  func(context.Context, *models.Config, ResearchRequest, utils.SearchFunc) (json.RawMessage, []string, string, error)
}

func NewResearchGraph(config *models.Config) *ResearchGraph {
	if config == nil {
		config = models.DefaultConfig()
	}
	return &ResearchGraph{
		config:    config,
		searchFunc: nil,
		runSmart:  runResearchSmart,
		runDepth:  runResearchDepth,
		runSearch: runResearchSearch,
	}
}

func (r *ResearchGraph) Run(ctx context.Context, req ResearchRequest) (*ResearchResult, error) {
	if req.SeedURL != "" && req.FollowSubpages {
		data, urls, failed, err := r.runDepth(ctx, r.config, req)
		if err != nil {
			return nil, err
		}
		return &ResearchResult{Mode: "depth", Result: data, Sources: urls, FailedURLs: failed, PagesUsed: len(urls)}, nil
	}
	if req.SeedURL != "" && !req.SearchFirst {
		data, urls, failed, err := r.runSmart(ctx, r.config, req)
		if err != nil {
			return nil, err
		}
		return &ResearchResult{Mode: "direct", Result: data, Sources: urls, FailedURLs: failed, PagesUsed: len(urls)}, nil
	}
	data, urls, query, err := r.runSearch(ctx, r.config, req, r.searchFunc)
	if err != nil {
		return nil, err
	}
	return &ResearchResult{Mode: "search", Result: data, Sources: urls, SearchQuery: query, PagesUsed: len(urls)}, nil
}

func runResearchSmart(ctx context.Context, cfg *models.Config, req ResearchRequest) (json.RawMessage, []string, []string, error) {
	scraper := NewSmartScraperGraph(req.Prompt, req.SeedURL, cfg, req.SchemaHint)
	data, err := scraper.Run(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	return data, []string{req.SeedURL}, nil, nil
}

func runResearchDepth(ctx context.Context, cfg *models.Config, req ResearchRequest) (json.RawMessage, []string, []string, error) {
	graph := NewDepthSearchGraph(DepthSearchConfig{
		Config:          cfg,
		Prompt:          req.Prompt,
		SchemaHint:      req.SchemaHint,
		Source:          req.SeedURL,
		MaxDepth:        req.MaxDepth,
		MaxLinksPerPage: req.MaxLinksPerPage,
		PerURLTimeout:   req.PerURLTimeout,
		AllowedDomains:  req.AllowedDomains,
		RestrictToHost:  req.RestrictToSeedHost,
		PathPrefixes:    req.PathPrefixes,
		IncludePatterns: req.IncludePatterns,
		ExcludePatterns: req.ExcludePatterns,
		MaxPages:        req.MaxPages,
	})
	data, err := graph.Run(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	return data, graph.GetVisitedURLs(), nil, nil
}

func runResearchSearch(ctx context.Context, cfg *models.Config, req ResearchRequest, searchFunc utils.SearchFunc) (json.RawMessage, []string, string, error) {
	graph := NewSearchGraph(SearchGraphConfig{
		Config:        cfg,
		Prompt:        req.Prompt,
		SchemaHint:    req.SchemaHint,
		MaxResults:    req.MaxResults,
		SearchFunc:    searchFunc,
		PerURLTimeout: req.PerURLTimeout,
	})
	data, err := graph.Run(ctx)
	if err != nil {
		return nil, nil, "", err
	}
	return data, graph.GetConsideredURLs(), extractSearchQuery(data), nil
}

func extractSearchQuery(data json.RawMessage) string {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return ""
	}
	if query, ok := payload["search_query"].(string); ok {
		return query
	}
	return ""
}