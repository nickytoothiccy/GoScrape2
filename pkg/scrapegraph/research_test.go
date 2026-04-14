package scrapegraph

import (
	"context"
	"encoding/json"
	"testing"

	"stealthfetch/internal/models"
	"stealthfetch/pkg/utils"
)

func TestResearchGraph_RunUsesDirectMode(t *testing.T) {
	g := NewResearchGraph(models.DefaultConfig())
	g.runSmart = func(context.Context, *models.Config, ResearchRequest) (json.RawMessage, []string, []string, error) {
		return json.RawMessage(`{"ok":true}`), []string{"https://a.com"}, nil, nil
	}
	out, err := g.Run(context.Background(), ResearchRequest{Prompt: "x", SeedURL: "https://a.com"})
	if err != nil || out.Mode != "direct" || out.PagesUsed != 1 {
		fatalResearch(t, out, err)
	}
}

func TestResearchGraph_RunUsesDepthMode(t *testing.T) {
	g := NewResearchGraph(models.DefaultConfig())
	g.runDepth = func(context.Context, *models.Config, ResearchRequest) (json.RawMessage, []string, []string, error) {
		return json.RawMessage(`{"ok":true}`), []string{"https://a.com/docs", "https://a.com/docs/p1"}, nil, nil
	}
	out, err := g.Run(context.Background(), ResearchRequest{Prompt: "x", SeedURL: "https://a.com/docs", FollowSubpages: true})
	if err != nil || out.Mode != "depth" || out.PagesUsed != 2 {
		fatalResearch(t, out, err)
	}
}

func TestResearchGraph_RunUsesSearchMode(t *testing.T) {
	g := NewResearchGraph(models.DefaultConfig())
	g.runSearch = func(context.Context, *models.Config, ResearchRequest, utils.SearchFunc) (json.RawMessage, []string, string, error) {
		return json.RawMessage(`{"ok":true}`), []string{"https://a.com"}, "test query", nil
	}
	out, err := g.Run(context.Background(), ResearchRequest{Prompt: "x", SearchFirst: true})
	if err != nil || out.Mode != "search" || out.SearchQuery != "test query" {
		fatalResearch(t, out, err)
	}
}

func fatalResearch(t *testing.T, out *ResearchResult, err error) {
	t.Helper()
	t.Fatalf("unexpected result: out=%+v err=%v", out, err)
}