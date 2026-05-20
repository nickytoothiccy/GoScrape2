// Package loaders provides different fetch backends
package loaders

import (
	"context"

	"stealthfetch/internal/models"
)

// Loader is the interface for all fetch backends
type Loader interface {
	// Load fetches content from the given source
	Load(ctx context.Context, source string) (*models.FetchResult, error)

	// Name returns the loader's identifier
	Name() string
}

// ScreenshotLoader is the interface for screenshot-capable backends.
type ScreenshotLoader interface {
	// Capture renders and captures a PNG screenshot from the source URL.
	Capture(ctx context.Context, source string) (*models.ScreenshotResult, error)

	// Name returns the loader's identifier.
	Name() string
}
