# Tech Context

## Environment
- OS: Windows 11
- Working directory: `c:\Users\Nicky\GoScrape2`
- Shell: PowerShell (pwsh)
- Note: Use `;` instead of `&&` for command chaining

## Go Setup
- Go version: 1.22
- Module: `stealthfetch`
- Primary development: Library first, HTTP API secondary

## Current Dependencies

### Installed
```go
require (
    github.com/go-rod/rod v0.116.2
    github.com/go-rod/stealth v0.4.9
    github.com/openai/openai-go/v3 v3.0.0
    github.com/refraction-networking/utls v1.6.6
    golang.org/x/net v0.34.0
)
```

### Indirect
- github.com/andybalholm/brotli
- github.com/cloudflare/circl
- github.com/klauspost/compress
- github.com/ysmood/fetchup, goob, got, gson, leakless (Rod deps)
- golang.org/x/crypto
- golang.org/x/sys
- golang.org/x/text

## Required Dependencies for Full Implementation

### Model Providers
```go
// Anthropic
github.com/anthropics/anthropic-sdk-go

// HuggingFace
github.com/hupe1980/go-huggingface

// Groq (uses OpenAI-compatible API)
// Can reuse OpenAI client with different base URL

// Google Gemini
github.com/google/generative-ai-go

// Ollama
github.com/ollama/ollama/api
```

### Browser Automation
```go
// Playwright
github.com/playwright-community/playwright-go

// Rod (already in webscrape, need to import)
github.com/go-rod/rod
github.com/go-rod/stealth
```

### Document Processing
```go
// PDF
github.com/ledongthuc/pdf
// or
github.com/pdfcpu/pdfcpu

// DOCX
baliance.com/gooxml
// or
github.com/nguyenthenguyen/docx
```

### Utilities
```go
// HTML parsing (already have via golang.org/x/net)
golang.org/x/net/html

// Text-to-Speech
github.com/hajimehoshi/oto (audio output)
// + TTS service API

// Image processing
github.com/disintegration/imaging

// HTTP client enhancements
github.com/go-resty/resty/v2
```

## Package Structure

```
stealthfetch/
├── go.mod
├── go.sum
├── internal/
│   ├── models/         # Shared types (Config, Results) ✅
│   ├── config/         # Configuration management
│   └── errors/         # Custom error types
├── pkg/
│   ├── graph/          # Core engine ✅ (loop protection, branching)
│   ├── scrapegraph/    # 3 graph workflows (SmartScraper, Search, DepthSearch)
│   ├── nodes/          # 9/30 nodes done
│   ├── llm/            # LLM providers (1/10 done, interface ready)
│   ├── loaders/        # Fetch backends (3/4: Local, UTLS, Rod)
│   ├── prompts/        # Template system ✅
│   ├── telemetry/      # Observability (0%)
│   ├── chunking/       # Text splitting ✅
│   ├── markdown/       # HTML conversion ✅
│   └── utils/          # Search utilities ✅
├── cmd/
│   └── server/         # HTTP API ✅ (5 endpoints)
├── examples/           # Usage examples
├── tests/              # Test suite (0%)
└── memory-bank/        # Project documentation ✅
```

## File Size Constraint
- Maximum 200 lines per file (enforced)
- Forces good separation of concerns
- Improves maintainability
- Easy code review

## Build Process

### Development
```bash
# Run server
cd cmd/server
go run main.go

# Build server
go build -o ../../stealthgraph.exe

# Run tests
go test ./...

# Run specific package tests
go test ./pkg/graph
```

### Library Usage
```go
import "stealthfetch/pkg/scrapegraph"

config := models.DefaultConfig()
config.LLMAPIKey = "sk-..."

scraper := scrapegraph.NewSmartScraperGraph(
    "Extract product names",
    "https://example.com",
    config,
    "",
)

result, err := scraper.Run(context.Background())
```

## ScrapeGraphAI Python Reference

### Location
`c:\Users\Nicky\webscrape\Reference_Material\Scrapegraph-ai\`

### Key Files to Study
```
scrapegraphai/
├── graphs/
│   ├── abstract_graph.py       # Base config & setup
│   ├── base_graph.py           # Execution engine
│   ├── smart_scraper_graph.py  # Primary graph ✅
│   ├── search_graph.py         # Search workflow
│   ├── depth_search_graph.py   # Recursive crawl
│   └── ... (23 more)
├── nodes/
│   ├── base_node.py            # Node interface
│   ├── fetch_node.py           # Fetch implementation ✅
│   ├── parse_node.py           # Parse implementation ✅
│   ├── generate_answer_node.py # Generate implementation ✅
│   ├── conditional_node.py     # Retry/branching
│   ├── reasoning_node.py       # Chain-of-thought
│   └── ... (25 more)
├── models/
│   ├── openai.py              # OpenAI integration ✅
│   ├── anthropic.py           # Claude integration
│   ├── hugging_face.py        # HF integration
│   └── ... (7 more)
├── prompts/
│   └── ... (template definitions)
├── utils/
│   └── ... (shared utilities)
└── telemetry/
    └── ... (observability)
```

## Testing Strategy

### Unit Tests
- Test each node independently
- Mock state and dependencies
- Verify input/output contracts

### Integration Tests
- Test complete graph flows
- Use real HTTP responses (saved)
- Verify end-to-end extraction

### Examples as Tests
- Every graph type has example usage
- Examples double as smoke tests
- Kept in `examples/` directory

## Performance Considerations

### Go Advantages
- Native concurrency for parallel fetching
- Lower memory footprint than Python
- Faster execution (no GIL)
- Static typing catches errors at compile time

### Optimization Targets
- Parallel node execution where possible
- Connection pooling for HTTP
- Efficient chunking algorithms
- Smart caching of LLM responses

## Development Workflow

1. Study ScrapeGraphAI Python implementation
2. Design Go equivalent (types, interfaces)
3. Implement in small, testable pieces
4. Keep files <200 lines
5. Add examples
6. Document in memory bank
7. Move to next component

## Known Issues
- PowerShell workspace syntax (@workspace:) doesn't work - use full paths
- Must use semicolons (;) not && for command chaining
- Windows paths need escaping in some contexts

## Next Dependencies to Add
Priority order based on implementation plan:
1. ~~Rod browser~~ ✅ DONE
2. PDF parser (github.com/ledongthuc/pdf or github.com/pdfcpu/pdfcpu)
3. DOCX parser (github.com/nguyenthenguyen/docx)
4. Anthropic SDK (github.com/anthropics/anthropic-sdk-go)
5. Ollama client (github.com/ollama/ollama/api)
