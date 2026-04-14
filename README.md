## StealthFetch - Go ScrapeGraph Library

A modular Go library for stealth web scraping with LLM-powered data extraction, inspired by ScrapeGraphAI.

### Architecture

```
pkg/
├── graph/          # Graph execution engine
├── nodes/          # Processing nodes (Fetch, Parse, Generate)
├── loaders/        # Fetch backends (UTLS, Local HTML)
├── chunking/       # Text splitting for LLM processing
├── markdown/       # HTML to text conversion
├── llm/            # OpenAI client with extraction logic
└── scrapegraph/    # High-level SmartScraperGraph API
```

### Features

- **Graph-based workflow**: Composable nodes with state management
- **Smart chunking**: Token-aware text splitting with map-reduce extraction
- **Multi-backend fetching**: UTLS TLS fingerprinting, Rod browser, or local HTML
- **Schema-aware extraction**: Structured JSON output with merge logic
- **Multi-page extraction**: Batch scrape multiple URLs with one workflow
- **Modular design**: All files under 200 lines for maintainability

### Current usable workflows

- `SmartScraperGraph` — single page extraction
- `SmartScraperMultiGraph` — multi-URL extraction with optional concatenation
- `SearchGraph` — search the web, scrape results, merge answers
- `DepthSearchGraph` — recursive crawling with link discovery

### Quick Start

**As a library:**

```go
import (
    "context"
    "stealthfetch/internal/models"
    "stealthfetch/pkg/scrapegraph"
)

config := models.DefaultConfig()
config.LLMAPIKey = "your-api-key"
config.LLMModel = "gpt-4o-mini"

scraper := scrapegraph.NewSmartScraperGraph(
    "Extract all product names and prices",
    "https://example.com/products",
    config,
    "",
)

result, err := scraper.Run(context.Background())
```

**Multi-URL extraction:**

```go
multi := scrapegraph.NewSmartScraperMultiGraph(scrapegraph.SmartScraperMultiConfig{
    Config:        config,
    Prompt:        "Extract title and summary as JSON",
    URLs:          []string{"https://example.com", "https://example.org"},
    ConcatResults: false,
})

result, err := multi.Run(context.Background())
failed := multi.GetFailedURLs()
```

**Real target: Hermes docs**

The Hermes docs site works well as both a single-page and multi-page target because it is a Docusaurus documentation site with stable, linkable pages.

Example pages:

- `https://hermes-agent.nousresearch.com/docs/`
- `https://hermes-agent.nousresearch.com/docs/getting-started/installation`
- `https://hermes-agent.nousresearch.com/docs/getting-started/quickstart`
- `https://hermes-agent.nousresearch.com/docs/user-guide/features/memory`
- `https://hermes-agent.nousresearch.com/docs/user-guide/features/skills`

Runnable example:

```bash
go run ./examples/hermes_docs
```

**As an HTTP service:**

```bash
export OPENAI_API_KEY=your-key
cd cmd/server
go run main.go
```

```bash
curl -X POST http://localhost:8899/scrape \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com",
    "prompt": "Extract the main heading",
    "model": "gpt-4o-mini"
  }'
```

### Graph Workflow

The SmartScraperGraph implements this workflow:

```
FetchNode → ParseNode → GenerateAnswerNode
   ↓            ↓              ↓
  html    →  chunks   →  extracted_data
```

1. **FetchNode**: Loads content (URL or local HTML)
2. **ParseNode**: Converts HTML to text, splits into chunks
3. **GenerateAnswerNode**: 
   - Single chunk: Direct LLM extraction
   - Multiple chunks: Map-reduce with merge pass

### Node System

Each node:
- Implements the `Node` interface
- Reads from and writes to shared `State`
- Validates inputs before execution
- Can be composed into custom graphs

### Configuration

```go
config := &models.Config{
    LLMModel:     "gpt-4o",           // OpenAI model
    LLMAPIKey:    "sk-...",            // API key
    Temperature:  0,                   // LLM temperature
    HTMLMaxChars: 50000,               // Max HTML to process
    ChunkSize:    8000,                // Chars per chunk
    ChunkOverlap: 200,                 // Overlap between chunks
    Verbose:      false,               // Debug logging
}
```

### HTTP Endpoints

- `POST /scrape` - Full workflow: fetch + extract
- `POST /fetch` - Fetch only (returns raw HTML)
- `POST /search` - Search + scrape + merge
- `POST /depth-search` - Recursive crawl + merge
- `GET /health` - Health check

### Next Steps

- [x] Add Rod browser loader for JS-heavy sites
- [x] Implement ConditionalNode for retry logic
- [x] Add ReasoningNode for chain-of-thought extraction
- [x] Create SearchGraph for multi-page crawling
- [ ] Add document loaders (PDF, DOCX)
- [ ] Add telemetry basics
- [ ] Expand test coverage with fixtures and graph smoke tests