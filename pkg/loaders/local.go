// Package loaders provides local HTML loading
package loaders

import (
	"context"
	"fmt"
	"strings"

	"stealthfetch/internal/models"
)

// LocalLoader handles pre-fetched HTML content
type LocalLoader struct{}

// NewLocalLoader creates a new local HTML loader
func NewLocalLoader() *LocalLoader {
	return &LocalLoader{}
}

// Name returns the loader identifier
func (l *LocalLoader) Name() string {
	return "local"
}

// Load treats the source as raw HTML content
func (l *LocalLoader) Load(ctx context.Context, source string) (*models.FetchResult, error) {
	// Validate it looks like HTML
	if !strings.HasPrefix(strings.TrimSpace(source), "<") {
		return nil, fmt.Errorf("source does not appear to be HTML")
	}

	return &models.FetchResult{
		HTML:        source,
		URL:         "local",
		StatusCode:  200,
		Headers:     make(map[string]string),
		ElapsedSecs: 0,
		Error:       nil,
	}, nil
}
