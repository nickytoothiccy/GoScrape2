// Package nodes provides LLM generation
package nodes

import (
	"context"
	"encoding/json"
	"fmt"

	"stealthfetch/pkg/graph"
	"stealthfetch/pkg/llm"
)

// GenerateAnswerNode performs LLM extraction
type GenerateAnswerNode struct {
	*graph.BaseNode
	llmClient  llm.LLM
	prompt     string
	schemaHint string
}

// NewGenerateAnswerNode creates a new generation node
func NewGenerateAnswerNode(llmClient llm.LLM, prompt, schemaHint string) *GenerateAnswerNode {
	return &GenerateAnswerNode{
		BaseNode: graph.NewBaseNode(
			"generate_answer",
			[]string{"chunks"},
			[]string{"extracted_data", "extract_result"},
		),
		llmClient:  llmClient,
		prompt:     prompt,
		schemaHint: schemaHint,
	}
}

// Execute performs extraction on chunks
func (n *GenerateAnswerNode) Execute(ctx context.Context, state *graph.State) error {
	// Validate inputs
	if err := n.ValidateInputs(state); err != nil {
		return err
	}

	// Get chunks from state
	chunks, ok := state.GetStringSlice("chunks")
	if !ok {
		return fmt.Errorf("chunks is not a string slice")
	}

	if len(chunks) == 0 {
		return fmt.Errorf("no chunks to process")
	}

	// Single chunk - direct extraction
	if len(chunks) == 1 {
		result, err := n.llmClient.Extract(ctx, chunks[0], n.prompt, n.schemaHint)
		if err != nil {
			return fmt.Errorf("extract: %w", err)
		}

		state.Set("extracted_data", result.Data)
		state.Set("extract_result", result)
		return nil
	}

	// Multiple chunks - map-reduce
	var chunkResults []json.RawMessage
	for i, chunk := range chunks {
		result, err := n.llmClient.Extract(ctx, chunk, n.prompt, n.schemaHint)
		if err != nil {
			return fmt.Errorf("extract chunk %d: %w", i, err)
		}
		chunkResults = append(chunkResults, result.Data)
	}

	// Merge results
	merged, err := n.llmClient.MergeExtractions(ctx, chunkResults, n.prompt)
	if err != nil {
		return fmt.Errorf("merge: %w", err)
	}

	state.Set("extracted_data", merged.Data)
	state.Set("extract_result", merged)

	return nil
}
