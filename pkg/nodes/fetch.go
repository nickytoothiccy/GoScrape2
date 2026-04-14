// Package nodes provides graph node implementations
package nodes

import (
	"context"
	"fmt"
	"strings"

	"stealthfetch/internal/models"
	"stealthfetch/pkg/graph"
	"stealthfetch/pkg/loaders"
)

// FetchNode handles content fetching from various sources
type FetchNode struct {
	*graph.BaseNode
	loader loaders.Loader
	config *models.Config
}

// NewFetchNode creates a fetch node with the given loader
func NewFetchNode(loader loaders.Loader, config *models.Config) *FetchNode {
	return &FetchNode{
		BaseNode: graph.NewBaseNode(
			"fetch",
			[]string{"url"},
			[]string{"html", "fetch_result"},
		),
		loader: loader,
		config: config,
	}
}

// Execute fetches content and stores it in state
func (n *FetchNode) Execute(ctx context.Context, state *graph.State) error {
	// Validate inputs
	if err := n.ValidateInputs(state); err != nil {
		return err
	}

	// Get URL from state
	source, ok := state.GetString("url")
	if !ok {
		return fmt.Errorf("url is not a string")
	}

	// Check if it's HTML content or URL
	trimmed := strings.TrimSpace(source)
	if strings.HasPrefix(trimmed, "<") {
		// Local HTML content
		localLoader := loaders.NewLocalLoader()
		result, err := localLoader.Load(ctx, source)
		if err != nil {
			return fmt.Errorf("local load: %w", err)
		}
		state.Set("html", result.HTML)
		state.Set("fetch_result", result)
		return nil
	}

	// Fetch from URL using configured loader
	result, err := n.loader.Load(ctx, source)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	if result.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d from %s", result.StatusCode, source)
	}

	// Store in state
	state.Set("html", result.HTML)
	state.Set("fetch_result", result)

	return nil
}
