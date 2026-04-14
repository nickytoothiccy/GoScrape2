// Package nodes provides the ConditionalNode for branching logic
package nodes

import (
	"context"
	"fmt"
	"log"

	"stealthfetch/pkg/graph"
)

// ConditionFunc evaluates state and returns true/false for branching
type ConditionFunc func(state *graph.State) bool

// ConditionalNode determines execution flow based on a condition
// It implements BranchNode to direct the graph to the true or false path
type ConditionalNode struct {
	*graph.BaseNode
	condition     ConditionFunc
	keyName       string // alternative: check if this key exists and is non-empty
	trueNodeName  string
	falseNodeName string
	lastResult    bool
	verbose       bool
}

// ConditionalConfig holds configuration for creating a ConditionalNode
type ConditionalConfig struct {
	// Name is the node's identifier (default: "conditional")
	Name string
	// Condition is a custom function to evaluate (takes priority over KeyName)
	Condition ConditionFunc
	// KeyName checks if this state key exists and has a non-empty value
	KeyName string
	// TrueNodeName is the node to execute if condition is true
	TrueNodeName string
	// FalseNodeName is the node to execute if condition is false
	FalseNodeName string
	// Verbose enables logging
	Verbose bool
}

// NewConditionalNode creates a new conditional branching node
func NewConditionalNode(cfg ConditionalConfig) *ConditionalNode {
	name := cfg.Name
	if name == "" {
		name = "conditional"
	}

	return &ConditionalNode{
		BaseNode: graph.NewBaseNode(
			name,
			nil, // no required inputs (depends on condition)
			nil, // no outputs (it's a control flow node)
		),
		condition:     cfg.Condition,
		keyName:       cfg.KeyName,
		trueNodeName:  cfg.TrueNodeName,
		falseNodeName: cfg.FalseNodeName,
		verbose:       cfg.Verbose,
	}
}

// Execute evaluates the condition and stores the result
func (n *ConditionalNode) Execute(ctx context.Context, state *graph.State) error {
	if n.trueNodeName == "" {
		return fmt.Errorf("conditional node: true_node_name not set")
	}

	// Evaluate condition
	if n.condition != nil {
		n.lastResult = n.condition(state)
	} else if n.keyName != "" {
		n.lastResult = n.evaluateKey(state)
	} else {
		return fmt.Errorf("conditional node: no condition or key_name configured")
	}

	if n.verbose {
		log.Printf("[conditional] condition result: %v (next: %s)",
			n.lastResult, n.nextNodeName())
	}

	// Store the branch decision in state for observability
	state.Set("_branch_result", n.lastResult)

	return nil
}

// NextNode returns the name of the next node based on condition result
// Implements the BranchNode interface
func (n *ConditionalNode) NextNode(state *graph.State) string {
	return n.nextNodeName()
}

func (n *ConditionalNode) nextNodeName() string {
	if n.lastResult {
		return n.trueNodeName
	}
	return n.falseNodeName
}

// evaluateKey checks if the key exists in state and has a non-empty value
func (n *ConditionalNode) evaluateKey(state *graph.State) bool {
	if !state.Has(n.keyName) {
		return false
	}

	val, ok := state.Get(n.keyName)
	if !ok {
		return false
	}

	// Check for common "empty" values
	switch v := val.(type) {
	case string:
		return v != ""
	case []string:
		return len(v) > 0
	case nil:
		return false
	default:
		return true
	}
}

// --- Common condition factory functions ---

// HasKey returns a ConditionFunc that checks if a key exists in state
func HasKey(key string) ConditionFunc {
	return func(state *graph.State) bool {
		return state.Has(key)
	}
}

// HasNonEmptyKey returns a ConditionFunc that checks for non-empty key
func HasNonEmptyKey(key string) ConditionFunc {
	return func(state *graph.State) bool {
		val, ok := state.GetString(key)
		return ok && val != ""
	}
}

// RetryCondition returns a ConditionFunc that allows N retries
// Tracks retry count in state under the given counter key
func RetryCondition(maxRetries int, counterKey string) ConditionFunc {
	return func(state *graph.State) bool {
		val, ok := state.Get(counterKey)
		if !ok {
			state.Set(counterKey, 1)
			return true // first retry
		}
		count, ok := val.(int)
		if !ok {
			return false
		}
		if count < maxRetries {
			state.Set(counterKey, count+1)
			return true
		}
		return false // exceeded max retries
	}
}
