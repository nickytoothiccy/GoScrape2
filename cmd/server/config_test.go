package main

import (
	"strings"
	"testing"

	"stealthfetch/pkg/llm"
)

var llmEnvKeys = []string{
	"LLM_PROVIDER", "LLM_API_KEY", "LLM_MODEL", "LLM_BASE_URL",
	"OPENAI_API_KEY", "OPENAI_MODEL", "OPENAI_BASE_URL",
	"OPENROUTER_API_KEY", "OPENROUTER_MODEL", "OPENROUTER_BASE_URL",
	"OPENROUTER_HTTP_REFERER", "OPENROUTER_APP_TITLE",
}

func TestLoadLLMConfigDefaultsToOpenAI(t *testing.T) {
	clearLLMEnv(t)
	t.Setenv("OPENAI_API_KEY", "openai-key")
	cfg, err := loadLLMConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.LLMProvider != llm.ProviderOpenAI || cfg.LLMAPIKey != "openai-key" || cfg.LLMModel != "gpt-4o" {
		t.Fatalf("unexpected config: %#v", cfg)
	}
}

func TestLoadLLMConfigOpenRouter(t *testing.T) {
	clearLLMEnv(t)
	t.Setenv("LLM_PROVIDER", "openrouter")
	t.Setenv("OPENROUTER_API_KEY", "router-key")
	t.Setenv("OPENROUTER_MODEL", "anthropic/claude-sonnet-4")
	t.Setenv("OPENROUTER_BASE_URL", "https://router.test/v1")
	t.Setenv("OPENROUTER_HTTP_REFERER", "https://goscrape.test")
	t.Setenv("OPENROUTER_APP_TITLE", "GoScrape2")
	cfg, err := loadLLMConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.LLMProvider != llm.ProviderOpenRouter || cfg.LLMAPIKey != "router-key" {
		t.Fatalf("unexpected provider config: %#v", cfg)
	}
	if cfg.LLMModel != "anthropic/claude-sonnet-4" || cfg.LLMBaseURL != "https://router.test/v1" {
		t.Fatalf("unexpected model config: %#v", cfg)
	}
	if cfg.LLMHTTPReferer != "https://goscrape.test" || cfg.LLMAppTitle != "GoScrape2" {
		t.Fatalf("unexpected attribution config: %#v", cfg)
	}
}

func TestLoadLLMConfigGenericValuesWin(t *testing.T) {
	clearLLMEnv(t)
	t.Setenv("LLM_PROVIDER", "openrouter")
	t.Setenv("LLM_API_KEY", "generic-key")
	t.Setenv("OPENROUTER_API_KEY", "provider-key")
	t.Setenv("LLM_MODEL", "openai/gpt-4o-mini")
	t.Setenv("OPENROUTER_MODEL", "other/model")
	cfg, err := loadLLMConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.LLMAPIKey != "generic-key" || cfg.LLMModel != "openai/gpt-4o-mini" {
		t.Fatalf("unexpected precedence: %#v", cfg)
	}
}

func TestLoadLLMConfigRequiresOpenRouterModel(t *testing.T) {
	clearLLMEnv(t)
	t.Setenv("LLM_PROVIDER", "openrouter")
	t.Setenv("OPENROUTER_API_KEY", "router-key")
	_, err := loadLLMConfigFromEnv()
	if err == nil || !strings.Contains(err.Error(), "OPENROUTER_MODEL") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadLLMConfigRejectsUnknownProvider(t *testing.T) {
	clearLLMEnv(t)
	t.Setenv("LLM_PROVIDER", "other")
	t.Setenv("LLM_API_KEY", "key")
	t.Setenv("LLM_MODEL", "model")
	_, err := loadLLMConfigFromEnv()
	if err == nil || !strings.Contains(err.Error(), "unsupported LLM provider") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func clearLLMEnv(t *testing.T) {
	t.Helper()
	for _, key := range llmEnvKeys {
		t.Setenv(key, "")
	}
}
