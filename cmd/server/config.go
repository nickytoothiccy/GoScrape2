package main

import (
	"fmt"
	"os"
	"strings"

	"stealthfetch/internal/models"
	"stealthfetch/pkg/llm"
)

func loadLLMConfigFromEnv() (*models.Config, error) {
	cfg := models.DefaultConfig()
	provider := strings.ToLower(strings.TrimSpace(os.Getenv("LLM_PROVIDER")))
	if provider == "" {
		provider = llm.ProviderOpenAI
	}
	cfg.LLMProvider = provider
	cfg.LLMAPIKey = firstEnv("LLM_API_KEY", providerKeyEnv(provider))
	cfg.LLMBaseURL = firstEnv("LLM_BASE_URL", providerBaseURLEnv(provider))
	cfg.LLMHTTPReferer = firstEnv("OPENROUTER_HTTP_REFERER")
	cfg.LLMAppTitle = firstEnv("OPENROUTER_APP_TITLE")

	model := firstEnv("LLM_MODEL", providerModelEnv(provider))
	if model != "" {
		cfg.LLMModel = model
	} else if provider == llm.ProviderOpenRouter {
		return nil, fmt.Errorf("LLM_MODEL or OPENROUTER_MODEL is required for openrouter")
	}
	if _, err := llm.NewClient(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func providerKeyEnv(provider string) string {
	if provider == llm.ProviderOpenRouter {
		return "OPENROUTER_API_KEY"
	}
	return "OPENAI_API_KEY"
}

func providerModelEnv(provider string) string {
	if provider == llm.ProviderOpenRouter {
		return "OPENROUTER_MODEL"
	}
	return "OPENAI_MODEL"
}

func providerBaseURLEnv(provider string) string {
	if provider == llm.ProviderOpenRouter {
		return "OPENROUTER_BASE_URL"
	}
	return "OPENAI_BASE_URL"
}
