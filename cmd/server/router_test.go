package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"stealthfetch/internal/models"
)

func TestMultiScrapeRequiresPOST(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/multi-scrape", nil)
	rr := httptest.NewRecorder()
	newMux(&Server{config: models.DefaultConfig()}).ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestMultiScrapeValidatesBody(t *testing.T) {
	body, _ := json.Marshal(MultiScrapeRequest{Prompt: "x"})
	req := httptest.NewRequest(http.MethodPost, "/multi-scrape", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	newMux(&Server{config: models.DefaultConfig()}).ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestDocumentScrapeRequiresPOST(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/document-scrape", nil)
	rr := httptest.NewRecorder()
	newMux(&Server{config: models.DefaultConfig()}).ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestDocumentScrapeValidatesBody(t *testing.T) {
	body, _ := json.Marshal(DocumentScrapeRequest{Prompt: "x"})
	req := httptest.NewRequest(http.MethodPost, "/document-scrape", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	newMux(&Server{config: models.DefaultConfig()}).ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}
}
