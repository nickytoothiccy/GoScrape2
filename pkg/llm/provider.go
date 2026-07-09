package llm

import (
	"fmt"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"stealthfetch/internal/models"
)

const (
	ProviderOpenAI     = "openai"
	ProviderOpenRouter = "openrouter"
	OpenRouterBaseURL  = "https://openrouter.ai/api/v1"
)

// NewClient creates the configured LLM provider.
func NewClient(config *models.Config) (LLM, error) {
	if config == nil {
		config = models.DefaultConfig()
	}
	clone := *config
	clone.LLMProvider = normalizeProvider(clone.LLMProvider)
	clone.LLMAPIKey = strings.TrimSpace(clone.LLMAPIKey)
	clone.LLMModel = strings.TrimSpace(clone.LLMModel)
	if clone.LLMAPIKey == "" {
		return nil, fmt.Errorf("%s API key is required", clone.LLMProvider)
	}
	if clone.LLMModel == "" {
		return nil, fmt.Errorf("%s model is required", clone.LLMProvider)
	}
	switch clone.LLMProvider {
	case ProviderOpenAI:
		return NewOpenAIClient(clone.LLMAPIKey, &clone), nil
	case ProviderOpenRouter:
		return NewOpenRouterClient(clone.LLMAPIKey, &clone), nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider %q", clone.LLMProvider)
	}
}

// NewOpenRouterClient creates an OpenAI-compatible OpenRouter client.
func NewOpenRouterClient(apiKey string, config *models.Config) *OpenAIClient {
	if config == nil {
		config = models.DefaultConfig()
	}
	baseURL := strings.TrimSpace(config.LLMBaseURL)
	if baseURL == "" {
		baseURL = OpenRouterBaseURL
	}
	opts := []option.RequestOption{option.WithBaseURL(baseURL)}
	if value := strings.TrimSpace(config.LLMHTTPReferer); value != "" {
		opts = append(opts, option.WithHeader("HTTP-Referer", value))
	}
	if value := strings.TrimSpace(config.LLMAppTitle); value != "" {
		opts = append(opts, option.WithHeader("X-OpenRouter-Title", value))
	}
	client := newOpenAICompatibleClient(apiKey, config, ProviderOpenRouter, opts)
	client.jsonOptions = []option.RequestOption{
		option.WithJSONSet("provider.require_parameters", true),
	}
	return client
}

func newOpenAICompatibleClient(apiKey string, config *models.Config, provider string, opts []option.RequestOption) *OpenAIClient {
	if config == nil {
		config = models.DefaultConfig()
	}
	allOpts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if baseURL := strings.TrimSpace(config.LLMBaseURL); baseURL != "" && provider == ProviderOpenAI {
		allOpts = append(allOpts, option.WithBaseURL(baseURL))
	}
	allOpts = append(allOpts, opts...)
	return &OpenAIClient{
		client: openai.NewClient(allOpts...), config: config, provider: provider,
	}
}

func normalizeProvider(provider string) string {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		return ProviderOpenAI
	}
	return provider
}
