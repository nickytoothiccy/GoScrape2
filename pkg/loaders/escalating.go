package loaders

import (
	"context"

	"stealthfetch/internal/models"
)

// EscalatingLoader tries a primary loader first and falls back when blocked.
type EscalatingLoader struct {
	primary  Loader
	fallback Loader
	detector func(*models.FetchResult) bool
}

// NewEscalatingLoader creates a loader that falls back when the primary result looks blocked.
func NewEscalatingLoader(primary, fallback Loader) *EscalatingLoader {
	return &EscalatingLoader{primary: primary, fallback: fallback, detector: IsLikelyBlocked}
}

func (l *EscalatingLoader) Name() string { return "auto" }

func (l *EscalatingLoader) Load(ctx context.Context, source string) (*models.FetchResult, error) {
	if l.primary == nil {
		return l.fallback.Load(ctx, source)
	}
	result, err := l.primary.Load(ctx, source)
	if l.fallback == nil {
		return result, err
	}
	if err != nil || l.detector(result) {
		return l.fallback.Load(ctx, source)
	}
	return result, nil
}
