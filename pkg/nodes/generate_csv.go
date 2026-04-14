package nodes

import "stealthfetch/pkg/llm"

// GenerateAnswerCSVNode performs CSV-focused extraction.
// Current parity slice reuses the standard extraction path until CSV-specific prompting is expanded.
type GenerateAnswerCSVNode struct{ *GenerateAnswerNode }

// NewGenerateAnswerCSVNode creates a CSV extraction node.
func NewGenerateAnswerCSVNode(llmClient llm.LLM, prompt, schemaHint string) *GenerateAnswerCSVNode {
	return &GenerateAnswerCSVNode{GenerateAnswerNode: NewGenerateAnswerNode(llmClient, prompt, schemaHint)}
}
