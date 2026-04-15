// Package nodes provides SearchLinkNode for link discovery from HTML
package nodes

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"stealthfetch/pkg/graph"
	"stealthfetch/pkg/llm"

	"golang.org/x/net/html"
)

// SearchLinkNode extracts links from an HTML page.
// Optionally uses LLM to filter links by relevance to the user's prompt.
type SearchLinkNode struct {
	*graph.BaseNode
	llmClient   llm.LLM
	prompt      string // user prompt for relevance filtering
	maxLinks    int
	filterByLLM bool
	verbose     bool
}

// SearchLinkConfig holds configuration for creating a SearchLinkNode
type SearchLinkConfig struct {
	LLMClient   llm.LLM
	Prompt      string // user prompt for relevance filtering
	MaxLinks    int    // max links to return, default 5
	FilterByLLM bool   // use LLM to filter relevant links, default true
	Verbose     bool
}

// NewSearchLinkNode creates a new search link node
func NewSearchLinkNode(cfg SearchLinkConfig) *SearchLinkNode {
	if cfg.MaxLinks <= 0 {
		cfg.MaxLinks = 5
	}
	return &SearchLinkNode{
		BaseNode: graph.NewBaseNode(
			"search_link",
			[]string{"html"},
			[]string{"links"},
		),
		llmClient:   cfg.LLMClient,
		prompt:      cfg.Prompt,
		maxLinks:    cfg.MaxLinks,
		filterByLLM: cfg.FilterByLLM,
		verbose:     cfg.Verbose,
	}
}

// Execute extracts links from the HTML in state
func (n *SearchLinkNode) Execute(ctx context.Context, state *graph.State) error {
	rawHTML, ok := state.GetString("html")
	if !ok {
		return fmt.Errorf("search_link: missing 'html' in state")
	}

	// Get the base URL for resolving relative links
	baseURL, _ := state.GetString("base_url")
	if baseURL == "" {
		baseURL, _ = state.GetString("url")
	}

	// Extract all links from HTML
	allLinks := extractLinks(rawHTML, baseURL)

	if n.verbose {
		log.Printf("[search_link] found %d raw links", len(allLinks))
	}

	// Deduplicate
	allLinks = deduplicateLinks(allLinks)

	if len(allLinks) == 0 {
		state.Set("links", []string{})
		return nil
	}

	// Filter by relevance if LLM is available
	var filtered []string
	if n.filterByLLM && n.llmClient != nil && n.prompt != "" {
		var err error
		filtered, err = n.filterLinksWithLLM(ctx, allLinks)
		if err != nil {
			if n.verbose {
				log.Printf("[search_link] LLM filter failed, using top links: %v", err)
			}
			filtered = topN(allLinks, n.maxLinks)
		}
	} else {
		filtered = topN(allLinks, n.maxLinks)
	}

	if n.verbose {
		log.Printf("[search_link] returning %d links", len(filtered))
	}

	state.Set("links", filtered)
	return nil
}

// filterLinksWithLLM asks the LLM to pick the most relevant links
func (n *SearchLinkNode) filterLinksWithLLM(ctx context.Context, links []string) ([]string, error) {
	// Build a numbered list of links
	var sb strings.Builder
	for i, link := range links {
		if i >= 50 { // cap at 50 to avoid token overflow
			break
		}
		fmt.Fprintf(&sb, "%d. %s\n", i+1, link)
	}

	systemPrompt := "You are a link relevance filter. Given a user's query and a list of URLs, " +
		"return ONLY the numbers of the most relevant URLs, one per line. " +
		"Return at most " + fmt.Sprintf("%d", n.maxLinks) + " numbers. Return ONLY numbers, nothing else."

	userPrompt := fmt.Sprintf("User query: %s\n\nLinks:\n%s", n.prompt, sb.String())

	response, err := n.llmClient.Generate(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

	// Parse the numbers from the response
	return parseNumberedLinks(response, links, n.maxLinks), nil
}

// extractLinks parses HTML and returns all href links
func extractLinks(rawHTML, baseURL string) []string {
	var links []string
	tokenizer := html.NewTokenizer(strings.NewReader(rawHTML))

	var base *url.URL
	if baseURL != "" {
		base, _ = url.Parse(baseURL)
	}

	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break
		}
		if tt != html.StartTagToken && tt != html.SelfClosingTagToken {
			continue
		}

		t := tokenizer.Token()
		if t.Data != "a" {
			continue
		}

		for _, attr := range t.Attr {
			if attr.Key != "href" {
				continue
			}
			href := strings.TrimSpace(attr.Val)
			if href == "" || strings.HasPrefix(href, "#") || strings.HasPrefix(href, "javascript:") {
				continue
			}

			// Resolve relative URLs
			if base != nil && !strings.HasPrefix(href, "http") {
				if ref, err := url.Parse(href); err == nil {
					href = base.ResolveReference(ref).String()
				}
			}

			// Only keep http/https links
			if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
				links = append(links, href)
			}
		}
	}

	return links
}

// deduplicateLinks removes duplicate URLs
func deduplicateLinks(links []string) []string {
	seen := make(map[string]bool)
	var unique []string
	for _, link := range links {
		normalized := strings.TrimRight(link, "/")
		if !seen[normalized] {
			seen[normalized] = true
			unique = append(unique, link)
		}
	}
	return unique
}

// parseNumberedLinks extracts link indices from LLM response
func parseNumberedLinks(response string, links []string, max int) []string {
	var result []string
	for _, line := range strings.Split(response, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var idx int
		if _, err := fmt.Sscanf(line, "%d", &idx); err == nil {
			if idx >= 1 && idx <= len(links) {
				result = append(result, links[idx-1])
				if len(result) >= max {
					break
				}
			}
		}
	}
	return result
}

// topN returns the first n items from a slice
func topN(items []string, n int) []string {
	if len(items) <= n {
		return items
	}
	return items[:n]
}
