package scrapegraph

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"stealthfetch/internal/models"
	"stealthfetch/pkg/graph"
	"stealthfetch/pkg/llm"
	"stealthfetch/pkg/loaders"
	"stealthfetch/pkg/nodes"
)

type stubLLM struct{}

func (stubLLM) Extract(context.Context, string, string, string) (*models.ExtractResult, error) {
	return &models.ExtractResult{Data: json.RawMessage(`{"ok":true}`), Model: "stub"}, nil
}
func (stubLLM) MergeExtractions(context.Context, []json.RawMessage, string) (*models.ExtractResult, error) {
	return &models.ExtractResult{Data: json.RawMessage(`{"ok":true}`), Model: "stub"}, nil
}
func (stubLLM) Generate(context.Context, string, string) (string, error) { return "", nil }
func (stubLLM) GenerateJSON(context.Context, string, string) (json.RawMessage, error) {
	return json.RawMessage(`{"ok":true}`), nil
}
func (stubLLM) ModelName() string { return "stub" }

func TestStructuredScraperRun_JSONLoader(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample.json")
	if err := os.WriteFile(path, []byte(`{"city":"Chioggia"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	s := structuredScraper{
		config:     models.DefaultConfig(),
		prompt:     "extract city",
		source:     path,
		name:       "json_scraper",
		loader:     loaders.NewJSONLoader(),
		newGenNode: func(c llm.LLM, p, sh string) graph.Node { return nodes.NewGenerateAnswerNode(stubLLM{}, p, sh) },
	}
	result, err := s.run(context.Background())
	if err != nil || string(result) != `{"ok":true}` {
		t.Fatalf("unexpected result=%s err=%v", string(result), err)
	}
}

func TestNewCSVScraperGraph_DefaultConfig(t *testing.T) {
	g := NewCSVScraperGraph("prompt", "sample.csv", nil, "")
	if g == nil || g.config == nil || g.loader == nil {
		t.Fatalf("expected initialized graph, got %#v", g)
	}
}
