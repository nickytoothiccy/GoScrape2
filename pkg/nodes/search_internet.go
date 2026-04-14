// Package nodes provides the SearchInternetNode for web search
package nodes

import (
	"context"
	"fmt"
	"log"
	"strings"

	"stealthfetch/pkg/graph"
	"stealthfetch/pkg/llm"
	"stealthfetch/pkg/utils"
)

// SearchInternetNode generates search queries from user input,
// searches the web, and returns discovered URLs
type SearchInternetNode struct {
	*graph.BaseNode
	llmClient  llm.LLM
	searchFunc utils.SearchFunc
	maxResults int
	verbose    bool
}

// SearchInternetConfig holds configuration for the SearchInternetNode
type SearchInternetConfig struct {
	LLMClient  llm.LLM
	SearchFunc utils.SearchFunc // nil defaults to DuckDuckGo
	MaxResults int              // default 3
	Verbose    bool
}

// NewSearchInternetNode creates a new search internet node
func NewSearchInternetNode(cfg SearchInternetConfig) *SearchInternetNode {
	searchFn := cfg.SearchFunc
	if searchFn == nil {
		searchFn = utils.DuckDuckGoSearch
	}

	maxResults := cfg.MaxResults
	if maxResults <= 0 {
		maxResults = 3
	}

	return &SearchInternetNode{
		BaseNode: graph.NewBaseNode(
			"search_internet",
			[]string{"user_prompt"},
			[]string{"urls", "search_results"},
		),
		llmClient:  cfg.LLMClient,
		searchFunc: searchFn,
		maxResults: maxResults,
		verbose:    cfg.Verbose,
	}
}

// Execute generates a search query and searches the web
func (n *SearchInternetNode) Execute(ctx context.Context, state *graph.State) error {
	if err := n.ValidateInputs(state); err != nil {
		return err
	}

	userPrompt, ok := state.GetString("user_prompt")
	if !ok {
		return fmt.Errorf("search_internet: user_prompt not found")
	}

	// Step 1: Use LLM to generate an optimized search query
	searchQuery, err := n.generateSearchQuery(ctx, userPrompt)
	if err != nil {
		// Fall back to using the prompt directly
		searchQuery = userPrompt
	}

	if n.verbose {
		log.Printf("[search_internet] search query: %s", searchQuery)
	}

	// Step 2: Perform the web search
	results, err := n.searchFunc(searchQuery, n.maxResults)
	if err != nil {
		return fmt.Errorf("search_internet: search failed: %w", err)
	}

	if len(results) == 0 {
		return fmt.Errorf("search_internet: no results found for query '%s'", searchQuery)
	}

	if n.verbose {
		log.Printf("[search_internet] found %d results", len(results))
		for i, r := range results {
			log.Printf("[search_internet]   %d: %s", i+1, r.URL)
		}
	}

	// Store results in state
	urls := utils.ExtractURLs(results)
	state.Set("urls", urls)
	state.Set("search_results", results)
	state.Set("search_query", searchQuery)

	return nil
}

// generateSearchQuery uses LLM to create an optimized search query
func (n *SearchInternetNode) generateSearchQuery(ctx context.Context, userPrompt string) (string, error) {
	systemPrompt := `You are a search query optimizer. Given a user's information need, generate a single effective search query. Return ONLY the search query text, nothing else. No quotes, no explanation.`

	userMsg := fmt.Sprintf("Generate a search query for: %s", userPrompt)

	result, err := n.llmClient.Generate(ctx, systemPrompt, userMsg)
	if err != nil {
		return "", err
	}

	// Clean up the result
	query := strings.TrimSpace(result)
	query = strings.Trim(query, `"'`)

	if query == "" {
		return userPrompt, nil
	}

	return query, nil
}
