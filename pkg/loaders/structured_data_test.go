package loaders

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStructuredDataLoader_LoadFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample.json")
	if err := os.WriteFile(path, []byte(`{"name":"chioggia"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := NewJSONLoader().Load(context.Background(), path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if result.Headers["x-loader"] != "json" || !strings.Contains(result.HTML, "chioggia") {
		t.Fatalf("unexpected result: %#v html=%q", result.Headers, result.HTML)
	}
	if result.Headers["x-structured-ext"] != ".json" {
		t.Fatalf("unexpected ext header: %#v", result.Headers)
	}
}

func TestStructuredDataLoader_LoadDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "b.csv"), []byte("name\nsecond"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.csv"), []byte("name\nfirst"), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := NewCSVLoader().Load(context.Background(), dir)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	html := result.HTML
	if !strings.Contains(html, "FILE: a.csv") || !strings.Contains(html, "FILE: b.csv") {
		t.Fatalf("expected both files in output, got %q", html)
	}
	if strings.Index(html, "FILE: a.csv") > strings.Index(html, "FILE: b.csv") {
		t.Fatalf("expected sorted directory output, got %q", html)
	}
}

func TestStructuredDataLoader_UnsupportedExtension(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := NewXMLLoader().Load(context.Background(), path)
	if err == nil || !strings.Contains(err.Error(), "unsupported extension") {
		t.Fatalf("expected unsupported extension error, got %v", err)
	}
}
