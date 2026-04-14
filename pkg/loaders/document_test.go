package loaders

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocumentLoader_UnsupportedExtension(t *testing.T) {
	path := filepath.Join(t.TempDir(), "note.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := NewDocumentLoader().Load(context.Background(), path)
	if err == nil || !strings.Contains(err.Error(), "unsupported extension") {
		t.Fatalf("expected unsupported extension error, got %v", err)
	}
}

func TestDocumentLoader_CustomExtractor(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample.pdf")
	if err := os.WriteFile(path, []byte("stub"), 0o600); err != nil {
		t.Fatal(err)
	}
	loader := &DocumentLoader{pdfExtractor: func(string) (string, error) { return "Hello <world>", nil }, docxExtractor: func(string) (string, error) { return "", nil }}
	result, err := loader.Load(context.Background(), path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if !strings.Contains(result.HTML, "Hello &lt;world&gt;") {
		t.Fatalf("expected escaped HTML, got %q", result.HTML)
	}
	if result.Headers["x-loader"] != "document" {
		t.Fatalf("unexpected headers: %#v", result.Headers)
	}
}

func TestDocumentLoader_DOCXExtraction(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample.docx")
	createTestDOCX(t, path, "Hello docx world")
	result, err := NewDocumentLoader().Load(context.Background(), path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if !strings.Contains(result.HTML, "Hello docx world") {
		t.Fatalf("expected extracted docx text, got %q", result.HTML)
	}
}

func createTestDOCX(t *testing.T, path, text string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	w, err := zw.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><document><body><p><r><t>` + text + `</t></r></p></body></document>`
	if _, err := w.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
}
