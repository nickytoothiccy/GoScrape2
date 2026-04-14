// Package nodes provides the ReasoningNode for chain-of-thought analysis
package nodes

import (
	"context"
	"fmt"
	"log"

	"stealthfetch/pkg/graph"
	"stealthfetch/pkg/llm"
	"stealthfetch/pkg/prompts"
)

// ReasoningNode performs chain-of-thought analysis before extraction
// It analyzes page content and generates a reasoning context that
// improves the quality of subsequent data extraction
type ReasoningNode struct {
	*graph.BaseNode
	llmClient      llm.LLM
	prompt         string // the user's extraction prompt
	schemaHint     string
	additionalInfo string
	verbose        bool
}

// ReasoningConfig holds configuration for creating a ReasoningNode
type ReasoningConfig struct {
	LLMClient      llm.LLM
	Prompt         string // user's extraction prompt
	SchemaHint     string // optional JSON schema hint
	AdditionalInfo string // optional extra context
	Verbose        bool
}

// NewReasoningNode creates a new chain-of-thought reasoning node
func NewReasoningNode(cfg ReasoningConfig) *ReasoningNode {
	return &ReasoningNode{
		BaseNode: graph.NewBaseNode(
			"reasoning",
			[]string{"chunks"},        // needs parsed content
			[]string{"reasoning_out"}, // produces reasoning analysis
		),
		llmClient:      cfg.LLMClient,
		prompt:         cfg.Prompt,
		schemaHint:     cfg.SchemaHint,
		additionalInfo: cfg.AdditionalInfo,
		verbose:        cfg.Verbose,
	}
}

// Execute performs chain-of-thought reasoning on the content
func (n *ReasoningNode) Execute(ctx context.Context, state *graph.State) error {
	if err := n.ValidateInputs(state); err != nil {
		return err
	}

	chunks, ok := state.GetStringSlice("chunks")
	if !ok || len(chunks) == 0 {
		return fmt.Errorf("reasoning: no chunks available")
	}

	// Use first chunk (or concatenated content) for reasoning
	content := chunks[0]
	if len(chunks) > 1 {
		content = n.summarizeChunks(chunks)
	}

	// Build reasoning prompt
	systemPrompt, userPrompt, err := n.buildPrompts(content)
	if err != nil {
		return fmt.Errorf("reasoning: build prompts: %w", err)
	}

	if n.verbose {
		log.Printf("[reasoning] analyzing content (%d chars) with %s",
			len(content), n.llmClient.ModelName())
	}

	// Call LLM for chain-of-thought analysis
	reasoning, err := n.llmClient.Generate(ctx, systemPrompt, userPrompt)
	if err != nil {
		return fmt.Errorf("reasoning: llm generate: %w", err)
	}

	if n.verbose {
		log.Printf("[reasoning] analysis complete (%d chars)", len(reasoning))
	}

	// Store reasoning output in state
	state.Set("reasoning_out", reasoning)

	// Also build and store the enhanced prompt for GenerateAnswerNode
	enhanced, err := n.buildEnhancedPrompt(reasoning, content)
	if err != nil {
		// Fall back to original prompt if template fails
		state.Set("enhanced_prompt", n.prompt)
	} else {
		state.Set("enhanced_prompt", enhanced)
	}

	return nil
}

// buildPrompts creates the system and user prompts for reasoning
func (n *ReasoningNode) buildPrompts(content string) (string, string, error) {
	// Get reasoning templates
	sysTmpl, ok := prompts.DefaultRegistry.Get("reasoning_system")
	if !ok {
		return n.fallbackSystemPrompt(), n.fallbackUserPrompt(content), nil
	}

	usrTmpl, ok := prompts.DefaultRegistry.Get("reasoning_user")
	if !ok {
		return n.fallbackSystemPrompt(), n.fallbackUserPrompt(content), nil
	}

	sysPrompt, err := sysTmpl.Render(nil)
	if err != nil {
		return n.fallbackSystemPrompt(), n.fallbackUserPrompt(content), nil
	}

	vars := map[string]string{
		"prompt":  n.prompt,
		"content": content,
	}
	usrPrompt, err := usrTmpl.Render(vars)
	if err != nil {
		return n.fallbackSystemPrompt(), n.fallbackUserPrompt(content), nil
	}

	return sysPrompt, usrPrompt, nil
}

// buildEnhancedPrompt creates a prompt enhanced with reasoning context
func (n *ReasoningNode) buildEnhancedPrompt(reasoning, content string) (string, error) {
	tmpl, ok := prompts.DefaultRegistry.Get("reasoning_guided_extraction_user")
	if !ok {
		return "", fmt.Errorf("template not found")
	}

	return tmpl.Render(map[string]string{
		"prompt":   n.prompt,
		"analysis": reasoning,
		"content":  content,
	})
}

// summarizeChunks joins chunk beginnings for a content overview
func (n *ReasoningNode) summarizeChunks(chunks []string) string {
	const maxChars = 8000
	var result string
	for _, chunk := range chunks {
		if len(result)+len(chunk) > maxChars {
			remaining := maxChars - len(result)
			if remaining > 0 {
				result += chunk[:remaining] + "\n[TRUNCATED]"
			}
			break
		}
		result += chunk + "\n---\n"
	}
	return result
}

func (n *ReasoningNode) fallbackSystemPrompt() string {
	return `You are an expert web content analyst. Analyze the content and reason about how to extract the requested information. Think step by step. Return your analysis as JSON.`
}

func (n *ReasoningNode) fallbackUserPrompt(content string) string {
	return fmt.Sprintf("## Request\n%s\n\n## Content\n%s\n\nAnalyze this content.",
		n.prompt, content)
}
