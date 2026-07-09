// Package main provides an HTTP API for the scrapegraph library
package main

import (
	"log"
	"net/http"
	"time"

	"stealthfetch/internal/envutil"
)

func main() {
	_ = envutil.LoadDotEnv(".env")

	config, err := loadLLMConfigFromEnv()
	if err != nil {
		log.Fatalf("LLM configuration: %v", err)
	}

	srv := &Server{config: config}
	mux := newMux(srv)

	addr := ":8899"
	log.Printf("stealthgraph server v2.0 listening on %s", addr)
	log.Printf("LLM provider=%s model=%s", config.LLMProvider, config.LLMModel)
	log.Printf("endpoints: POST /scrape | POST /document-scrape | POST /multi-scrape | POST /search | POST /depth-search | POST /fetch | POST /screenshot | GET /health")

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 180 * time.Second,
	}

	if err = server.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
