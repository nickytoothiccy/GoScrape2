// Package utils provides shared utility functions
package utils

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// SearchResult represents a single search result
type SearchResult struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// SearchFunc defines the interface for web search implementations
type SearchFunc func(query string, maxResults int) ([]SearchResult, error)

// DuckDuckGoSearch performs a search using DuckDuckGo HTML
func DuckDuckGoSearch(query string, maxResults int) ([]SearchResult, error) {
	if maxResults <= 0 {
		maxResults = 5
	}

	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s",
		url.QueryEscape(query))

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("search request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search fetch: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("search read: %w", err)
	}

	return parseDDGResults(string(body), maxResults), nil
}

// parseDDGResults extracts URLs from DuckDuckGo HTML results
func parseDDGResults(html string, maxResults int) []SearchResult {
	var results []SearchResult

	// DuckDuckGo HTML results have links in <a class="result__a" href="...">
	linkPattern := regexp.MustCompile(`<a[^>]+class="result__a"[^>]+href="([^"]+)"[^>]*>([^<]*)`)
	matches := linkPattern.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(results) >= maxResults {
			break
		}
		if len(match) >= 3 {
			rawURL := match[1]
			title := strings.TrimSpace(match[2])

			// DDG wraps URLs in a redirect, extract the actual URL
			actualURL := extractDDGURL(rawURL)
			if actualURL != "" && !strings.Contains(actualURL, "duckduckgo.com") {
				results = append(results, SearchResult{
					URL:   actualURL,
					Title: title,
				})
			}
		}
	}

	// Fallback: try to find any reasonable URLs if the pattern didn't match
	if len(results) == 0 {
		urlPattern := regexp.MustCompile(`href="(https?://[^"]+)"`)
		matches := urlPattern.FindAllStringSubmatch(html, -1)
		seen := make(map[string]bool)
		for _, match := range matches {
			if len(results) >= maxResults {
				break
			}
			if len(match) >= 2 {
				u := match[1]
				if !strings.Contains(u, "duckduckgo.com") && !seen[u] {
					seen[u] = true
					results = append(results, SearchResult{URL: u})
				}
			}
		}
	}

	return results
}

// extractDDGURL extracts the actual URL from DDG redirect URL
func extractDDGURL(rawURL string) string {
	// DDG uses //duckduckgo.com/l/?uddg=ENCODED_URL&... format
	if strings.Contains(rawURL, "uddg=") {
		parsed, err := url.Parse(rawURL)
		if err == nil {
			uddg := parsed.Query().Get("uddg")
			if uddg != "" {
				return uddg
			}
		}
	}

	// If it's already a direct URL
	if strings.HasPrefix(rawURL, "http") {
		return rawURL
	}

	return ""
}

// ExtractURLs returns just the URLs from search results
func ExtractURLs(results []SearchResult) []string {
	urls := make([]string, 0, len(results))
	for _, r := range results {
		urls = append(urls, r.URL)
	}
	return urls
}
