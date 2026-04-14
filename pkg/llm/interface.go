// Package llm provides LLM client abstractions and implementations
package llm

import (
	"context"
	"encoding/json"

	"stealthfetch/internal/models"
)

// LLM defines the interface for all language model providers
type LLM interface {
	// Extract performs data extraction from content using a prompt
	Extract(ctx context.Context, content, prompt, schemaHint string) (*models.ExtractResult, error)

	// MergeExtractions combines multiple chunk extraction results
	MergeExtractions(ctx context.Context, chunks []json.RawMessage, originalPrompt string) (*models.ExtractResult, error)

	// Generate sends a prompt and returns the raw text response
	Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error)

	// GenerateJSON sends a prompt and returns a JSON response
	GenerateJSON(ctx context.Context, systemPrompt, userPrompt string) (json.RawMessage, error)

	// ModelName returns the name of the model being used
	ModelName() string
}
