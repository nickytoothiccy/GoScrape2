// Package loaders provides document loading helpers for local PDF/DOCX files.
package loaders

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"stealthfetch/internal/models"
)

// DocumentLoader loads local documents and returns extracted text in HTML.
type DocumentLoader struct {
	pdfExtractor  func(string) (string, error)
	docxExtractor func(string) (string, error)
}

// NewDocumentLoader creates a document loader with builtin extractors.
func NewDocumentLoader() *DocumentLoader {
	return &DocumentLoader{pdfExtractor: extractPDFText, docxExtractor: extractDOCXText}
}

// Name returns the loader identifier.
func (l *DocumentLoader) Name() string { return "document" }

// Load extracts local file text and wraps it as minimal HTML.
func (l *DocumentLoader) Load(_ context.Context, source string) (*models.FetchResult, error) {
	start := time.Now()
	info, err := os.Stat(source)
	if err != nil {
		return nil, fmt.Errorf("document load: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("document load: source is a directory")
	}
	ext := strings.ToLower(filepath.Ext(source))
	var text string
	switch ext {
	case ".pdf":
		text, err = l.pdfExtractor(source)
	case ".docx":
		text, err = l.docxExtractor(source)
	default:
		return nil, fmt.Errorf("document load: unsupported extension %q", ext)
	}
	if err != nil {
		return nil, err
	}
	html := "<html><body><pre>" + escapeHTML(text) + "</pre></body></html>"
	return &models.FetchResult{HTML: html, URL: source, StatusCode: 200, Headers: map[string]string{"x-loader": "document", "x-document-ext": ext}, ElapsedSecs: time.Since(start).Seconds()}, nil
}

func escapeHTML(s string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return replacer.Replace(s)
}
