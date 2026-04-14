// Package llm provides LLM client functionality
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

// OpenAIClient wraps the OpenAI API
type OpenAIClient struct {
	client openai.Client
	config *models.Config
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(apiKey string, config *models.Config) *OpenAIClient {
	client := openai.NewClient(option.WithAPIKey(apiKey))
	return &OpenAIClient{
		client: client,
		config: config,
	}
}

// Extract performs LLM-based data extraction
func (c *OpenAIClient) Extract(ctx context.Context, content, prompt, schemaHint string) (*models.ExtractResult, error) {
	userMsg := fmt.Sprintf("## Extraction Prompt\n%s\n\n## Page Content\n%s", prompt, content)
	if schemaHint != "" {
		userMsg += fmt.Sprintf("\n\n## Expected Schema\n%s", schemaHint)
	}

	start := time.Now()

	temp := c.config.Temperature
	chat, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userMsg),
		},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &openai.ResponseFormatJSONObjectParam{},
		},
		Model:       c.config.LLMModel,
		Temperature: openai.Float(temp),
	})
	if err != nil {
		return nil, fmt.Errorf("openai: %w", err)
	}

	rawContent := chat.Choices[0].Message.Content

	return &models.ExtractResult{
		Data:        json.RawMessage(rawContent),
		Model:       c.config.LLMModel,
		ElapsedSecs: time.Since(start).Seconds(),
		TokensUsed:  int(chat.Usage.TotalTokens),
		Error:       nil,
	}, nil
}

// ModelName returns the model name
func (c *OpenAIClient) ModelName() string {
	return c.config.LLMModel
}

// Generate sends a prompt and returns raw text
func (c *OpenAIClient) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	chat, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		},
		Model:       c.config.LLMModel,
		Temperature: openai.Float(c.config.Temperature),
	})
	if err != nil {
		return "", fmt.Errorf("openai generate: %w", err)
	}
	return chat.Choices[0].Message.Content, nil
}

// GenerateJSON sends a prompt and returns JSON
func (c *OpenAIClient) GenerateJSON(ctx context.Context, systemPrompt, userPrompt string) (json.RawMessage, error) {
	chat, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &openai.ResponseFormatJSONObjectParam{},
		},
		Model:       c.config.LLMModel,
		Temperature: openai.Float(c.config.Temperature),
	})
	if err != nil {
		return nil, fmt.Errorf("openai generate json: %w", err)
	}
	return json.RawMessage(chat.Choices[0].Message.Content), nil
}

// MergeExtractions combines multiple chunk extractions
func (c *OpenAIClient) MergeExtractions(ctx context.Context, chunks []json.RawMessage, originalPrompt string) (*models.ExtractResult, error) {
	// Build merge prompt
	var sb strings.Builder
	sb.WriteString("## Task\n")
	sb.WriteString("Merge the following extraction results into a single coherent result.\n\n")
	sb.WriteString("## Original Prompt\n")
	sb.WriteString(originalPrompt)
	sb.WriteString("\n\n## Chunk Results\n")

	for i, chunk := range chunks {
		sb.WriteString(fmt.Sprintf("### Chunk %d\n", i+1))
		sb.WriteString(string(chunk))
		sb.WriteString("\n\n")
	}

	start := time.Now()

	chat, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("You are a data merging engine. Combine multiple extraction results into a single, deduplicated result."),
			openai.UserMessage(sb.String()),
		},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &openai.ResponseFormatJSONObjectParam{},
		},
		Model:       c.config.LLMModel,
		Temperature: openai.Float(0),
	})
	if err != nil {
		return nil, fmt.Errorf("merge: %w", err)
	}

	return &models.ExtractResult{
		Data:        json.RawMessage(chat.Choices[0].Message.Content),
		Model:       c.config.LLMModel,
		ElapsedSecs: time.Since(start).Seconds(),
		TokensUsed:  int(chat.Usage.TotalTokens),
		Error:       nil,
	}, nil
}
