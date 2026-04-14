// Package nodes provides the MergeAnswersNode for combining multi-source results
package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"stealthfetch/pkg/graph"
	"stealthfetch/pkg/llm"
)

// MergeAnswersNode combines extraction results from multiple sources
// into a single coherent answer. Used by SearchGraph and Multi graphs.
type MergeAnswersNode struct {
	*graph.BaseNode
	llmClient llm.LLM
	prompt    string
	verbose   bool
}

// NewMergeAnswersNode creates a new merge answers node
func NewMergeAnswersNode(llmClient llm.LLM, prompt string, verbose bool) *MergeAnswersNode {
	return &MergeAnswersNode{
		BaseNode: graph.NewBaseNode(
			"merge_answers",
			[]string{"answers"}, // list of extraction results
			[]string{"extracted_data", "extract_result"}, // merged output
		),
		llmClient: llmClient,
		prompt:    prompt,
		verbose:   verbose,
	}
}

// Execute merges multiple extraction results into one
func (n *MergeAnswersNode) Execute(ctx context.Context, state *graph.State) error {
	if err := n.ValidateInputs(state); err != nil {
		return err
	}

	// Get the answers to merge
	answersRaw, ok := state.Get("answers")
	if !ok {
		return fmt.Errorf("merge_answers: no answers in state")
	}

	answers, ok := answersRaw.([]json.RawMessage)
	if !ok {
		return fmt.Errorf("merge_answers: answers is not []json.RawMessage")
	}

	if len(answers) == 0 {
		return fmt.Errorf("merge_answers: empty answers list")
	}

	// Single answer: pass through directly
	if len(answers) == 1 {
		state.Set("extracted_data", answers[0])
		return nil
	}

	if n.verbose {
		log.Printf("[merge_answers] merging %d results with %s",
			len(answers), n.llmClient.ModelName())
	}

	// Build the merge prompt
	userPrompt := n.buildMergePrompt(answers)

	systemPrompt := `You are a data merging engine. Combine multiple extraction results from different web pages into a single, deduplicated, coherent JSON result. Preserve all unique data points and resolve conflicts by keeping the most complete version.`

	// Call LLM to merge
	merged, err := n.llmClient.GenerateJSON(ctx, systemPrompt, userPrompt)
	if err != nil {
		return fmt.Errorf("merge_answers: llm merge: %w", err)
	}

	// Attach source URLs if available
	merged = n.attachSources(state, merged)

	if n.verbose {
		log.Printf("[merge_answers] merge complete (%d bytes)", len(merged))
	}

	state.Set("extracted_data", merged)
	return nil
}

// buildMergePrompt creates the prompt for merging multiple results
func (n *MergeAnswersNode) buildMergePrompt(answers []json.RawMessage) string {
	var sb strings.Builder
	sb.WriteString("## Task\n")
	sb.WriteString("Merge the following extraction results into a single coherent result.\n\n")
	sb.WriteString("## Original Prompt\n")
	sb.WriteString(n.prompt)
	sb.WriteString("\n\n## Results from Different Sources\n")

	for i, answer := range answers {
		sb.WriteString(fmt.Sprintf("### Source %d\n", i+1))
		sb.WriteString(string(answer))
		sb.WriteString("\n\n")
	}

	return sb.String()
}

// attachSources adds source URLs to the merged result if available
func (n *MergeAnswersNode) attachSources(state *graph.State, data json.RawMessage) json.RawMessage {
	urls, ok := state.GetStringSlice("urls")
	if !ok || len(urls) == 0 {
		urls, ok = state.GetStringSlice("considered_urls")
		if !ok || len(urls) == 0 {
			return data
		}
	}

	// Try to add sources to the JSON
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return data // can't parse, return as-is
	}

	result["sources"] = urls
	enriched, err := json.Marshal(result)
	if err != nil {
		return data
	}

	return json.RawMessage(enriched)
}
