// Package loaders provides Rod browser-based fetching with stealth
package loaders

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/stealth"

	"stealthfetch/internal/models"
)

// RodLoader fetches pages using a headless Chrome browser via Rod.
// Handles JavaScript-rendered pages and Cloudflare challenges.
type RodLoader struct {
	headless bool
	waitSecs int
	timeout  time.Duration
	verbose  bool
}

// RodConfig holds configuration for creating a RodLoader
type RodConfig struct {
	Headless bool          // default true
	WaitSecs int           // extra wait after page load, default 3
	Timeout  time.Duration // per-page timeout, default 30s
	Verbose  bool
}

// NewRodLoader creates a new Rod browser loader
func NewRodLoader(cfg RodConfig) *RodLoader {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.WaitSecs == 0 {
		cfg.WaitSecs = 3
	}
	return &RodLoader{
		headless: cfg.Headless,
		waitSecs: cfg.WaitSecs,
		timeout:  cfg.Timeout,
		verbose:  cfg.Verbose,
	}
}

// NewDefaultRodLoader creates a RodLoader with sensible defaults
func NewDefaultRodLoader(verbose bool) *RodLoader {
	return &RodLoader{
		headless: true,
		waitSecs: 3,
		timeout:  30 * time.Second,
		verbose:  verbose,
	}
}

// Name returns the loader identifier
func (l *RodLoader) Name() string {
	return "rod"
}

// Load fetches the URL using a headless Chrome browser with stealth patches
func (l *RodLoader) Load(ctx context.Context, source string) (*models.FetchResult, error) {
	start := time.Now()

	// Find Chrome
	path, exists := launcher.LookPath()
	if !exists {
		return nil, fmt.Errorf("Chrome/Chromium not found on system")
	}

	// Launch browser with stealth-friendly settings
	u := launcher.New().
		Bin(path).
		Leakless(false). // avoid Windows Defender blocking leakless.exe
		Headless(l.headless).
		Set("disable-blink-features", "AutomationControlled").
		Set("no-first-run").
		Set("no-default-browser-check").
		Set("window-size", "1366,768").
		MustLaunch()

	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	browser = browser.Timeout(l.timeout)

	// Create stealth page
	page := stealth.MustPage(browser)
	defer page.MustClose()

	// Navigate
	if err := page.Navigate(source); err != nil {
		return l.errorResult(source, start, fmt.Errorf("navigation: %w", err)), nil
	}

	// Wait for load
	if err := page.WaitLoad(); err != nil {
		return l.errorResult(source, start, fmt.Errorf("wait load: %w", err)), nil
	}

	// Extra wait for dynamic content
	time.Sleep(time.Duration(l.waitSecs) * time.Second)

	// Check for Cloudflare challenge and wait if needed
	l.handleCloudflare(page, source)

	// Get final HTML
	html, err := page.HTML()
	if err != nil {
		return nil, fmt.Errorf("get HTML: %w", err)
	}

	return &models.FetchResult{
		HTML:        html,
		URL:         source,
		StatusCode:  200,
		Headers:     map[string]string{"x-loader": "rod"},
		ElapsedSecs: time.Since(start).Seconds(),
	}, nil
}

// handleCloudflare detects and waits for CF challenges to resolve
func (l *RodLoader) handleCloudflare(page *rod.Page, url string) {
	title, err := page.Eval(`() => document.title`)
	if err != nil {
		return
	}
	titleStr := title.Value.String()

	if titleStr != "Just a moment..." && titleStr != "Just a moment" {
		return
	}

	if l.verbose {
		log.Printf("[rod] Cloudflare challenge detected for %s, waiting...", url)
	}

	deadline := time.Now().Add(l.timeout)
	for time.Now().Before(deadline) {
		time.Sleep(2 * time.Second)
		title, err = page.Eval(`() => document.title`)
		if err != nil {
			return
		}
		titleStr = title.Value.String()
		if titleStr != "Just a moment..." && titleStr != "Just a moment" {
			if l.verbose {
				log.Printf("[rod] Challenge resolved! Title: %s", titleStr)
			}
			time.Sleep(2 * time.Second) // extra wait for JS rendering
			return
		}
	}

	if l.verbose {
		log.Printf("[rod] Cloudflare challenge did not resolve within timeout for %s", url)
	}
}

// errorResult creates a FetchResult representing a non-fatal fetch error.
// Returns a result with error info rather than failing hard, so the graph
// can decide whether to retry or skip.
func (l *RodLoader) errorResult(url string, start time.Time, err error) *models.FetchResult {
	return &models.FetchResult{
		HTML:        "",
		URL:         url,
		StatusCode:  0,
		ElapsedSecs: time.Since(start).Seconds(),
		Error:       err,
	}
}
