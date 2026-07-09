// Package llm provides LLM client functionality.
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"stealthfetch/internal/models"
)

const systemPrompt = `You are a precise data extraction engine. Given content from a webpage and a user's extraction prompt, extract the requested information and return it as valid JSON.

Rules:
- Return ONLY valid JSON, no markdown fences, no explanation
- If the requested data is not found, return {"error": "not_found", "reason": "..."}
- Use arrays for lists of items
- Use descriptive key names
- Preserve URLs as absolute paths when possible`

// OpenAIClient wraps OpenAI-compatible chat completion APIs.
type OpenAIClient struct {
	client      openai.Client
	config      *models.Config
	provider    string
	jsonOptions []option.RequestOption
}

// NewOpenAIClient creates a client for the OpenAI API.
func NewOpenAIClient(apiKey string, config *models.Config) *OpenAIClient {
	return newOpenAICompatibleClient(apiKey, config, ProviderOpenAI, nil)
}

// Extract performs LLM-based data extraction.
func (c *OpenAIClient) Extract(ctx context.Context, content, prompt, schemaHint string) (*models.ExtractResult, error) {
	userMsg := fmt.Sprintf("## Extraction Prompt\n%s\n\n## Page Content\n%s", prompt, content)
	if schemaHint != "" {
		userMsg += fmt.Sprintf("\n\n## Expected Schema\n%s", schemaHint)
	}
	start := time.Now()
	chat, raw, err := c.complete(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt), openai.UserMessage(userMsg),
		},
		ResponseFormat: jsonResponseFormat(),
		Model:          c.config.LLMModel,
		Temperature:    openai.Float(c.config.Temperature),
	}, true)
	if err != nil {
		return nil, fmt.Errorf("%s extract: %w", c.provider, err)
	}
	return &models.ExtractResult{
		Data: json.RawMessage(raw), Model: c.config.LLMModel,
		ElapsedSecs: time.Since(start).Seconds(), TokensUsed: int(chat.Usage.TotalTokens),
	}, nil
}

// ModelName returns the configured model name.
func (c *OpenAIClient) ModelName() string { return c.config.LLMModel }

// Generate sends a prompt and returns raw text.
func (c *OpenAIClient) Generate(ctx context.Context, system, user string) (string, error) {
	_, raw, err := c.complete(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(system), openai.UserMessage(user),
		},
		Model: c.config.LLMModel, Temperature: openai.Float(c.config.Temperature),
	}, false)
	if err != nil {
		return "", fmt.Errorf("%s generate: %w", c.provider, err)
	}
	return raw, nil
}

// GenerateJSON sends a prompt and returns validated JSON.
func (c *OpenAIClient) GenerateJSON(ctx context.Context, system, user string) (json.RawMessage, error) {
	_, raw, err := c.complete(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(system), openai.UserMessage(user),
		},
		ResponseFormat: jsonResponseFormat(),
		Model:          c.config.LLMModel,
		Temperature:    openai.Float(c.config.Temperature),
	}, true)
	if err != nil {
		return nil, fmt.Errorf("%s generate json: %w", c.provider, err)
	}
	return json.RawMessage(raw), nil
}

// MergeExtractions combines multiple chunk extraction results.
func (c *OpenAIClient) MergeExtractions(ctx context.Context, chunks []json.RawMessage, originalPrompt string) (*models.ExtractResult, error) {
	var sb strings.Builder
	sb.WriteString("## Task\nMerge the following extraction results into a single coherent result.\n\n## Original Prompt\n")
	sb.WriteString(originalPrompt)
	sb.WriteString("\n\n## Chunk Results\n")
	for i, chunk := range chunks {
		fmt.Fprintf(&sb, "### Chunk %d\n%s\n\n", i+1, chunk)
	}
	start := time.Now()
	chat, raw, err := c.complete(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("You are a data merging engine. Combine extraction results into one deduplicated JSON result."),
			openai.UserMessage(sb.String()),
		},
		ResponseFormat: jsonResponseFormat(), Model: c.config.LLMModel, Temperature: openai.Float(0),
	}, true)
	if err != nil {
		return nil, fmt.Errorf("%s merge: %w", c.provider, err)
	}
	return &models.ExtractResult{
		Data: json.RawMessage(raw), Model: c.config.LLMModel,
		ElapsedSecs: time.Since(start).Seconds(), TokensUsed: int(chat.Usage.TotalTokens),
	}, nil
}

func (c *OpenAIClient) complete(ctx context.Context, params openai.ChatCompletionNewParams, jsonMode bool) (*openai.ChatCompletion, string, error) {
	opts := []option.RequestOption(nil)
	if jsonMode {
		opts = c.jsonOptions
	}
	chat, err := c.client.Chat.Completions.New(ctx, params, opts...)
	if err != nil {
		return nil, "", err
	}
	if len(chat.Choices) == 0 {
		return nil, "", fmt.Errorf("response contained no choices")
	}
	raw := strings.TrimSpace(chat.Choices[0].Message.Content)
	if raw == "" {
		return nil, "", fmt.Errorf("response content was empty")
	}
	if jsonMode && !json.Valid([]byte(raw)) {
		return nil, "", fmt.Errorf("response was not valid JSON")
	}
	return chat, raw, nil
}

func jsonResponseFormat() openai.ChatCompletionNewParamsResponseFormatUnion {
	return openai.ChatCompletionNewParamsResponseFormatUnion{
		OfJSONObject: &openai.ResponseFormatJSONObjectParam{},
	}
}
