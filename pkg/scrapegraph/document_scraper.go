package scrapegraph

import (
	"context"
	"encoding/json"
	"fmt"

	"stealthfetch/internal/models"
	"stealthfetch/pkg/graph"
	"stealthfetch/pkg/llm"
	"stealthfetch/pkg/loaders"
	"stealthfetch/pkg/nodes"
	"stealthfetch/pkg/telemetry"
)

// DocumentScraperGraph extracts structured data from local PDF/DOCX files.
type DocumentScraperGraph struct {
	config     *models.Config
	prompt     string
	schemaHint string
	source     string
}

// NewDocumentScraperGraph creates a new document scraper graph.
func NewDocumentScraperGraph(prompt, source string, config *models.Config, schemaHint string) *DocumentScraperGraph {
	if config == nil {
		config = models.DefaultConfig()
	}
	return &DocumentScraperGraph{config: config, prompt: prompt, schemaHint: schemaHint, source: source}
}

// Run executes the document extraction workflow.
func (d *DocumentScraperGraph) Run(ctx context.Context) (json.RawMessage, error) {
	start := telemetry.Start()
	var runErr error
	defer func() { telemetry.LogGraph("document_scraper", start, runErr) }()
	g := graph.NewGraph(d.config)
	llmClient := llm.NewOpenAIClient(d.config.LLMAPIKey, d.config)
	fetchNode := nodes.NewFetchNode(loaders.NewDocumentLoader(), d.config)
	parseNode := nodes.NewParseNode(d.config)
	generateNode := nodes.NewGenerateAnswerNode(llmClient, d.prompt, d.schemaHint)
	g.AddNode(fetchNode)
	g.AddNode(parseNode)
	g.AddNode(generateNode)
	if err := g.AddEdge("fetch", "parse"); err != nil {
		runErr = err
		return nil, err
	}
	if err := g.AddEdge("parse", "generate_answer"); err != nil {
		runErr = err
		return nil, err
	}
	state := graph.NewState()
	state.Set("url", d.source)
	if err := g.Execute(ctx, state); err != nil {
		runErr = fmt.Errorf("execute: %w", err)
		return nil, runErr
	}
	result, ok := state.GetJSON("extracted_data")
	if !ok {
		runErr = fmt.Errorf("no extracted data in state")
		return nil, runErr
	}
	return result, nil
}
