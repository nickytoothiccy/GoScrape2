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

// JSONScraperGraph extracts structured answers from JSON files or directories.
type JSONScraperGraph struct{ structuredScraper }

// NewJSONScraperGraph creates a new JSON scraper graph.
func NewJSONScraperGraph(prompt, source string, config *models.Config, schemaHint string) *JSONScraperGraph {
	if config == nil {
		config = models.DefaultConfig()
	}
	return &JSONScraperGraph{structuredScraper{config: config, prompt: prompt, schemaHint: schemaHint, source: source, name: "json_scraper", loader: loaders.NewJSONLoader(), newGenNode: func(c llm.LLM, p, s string) graph.Node { return nodes.NewGenerateAnswerNode(c, p, s) }}}
}

// Run executes the JSON extraction workflow.
func (g *JSONScraperGraph) Run(ctx context.Context) (json.RawMessage, error) {
	return g.structuredScraper.run(ctx)
}
