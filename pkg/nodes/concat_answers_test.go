package nodes

import (
	"context"
	"encoding/json"
	"testing"

	"stealthfetch/pkg/graph"
)

func TestConcatAnswersNode_MixedInputs(t *testing.T) {
	node := NewConcatAnswersNode(ConcatAnswersConfig{})
	state := graph.NewState()

	state.Set("answers", []json.RawMessage{
		json.RawMessage(`[{"id":1},{"id":2}]`),
		json.RawMessage(`{"id":3}`),
		json.RawMessage(`null`),
		json.RawMessage(`   `),
	})

	if err := node.Execute(context.Background(), state); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	raw, ok := state.GetJSON("extracted_data")
	if !ok {
		t.Fatalf("expected extracted_data in state")
	}

	if string(raw) != `[{"id":1},{"id":2},{"id":3}]` {
		t.Fatalf("unexpected output: %s", raw)
	}
}

func TestConcatAnswersNode_InvalidJSON(t *testing.T) {
	node := NewConcatAnswersNode(ConcatAnswersConfig{})
	state := graph.NewState()
	state.Set("answers", []json.RawMessage{json.RawMessage(`{"ok":1}`), json.RawMessage(`{bad`)})

	err := node.Execute(context.Background(), state)
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestConcatAnswersNode_UnsupportedType(t *testing.T) {
	node := NewConcatAnswersNode(ConcatAnswersConfig{})
	state := graph.NewState()
	state.Set("answers", []int{1, 2, 3})

	err := node.Execute(context.Background(), state)
	if err == nil {
		t.Fatal("expected error for unsupported answers type")
	}
}

func TestConcatAnswersNode_StringInputs(t *testing.T) {
	node := NewConcatAnswersNode(ConcatAnswersConfig{})
	state := graph.NewState()
	state.Set("answers", []string{`[{"id":1}]`, `{"id":2}`})

	if err := node.Execute(context.Background(), state); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	raw, ok := state.GetJSON("extract_result")
	if !ok {
		t.Fatalf("expected extract_result in state")
	}

	if string(raw) != `[{"id":1},{"id":2}]` {
		t.Fatalf("unexpected output: %s", raw)
	}
}

func TestConcatAnswersNode_EmptyInputs(t *testing.T) {
	node := NewConcatAnswersNode(ConcatAnswersConfig{})
	state := graph.NewState()
	state.Set("answers", []json.RawMessage{})

	if err := node.Execute(context.Background(), state); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	raw, ok := state.GetJSON("extracted_data")
	if !ok {
		t.Fatalf("expected extracted_data in state")
	}

	if string(raw) != `[]` {
		t.Fatalf("unexpected output: %s", raw)
	}
}
