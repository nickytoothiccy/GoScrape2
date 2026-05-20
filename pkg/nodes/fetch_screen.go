package nodes

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"stealthfetch/internal/models"
	"stealthfetch/pkg/graph"
	"stealthfetch/pkg/loaders"
)

// FetchScreenNode captures a rendered screenshot for a URL source.
type FetchScreenNode struct {
	*graph.BaseNode
	loader loaders.ScreenshotLoader
}

// NewFetchScreenNode creates a screenshot node with the given loader.
func NewFetchScreenNode(loader loaders.ScreenshotLoader, _ *models.Config) *FetchScreenNode {
	return &FetchScreenNode{
		BaseNode: graph.NewBaseNode(
			"fetch_screen",
			[]string{"url"},
			[]string{"screenshot_png", "screenshot_base64", "screenshot_result"},
		),
		loader: loader,
	}
}

// Execute captures a screenshot and stores raw bytes and base64 data in state.
func (n *FetchScreenNode) Execute(ctx context.Context, state *graph.State) error {
	if err := n.ValidateInputs(state); err != nil {
		return err
	}
	source, ok := state.GetString("url")
	if !ok {
		return fmt.Errorf("url is not a string")
	}
	trimmed := strings.TrimSpace(source)
	if trimmed == "" {
		return fmt.Errorf("url is required")
	}
	if strings.HasPrefix(trimmed, "<") {
		return fmt.Errorf("screenshot capture requires URL source")
	}

	result, err := n.loader.Capture(ctx, source)
	if err != nil {
		return fmt.Errorf("capture screenshot: %w", err)
	}
	if result == nil || len(result.PNG) == 0 {
		return fmt.Errorf("empty screenshot result")
	}

	state.Set("screenshot_png", result.PNG)
	state.Set("screenshot_base64", base64.StdEncoding.EncodeToString(result.PNG))
	state.Set("screenshot_result", result)
	return nil
}
