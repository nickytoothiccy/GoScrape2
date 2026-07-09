package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"stealthfetch/internal/models"
)

func TestNewClientDefaultsToOpenAI(t *testing.T) {
	cfg := models.DefaultConfig()
	cfg.LLMAPIKey = "test-key"
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if client.ModelName() != "gpt-4o" {
		t.Fatalf("model = %q", client.ModelName())
	}
}

func TestNewClientValidatesProviderConfiguration(t *testing.T) {
	tests := []struct {
		name string
		cfg  *models.Config
		want string
	}{
		{"missing key", &models.Config{LLMProvider: ProviderOpenRouter, LLMModel: "openai/gpt-4o-mini"}, "API key is required"},
		{"missing model", &models.Config{LLMProvider: ProviderOpenRouter, LLMAPIKey: "key"}, "model is required"},
		{"unknown provider", &models.Config{LLMProvider: "other", LLMAPIKey: "key", LLMModel: "model"}, "unsupported LLM provider"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.cfg)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestOpenRouterJSONRequestContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/chat/completions" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer router-key" {
			t.Errorf("authorization = %q", got)
		}
		if got := r.Header.Get("HTTP-Referer"); got != "https://goscrape.test" {
			t.Errorf("referer = %q", got)
		}
		if got := r.Header.Get("X-OpenRouter-Title"); got != "GoScrape2" {
			t.Errorf("title = %q", got)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		assertNestedValue(t, body, "response_format", "type", "json_object")
		assertNestedValue(t, body, "provider", "require_parameters", true)
		if body["model"] != "anthropic/claude-sonnet-4" {
			t.Errorf("model = %#v", body["model"])
		}
		writeCompletion(t, w, `{"ok":true}`)
	}))
	defer server.Close()

	cfg := models.DefaultConfig()
	cfg.LLMProvider = ProviderOpenRouter
	cfg.LLMAPIKey = "router-key"
	cfg.LLMModel = "anthropic/claude-sonnet-4"
	cfg.LLMBaseURL = server.URL + "/api/v1"
	cfg.LLMHTTPReferer = "https://goscrape.test"
	cfg.LLMAppTitle = "GoScrape2"
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.GenerateJSON(context.Background(), "system", "Return JSON")
	if err != nil || string(result) != `{"ok":true}` {
		t.Fatalf("result=%s err=%v", result, err)
	}
}

func TestOpenRouterGenerateOmitsJSONRouting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		if _, ok := body["response_format"]; ok {
			t.Error("plain generation included response_format")
		}
		if _, ok := body["provider"]; ok {
			t.Error("plain generation included provider requirements")
		}
		writeCompletion(t, w, "plain text")
	}))
	defer server.Close()
	client := openRouterTestClient(t, server.URL)
	result, err := client.Generate(context.Background(), "system", "user")
	if err != nil || result != "plain text" {
		t.Fatalf("result=%q err=%v", result, err)
	}
}

func TestOpenRouterRejectsInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeCompletion(t, w, "not-json")
	}))
	defer server.Close()
	_, err := openRouterTestClient(t, server.URL).GenerateJSON(context.Background(), "system", "Return JSON")
	if err == nil || !strings.Contains(err.Error(), "not valid JSON") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOpenRouterRejectsEmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "test", "object": "chat.completion", "created": 1,
			"model": "test", "choices": []any{},
		})
	}))
	defer server.Close()
	_, err := openRouterTestClient(t, server.URL).Generate(context.Background(), "system", "user")
	if err == nil || !strings.Contains(err.Error(), "no choices") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func openRouterTestClient(t *testing.T, baseURL string) LLM {
	t.Helper()
	cfg := models.DefaultConfig()
	cfg.LLMProvider, cfg.LLMAPIKey = ProviderOpenRouter, "key"
	cfg.LLMModel, cfg.LLMBaseURL = "openai/gpt-4o-mini", baseURL
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func writeCompletion(t *testing.T, w http.ResponseWriter, content string) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]any{
		"id": "test", "object": "chat.completion", "created": 1, "model": "test",
		"choices": []any{map[string]any{"index": 0, "message": map[string]any{"role": "assistant", "content": content}, "finish_reason": "stop"}},
		"usage":   map[string]any{"prompt_tokens": 1, "completion_tokens": 1, "total_tokens": 2},
	})
	if err != nil {
		t.Errorf("encode response: %v", err)
	}
}

func assertNestedValue(t *testing.T, body map[string]any, object, key string, want any) {
	t.Helper()
	nested, ok := body[object].(map[string]any)
	if !ok || nested[key] != want {
		t.Fatalf("%s.%s = %#v, want %#v", object, key, nested[key], want)
	}
}
