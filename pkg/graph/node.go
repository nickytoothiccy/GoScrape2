// Package graph provides the core graph execution engine
package graph

import (
	"context"
	"fmt"
)

// Node represents a single processing step in the graph
type Node interface {
	// Name returns the node's identifier
	Name() string

	// Execute runs the node's logic with the given state
	Execute(ctx context.Context, state *State) error

	// InputKeys returns the state keys this node reads from
	InputKeys() []string

	// OutputKeys returns the state keys this node writes to
	OutputKeys() []string
}

// BranchNode is a node that can direct execution flow
// After execution, it returns the name of the next node to run
type BranchNode interface {
	Node

	// NextNode returns the name of the next node to execute
	// Called after Execute completes successfully
	NextNode(state *State) string
}

// BaseNode provides common node functionality
type BaseNode struct {
	name       string
	inputKeys  []string
	outputKeys []string
}

// NewBaseNode creates a new base node
func NewBaseNode(name string, inputs, outputs []string) *BaseNode {
	return &BaseNode{
		name:       name,
		inputKeys:  inputs,
		outputKeys: outputs,
	}
}

// Name returns the node's name
func (n *BaseNode) Name() string {
	return n.name
}

// InputKeys returns input state keys
func (n *BaseNode) InputKeys() []string {
	return n.inputKeys
}

// OutputKeys returns output state keys
func (n *BaseNode) OutputKeys() []string {
	return n.outputKeys
}

// ValidateInputs checks if required inputs are present in state
func (n *BaseNode) ValidateInputs(state *State) error {
	for _, key := range n.inputKeys {
		if !state.Has(key) {
			return fmt.Errorf("node %s: missing required input key '%s'", n.name, key)
		}
	}
	return nil
}
