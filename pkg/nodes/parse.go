// Package nodes provides parsing and chunking
package nodes

import (
	"context"
	"fmt"

	"stealthfetch/internal/models"
	"stealthfetch/pkg/chunking"
	"stealthfetch/pkg/graph"
	"stealthfetch/pkg/markdown"
)

// ParseNode converts HTML to text and chunks it
type ParseNode struct {
	*graph.BaseNode
	config  *models.Config
	chunker *chunking.Chunker
}

// NewParseNode creates a new parse node
func NewParseNode(config *models.Config) *ParseNode {
	chunker := chunking.NewChunker(config.ChunkSize, config.ChunkOverlap)
	return &ParseNode{
		BaseNode: graph.NewBaseNode(
			"parse",
			[]string{"html"},
			[]string{"text", "chunks", "parse_result"},
		),
		config:  config,
		chunker: chunker,
	}
}

// Execute converts HTML to text and chunks it
func (n *ParseNode) Execute(ctx context.Context, state *graph.State) error {
	// Validate inputs
	if err := n.ValidateInputs(state); err != nil {
		return err
	}

	// Get HTML from state
	html, ok := state.GetString("html")
	if !ok {
		return fmt.Errorf("html is not a string")
	}

	// Convert to text
	text := markdown.HTMLToText(html, n.config.HTMLMaxChars)

	// Chunk the text
	chunks := n.chunker.Split(text)

	// Create result
	result := &models.ParseResult{
		Chunks:     chunks,
		FullText:   text,
		ChunkCount: len(chunks),
		Error:      nil,
	}

	// Store in state
	state.Set("text", text)
	state.Set("chunks", chunks)
	state.Set("parse_result", result)

	return nil
}
