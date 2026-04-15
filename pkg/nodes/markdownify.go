package nodes

import (
	"context"
	"fmt"

	"stealthfetch/internal/models"
	"stealthfetch/pkg/graph"
	"stealthfetch/pkg/markdown"
)

// MarkdownifyNode converts HTML into Markdown.
type MarkdownifyNode struct {
	*graph.BaseNode
	config *models.Config
}

// NewMarkdownifyNode creates a new Markdownify node.
func NewMarkdownifyNode(config *models.Config) *MarkdownifyNode {
	if config == nil {
		config = models.DefaultConfig()
	}
	return &MarkdownifyNode{
		BaseNode: graph.NewBaseNode("markdownify", []string{"html"}, []string{"markdown"}),
		config:   config,
	}
}

// Execute converts HTML in state to Markdown.
func (n *MarkdownifyNode) Execute(ctx context.Context, state *graph.State) error {
	_ = ctx
	if err := n.ValidateInputs(state); err != nil {
		return err
	}
	rawHTML, ok := state.GetString("html")
	if !ok {
		return fmt.Errorf("markdownify: html is not a string")
	}
	state.Set("markdown", markdown.HTMLToMarkdown(rawHTML, n.config.HTMLMaxChars))
	return nil
}
