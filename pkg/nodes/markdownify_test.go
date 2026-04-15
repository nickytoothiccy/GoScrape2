package nodes

import (
	"context"
	"strings"
	"testing"

	"stealthfetch/internal/models"
	"stealthfetch/pkg/graph"
)

func TestMarkdownifyNodeExecute(t *testing.T) {
	state := graph.NewState()
	state.Set("html", `<html><body><h1>Title</h1><p>Hello <a href="https://example.com">world</a></p><ul><li>One</li></ul></body></html>`)
	node := NewMarkdownifyNode(models.DefaultConfig())
	if err := node.Execute(context.Background(), state); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	out, ok := state.GetString("markdown")
	if !ok {
		t.Fatal("expected markdown in state")
	}
	for _, want := range []string{"# Title", "Hello world (https://example.com)", "- One"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in output, got %q", want, out)
		}
	}
}

func TestMarkdownifyNodeMissingHTML(t *testing.T) {
	node := NewMarkdownifyNode(models.DefaultConfig())
	err := node.Execute(context.Background(), graph.NewState())
	if err == nil {
		t.Fatal("expected error for missing html")
	}
}
