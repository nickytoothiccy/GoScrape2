package scrapegraph

import (
	"context"
	"encoding/json"

	"stealthfetch/internal/models"
	"stealthfetch/pkg/graph"
	"stealthfetch/pkg/llm"
	"stealthfetch/pkg/loaders"
	"stealthfetch/pkg/nodes"
)

// CSVScraperGraph extracts structured answers from CSV files or directories.
type CSVScraperGraph struct{ structuredScraper }

// NewCSVScraperGraph creates a new CSV scraper graph.
func NewCSVScraperGraph(prompt, source string, config *models.Config, schemaHint string) *CSVScraperGraph {
	if config == nil {
		config = models.DefaultConfig()
	}
	return &CSVScraperGraph{structuredScraper{config: config, prompt: prompt, schemaHint: schemaHint, source: source, name: "csv_scraper", loader: loaders.NewCSVLoader(), newGenNode: func(c llm.LLM, p, s string) graph.Node { return nodes.NewGenerateAnswerCSVNode(c, p, s) }}}
}

// Run executes the CSV extraction workflow.
func (g *CSVScraperGraph) Run(ctx context.Context) (json.RawMessage, error) {
	return g.structuredScraper.run(ctx)
}
