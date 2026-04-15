package loaders

import (
	"context"
	"testing"

	"stealthfetch/internal/models"
)

type stubLoader struct {
	name   string
	result *models.FetchResult
	err    error
	calls  int
}

func (l *stubLoader) Name() string { return l.name }

func (l *stubLoader) Load(context.Context, string) (*models.FetchResult, error) {
	l.calls++
	return l.result, l.err
}

func TestIsLikelyBlocked(t *testing.T) {
	tests := []struct {
		name   string
		result *models.FetchResult
		want   bool
	}{
		{name: "status block", result: &models.FetchResult{StatusCode: 403}, want: true},
		{name: "marker block", result: &models.FetchResult{StatusCode: 200, HTML: "<title>Just a moment...</title>"}, want: true},
		{name: "normal page", result: &models.FetchResult{StatusCode: 200, HTML: "<html><body>ok</body></html>"}, want: false},
	}
	for _, tt := range tests {
		if got := IsLikelyBlocked(tt.result); got != tt.want {
			t.Fatalf("%s: got %v want %v", tt.name, got, tt.want)
		}
	}
}

func TestEscalatingLoaderFallsBackOnBlockedResult(t *testing.T) {
	primary := &stubLoader{name: "utls", result: &models.FetchResult{StatusCode: 403, HTML: "blocked"}}
	fallback := &stubLoader{name: "rod", result: &models.FetchResult{StatusCode: 200, HTML: "ok"}}
	loader := NewEscalatingLoader(primary, fallback)
	result, err := loader.Load(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if result == nil || result.HTML != "ok" {
		t.Fatalf("expected fallback result, got %#v", result)
	}
	if primary.calls != 1 || fallback.calls != 1 {
		t.Fatalf("expected one primary and one fallback call, got %d/%d", primary.calls, fallback.calls)
	}
}

func TestEscalatingLoaderKeepsPrimaryWhenHealthy(t *testing.T) {
	primary := &stubLoader{name: "utls", result: &models.FetchResult{StatusCode: 200, HTML: "ok"}}
	fallback := &stubLoader{name: "rod", result: &models.FetchResult{StatusCode: 200, HTML: "fallback"}}
	loader := NewEscalatingLoader(primary, fallback)
	result, err := loader.Load(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if result == nil || result.HTML != "ok" {
		t.Fatalf("expected primary result, got %#v", result)
	}
	if primary.calls != 1 || fallback.calls != 0 {
		t.Fatalf("expected one primary and zero fallback calls, got %d/%d", primary.calls, fallback.calls)
	}
}
