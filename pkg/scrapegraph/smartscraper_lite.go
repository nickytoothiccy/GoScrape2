package scrapegraph

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"stealthfetch/internal/models"
	"stealthfetch/pkg/graph"
	"stealthfetch/pkg/llm"
	"stealthfetch/pkg/loaders"
	"stealthfetch/pkg/nodes"
	"stealthfetch/pkg/telemetry"
)

const (
	defaultLiteHTMLMaxChars = 15000
	defaultLiteChunkSize    = 4000
)

// SmartScraperLiteGraph is a lighter-weight single-page extraction workflow.
type SmartScraperLiteGraph struct {
	config     *models.Config
	prompt     string
	schemaHint string
	source     string
	loader     loaders.Loader
	llmClient  llm.LLM
}

// NewSmartScraperLiteGraph creates a new lightweight scraper graph.
func NewSmartScraperLiteGraph(prompt, source string, config *models.Config, schemaHint string) *SmartScraperLiteGraph {
	config = buildLiteConfig(config)
	return &SmartScraperLiteGraph{
		config:     config,
		prompt:     prompt,
		schemaHint: schemaHint,
		source:     source,
		loader:     chooseSmartScraperLoader(source, config),
		llmClient:  llm.NewOpenAIClient(config.LLMAPIKey, config),
	}
}

// Run executes the lightweight extraction workflow.
func (s *SmartScraperLiteGraph) Run(ctx context.Context) (json.RawMessage, error) {
	start := telemetry.Start()
	var runErr error
	defer func() { telemetry.LogGraph("smart_scraper_lite", start, runErr) }()
	g := graph.NewGraph(s.config)
	g.AddNode(nodes.NewFetchNode(s.loader, s.config))
	g.AddNode(nodes.NewParseNode(s.config))
	g.AddNode(nodes.NewGenerateAnswerNode(s.llmClient, s.prompt, s.schemaHint))
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
	if !isMeaningfulResult(result) {
		runErr = fmt.Errorf("extraction returned invalid/empty result")
		return nil, runErr
	}
	return result, nil
}

func buildLiteConfig(config *models.Config) *models.Config {
	if config == nil {
		config = models.DefaultConfig()
	}
	clone := *config
	if clone.HTMLMaxChars == 0 || clone.HTMLMaxChars > defaultLiteHTMLMaxChars {
		clone.HTMLMaxChars = defaultLiteHTMLMaxChars
	}
	if clone.ChunkSize == 0 || clone.ChunkSize > defaultLiteChunkSize {
		clone.ChunkSize = defaultLiteChunkSize
	}
	if clone.ChunkOverlap > clone.ChunkSize/4 {
		clone.ChunkOverlap = clone.ChunkSize / 4
	}
	return &clone
}

func chooseSmartScraperLoader(source string, config *models.Config) loaders.Loader {
	return loaders.NewFetchLoader(source, config)
}

func isMeaningfulResult(data json.RawMessage) bool {
	if len(data) == 0 {
		return false
	}
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "{}" || trimmed == "[]" || trimmed == "null" {
		return false
	}
	return !(strings.Contains(trimmed, `"error"`) && strings.Contains(trimmed, `"not_found"`))
}
