// Package nodes provides deterministic answer concatenation without LLM usage.
package nodes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"stealthfetch/pkg/graph"
)

// ConcatAnswersNode concatenates multiple extraction outputs into one JSON array.
// Unlike MergeAnswersNode, this performs no LLM synthesis.
type ConcatAnswersNode struct {
	*graph.BaseNode
	inputKey  string
	outputKey string
	verbose   bool
}

// ConcatAnswersConfig controls how ConcatAnswersNode reads/writes state.
type ConcatAnswersConfig struct {
	InputKey  string // default: "answers"
	OutputKey string // default: "extracted_data"
	Verbose   bool
}

// NewConcatAnswersNode creates a new deterministic concatenation node.
func NewConcatAnswersNode(cfg ConcatAnswersConfig) *ConcatAnswersNode {
	if cfg.InputKey == "" {
		cfg.InputKey = "answers"
	}
	if cfg.OutputKey == "" {
		cfg.OutputKey = "extracted_data"
	}

	return &ConcatAnswersNode{
		BaseNode: graph.NewBaseNode(
			"concat_answers",
			[]string{cfg.InputKey},
			[]string{cfg.OutputKey, "extract_result"},
		),
		inputKey:  cfg.InputKey,
		outputKey: cfg.OutputKey,
		verbose:   cfg.Verbose,
	}
}

// Execute concatenates state answers into a single JSON array.
func (n *ConcatAnswersNode) Execute(_ context.Context, state *graph.State) error {
	if err := n.ValidateInputs(state); err != nil {
		return err
	}

	raw, _ := state.Get(n.inputKey)
	answers, err := asRawMessages(raw)
	if err != nil {
		return fmt.Errorf("concat_answers: %w", err)
	}

	merged := make([]interface{}, 0, len(answers))
	for i, answer := range answers {
		trimmed := bytes.TrimSpace(answer)
		if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
			continue
		}

		var parsed interface{}
		if err := json.Unmarshal(trimmed, &parsed); err != nil {
			return fmt.Errorf("concat_answers: invalid JSON at index %d: %w", i, err)
		}

		if arr, ok := parsed.([]interface{}); ok {
			merged = append(merged, arr...)
			continue
		}
		merged = append(merged, parsed)
	}

	out, err := json.Marshal(merged)
	if err != nil {
		return fmt.Errorf("concat_answers: marshal merged output: %w", err)
	}

	result := json.RawMessage(out)
	state.Set(n.outputKey, result)
	state.Set("extract_result", result)

	if n.verbose {
		log.Printf("[concat_answers] concatenated %d items into %d top-level entries",
			len(answers), len(merged))
	}

	return nil
}

func asRawMessages(v interface{}) ([]json.RawMessage, error) {
	switch typed := v.(type) {
	case []json.RawMessage:
		return typed, nil
	case []string:
		out := make([]json.RawMessage, 0, len(typed))
		for _, s := range typed {
			out = append(out, json.RawMessage(s))
		}
		return out, nil
	default:
		return nil, fmt.Errorf("'%v' is not []json.RawMessage or []string", v)
	}
}
