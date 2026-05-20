package nodes

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"stealthfetch/internal/models"
	"stealthfetch/pkg/graph"
)

type stubScreenshotLoader struct {
	result *models.ScreenshotResult
	err    error
	calls  int
}

func (l *stubScreenshotLoader) Name() string { return "stub_screenshot" }

func (l *stubScreenshotLoader) Capture(context.Context, string) (*models.ScreenshotResult, error) {
	l.calls++
	return l.result, l.err
}

func TestFetchScreenNodeExecute(t *testing.T) {
	loader := &stubScreenshotLoader{
		result: &models.ScreenshotResult{
			PNG: []byte{0x89, 0x50, 0x4E, 0x47},
			URL: "https://example.com",
		},
	}
	node := NewFetchScreenNode(loader, models.DefaultConfig())
	state := graph.NewState()
	state.Set("url", "https://example.com")
	if err := node.Execute(context.Background(), state); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if loader.calls != 1 {
		t.Fatalf("expected one loader call, got %d", loader.calls)
	}
	b64, ok := state.GetString("screenshot_base64")
	if !ok {
		t.Fatal("expected screenshot_base64 in state")
	}
	if b64 != base64.StdEncoding.EncodeToString(loader.result.PNG) {
		t.Fatalf("unexpected base64 screenshot: %s", b64)
	}
}

func TestFetchScreenNodeMissingURL(t *testing.T) {
	node := NewFetchScreenNode(&stubScreenshotLoader{}, models.DefaultConfig())
	err := node.Execute(context.Background(), graph.NewState())
	if err == nil {
		t.Fatal("expected error for missing url")
	}
}

func TestFetchScreenNodeRejectsHTMLInput(t *testing.T) {
	node := NewFetchScreenNode(&stubScreenshotLoader{}, models.DefaultConfig())
	state := graph.NewState()
	state.Set("url", "<html><body>hi</body></html>")
	err := node.Execute(context.Background(), state)
	if err == nil {
		t.Fatal("expected html-input error")
	}
}

func TestFetchScreenNodeLoaderError(t *testing.T) {
	node := NewFetchScreenNode(&stubScreenshotLoader{err: errors.New("boom")}, models.DefaultConfig())
	state := graph.NewState()
	state.Set("url", "https://example.com")
	err := node.Execute(context.Background(), state)
	if err == nil {
		t.Fatal("expected loader error")
	}
}
