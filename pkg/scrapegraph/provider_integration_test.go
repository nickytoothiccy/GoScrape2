package scrapegraph

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"stealthfetch/internal/models"
)

func TestSmartScraperGraphUsesConfiguredOpenRouterProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		provider, ok := body["provider"].(map[string]any)
		if !ok || provider["require_parameters"] != true {
			t.Errorf("provider requirements = %#v", body["provider"])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "test", "object": "chat.completion", "created": 1, "model": "test",
			"choices": []any{map[string]any{
				"index": 0, "message": map[string]any{"role": "assistant", "content": `{"title":"Provider Test"}`},
				"finish_reason": "stop",
			}},
			"usage": map[string]any{"prompt_tokens": 1, "completion_tokens": 1, "total_tokens": 2},
		})
	}))
	defer server.Close()

	cfg := models.DefaultConfig()
	cfg.LLMProvider = "openrouter"
	cfg.LLMAPIKey = "router-key"
	cfg.LLMModel = "openai/gpt-4o-mini"
	cfg.LLMBaseURL = server.URL
	graph := NewSmartScraperGraph("extract title", `<html><h1>Provider Test</h1></html>`, cfg, "")
	result, err := graph.Run(context.Background())
	if err != nil || string(result) != `{"title":"Provider Test"}` {
		t.Fatalf("result=%s err=%v", result, err)
	}
}
