package scrapegraph

import "testing"

func TestNormalizeURL(t *testing.T) {
	got := normalizeURL("HTTPS://Example.com:443/docs/start/#intro")
	if got != "https://example.com/docs/start" {
		t.Fatalf("unexpected normalized url: %s", got)
	}
}

func TestDepthSearchShouldVisitURL(t *testing.T) {
	d := &DepthSearchGraph{
		source:         "https://example.com/docs",
		restrictHost:   true,
		allowedDomains: []string{"example.com"},
		pathPrefixes:   []string{"/docs"},
		excludePatterns: []string{"changelog"},
	}
	if !d.shouldVisitURL("https://example.com/docs/getting-started") {
		t.Fatal("expected docs page to be allowed")
	}
	if d.shouldVisitURL("https://other.com/docs/getting-started") {
		t.Fatal("expected foreign host to be blocked")
	}
	if d.shouldVisitURL("https://example.com/blog/post") {
		t.Fatal("expected non-matching path prefix to be blocked")
	}
	if d.shouldVisitURL("https://example.com/docs/changelog") {
		t.Fatal("expected exclude pattern to be blocked")
	}
}