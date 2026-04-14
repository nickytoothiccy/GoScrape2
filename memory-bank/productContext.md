# Product Context

## Why GoScrape2 Exists

### Problem Statement
ScrapeGraphAI is an excellent Python library for LLM-powered web scraping, but:
- Python has performance limitations (GIL, memory overhead)
- Type safety is limited (even with type hints)
- Deployment is complex (dependencies, virtual envs)
- Concurrency is harder to reason about
- Native binary distribution would be better

### Solution
A complete Go reimplementation that:
- Maintains 100% functional parity with ScrapeGraphAI
- Leverages Go's type safety, concurrency, and performance
- Provides both library and HTTP API interfaces
- Compiles to single native binary
- Runs anywhere without dependencies

## Target Users

### Primary: Developers
- Building web scraping pipelines
- Need LLM-powered data extraction
- Want type-safe, performant solution
- Prefer Go ecosystem over Python

### Secondary: DevOps/SRE
- Operating scraping infrastructure
- Need reliable, containerizable service
- Want observability and metrics
- Require production-grade error handling

### Tertiary: Researchers
- Experimenting with scraping techniques
- Testing different LLM models
- Comparing extraction strategies
- Need flexible, extensible framework

## Use Cases

### 1. E-commerce Data Extraction
```go
// Extract product details from multiple retailers
scraper := graphs.NewSmartScraperGraph(
    "Extract product name, price, rating, reviews",
    productURL,
    config,
    `{"name": "str", "price": "float", "rating": "float", "reviews": "int"}`,
)
```

### 2. News Aggregation
```go
// Search multiple news sites and extract articles
scraper := graphs.NewSearchGraph(
    "Find articles about AI regulation published this week",
    []string{"reuters.com", "techcrunch.com"},
    config,
)
```

### 3. Competitive Intelligence
```go
// Monitor competitor websites for changes
scraper := graphs.NewDepthSearchGraph(
    "Extract all product launches and pricing changes",
    competitorURL,
    2, // depth
    config,
)
```

### 4. Document Processing
```go
// Extract structured data from PDFs
scraper := graphs.NewDocumentScraperGraph(
    "Extract financial tables and key metrics",
    "report.pdf",
    config,
)
```

### 5. Code Generation
```go
// Generate custom scraping scripts
scraper := graphs.NewCodeGeneratorGraph(
    "Create a scraper for real estate listings",
    exampleURL,
    config,
)
```

## Value Propositions

### Compared to ScrapeGraphAI (Python)
**Advantages:**
- ✅ 10-100x faster execution
- ✅ Lower memory footprint
- ✅ Single binary deployment
- ✅ Built-in concurrency (goroutines)
- ✅ Compile-time type checking
- ✅ No dependency hell

**Parity:**
- ✅ Same graph types
- ✅ Same node library
- ✅ Same model support
- ✅ Same prompt system
- ✅ Same extraction quality

**Trade-offs:**
- ⚠️ Smaller ecosystem than Python
- ⚠️ Fewer pre-built ML integrations
- ⚠️ Python still preferred by some data scientists

### Compared to Custom Scrapers
**Advantages:**
- ✅ LLM-powered extraction (vs brittle XPath)
- ✅ Handles dynamic content changes
- ✅ Multi-model support (fallback strategies)
- ✅ Built-in retry and error handling
- ✅ Standardized architecture

### Compared to SaaS Scraping Services
**Advantages:**
- ✅ Self-hosted (data privacy)
- ✅ No usage-based pricing
- ✅ Full control over models
- ✅ Customizable workflows
- ✅ Open source

## Success Metrics

### Functional Completeness
- [ ] 100% graph type parity (26/26)
- [ ] 100% node type parity (30/30)
- [ ] 100% model provider parity (10/10)
- [ ] Full prompt system
- [ ] Complete telemetry

### Performance Targets
- [ ] <500ms avg extraction time (simple pages)
- [ ] <5s avg extraction time (complex pages)
- [ ] Support 100+ concurrent scrapes
- [ ] <100MB memory per scrape

### Quality Targets
- [ ] 95%+ extraction accuracy (vs ScrapeGraphAI)
- [ ] Handles 99% of sites ScrapeGraphAI handles
- [ ] Zero crashes in production workloads
- [ ] Graceful degradation on errors

### Developer Experience
- [ ] Complete API documentation
- [ ] 50+ working examples
- [ ] Type-safe interfaces
- [ ] Clear error messages
- [ ] Easy onboarding (<10 min)

## Competitive Landscape

### Direct Competitors
1. **ScrapeGraphAI (Python)** - The original, mature ecosystem
2. **Crawl4AI** - Python, focused on LLM-ready data
3. **Langchain Document Loaders** - Part of larger framework

### Indirect Competitors
1. **BeautifulSoup + Custom LLM** - DIY approach
2. **Scrapy + LLM plugin** - Traditional crawler + AI
3. **Browserbase/BrightData** - SaaS solutions
4. **Apify** - Cloud scraping platform

### Our Niche
- **Performance-critical Go applications**
- **Type-safe enterprise deployments**
- **Self-hosted AI scraping**
- **High-concurrency workloads**

## Feature Roadmap

### Phase 1: Foundation (DONE)
- ✅ Core graph engine
- ✅ Basic nodes (Fetch, Parse, Generate)
- ✅ SmartScraperGraph
- ✅ OpenAI integration
- ✅ HTTP API

### Phase 2: Essential Nodes (Q1 2026)
- [ ] ConditionalNode (retry)
- [ ] ReasoningNode (CoT)
- [ ] SearchInternetNode
- [ ] All parsing variants

### Phase 3: Core Graphs (Q1 2026)
- [ ] SearchGraph
- [ ] DepthSearchGraph
- [ ] DocumentScraperGraph
- [ ] Multi-variants

### Phase 4: Model Diversity (Q2 2026)
- [ ] Anthropic (Claude)
- [ ] HuggingFace
- [ ] Groq
- [ ] Gemini
- [ ] Local models (Ollama)

### Phase 5: Advanced Features (Q2 2026)
- [ ] Prompt template system
- [ ] Telemetry/observability
- [ ] Document loaders
- [ ] Browser automation
- [ ] Multi-modal support

### Phase 6: Enterprise Features (Q3 2026)
- [ ] Caching layer
- [ ] Rate limiting
- [ ] Distributed execution
- [ ] Web UI dashboard
- [ ] SaaS deployment option

## Design Philosophy

### Core Principles
1. **Parity over innovation** - Match ScrapeGraphAI first, innovate second
2. **Type safety everywhere** - Leverage Go's type system
3. **Explicit over implicit** - No magic, clear control flow
4. **Library first, API second** - Programmatic use is primary
5. **Modularity** - Every component is swappable

### Non-Goals
- ❌ Not trying to replace Scrapy for simple scraping
- ❌ Not building a no-code scraping GUI (yet)
- ❌ Not competing with enterprise SaaS on convenience
- ❌ Not trying to be "better" than Python for prototyping

### Quality Standards
- All files <200 lines
- 90%+ test coverage
- Zero known memory leaks
- Graceful error handling
- Clear documentation

## Integration Patterns

### As a Library
```go
import "stealthfetch/pkg/graphs"

scraper := graphs.NewSmartScraperGraph(...)
result, err := scraper.Run(ctx)
```

### As a Service
```bash
./stealthgraph &
curl -X POST localhost:8899/scrape -d '{"url":"...", "prompt":"..."}'
```

### As a Container
```dockerfile
FROM golang:1.22-alpine AS builder
COPY . /build
RUN go build -o stealthgraph cmd/server/main.go

FROM alpine:latest
COPY --from=builder /build/stealthgraph /
CMD ["/stealthgraph"]
```

### With Existing Go Apps
```go
// Embed in existing service
import "stealthfetch/pkg/scrapegraph"

func handleExtraction(url, prompt string) (json.RawMessage, error) {
    scraper := scrapegraph.NewSmartScraperGraph(prompt, url, cfg, "")
    return scraper.Run(context.Background())
}
```

This is a **production-grade library**, not a toy project.