package main

import "net/http"

func newMux(srv *Server) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/scrape", srv.handleScrape)
	mux.HandleFunc("/document-scrape", srv.handleDocumentScrape)
	mux.HandleFunc("/multi-scrape", srv.handleMultiScrape)
	mux.HandleFunc("/search", srv.handleSearch)
	mux.HandleFunc("/depth-search", srv.handleDepthSearch)
	mux.HandleFunc("/fetch", srv.handleFetch)
	mux.HandleFunc("/health", srv.handleHealth)
	return mux
}
