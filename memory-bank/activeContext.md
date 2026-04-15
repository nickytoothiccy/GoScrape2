# Active Context

## Current State
GoScrape2 has completed another meaningful implementation slice and is now approximately **~43% complete** compared to full ScrapeGraphAI.

## What Was Built This Session

### 30. SmartScraperLiteGraph Parity Slice (`pkg/scrapegraph/smartscraper_lite.go`)
- Added `SmartScraperLiteGraph` as the canonical lightweight single-page extraction graph
- Reused the standard fetch → parse → extract workflow while applying lite-oriented defaults for HTML size and chunk size
- Added focused graph tests covering default config shaping and deterministic local-HTML execution with a stub LLM
- Added runnable example under `examples/smartscraper_lite/`
- **Impact:** graph parity moved forward with the expected lightweight variant of the core SmartScraper workflow, using a dedicated public graph rather than an internal-only configuration tweak

## What Was Built This Session

### 29. SearchLinkGraph Parity Slice (`pkg/scrapegraph/search_link.go`)
- Added `SearchLinkGraph` as a dedicated high-level graph for fetch → link discovery workflows
- Reused the existing `SearchLinkNode` with graph-level loader selection for local HTML, UTLS, or Rod fetching
- Added focused graph tests for deterministic link extraction and relative-link resolution
- Added runnable example under `examples/search_link/`
- **Impact:** parity moved forward by exposing an existing internal link-discovery primitive as a canonical user-facing graph

### 28. Markdownify Parity Slice (`pkg/markdown/markdown.go`, `pkg/nodes/markdownify.go`, `pkg/scrapegraph/markdownify.go`)
- Added lightweight HTML→Markdown conversion support alongside the existing HTML→text utility
- Added `MarkdownifyNode` as a canonical transformation node that reads `html` and writes `markdown`
- Added `MarkdownifyGraph` as a dedicated high-level graph for converting a page or local HTML snippet directly into Markdown output
- Added focused node/graph tests plus a runnable example under `examples/markdownify/`
- **Impact:** parity moved forward on both the missing node inventory and missing graph inventory using a native ScrapeGraphAI-style slice rather than more orchestration-only work

### 26. Structured-Data Graph Parity Slice (`pkg/loaders/structured_data*.go`, `pkg/scrapegraph/*_scraper.go`)
- Added canonical parity-oriented loaders for local/directory-based `JSON`, `XML`, and `CSV` sources
- New `StructuredDataLoader` supports both single-file inputs and directory aggregation with stable file ordering
- Added `JSONScraperGraph`, `XMLScraperGraph`, and `CSVScraperGraph` as dedicated high-level graph types matching the ScrapeGraphAI graph inventory
- Introduced `GenerateAnswerCSVNode` as a minimal parity placeholder that currently reuses the standard extraction flow until CSV-specific prompting is expanded further
- **Impact:** graph parity moved forward with real ScrapeGraphAI-native graph types instead of custom orchestration surface work

### 27. Structured-Data Test Coverage
- Added loader tests covering:
  - single-file JSON loading
  - directory-based CSV aggregation with deterministic order
  - unsupported-extension rejection
- Added graph-level tests covering:
  - shared structured scraper execution path
  - default config initialization for CSV graph construction
- Verified targeted package tests pass for:
  - `./pkg/loaders`
  - `./pkg/scrapegraph`
  - `./pkg/nodes`
  - `./cmd/server`
- **Impact:** the new parity slice is regression-protected and validated without expanding provider scope

## What Was Built This Session

### 21. DepthSearchGraph Crawl Guardrails (`pkg/scrapegraph/depth_search*.go`)
- Added crawl policy controls to `DepthSearchGraph`
- New options include: `AllowedDomains`, `RestrictToHost`, `PathPrefixes`, `IncludePatterns`, `ExcludePatterns`, and `MaxPages`
- Added URL normalization and junk-link rejection helpers to reduce noisy traversal
- Crawl now stops once `MaxPages` is reached and deduplicates via normalized URLs
- **Impact:** recursive crawling is now safer and more practical for large docs/help-center sites and broader site trees

### 22. Unified Library Entrypoint (`pkg/scrapegraph/research.go`)
- Added `ResearchGraph` as a high-level orchestration API for integration into other Go applications
- Added `ResearchRequest` and `ResearchResult` types
- `ResearchGraph` selects between:
  - direct single-page extraction
  - depth crawl extraction from a seed URL
  - search-first extraction flow
- Returns integration-friendly metadata including mode, sources, failed URLs, and pages used
- **Impact:** library now has a cleaner single entrypoint for the core natural-language research/extraction use case without replacing the lower-level graph APIs

### 23. Targeted Tests for New Phase-1 Parity Slice
- Added `research_test.go` covering direct, depth, and search orchestration mode selection
- Added `depth_search_policy_test.go` covering URL normalization and crawl policy enforcement
- Verified package tests pass for:
  - `./pkg/scrapegraph`
  - `./cmd/server`
- **Impact:** new orchestration and crawl-guardrail behavior now has regression protection

### 24. Hermes Live Validation + Saved Output
- Updated `examples/hermes_docs/main.go` so the site-wide `ResearchGraph` result is persisted to `examples/hermes_docs/output.json`
- Executed the example live against `https://hermes-agent.nousresearch.com/docs`
- Confirmed the flow works end-to-end: root page → subpage crawl → merged structured JSON → saved output file
- **Observed limitation:** output quality was below expectation, but the run used `gpt-4o-mini` as a lightweight/legacy validation model rather than a stronger extraction model
- **Interpretation:** current limitation appears to be a combination of crawl breadth and model quality, not a failure of the underlying web-scraping pipeline

### 25. Hermes Example Crawl Tuning (`examples/hermes_docs/main.go`)
- Replaced the previously hard-coded conservative research crawl settings with a named balanced profile helper
- New default profile uses `MaxDepth: 2`, `MaxPages: 20`, and `MaxLinksPerPage: 10`
- Added console output showing the active research crawl profile before the site-wide run starts
- **Impact:** the Hermes docs example now better reflects an intentional “balanced coverage” crawl instead of looking accidentally shallow, while keeping host/path guardrails intact

### 1. Infrastructure Hardening (`pkg/graph/graph.go`)
- Added `MaxIterations = 100` constant to prevent infinite loops
- Graph executor now counts iterations and bails out with clear error if exceeded
- BranchNode cycles are the typical cause — now safely caught
- Verbose logging includes iteration count
- **Impact:** Production-safe graph execution, no runaway loops

### 2. SearchGraph Timeout Support (`pkg/scrapegraph/search.go`)
- Added `Timeout` field — overall timeout for entire search workflow
- Added `PerURLTimeout` field — per-URL scrape limit (default 60s)
- Each URL scrape gets its own `context.WithTimeout`
- If overall context expires mid-scrape, returns partial results gracefully
- **Impact:** Production-safe search with predictable timeouts

### 3. Rod Browser Loader (`pkg/loaders/rod.go`)
- Full Rod + stealth integration implementing `Loader` interface
- Headless Chrome with anti-detection patches
- Automatic Cloudflare challenge detection and waiting
- Configurable wait times, timeouts, headless mode
- `NewDefaultRodLoader(verbose)` for quick setup
- Windows-friendly: `Leakless(false)` to avoid Defender issues
- **Impact:** JavaScript-rendered pages now supported, CF bypass

### 4. Config Headless Flag (`internal/models/models.go`)
- Added `Headless bool` to Config struct
- SmartScraperGraph auto-selects Rod vs UTLS based on flag
- Default: `false` (UTLS for speed), set `true` for JS pages
- **Impact:** Single config flag switches fetching strategy

### 5. GraphIteratorNode (`pkg/nodes/graph_iterator.go`)
- Runs a `GraphFactory` function for each item in a string list
- Configurable input/output state keys, per-item timeout
- Graceful failure handling — skips failed items, continues
- Reports succeeded/failed counts, stores failed items in state
- **Impact:** Architectural enabler for all Multi-graph variants

### 6. SearchLinkNode (`pkg/nodes/search_link.go`)
- Extracts all `<a href>` links from HTML using `golang.org/x/net/html`
- Resolves relative URLs against base URL
- Deduplicates links
- Optional LLM relevance filtering — sends numbered link list to LLM, picks most relevant
- Graceful fallback to top-N if LLM fails
- **Impact:** Core component for DepthSearchGraph and link discovery

### 7. DepthSearchGraph (`pkg/scrapegraph/depth_search.go`)
- Recursive crawling from a seed URL up to configurable depth
- Per-depth: scrape page → discover links → recurse
- Visited URL tracking prevents cycles
- LLM-based link relevance filtering (optional)
- Per-URL and overall timeout support
- Merges all results via MergeAnswersNode
- Attaches metadata: sources, pages_crawled, max_depth
- **Impact:** Third graph type, enables site-wide data extraction

### 8. HTTP Server Update (`cmd/server/main.go`)
- New `POST /depth-search` endpoint
- Accepts: url, prompt, max_depth, max_links_per_page, filter_by_llm, headless
- Returns: result, visited_urls, model_used, total_time_s
- **Impact:** DepthSearchGraph accessible via HTTP API

### 9. Multi-URL Extraction Workflow (`pkg/scrapegraph/smartscraper_multi.go`)
- Added `SmartScraperMultiGraph` for batch scraping many URLs with one prompt
- Reuses `GraphIteratorNode` to run `SmartScraperGraph` once per URL
- Supports either separate per-URL results or deterministic concatenation via `ConcatAnswersNode`
- Tracks failed URLs for graceful partial-success behavior
- Added runnable example at `examples/multi_scrape.go`
- **Impact:** Package now has a practical, library-first multi-page scraping workflow

### 10. Hermes Docs Test Target
- Identified `https://hermes-agent.nousresearch.com/docs/` as a concrete real-world validation target
- Confirmed the site is a Docusaurus docs site with stable, scrapeable internal pages
- Added dedicated runnable example at `examples/hermes_docs/main.go`
- Example covers both single-page extraction and multi-page extraction on Hermes docs
- **Impact:** Project now has an immediately useful real-world test/demo target instead of only generic example.com samples

## File Summary

### New Files Created (4)
| File | Purpose | Lines |
|------|---------|-------|
| `pkg/loaders/rod.go` | Rod headless browser loader | ~175 |
| `pkg/nodes/graph_iterator.go` | Sub-graph execution per item | ~135 |
| `pkg/nodes/search_link.go` | HTML link extraction + LLM filter | ~210 |
| `pkg/scrapegraph/depth_search.go` | Recursive crawling workflow | ~235 |

### Modified Files (5)
| File | Changes |
|------|---------|
| `pkg/graph/graph.go` | MaxIterations loop protection, iteration counting |
| `pkg/scrapegraph/search.go` | Timeout + PerURLTimeout support, context cancellation |
| `pkg/scrapegraph/smartscraper.go` | Rod/UTLS auto-selection via config.Headless |
| `internal/models/models.go` | Added Headless config field |
| `cmd/server/main.go` | Added /depth-search endpoint |

### Dependencies Added
| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/go-rod/rod` | v0.116.2 | Headless Chrome automation |
| `github.com/go-rod/stealth` | v0.4.9 | Anti-detection patches |

## Architecture Decisions Made

### Rod Loader Design
- Implements `Loader` interface — drop-in replacement for UTLS
- Launches fresh browser per request (no connection pooling yet)
- Cloudflare detection via page title check
- Returns `FetchResult` with error info rather than failing hard on navigation errors

### DepthSearchGraph Design
- Recursive approach (not graph-node-based) — simpler for depth tracking
- Visited URL set prevents infinite loops
- Links discovered via separate fetch (re-fetches page for link extraction)
- LLM link filtering is optional but recommended for focused crawling
- All results merged at end via MergeAnswersNode

### GraphIteratorNode Design
- Takes a `GraphFactory func(ctx, item) (json.RawMessage, error)`
- Factory pattern allows any graph type to be used as the sub-graph
- Per-item timeouts via context
- Failed items tracked separately in state

## Next Session Goals

### Week 3 Priority
1. **Document loaders** — PDF and DOCX parsing
2. **Telemetry basics** — Structured logging, timing metrics
3. HTTP endpoint for multi-URL scraping
4. Add smoke tests for SmartScraperMultiGraph

### Infrastructure Improvements
- Add connection pooling to Rod loader (reuse browser across requests)
- Add parallel execution option to GraphIteratorNode (goroutines)
- Optimize DepthSearchGraph to reuse fetched HTML for link discovery
- Consider adding `CacheLoader` wrapper for repeated URL fetches

## What Was Built This Session

### 11. Multi-Scrape HTTP Endpoint (`cmd/server/`)
- Added `POST /multi-scrape` endpoint exposing `SmartScraperMultiGraph`
- Supports `urls`, `prompt`, `schema_hint`, `model`, `headless`, `concat_results`, and `per_url_timeout_ms`
- Returns both `result` and `failed_urls` for partial-success workflows
- **Impact:** batch scraping is now available over HTTP, not just as a library/example

### 12. Basic HTTP Telemetry (`pkg/telemetry/http.go`)
- Added lightweight structured HTTP timing logs
- Handlers now log handler name, method, path, status, duration, and error
- **Impact:** first telemetry baseline now exists without adding heavy observability dependencies

### 13. Server Refactor for Maintainability (`cmd/server/*.go`)
- Split server code into `types.go`, `handlers.go`, `router.go`, and `http_helpers.go`
- Reduced monolithic `main.go` to server bootstrapping only
- **Impact:** keeps files under the project size constraint and makes future endpoint additions easier

### 14. Document Loader Groundwork (`pkg/loaders/document*.go`)
- Added `DocumentLoader` implementing the existing `Loader` interface
- Supports local `.pdf` and `.docx` file detection and wraps extracted text as minimal HTML for `ParseNode` reuse
- Added stub extractor functions for PDF/DOCX so dependency wiring can be added next without changing the loader contract
- Added unit tests for unsupported extensions and custom extractor injection
- **Impact:** document ingestion architecture is now in place, ready for real parser dependencies

### 15. HTTP Router Smoke Tests (`cmd/server/router_test.go`)
- Added tests validating `/multi-scrape` method enforcement and request validation
- **Impact:** basic regression protection now exists for the new endpoint surface

## Current Notes
- `go test ./...` still fails because `examples/` contains multiple standalone `main` programs in one package (`basic.go`, `test_workflow.go`), which is a pre-existing layout issue unrelated to the new changes.
- Targeted package tests for `cmd/server`, `pkg/loaders`, and `pkg/nodes` pass.
- New targeted tests also pass for `pkg/scrapegraph` and `cmd/server` after adding `ResearchGraph` and crawl guardrails.
- Hermes live validation now also writes output to `examples/hermes_docs/output.json` for inspection and comparison across future tuning passes.
- User feedback: output quality was less than expected, but current live validation used a weaker model tier for cost/speed; quality should be re-evaluated after testing with a stronger model.
- Hermes example defaults are now tuned toward a clearer balanced crawl profile, but still need a fresh live rerun to compare breadth/quality against the previous shallow output.

## Current Focus
- Continue parity-focused Go port work against the Python ScrapeGraphAI reference, not a reinvention of product scope
- Prioritize canonical ScrapeGraphAI graph and node coverage over custom orchestration surfaces
- Next likely step: add more parity-native graph families such as structured-data multi variants, screenshot/lite/search-link graphs, or Markdownify-related functionality
- Near-term quality task: rerun the Hermes balanced profile and compare pages used/output quality before deciding whether to increase depth further or relax link filtering
- Near-term maintenance task: keep memory-bank priorities aligned with parity-first implementation order
- Next likely step after SmartScraperLiteGraph: begin screenshot parity groundwork with `FetchScreenNode` and `ScreenshotScraperGraph`, then continue with other parity-native lightweight/multi variants

## What Was Built This Session

### 16. Real Document Parsing (`pkg/loaders/document*.go`)
- Replaced placeholder PDF extraction with real parsing using `github.com/ledongthuc/pdf`
- Added native DOCX text extraction by reading `word/document.xml` from the zip archive and decoding XML text content
- Kept the existing `DocumentLoader` contract so document input still flows through `FetchNode` → `ParseNode`
- Added DOCX-focused unit coverage in addition to existing unsupported-extension and custom-extractor tests
- **Impact:** document ingestion is now functional for real `.pdf` and `.docx` files instead of architectural stub-only groundwork

### 17. DocumentScraperGraph (`pkg/scrapegraph/document_scraper.go`)
- Added a dedicated high-level graph for local document extraction
- Reuses the same graph pattern as SmartScraperGraph: fetch document → parse text → run LLM extraction
- Added graph-level telemetry timing around execution
- **Impact:** project now has a true document workflow, not just a low-level loader

### 18. Document HTTP API (`cmd/server/document_handler.go`)
- Added `POST /document-scrape`
- Accepts `path`, `prompt`, `model`, and `schema_hint`
- Returns `result`, `path`, `model_used`, and `total_time_s`
- Added request/response types and router coverage
- **Impact:** document extraction is now available through the service API as well as the library surface

### 19. Graph/Node Telemetry Baseline (`pkg/telemetry/graph.go`, `pkg/graph/graph.go`, `pkg/scrapegraph/*.go`)
- Added lightweight graph execution logs with duration and error fields
- Added per-node execution timing inside the graph executor
- Added graph-level timing to SmartScraperGraph and DocumentScraperGraph
- **Impact:** observability is now expanding beyond HTTP-only timing into core execution flow

### 20. Document Example + Validation
- Added runnable example at `examples/document_scrape/main.go`
- Ran `go mod tidy`
- Verified targeted tests pass for:
  - `./pkg/loaders`
  - `./pkg/scrapegraph`
  - `./cmd/server`
- **Impact:** new document workflow is both demoable and regression-protected at the targeted package level