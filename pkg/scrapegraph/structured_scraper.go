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

type structuredScraper struct {
	config     *models.Config
	prompt     string
	schemaHint string
	source     string
	name       string
	loader     loaders.Loader
	newGenNode func(llm.LLM, string, string) graph.Node
}

func (s *structuredScraper) run(ctx context.Context) (json.RawMessage, error) {
	start := telemetry.Start()
	var runErr error
	defer func() { telemetry.LogGraph(s.name, start, runErr) }()
	g := graph.NewGraph(s.config)
	llmClient := llm.NewOpenAIClient(s.config.LLMAPIKey, s.config)
	fetchNode := nodes.NewFetchNode(s.loader, s.config)
	parseNode := nodes.NewParseNode(s.config)
	generateNode := s.newGenNode(llmClient, s.prompt, s.schemaHint)
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
	state.Set("url", s.source)
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
