package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/html"
	"golang.org/x/net/http2"
)

// ==================== MODELS ====================

type FetchRequest struct {
	URL       string            `json:"url"`
	Method    string            `json:"method,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	Body      string            `json:"body,omitempty"`
	Profile   string            `json:"profile,omitempty"`
	ProxyURL  string            `json:"proxy_url,omitempty"`
	TimeoutMs int               `json:"timeout_ms,omitempty"`
}

type FetchResponse struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
	Profile string            `json:"profile_used"`
	Elapsed float64           `json:"elapsed_s"`
}

type ExtractRequest struct {
	HTML       string `json:"html"`
	Prompt     string `json:"prompt"`
	Model      string `json:"model,omitempty"`
	SchemaHint string `json:"schema_hint,omitempty"`
	Structured bool   `json:"structured,omitempty"`
}

type ExtractResponse struct {
	Result  json.RawMessage `json:"result"`
	Model   string          `json:"model_used"`
	Elapsed float64         `json:"elapsed_s"`
}

type ScrapeRequest struct {
	URL        string            `json:"url"`
	Prompt     string            `json:"prompt"`
	Model      string            `json:"model,omitempty"`
	SchemaHint string            `json:"schema_hint,omitempty"`
	Structured bool              `json:"structured,omitempty"`
	Profile    string            `json:"profile,omitempty"`
	ProxyURL   string            `json:"proxy_url,omitempty"`
	TimeoutMs  int               `json:"timeout_ms,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
}

type ScrapeResponse struct {
	Result      json.RawMessage `json:"result"`
	FetchStatus int             `json:"fetch_status"`
	Profile     string          `json:"profile_used"`
	Model       string          `json:"model_used"`
	FetchTime   float64         `json:"fetch_time_s"`
	ExtractTime float64         `json:"extract_time_s"`
	TotalTime   float64         `json:"total_time_s"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// ==================== TLS PROFILES ====================

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
	"safari": {
		Name:      "safari",
		ClientID:  &utls.HelloSafari_Auto,
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_4_1) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4.1 Safari/605.1.15",
		Headers: [][2]string{
			{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"},
			{"accept-language", "en-US,en;q=0.9"},
			{"accept-encoding", "gzip, deflate, br"},
		},
	},
}

var profileKeys = []string{"chrome", "firefox", "safari"}

func pickProfile(name string) BrowserProfile {
	if name == "" || name == "random" {
		return profiles[profileKeys[rand.Intn(len(profileKeys))]]
	}
	if p, ok := profiles[strings.ToLower(name)]; ok {
		return p
	}
	return profiles["chrome"]
}

// ==================== STEALTH TRANSPORT ====================

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

func buildTransport(profile BrowserProfile, proxyURL string) *http.Transport {
	t := &http.Transport{
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return utlsDialTLS(ctx, network, addr, profile.ClientID)
		},
		DialContext: (&net.Dialer{Timeout: 15 * time.Second}).DialContext,
	}

	if proxyURL != "" {
		if u, err := url.Parse(proxyURL); err == nil {
			t.Proxy = http.ProxyURL(u)
		}
	}

	if err := http2.ConfigureTransport(t); err != nil {
		log.Printf("warn: http2 configure failed: %v", err)
	}

	return t
}

// ==================== FETCH ====================

func doFetch(req FetchRequest) (*FetchResponse, error) {
	profile := pickProfile(req.Profile)

	method := req.Method
	if method == "" {
		method = "GET"
	}
	timeout := time.Duration(req.TimeoutMs) * time.Millisecond
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	transport := buildTransport(profile, req.ProxyURL)
	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	var bodyReader io.Reader
	if req.Body != "" {
		bodyReader = strings.NewReader(req.Body)
	}

	httpReq, err := http.NewRequest(method, req.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("User-Agent", profile.UserAgent)
	for _, h := range profile.Headers {
		httpReq.Header.Set(h[0], h[1])
	}
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	start := time.Now()
	resp, err := client.Do(httpReq)
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

	return &FetchResponse{
		Status:  resp.StatusCode,
		Headers: headers,
		Body:    string(body),
		Profile: profile.Name,
		Elapsed: time.Since(start).Seconds(),
	}, nil
}

// ==================== HTML CLEANING ====================

var stripTags = map[string]bool{
	"script": true, "style": true, "noscript": true, "svg": true,
	"iframe": true, "link": true, "meta": true, "head": true,
}

func cleanHTML(rawHTML string, maxChars int) string {
	if maxChars == 0 {
		maxChars = 50000
	}

	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return rawHTML
	}

	var sb strings.Builder
	var extract func(*html.Node)
	extract = func(n *html.Node) {
		if n.Type == html.ElementNode && stripTags[n.Data] {
			return
		}
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				sb.WriteString(text)
				sb.WriteString("\n")
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}
	extract(doc)

	result := sb.String()
	for strings.Contains(result, "\n\n\n") {
		result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
	}
	if len(result) > maxChars {
		result = result[:maxChars] + "\n\n[TRUNCATED]"
	}

	return result
}

// ==================== LLM EXTRACTION (openai-go) ====================

const systemPrompt = `You are a precise data extraction engine. Given content from a webpage and a user's extraction prompt, extract the requested information and return it as valid JSON.

Rules:
- Return ONLY valid JSON, no markdown fences, no explanation
- If the requested data is not found, return {"error": "not_found", "reason": "..."}
- Use arrays for lists of items
- Use descriptive key names
- Preserve URLs as absolute paths when possible`

func doExtract(ctx context.Context, oaiClient openai.Client, req ExtractRequest) (*ExtractResponse, error) {
	model := req.Model
	if model == "" {
		model = "gpt-4o"
	}

	var content string
	if req.Structured {
		content = req.HTML
	} else {
		content = cleanHTML(req.HTML, 50000)
	}

	userMsg := fmt.Sprintf("## Extraction Prompt\n%s\n\n## Page Content\n%s", req.Prompt, content)
	if req.SchemaHint != "" {
		userMsg += fmt.Sprintf("\n\n## Expected Schema\n%s", req.SchemaHint)
	}

	start := time.Now()

	chat, err := oaiClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userMsg),
		},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &openai.ResponseFormatJSONObjectParam{},
		},
		Model:       model,
		Temperature: openai.Float(0),
	})
	if err != nil {
		return nil, fmt.Errorf("openai: %w", err)
	}

	rawContent := chat.Choices[0].Message.Content

	return &ExtractResponse{
		Result:  json.RawMessage(rawContent),
		Model:   model,
		Elapsed: time.Since(start).Seconds(),
	}, nil
}

// ==================== HTTP HANDLERS ====================

type Server struct {
	oaiClient openai.Client
}

func (s *Server) handleFetch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "POST only")
		return
	}

	var req FetchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.URL == "" {
		writeError(w, http.StatusBadRequest, "url is required")
		return
	}

	resp, err := doFetch(req)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleExtract(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "POST only")
		return
	}

	var req ExtractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.HTML == "" || req.Prompt == "" {
		writeError(w, http.StatusBadRequest, "html and prompt are required")
		return
	}

	resp, err := doExtract(r.Context(), s.oaiClient, req)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleScrape(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "POST only")
		return
	}

	var req ScrapeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.URL == "" || req.Prompt == "" {
		writeError(w, http.StatusBadRequest, "url and prompt are required")
		return
	}

	totalStart := time.Now()

	fetchResp, err := doFetch(FetchRequest{
		URL:       req.URL,
		Profile:   req.Profile,
		ProxyURL:  req.ProxyURL,
		TimeoutMs: req.TimeoutMs,
		Headers:   req.Headers,
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, fmt.Sprintf("fetch failed: %s", err))
		return
	}
	if fetchResp.Status >= 400 {
		writeError(w, http.StatusBadGateway, fmt.Sprintf("HTTP %d from target", fetchResp.Status))
		return
	}

	extractResp, err := doExtract(r.Context(), s.oaiClient, ExtractRequest{
		HTML:       fetchResp.Body,
		Prompt:     req.Prompt,
		Model:      req.Model,
		SchemaHint: req.SchemaHint,
		Structured: req.Structured,
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, fmt.Sprintf("extract failed: %s", err))
		return
	}

	writeJSON(w, http.StatusOK, ScrapeResponse{
		Result:      extractResp.Result,
		FetchStatus: fetchResp.Status,
		Profile:     fetchResp.Profile,
		Model:       extractResp.Model,
		FetchTime:   fetchResp.Elapsed,
		ExtractTime: extractResp.Elapsed,
		TotalTime:   time.Since(totalStart).Seconds(),
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"version": "0.2.0",
	})
}

// ==================== HELPERS ====================

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{Error: msg})
}

// ==================== MAIN ====================

func main() {
	rand.Seed(time.Now().UnixNano())

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY env var required")
	}

	oaiClient := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	srv := &Server{oaiClient: oaiClient}

	mux := http.NewServeMux()
	mux.HandleFunc("/fetch", srv.handleFetch)
	mux.HandleFunc("/extract", srv.handleExtract)
	mux.HandleFunc("/scrape", srv.handleScrape)
	mux.HandleFunc("/health", srv.handleHealth)

	addr := ":8899"
	log.Printf("stealthgraph listening on %s", addr)
	log.Printf("endpoints: POST /fetch | POST /extract | POST /scrape | GET /health")

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 180 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
