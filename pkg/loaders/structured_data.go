package loaders

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"stealthfetch/internal/models"
)

// StructuredDataLoader loads local JSON, XML, or CSV files and directories.
type StructuredDataLoader struct{ ext string }

// NewJSONLoader creates a loader for JSON files/directories.
func NewJSONLoader() *StructuredDataLoader { return &StructuredDataLoader{ext: ".json"} }

// NewXMLLoader creates a loader for XML files/directories.
func NewXMLLoader() *StructuredDataLoader { return &StructuredDataLoader{ext: ".xml"} }

// NewCSVLoader creates a loader for CSV files/directories.
func NewCSVLoader() *StructuredDataLoader { return &StructuredDataLoader{ext: ".csv"} }

// Name returns the loader identifier.
func (l *StructuredDataLoader) Name() string { return strings.TrimPrefix(l.ext, ".") }

// Load reads a single structured-data file or all matching files from a directory.
func (l *StructuredDataLoader) Load(_ context.Context, source string) (*models.FetchResult, error) {
	start := time.Now()
	info, err := os.Stat(source)
	if err != nil {
		return nil, fmt.Errorf("%s load: %w", l.Name(), err)
	}
	var text string
	if info.IsDir() {
		text, err = l.loadDir(source)
	} else {
		text, err = l.loadFile(source)
	}
	if err != nil {
		return nil, err
	}
	html := "<html><body><pre>" + escapeHTML(text) + "</pre></body></html>"
	headers := map[string]string{"x-loader": l.Name(), "x-structured-ext": l.ext}
	return &models.FetchResult{HTML: html, URL: source, StatusCode: 200, Headers: headers, ElapsedSecs: time.Since(start).Seconds()}, nil
}

func (l *StructuredDataLoader) loadDir(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("%s load: %w", l.Name(), err)
	}
	var paths []string
	for _, entry := range entries {
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != l.ext {
			continue
		}
		paths = append(paths, filepath.Join(dir, entry.Name()))
	}
	if len(paths) == 0 {
		return "", fmt.Errorf("%s load: no %s files found in directory", l.Name(), l.ext)
	}
	sort.Strings(paths)
	parts := make([]string, 0, len(paths))
	for _, path := range paths {
		text, err := l.loadFile(path)
		if err != nil {
			return "", err
		}
		parts = append(parts, "FILE: "+filepath.Base(path)+"\n"+text)
	}
	return strings.Join(parts, "\n\n"), nil
}

func (l *StructuredDataLoader) loadFile(path string) (string, error) {
	if strings.ToLower(filepath.Ext(path)) != l.ext {
		return "", fmt.Errorf("%s load: unsupported extension %q", l.Name(), filepath.Ext(path))
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("%s load: %w", l.Name(), err)
	}
	text := strings.TrimSpace(string(data))
	if text == "" {
		return "", fmt.Errorf("%s load: empty file", l.Name())
	}
	return text, nil
}
