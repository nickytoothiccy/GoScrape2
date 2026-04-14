package main

import (
	"encoding/json"

	"stealthfetch/internal/models"
)

type Server struct{ config *models.Config }

type ScrapeRequest struct {
	URL        string `json:"url"`
	Prompt     string `json:"prompt"`
	Model      string `json:"model,omitempty"`
	SchemaHint string `json:"schema_hint,omitempty"`
	Headless   bool   `json:"headless,omitempty"`
}

type ScrapeResponse struct {
	Result    json.RawMessage `json:"result"`
	Model     string          `json:"model_used"`
	TotalTime float64         `json:"total_time_s"`
}

type MultiScrapeRequest struct {
	URLs          []string `json:"urls"`
	Prompt        string   `json:"prompt"`
	Model         string   `json:"model,omitempty"`
	SchemaHint    string   `json:"schema_hint,omitempty"`
	Headless      bool     `json:"headless,omitempty"`
	ConcatResults bool     `json:"concat_results,omitempty"`
	PerURLTimeout int      `json:"per_url_timeout_ms,omitempty"`
}

type MultiScrapeResponse struct {
	Result     json.RawMessage `json:"result"`
	URLs       []string        `json:"urls"`
	FailedURLs []string        `json:"failed_urls,omitempty"`
	Model      string          `json:"model_used"`
	TotalTime  float64         `json:"total_time_s"`
}

type DocumentScrapeRequest struct {
	Path       string `json:"path"`
	Prompt     string `json:"prompt"`
	Model      string `json:"model,omitempty"`
	SchemaHint string `json:"schema_hint,omitempty"`
}

type DocumentScrapeResponse struct {
	Result    json.RawMessage `json:"result"`
	Path      string          `json:"path"`
	Model     string          `json:"model_used"`
	TotalTime float64         `json:"total_time_s"`
}

type SearchRequest struct {
	Prompt     string `json:"prompt"`
	Model      string `json:"model,omitempty"`
	SchemaHint string `json:"schema_hint,omitempty"`
	MaxResults int    `json:"max_results,omitempty"`
}

type SearchResponse struct {
	Result    json.RawMessage `json:"result"`
	URLs      []string        `json:"urls_scraped"`
	Model     string          `json:"model_used"`
	TotalTime float64         `json:"total_time_s"`
}

type DepthSearchRequest struct {
	URL             string `json:"url"`
	Prompt          string `json:"prompt"`
	Model           string `json:"model,omitempty"`
	SchemaHint      string `json:"schema_hint,omitempty"`
	MaxDepth        int    `json:"max_depth,omitempty"`
	MaxLinksPerPage int    `json:"max_links_per_page,omitempty"`
	FilterByLLM     bool   `json:"filter_by_llm,omitempty"`
	Headless        bool   `json:"headless,omitempty"`
}

type DepthSearchResponse struct {
	Result      json.RawMessage `json:"result"`
	VisitedURLs []string        `json:"visited_urls"`
	Model       string          `json:"model_used"`
	TotalTime   float64         `json:"total_time_s"`
}

type FetchRequest struct {
	URL      string `json:"url"`
	Profile  string `json:"profile,omitempty"`
	Timeout  int    `json:"timeout_ms,omitempty"`
	Headless bool   `json:"headless,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
