// Package models contains shared types used across the library
package models

import "encoding/json"

// Config holds global configuration for graph execution
type Config struct {
	LLMModel      string
	LLMAPIKey     string
	Temperature   float64
	MaxTokens     int
	Verbose       bool
	HTMLMaxChars  int
	ChunkSize     int
	ChunkOverlap  int
	Headless      bool   // use headless browser (Rod) instead of HTTP for fetching
	FetchStrategy string // "utls", "rod", or "auto"; empty preserves Headless-based behavior
}

// DefaultConfig returns sensible defaults
func DefaultConfig() *Config {
	return &Config{
		LLMModel:      "gpt-4o",
		Temperature:   0,
		MaxTokens:     4000,
		Verbose:       false,
		HTMLMaxChars:  50000,
		ChunkSize:     8000,
		ChunkOverlap:  200,
		Headless:      false, // default to UTLS HTTP fetching
		FetchStrategy: "",
	}
}

// FetchResult contains the result of a fetch operation
type FetchResult struct {
	HTML        string
	URL         string
	StatusCode  int
	Headers     map[string]string
	ElapsedSecs float64
	Error       error
}

// ParseResult contains parsed and chunked content
type ParseResult struct {
	Chunks     []string
	FullText   string
	ChunkCount int
	Error      error
}

// ExtractResult contains LLM extraction output
type ExtractResult struct {
	Data        json.RawMessage
	Model       string
	ElapsedSecs float64
	TokensUsed  int
	Error       error
}

// NodeInput is the interface for node execution input
type NodeInput interface{}

// NodeOutput is the interface for node execution output
type NodeOutput interface{}
