package scrapegraph

import (
	"context"
	"fmt"

	"stealthfetch/internal/models"
	"stealthfetch/pkg/graph"
	"stealthfetch/pkg/loaders"
	"stealthfetch/pkg/nodes"
	"stealthfetch/pkg/telemetry"
)

// MarkdownifyGraph converts a source page or HTML snippet into Markdown.
type MarkdownifyGraph struct {
	config *models.Config
	source string
	loader loaders.Loader
}

// NewMarkdownifyGraph creates a new Markdownify graph.
func NewMarkdownifyGraph(source string, config *models.Config) *MarkdownifyGraph {
	if config == nil {
		config = models.DefaultConfig()
	}
	return &MarkdownifyGraph{config: config, source: source, loader: chooseMarkdownLoader(source, config)}
}

// Run executes the Markdown conversion workflow.
func (m *MarkdownifyGraph) Run(ctx context.Context) (string, error) {
	start := telemetry.Start()
	var runErr error
	defer func() { telemetry.LogGraph("markdownify", start, runErr) }()
	g := graph.NewGraph(m.config)
	fetchNode := nodes.NewFetchNode(m.loader, m.config)
	markdownNode := nodes.NewMarkdownifyNode(m.config)
	g.AddNode(fetchNode)
	g.AddNode(markdownNode)
	if err := g.AddEdge("fetch", "markdownify"); err != nil {
		runErr = err
		return "", err
	}
	state := graph.NewState()
	state.Set("url", m.source)
	if err := g.Execute(ctx, state); err != nil {
		runErr = fmt.Errorf("execute: %w", err)
		return "", runErr
	}
	result, ok := state.GetString("markdown")
	if !ok {
		runErr = fmt.Errorf("no markdown in state")
		return "", runErr
	}
	return result, nil
}

func chooseMarkdownLoader(source string, config *models.Config) loaders.Loader {
	return loaders.NewFetchLoader(source, config)
}
