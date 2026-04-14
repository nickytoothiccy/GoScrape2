// Package nodes provides the GraphIteratorNode for running sub-graphs on lists
package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"stealthfetch/pkg/graph"
)

// GraphFactory creates and runs a sub-graph for a single item.
// Returns the extracted JSON result or an error.
type GraphFactory func(ctx context.Context, item string) (json.RawMessage, error)

// GraphIteratorNode runs a sub-graph for each item in a list.
// This is the core enabler for all Multi-graph variants and
// depth-based scraping workflows.
type GraphIteratorNode struct {
	*graph.BaseNode
	factory        GraphFactory
	inputKey       string        // state key containing []string items
	outputKey      string        // state key to store []json.RawMessage results
	perItemTimeout time.Duration // timeout per item, 0 = inherit parent ctx
	verbose        bool
}

// GraphIteratorConfig holds configuration for creating a GraphIteratorNode
type GraphIteratorConfig struct {
	Factory        GraphFactory
	InputKey       string        // default "urls"
	OutputKey      string        // default "answers"
	PerItemTimeout time.Duration // default 60s
	Verbose        bool
}

// NewGraphIteratorNode creates a new graph iterator node
func NewGraphIteratorNode(cfg GraphIteratorConfig) *GraphIteratorNode {
	if cfg.InputKey == "" {
		cfg.InputKey = "urls"
	}
	if cfg.OutputKey == "" {
		cfg.OutputKey = "answers"
	}
	if cfg.PerItemTimeout == 0 {
		cfg.PerItemTimeout = 60 * time.Second
	}

	return &GraphIteratorNode{
		BaseNode: graph.NewBaseNode(
			"graph_iterator",
			[]string{cfg.InputKey},
			[]string{cfg.OutputKey},
		),
		factory:        cfg.Factory,
		inputKey:       cfg.InputKey,
		outputKey:      cfg.OutputKey,
		perItemTimeout: cfg.PerItemTimeout,
		verbose:        cfg.Verbose,
	}
}

// Execute iterates over items and runs the sub-graph factory for each
func (n *GraphIteratorNode) Execute(ctx context.Context, state *graph.State) error {
	items, ok := state.GetStringSlice(n.inputKey)
	if !ok {
		return fmt.Errorf("graph_iterator: missing or invalid '%s' in state", n.inputKey)
	}

	if len(items) == 0 {
		state.Set(n.outputKey, []json.RawMessage{})
		return nil
	}

	if n.verbose {
		log.Printf("[graph_iterator] processing %d items from '%s'", len(items), n.inputKey)
	}

	var results []json.RawMessage
	var failedURLs []string

	for i, item := range items {
		// Check parent context
		select {
		case <-ctx.Done():
			if n.verbose {
				log.Printf("[graph_iterator] context cancelled at item %d/%d", i+1, len(items))
			}
			if len(results) > 0 {
				break
			}
			return ctx.Err()
		default:
		}

		if n.verbose {
			log.Printf("[graph_iterator] item %d/%d: %s", i+1, len(items), truncate(item, 80))
		}

		// Per-item timeout
		itemCtx, cancel := context.WithTimeout(ctx, n.perItemTimeout)
		result, err := n.factory(itemCtx, item)
		cancel()

		if err != nil {
			if n.verbose {
				log.Printf("[graph_iterator] item %d failed: %v", i+1, err)
			}
			failedURLs = append(failedURLs, item)
			continue
		}

		results = append(results, result)
	}

	if n.verbose {
		log.Printf("[graph_iterator] completed: %d succeeded, %d failed",
			len(results), len(failedURLs))
	}

	if len(results) == 0 {
		return fmt.Errorf("graph_iterator: all %d items failed", len(items))
	}

	state.Set(n.outputKey, results)
	if len(failedURLs) > 0 {
		state.Set("failed_items", failedURLs)
	}

	return nil
}

// truncate shortens a string for logging
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
