package loaders

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"

	"stealthfetch/internal/models"
)

// Capture renders and captures a PNG screenshot using Rod.
func (l *RodLoader) Capture(ctx context.Context, source string) (*models.ScreenshotResult, error) {
	start := time.Now()
	path, exists := launcher.LookPath()
	if !exists {
		return nil, fmt.Errorf("Chrome/Chromium not found on system")
	}

	u := launcher.New().
		Bin(path).
		Leakless(false).
		Headless(l.headless).
		Set("no-sandbox").
		Set("disable-dev-shm-usage").
		Set("disable-blink-features", "AutomationControlled").
		Set("no-first-run").
		Set("no-default-browser-check").
		Set("window-size", "1366,768").
		MustLaunch()

	browser := rod.New().ControlURL(u).MustConnect()
	defer func() { _ = browser.Close() }()
	runCtx, cancel := context.WithTimeout(ctx, l.timeout)
	defer cancel()
	browser = browser.Context(runCtx)

	page := stealth.MustPage(browser)
	defer func() { _ = page.Close() }()
	if err := page.Navigate(source); err != nil {
		return nil, fmt.Errorf("navigation: %w", err)
	}
	if err := page.WaitLoad(); err != nil {
		return nil, fmt.Errorf("wait load: %w", err)
	}
	if err := sleepWithContext(ctx, time.Duration(l.waitSecs)*time.Second); err != nil {
		return nil, err
	}

	l.handleCloudflare(page, source)
	png, err := page.Screenshot(true, &proto.PageCaptureScreenshot{
		Format: proto.PageCaptureScreenshotFormatPng,
	})
	if err != nil {
		return nil, fmt.Errorf("capture screenshot: %w", err)
	}

	return &models.ScreenshotResult{
		PNG:         png,
		URL:         source,
		ElapsedSecs: time.Since(start).Seconds(),
	}, nil
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
