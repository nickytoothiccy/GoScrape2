// Package graph implements the graph execution engine
package graph

import (
	"context"
	"fmt"
	"log"

	"stealthfetch/internal/models"
	"stealthfetch/pkg/telemetry"
)

// Graph represents a workflow of connected nodes
type Graph struct {
	nodes   []Node
	nodeMap map[string]Node     // name -> node for fast lookup
	edges   map[string][]string // node name -> list of next nodes
	config  *models.Config
}

// NewGraph creates a new graph with the given config
func NewGraph(config *models.Config) *Graph {
	if config == nil {
		config = models.DefaultConfig()
	}
	return &Graph{
		nodes:   make([]Node, 0),
		nodeMap: make(map[string]Node),
		edges:   make(map[string][]string),
		config:  config,
	}
}

// AddNode adds a node to the graph
func (g *Graph) AddNode(node Node) {
	g.nodes = append(g.nodes, node)
	g.nodeMap[node.Name()] = node
}

// AddEdge adds a directed edge between nodes
func (g *Graph) AddEdge(from, to string) error {
	if _, ok := g.nodeMap[from]; !ok {
		return fmt.Errorf("source node '%s' not found", from)
	}
	if _, ok := g.nodeMap[to]; !ok {
		return fmt.Errorf("target node '%s' not found", to)
	}

	g.edges[from] = append(g.edges[from], to)
	return nil
}

// GetNode returns a node by name
func (g *Graph) GetNode(name string) (Node, bool) {
	n, ok := g.nodeMap[name]
	return n, ok
}

// MaxIterations is the maximum number of node executions before
// the graph executor bails out to prevent infinite loops.
// BranchNode cycles are the typical cause of runaway execution.
const MaxIterations = 100

// Execute runs the graph with the given initial state.
// Supports both sequential execution and branching via BranchNode.
// Enforces MaxIterations to prevent infinite loops from branch cycles.
func (g *Graph) Execute(ctx context.Context, state *State) error {
	if len(g.nodes) == 0 {
		return nil
	}

	// Start with the first node
	current := g.nodes[0]
	iterations := 0

	for current != nil {
		iterations++
		if iterations > MaxIterations {
			return fmt.Errorf("graph exceeded %d iterations — possible infinite loop (last node: %s)", MaxIterations, current.Name())
		}

		if g.config.Verbose {
			log.Printf("[graph] executing node: %s (iteration %d)", current.Name(), iterations)
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		nodeStart := telemetry.Start()
		err := current.Execute(ctx, state)
		telemetry.LogNode(current.Name(), nodeStart, err)
		if err != nil {
			return fmt.Errorf("node %s failed: %w", current.Name(), err)
		}

		if g.config.Verbose {
			log.Printf("[graph] node %s completed", current.Name())
		}

		// Determine the next node
		current = g.nextNode(current, state)
	}

	return nil
}

// nextNode determines which node to execute next
func (g *Graph) nextNode(current Node, state *State) Node {
	// If this is a BranchNode, use its NextNode method
	if bn, ok := current.(BranchNode); ok {
		nextName := bn.NextNode(state)
		if nextName == "" {
			return nil
		}
		if next, ok := g.nodeMap[nextName]; ok {
			return next
		}
		return nil
	}

	// Otherwise, follow edges (take first edge as default path)
	edges := g.edges[current.Name()]
	if len(edges) == 0 {
		// No edges: try sequential fallback (next in order)
		return g.nextSequential(current)
	}

	if next, ok := g.nodeMap[edges[0]]; ok {
		return next
	}
	return nil
}

// nextSequential returns the next node in insertion order
func (g *Graph) nextSequential(current Node) Node {
	for i, n := range g.nodes {
		if n.Name() == current.Name() && i+1 < len(g.nodes) {
			return g.nodes[i+1]
		}
	}
	return nil
}

// Config returns the graph's configuration
func (g *Graph) Config() *models.Config {
	return g.config
}

// Nodes returns all nodes in the graph
func (g *Graph) Nodes() []Node {
	return g.nodes
}
