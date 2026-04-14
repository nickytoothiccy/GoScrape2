// Package loaders provides UTLS-based HTTP fetching
package loaders

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"

	"stealthfetch/internal/models"
)

// BrowserProfile defines TLS fingerprint and headers
type BrowserProfile struct {
	Name      string
	ClientID  *utls.ClientHelloID
	UserAgent string
	Headers   [][2]string
}

var profiles = map[string]BrowserProfile{
	"chrome": {
		Name:      "chrome",
		ClientID:  &utls.HelloChrome_Auto,
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
		Headers: [][2]string{
			{"sec-ch-ua", `"Chromium";v="124", "Google Chrome";v="124", "Not-A.Brand";v="99"`},
			{"sec-ch-ua-mobile", "?0"},
			{"sec-ch-ua-platform", `"Windows"`},
			{"upgrade-insecure-requests", "1"},
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8"},
			{"sec-fetch-site", "none"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-user", "?1"},
			{"sec-fetch-dest", "document"},
			{"accept-encoding", "gzip, deflate, br"},
			{"accept-language", "en-US,en;q=0.9"},
		},
	},
	"firefox": {
		Name:      "firefox",
		ClientID:  &utls.HelloFirefox_Auto,
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:125.0) Gecko/20100101 Firefox/125.0",
		Headers: [][2]string{
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8"},
			{"accept-language", "en-US,en;q=0.5"},
			{"accept-encoding", "gzip, deflate, br"},
			{"upgrade-insecure-requests", "1"},
			{"sec-fetch-dest", "document"},
			{"sec-fetch-mode", "navigate"},
			{"sec-fetch-site", "none"},
			{"sec-fetch-user", "?1"},
		},
	},
}

var profileKeys = []string{"chrome", "firefox"}

// UTLSLoader fetches using UTLS for TLS fingerprinting
type UTLSLoader struct {
	profile  string
	proxyURL string
	timeout  time.Duration
}

// NewUTLSLoader creates a new UTLS HTTP loader
func NewUTLSLoader(profile, proxyURL string, timeout time.Duration) *UTLSLoader {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &UTLSLoader{
		profile:  profile,
		proxyURL: proxyURL,
		timeout:  timeout,
	}
}

// Name returns the loader identifier
func (l *UTLSLoader) Name() string {
	return "utls"
}

// Load fetches the URL using UTLS
func (l *UTLSLoader) Load(ctx context.Context, source string) (*models.FetchResult, error) {
	start := time.Now()

	profile := l.pickProfile()
	transport := l.buildTransport(profile)

	client := &http.Client{
		Transport: transport,
		Timeout:   l.timeout,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", source, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", profile.UserAgent)
	for _, h := range profile.Headers {
		req.Header.Set(h[0], h[1])
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	headers := make(map[string]string)
	for k := range resp.Header {
		headers[strings.ToLower(k)] = resp.Header.Get(k)
	}

	return &models.FetchResult{
		HTML:        string(body),
		URL:         source,
		StatusCode:  resp.StatusCode,
		Headers:     headers,
		ElapsedSecs: time.Since(start).Seconds(),
		Error:       nil,
	}, nil
}

func (l *UTLSLoader) pickProfile() BrowserProfile {
	if l.profile == "" || l.profile == "random" {
		return profiles[profileKeys[rand.Intn(len(profileKeys))]]
	}
	if p, ok := profiles[strings.ToLower(l.profile)]; ok {
		return p
	}
	return profiles["chrome"]
}

func (l *UTLSLoader) buildTransport(profile BrowserProfile) *http.Transport {
	t := &http.Transport{
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return utlsDialTLS(ctx, network, addr, profile.ClientID)
		},
		DialContext: (&net.Dialer{Timeout: 15 * time.Second}).DialContext,
	}

	if l.proxyURL != "" {
		if u, err := url.Parse(l.proxyURL); err == nil {
			t.Proxy = http.ProxyURL(u)
		}
	}

	http2.ConfigureTransport(t)
	return t
}

func utlsDialTLS(ctx context.Context, network, addr string, clientHello *utls.ClientHelloID) (net.Conn, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}

	dialer := &net.Dialer{Timeout: 15 * time.Second}
	rawConn, err := dialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	tlsConn := utls.UClient(rawConn, &utls.Config{
		ServerName:         host,
		InsecureSkipVerify: false,
	}, *clientHello)

	if err := tlsConn.HandshakeContext(ctx); err != nil {
		rawConn.Close()
		return nil, fmt.Errorf("tls handshake: %w", err)
	}

	return tlsConn, nil
}
