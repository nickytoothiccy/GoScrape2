package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"stealthfetch/pkg/scrapegraph"
	"stealthfetch/pkg/telemetry"
)

func (s *Server) handleDocumentScrape(w http.ResponseWriter, r *http.Request) {
	start := telemetry.Start()
	status := http.StatusOK
	var loggedErr error
	defer func() { telemetry.LogHTTP(r, status, start, "document_scrape", loggedErr) }()
	if r.Method != http.MethodPost {
		status = http.StatusMethodNotAllowed
		loggedErr = fmt.Errorf("POST only")
		writeError(w, status, loggedErr.Error())
		return
	}
	var req DocumentScrapeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		status = http.StatusBadRequest
		loggedErr = err
		writeError(w, status, err.Error())
		return
	}
	if req.Path == "" || req.Prompt == "" {
		status = http.StatusBadRequest
		loggedErr = fmt.Errorf("path and prompt are required")
		writeError(w, status, loggedErr.Error())
		return
	}
	cfg := *s.config
	if req.Model != "" {
		cfg.LLMModel = req.Model
	}
	graph := scrapegraph.NewDocumentScraperGraph(req.Prompt, req.Path, &cfg, req.SchemaHint)
	result, err := graph.Run(r.Context())
	if err != nil {
		status = http.StatusBadGateway
		loggedErr = err
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, status, DocumentScrapeResponse{Result: result, Path: req.Path, Model: cfg.LLMModel, TotalTime: time.Since(start).Seconds()})
}
