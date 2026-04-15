package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"stealthfetch/internal/hermesenv"
	"stealthfetch/internal/models"
	"stealthfetch/pkg/loaders"
	"stealthfetch/pkg/scrapegraph"
)

type cliError struct {
	Error string `json:"error"`
}

type fetchOutput struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
	Elapsed float64           `json:"elapsed_s"`
}

func main() {
	_ = hermesenv.LoadEnv()

	if len(os.Args) < 2 {
		failf("usage: hermes-skill <command> [args]")
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "fetch":
		must(runFetch(args))
	case "search":
		must(runSearch(args))
	case "scrape":
		must(runScrape(args))
	case "multi-scrape":
		must(runMultiScrape(args))
	case "depth-search":
		must(runDepthSearch(args))
	case "document-scrape":
		must(runDocumentScrape(args))
	case "research":
		must(runResearch(args))
	default:
		failf("unknown command: %s", command)
	}
}

func runFetch(args []string) error {
	fs := flag.NewFlagSet("fetch", flag.ContinueOnError)
	url := fs.String("url", "", "Target URL")
	profile := fs.String("profile", "chrome", "Browser fingerprint profile")
	timeoutMS := fs.Int("timeout-ms", 30000, "Timeout in milliseconds")
	headless := fs.Bool("headless", false, "Use Rod instead of UTLS")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*url) == "" {
		return errors.New("fetch requires --url")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeoutMS)*time.Millisecond)
	defer cancel()

	var loader loaders.Loader
	if *headless {
		loader = loaders.NewDefaultRodLoader(false)
	} else {
		loader = loaders.NewUTLSLoader(*profile, "", time.Duration(*timeoutMS)*time.Millisecond)
	}

	result, err := loader.Load(ctx, *url)
	if err != nil {
		return err
	}
	return writeJSON(fetchOutput{
		Status:  result.StatusCode,
		Headers: result.Headers,
		Body:    result.HTML,
		Elapsed: result.ElapsedSecs,
	})
}

func runSearch(args []string) error {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	prompt := fs.String("prompt", "", "Search prompt")
	model := fs.String("model", "", "Override extraction model")
	schema := fs.String("schema", "", "Optional schema hint")
	maxResults := fs.Int("max-results", 5, "Maximum search results")
	perURLTimeoutMS := fs.Int("per-url-timeout-ms", 60000, "Per URL timeout in milliseconds")
	headless := fs.Bool("headless", false, "Use headless browser for page fetches")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*prompt) == "" {
		return errors.New("search requires --prompt")
	}

	cfg := baseConfig(*model, *headless)
	graph := scrapegraph.NewSearchGraph(scrapegraph.SearchGraphConfig{
		Config:        cfg,
		Prompt:        *prompt,
		SchemaHint:    *schema,
		MaxResults:    *maxResults,
		PerURLTimeout: time.Duration(*perURLTimeoutMS) * time.Millisecond,
	})
	result, err := graph.Run(context.Background())
	if err != nil {
		return err
	}

	return writeJSON(map[string]any{
		"result":       json.RawMessage(result),
		"urls_scraped": graph.GetConsideredURLs(),
		"model_used":   cfg.LLMModel,
	})
}

func runScrape(args []string) error {
	fs := flag.NewFlagSet("scrape", flag.ContinueOnError)
	url := fs.String("url", "", "Target URL")
	prompt := fs.String("prompt", "", "Extraction prompt")
	model := fs.String("model", "", "Override extraction model")
	schema := fs.String("schema", "", "Optional schema hint")
	headless := fs.Bool("headless", false, "Use headless browser")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*url) == "" || strings.TrimSpace(*prompt) == "" {
		return errors.New("scrape requires --url and --prompt")
	}

	cfg := baseConfig(*model, *headless)
	graph := scrapegraph.NewSmartScraperGraph(*prompt, *url, cfg, *schema)
	result, err := graph.Run(context.Background())
	if err != nil {
		return err
	}
	return writeJSON(map[string]any{
		"result":     json.RawMessage(result),
		"source_url": *url,
		"model_used": cfg.LLMModel,
	})
}

func runMultiScrape(args []string) error {
	fs := flag.NewFlagSet("multi-scrape", flag.ContinueOnError)
	urls := fs.String("urls", "", "Comma-separated URLs")
	prompt := fs.String("prompt", "", "Extraction prompt")
	model := fs.String("model", "", "Override extraction model")
	schema := fs.String("schema", "", "Optional schema hint")
	headless := fs.Bool("headless", false, "Use headless browser")
	concatResults := fs.Bool("concat-results", false, "Concatenate results into one JSON array")
	perURLTimeoutMS := fs.Int("per-url-timeout-ms", 60000, "Per URL timeout in milliseconds")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}

	urlList := splitCSV(*urls)
	if len(urlList) == 0 || strings.TrimSpace(*prompt) == "" {
		return errors.New("multi-scrape requires --urls and --prompt")
	}

	cfg := baseConfig(*model, *headless)
	graph := scrapegraph.NewSmartScraperMultiGraph(scrapegraph.SmartScraperMultiConfig{
		Config:        cfg,
		Prompt:        *prompt,
		SchemaHint:    *schema,
		URLs:          urlList,
		ConcatResults: *concatResults,
		PerURLTimeout: time.Duration(*perURLTimeoutMS) * time.Millisecond,
	})
	result, err := graph.Run(context.Background())
	if err != nil {
		return err
	}
	return writeJSON(map[string]any{
		"result":      json.RawMessage(result),
		"urls":        urlList,
		"failed_urls": graph.GetFailedURLs(),
		"model_used":  cfg.LLMModel,
	})
}

func runDepthSearch(args []string) error {
	fs := flag.NewFlagSet("depth-search", flag.ContinueOnError)
	url := fs.String("url", "", "Seed URL")
	prompt := fs.String("prompt", "", "Extraction prompt")
	model := fs.String("model", "", "Override extraction model")
	schema := fs.String("schema", "", "Optional schema hint")
	headless := fs.Bool("headless", false, "Use headless browser")
	maxDepth := fs.Int("max-depth", 2, "Maximum crawl depth")
	maxLinks := fs.Int("max-links-per-page", 3, "Maximum followed links per page")
	maxPages := fs.Int("max-pages", 0, "Hard cap on visited pages")
	perURLTimeoutMS := fs.Int("per-url-timeout-ms", 60000, "Per URL timeout in milliseconds")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*url) == "" || strings.TrimSpace(*prompt) == "" {
		return errors.New("depth-search requires --url and --prompt")
	}

	cfg := baseConfig(*model, *headless)
	graph := scrapegraph.NewDepthSearchGraph(scrapegraph.DepthSearchConfig{
		Config:          cfg,
		Prompt:          *prompt,
		SchemaHint:      *schema,
		Source:          *url,
		MaxDepth:        *maxDepth,
		MaxLinksPerPage: *maxLinks,
		MaxPages:        *maxPages,
		PerURLTimeout:   time.Duration(*perURLTimeoutMS) * time.Millisecond,
	})
	result, err := graph.Run(context.Background())
	if err != nil {
		return err
	}
	return writeJSON(map[string]any{
		"result":       json.RawMessage(result),
		"visited_urls": graph.GetVisitedURLs(),
		"model_used":   cfg.LLMModel,
	})
}

func runDocumentScrape(args []string) error {
	fs := flag.NewFlagSet("document-scrape", flag.ContinueOnError)
	path := fs.String("path", "", "Document path")
	prompt := fs.String("prompt", "", "Extraction prompt")
	model := fs.String("model", "", "Override extraction model")
	schema := fs.String("schema", "", "Optional schema hint")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*path) == "" || strings.TrimSpace(*prompt) == "" {
		return errors.New("document-scrape requires --path and --prompt")
	}

	cfg := baseConfig(*model, false)
	graph := scrapegraph.NewDocumentScraperGraph(*prompt, *path, cfg, *schema)
	result, err := graph.Run(context.Background())
	if err != nil {
		return err
	}
	return writeJSON(map[string]any{
		"result":     json.RawMessage(result),
		"path":       *path,
		"model_used": cfg.LLMModel,
	})
}

func runResearch(args []string) error {
	fs := flag.NewFlagSet("research", flag.ContinueOnError)
	prompt := fs.String("prompt", "", "Research prompt")
	seedURL := fs.String("seed-url", "", "Optional seed URL")
	model := fs.String("model", "", "Override extraction model")
	schema := fs.String("schema", "", "Optional schema hint")
	headless := fs.Bool("headless", false, "Use headless browser")
	searchFirst := fs.Bool("search-first", false, "Search before scraping")
	followSubpages := fs.Bool("follow-subpages", false, "Recursively follow discovered links")
	maxDepth := fs.Int("max-depth", 2, "Maximum crawl depth")
	maxPages := fs.Int("max-pages", 0, "Hard cap on visited pages")
	maxResults := fs.Int("max-results", 5, "Maximum search results")
	maxLinks := fs.Int("max-links-per-page", 3, "Maximum followed links per page")
	perURLTimeoutMS := fs.Int("per-url-timeout-ms", 60000, "Per URL timeout in milliseconds")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*prompt) == "" {
		return errors.New("research requires --prompt")
	}

	cfg := baseConfig(*model, *headless)
	graph := scrapegraph.NewResearchGraph(cfg)
	result, err := graph.Run(context.Background(), scrapegraph.ResearchRequest{
		Prompt:          *prompt,
		SeedURL:         strings.TrimSpace(*seedURL),
		SchemaHint:      *schema,
		SearchFirst:     *searchFirst,
		FollowSubpages:  *followSubpages,
		MaxDepth:        *maxDepth,
		MaxPages:        *maxPages,
		MaxResults:      *maxResults,
		MaxLinksPerPage: *maxLinks,
		PerURLTimeout:   time.Duration(*perURLTimeoutMS) * time.Millisecond,
	})
	if err != nil {
		return err
	}
	return writeJSON(result)
}

func baseConfig(model string, headless bool) *models.Config {
	cfg := hermesenv.DefaultConfig()
	if strings.TrimSpace(model) != "" {
		cfg.LLMModel = strings.TrimSpace(model)
	}
	cfg.Headless = headless
	return cfg
}

func splitCSV(values string) []string {
	parts := strings.Split(values, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func writeJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func must(err error) {
	if err == nil {
		return
	}
	failf("%v", err)
}

func failf(format string, args ...any) {
	_ = writeJSON(cliError{Error: fmt.Sprintf(format, args...)})
	os.Exit(1)
}
