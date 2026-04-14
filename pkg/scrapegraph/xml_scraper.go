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

// XMLScraperGraph extracts structured answers from XML files or directories.
type XMLScraperGraph struct{ structuredScraper }

// NewXMLScraperGraph creates a new XML scraper graph.
func NewXMLScraperGraph(prompt, source string, config *models.Config, schemaHint string) *XMLScraperGraph {
	if config == nil {
		config = models.DefaultConfig()
	}
	return &XMLScraperGraph{structuredScraper{config: config, prompt: prompt, schemaHint: schemaHint, source: source, name: "xml_scraper", loader: loaders.NewXMLLoader(), newGenNode: func(c llm.LLM, p, s string) graph.Node { return nodes.NewGenerateAnswerNode(c, p, s) }}}
}

// Run executes the XML extraction workflow.
func (g *XMLScraperGraph) Run(ctx context.Context) (json.RawMessage, error) {
	return g.structuredScraper.run(ctx)
}
