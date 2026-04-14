# Architecture: ScrapeGraphAI Port to Go

## Comparison with ScrapeGraphAI

### Core Components Implemented

| ScrapeGraphAI (Python) | StealthFetch (Go) | Status |
|------------------------|-------------------|--------|
| `AbstractGraph` | `graph.Graph` | ✅ Complete |
| `BaseGraph` | `graph.Graph` execution engine | ✅ Complete |
| `State` | `graph.State` with thread-safety | ✅ Complete |
| `Node` interface | `graph.Node` interface | ✅ Complete |
| `SmartScraperGraph` | `scrapegraph.SmartScraperGraph` | ✅ Complete |
| `FetchNode` | `nodes.FetchNode` | ✅ Complete |
| `ParseNode` | `nodes.ParseNode` | ✅ Complete |
| `GenerateAnswerNode` | `nodes.GenerateAnswerNode` | ✅ Complete |
| Multi-backend loading | `loaders.Loader` interface | ✅ Complete |
| Chunking logic | `chunking.Chunker` | ✅ Complete |
| HTML→text | `markdown.HTMLToText` | ✅ Complete |
| LLM client | `llm.OpenAIClient` | ✅ Complete |
| Map-reduce extraction | `llm.MergeExtractions` | ✅ Complete |

### What's Different

**Improvements:**
- **Type safety**: Strong typing throughout vs Python dicts
- **Concurrency**: Thread-safe state with sync.RWMutex
- **Modularity**: All files under 200 lines
- **Explicitness**: No hidden config magic, clear interfaces
- **Performance**: Native Go vs Python interpretation

**Not Yet Implemented:**
- `ConditionalNode` for retry logic
- `ReasoningNode` for chain-of-thought
- `SearchGraph`, `DepthGraph`, `DocumentGraph`
- Rod browser loader (webscrape already has this)
- Prompt templates and DSL

## Package Structure

```
stealthfetch/
├── internal/
│   └── models/         # Shared types (Config, Results)
├── pkg/
│   ├── graph/          # Execution engine
│   │   ├── graph.go    # Graph with node orchestration
│   │   ├── node.go     # Node interface + BaseNode
│   │   └── state.go    # Thread-safe state management
│   ├── nodes/          # Processing nodes
│   │   ├── fetch.go    # Multi-source content loading
│   │   ├── parse.go    # HTML→text + chunking
│   │   └── generate.go # LLM extraction + merge
│   ├── loaders/        # Fetch backends
│   │   ├── loader.go   # Loader interface
│   │   ├── local.go    # Local HTML passthrough
│   │   └── utls.go     # UTLS TLS fingerprinting
│   ├── chunking/       # Text splitting
│   │   └── chunking.go # Token-aware chunking
│   ├── markdown/       # HTML conversion
│   │   └── markdown.go # HTML→clean text
│   ├── llm/            # LLM client
│   │   └── openai.go   # OpenAI with merge logic
│   └── scrapegraph/    # High-level API
│       └── smartscraper.go # SmartScraperGraph workflow
├── cmd/
│   └── server/         # HTTP wrapper
│       └── main.go     # REST API endpoints
└── examples/           # Usage examples
    ├── basic.go
    └── test_workflow.go
```

## Data Flow

### SmartScraperGraph Execution

```
1. User creates scraper:
   NewSmartScraperGraph(prompt, source, config, schema)

2. buildGraph() constructs workflow:
   FetchNode → ParseNode → GenerateAnswerNode

3. Run() executes:
   state.Set("url", source)
   → graph.Execute(ctx, state)
   → return state.GetJSON("extracted_data")

4. Each node:
   - Validates inputs exist in state
   - Executes logic
   - Writes outputs to state
   - Returns error or nil
```

### State Keys

| Key | Type | Set By | Used By |
|-----|------|--------|---------|
| `url` | string | User | FetchNode |
| `html` | string | FetchNode | ParseNode |
| `fetch_result` | FetchResult | FetchNode | - |
| `text` | string | ParseNode | - |
| `chunks` | []string | ParseNode | GenerateAnswerNode |
| `parse_result` | ParseResult | ParseNode | - |
| `extracted_data` | json.RawMessage | GenerateAnswerNode | User |
| `extract_result` | ExtractResult | GenerateAnswerNode | - |

## Extraction Strategies

### Single Chunk
```
FetchNode: URL → HTML (155KB)
ParseNode: HTML → Text (30KB) → [single chunk]
GenerateAnswerNode:
  - Extract(chunk, prompt) → JSON
  - Return directly
```

### Multi-Chunk (Map-Reduce)
```
FetchNode: URL → HTML (500KB)
ParseNode: HTML → Text (200KB) → [chunk1, chunk2, chunk3]
GenerateAnswerNode:
  - Extract(chunk1, prompt) → result1
  - Extract(chunk2, prompt) → result2
  - Extract(chunk3, prompt) → result3
  - MergeExtractions([result1, result2, result3]) → merged JSON
  - Return merged
```

## Configuration

All configuration through typed structs:

```go
type Config struct {
    LLMModel      string  // "gpt-4o-mini"
    LLMAPIKey     string  // OpenAI key
    Temperature   float64 // 0 for deterministic
    MaxTokens     int     // 4000
    Verbose       bool    // Debug logging
    HTMLMaxChars  int     // 50000 max HTML to process
    ChunkSize     int     // 8000 chars per chunk
    ChunkOverlap  int     // 200 chars overlap
}
```

## Next Phase: Rod Integration

To match webscrape's stealth capabilities:

```go
// pkg/loaders/rod.go
type RodLoader struct {
    headless bool
    timeout  time.Duration
}

func (l *RodLoader) Load(ctx context.Context, source string) (*models.FetchResult, error) {
    browser := rod.New().Timeout(l.timeout)
    if l.headless {
        browser = browser.MustLaunch()
    } else {
        browser = browser.NoDefaultDevice().MustLaunch()
    }
    defer browser.MustClose()
    
    page := stealth.MustPage(browser)
    page.MustNavigate(source).MustWaitLoad()
    
    // CF detection...
    html := page.MustHTML()
    
    return &models.FetchResult{
        HTML: html,
        // ...
    }, nil
}
```

Then in SmartScraperGraph, allow loader selection:

```go
scraper := scrapegraph.NewSmartScraperGraphWithLoader(
    prompt,
    url,
    config,
    loaders.NewRodLoader(true, 30*time.Second),
)
```

This preserves the stealth path from webscrape while adding pure Go extraction.