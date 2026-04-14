package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"stealthfetch/pkg/loaders"
	"stealthfetch/pkg/scrapegraph"
	"stealthfetch/pkg/telemetry"
)

func (s *Server) handleScrape(w http.ResponseWriter, r *http.Request) {
	start := telemetry.Start()
	status := http.StatusOK
	var loggedErr error
	defer func() { telemetry.LogHTTP(r, status, start, "scrape", loggedErr) }()
	if r.Method != http.MethodPost {
		status = http.StatusMethodNotAllowed
		loggedErr = fmt.Errorf("POST only")
		writeError(w, status, loggedErr.Error())
		return
	}
	var req ScrapeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		status = http.StatusBadRequest
		loggedErr = err
		writeError(w, status, err.Error())
		return
	}
	if req.URL == "" || req.Prompt == "" {
		status = http.StatusBadRequest
		loggedErr = fmt.Errorf("url and prompt are required")
		writeError(w, status, loggedErr.Error())
		return
	}
	cfg := *s.config
	if req.Model != "" {
		cfg.LLMModel = req.Model
	}
	cfg.Headless = req.Headless
	scraper := scrapegraph.NewSmartScraperGraph(req.Prompt, req.URL, &cfg, req.SchemaHint)
	result, err := scraper.Run(r.Context())
	if err != nil {
		status = http.StatusBadGateway
		loggedErr = err
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, status, ScrapeResponse{Result: result, Model: cfg.LLMModel, TotalTime: time.Since(start).Seconds()})
}

func (s *Server) handleMultiScrape(w http.ResponseWriter, r *http.Request) {
	start := telemetry.Start()
	status := http.StatusOK
	var loggedErr error
	defer func() { telemetry.LogHTTP(r, status, start, "multi_scrape", loggedErr) }()
	if r.Method != http.MethodPost {
		status = http.StatusMethodNotAllowed
		loggedErr = fmt.Errorf("POST only")
		writeError(w, status, loggedErr.Error())
		return
	}
	var req MultiScrapeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		status = http.StatusBadRequest
		loggedErr = err
		writeError(w, status, err.Error())
		return
	}
	if len(req.URLs) == 0 || req.Prompt == "" {
		status = http.StatusBadRequest
		loggedErr = fmt.Errorf("urls and prompt are required")
		writeError(w, status, loggedErr.Error())
		return
	}
	cfg := *s.config
	if req.Model != "" {
		cfg.LLMModel = req.Model
	}
	cfg.Headless = req.Headless
	perURL := time.Duration(req.PerURLTimeout) * time.Millisecond
	multi := scrapegraph.NewSmartScraperMultiGraph(scrapegraph.SmartScraperMultiConfig{
		Config:        &cfg,
		Prompt:        req.Prompt,
		SchemaHint:    req.SchemaHint,
		URLs:          req.URLs,
		PerURLTimeout: perURL,
		ConcatResults: req.ConcatResults,
	})
	result, err := multi.Run(r.Context())
	if err != nil {
		status = http.StatusBadGateway
		loggedErr = err
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, status, MultiScrapeResponse{Result: result, URLs: req.URLs, FailedURLs: multi.GetFailedURLs(), Model: cfg.LLMModel, TotalTime: time.Since(start).Seconds()})
}

func (s *Server) handleFetch(w http.ResponseWriter, r *http.Request) {
	start := telemetry.Start()
	status := http.StatusOK
	var loggedErr error
	defer func() { telemetry.LogHTTP(r, status, start, "fetch", loggedErr) }()
	if r.Method != http.MethodPost {
		status = http.StatusMethodNotAllowed
		loggedErr = fmt.Errorf("POST only")
		writeError(w, status, loggedErr.Error())
		return
	}
	var req FetchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		status = http.StatusBadRequest
		loggedErr = err
		writeError(w, status, err.Error())
		return
	}
	if req.URL == "" {
		status = http.StatusBadRequest
		loggedErr = fmt.Errorf("url is required")
		writeError(w, status, loggedErr.Error())
		return
	}
	timeout := time.Duration(req.Timeout) * time.Millisecond
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	var loader loaders.Loader
	if req.Headless {
		loader = loaders.NewDefaultRodLoader(false)
	} else {
		loader = loaders.NewUTLSLoader(req.Profile, "", timeout)
	}
	result, err := loader.Load(context.Background(), req.URL)
	if err != nil {
		status = http.StatusBadGateway
		loggedErr = err
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, status, map[string]interface{}{"status": result.StatusCode, "headers": result.Headers, "body": result.HTML, "elapsed": result.ElapsedSecs})
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	start := telemetry.Start()
	status := http.StatusOK
	var loggedErr error
	defer func() { telemetry.LogHTTP(r, status, start, "search", loggedErr) }()
	if r.Method != http.MethodPost {
		status = http.StatusMethodNotAllowed
		loggedErr = fmt.Errorf("POST only")
		writeError(w, status, loggedErr.Error())
		return
	}
	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		status = http.StatusBadRequest
		loggedErr = err
		writeError(w, status, err.Error())
		return
	}
	if req.Prompt == "" {
		status = http.StatusBadRequest
		loggedErr = fmt.Errorf("prompt is required")
		writeError(w, status, loggedErr.Error())
		return
	}
	cfg := *s.config
	if req.Model != "" {
		cfg.LLMModel = req.Model
	}
	sg := scrapegraph.NewSearchGraph(scrapegraph.SearchGraphConfig{Config: &cfg, Prompt: req.Prompt, SchemaHint: req.SchemaHint, MaxResults: req.MaxResults})
	result, err := sg.Run(r.Context())
	if err != nil {
		status = http.StatusBadGateway
		loggedErr = err
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, status, SearchResponse{Result: result, URLs: sg.GetConsideredURLs(), Model: cfg.LLMModel, TotalTime: time.Since(start).Seconds()})
}

func (s *Server) handleDepthSearch(w http.ResponseWriter, r *http.Request) {
	start := telemetry.Start()
	status := http.StatusOK
	var loggedErr error
	defer func() { telemetry.LogHTTP(r, status, start, "depth_search", loggedErr) }()
	if r.Method != http.MethodPost {
		status = http.StatusMethodNotAllowed
		loggedErr = fmt.Errorf("POST only")
		writeError(w, status, loggedErr.Error())
		return
	}
	var req DepthSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		status = http.StatusBadRequest
		loggedErr = err
		writeError(w, status, err.Error())
		return
	}
	if req.URL == "" || req.Prompt == "" {
		status = http.StatusBadRequest
		loggedErr = fmt.Errorf("url and prompt are required")
		writeError(w, status, loggedErr.Error())
		return
	}
	cfg := *s.config
	if req.Model != "" {
		cfg.LLMModel = req.Model
	}
	cfg.Headless = req.Headless
	dsg := scrapegraph.NewDepthSearchGraph(scrapegraph.DepthSearchConfig{Config: &cfg, Prompt: req.Prompt, SchemaHint: req.SchemaHint, Source: req.URL, MaxDepth: req.MaxDepth, MaxLinksPerPage: req.MaxLinksPerPage, FilterByLLM: req.FilterByLLM})
	result, err := dsg.Run(r.Context())
	if err != nil {
		status = http.StatusBadGateway
		loggedErr = err
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, status, DepthSearchResponse{Result: result, VisitedURLs: dsg.GetVisitedURLs(), Model: cfg.LLMModel, TotalTime: time.Since(start).Seconds()})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	start := telemetry.Start()
	defer telemetry.LogHTTP(r, http.StatusOK, start, "health", nil)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": "1.0.0"})
}
