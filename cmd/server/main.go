// Package main provides an HTTP API for the scrapegraph library
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"stealthfetch/internal/envutil"
	"stealthfetch/internal/models"
)

func main() {
	_ = envutil.LoadDotEnv(".env")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY env var required")
	}

	config := models.DefaultConfig()
	config.LLMAPIKey = apiKey

	srv := &Server{config: config}
	mux := newMux(srv)

	addr := ":8899"
	log.Printf("stealthgraph server v2.0 listening on %s", addr)
	log.Printf("endpoints: POST /scrape | POST /document-scrape | POST /multi-scrape | POST /search | POST /depth-search | POST /fetch | GET /health")

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 180 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
