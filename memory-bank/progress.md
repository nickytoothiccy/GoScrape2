# Progress

## Phase 1: Core Infrastructure ✅ (100%)

**Completed:**
- ✅ Graph execution engine (`pkg/graph/`)
  - Graph struct with node orchestration
  - State management with thread safety
  - Node interface definition
  - Edge/dependency management
  - BranchNode interface for conditional execution
  - Name-based node lookup (nodeMap)
  - **NEW: MaxIterations loop protection (100 iteration limit)**
- ✅ Basic node implementations (10/30)
  - FetchNode (multi-backend support)
  - ParseNode (HTML→text, chunking)
  - GenerateAnswerNode (single + map-reduce)
  - ConditionalNode (branching logic, retry support)
  - ReasoningNode (chain-of-thought analysis)
  - MergeAnswersNode (multi-source result merging)
  - SearchInternetNode (LLM query + web search)
  - **NEW: GraphIteratorNode (sub-graph execution per item)**
  - **NEW: SearchLinkNode (HTML link extraction + LLM relevance filtering)**
  - **NEW: MarkdownifyNode (HTML→Markdown transformation)**
- ✅ Loader system (4/4)
  - Local HTML passthrough
  - UTLS TLS fingerprinting
  - **NEW: Rod headless browser (stealth, Cloudflare bypass)**
  - **NEW: Document loader with real PDF/DOCX extraction**
  - **NEW: Shared fetch strategy factory + UTLS→Rod auto-escalation on likely block pages**
- ✅ Supporting packages
  - Chunking (token-aware splitting)
  - Markdown (HTML→text conversion)
  - Prompt templates (`pkg/prompts/`)
  - Search utilities (`pkg/utils/`)
- ✅ SmartScraperGraph (with retry logic, Rod support)
- ✅ SearchGraph (with per-URL + overall timeouts)
- ✅ **NEW: DepthSearchGraph** (3/26 graphs)
- ✅ **NEW: SmartScraperMultiGraph** (4/26 graphs)
- ✅ **NEW: DocumentScraperGraph** (5/26 graphs)
- ✅ **NEW: JSONScraperGraph** (6/26 graphs)
- ✅ **NEW: XMLScraperGraph** (7/26 graphs)
- ✅ **NEW: CSVScraperGraph** (8/26 graphs)
- ✅ **NEW: MarkdownifyGraph** (9/26 graphs)
- ✅ **NEW: SearchLinkGraph** (10/26 graphs)
- ✅ **NEW: SmartScraperLiteGraph** (11/26 graphs)
- ✅ **NEW: ResearchGraph** (high-level orchestration entrypoint over existing workflows)
- ✅ LLM interface abstraction (`pkg/llm/interface.go`)
- ✅ HTTP server wrapper
	- `/scrape`, `/document-scrape`, `/multi-scrape`, `/search`, `/depth-search`, `/fetch`, `/health` endpoints
- ✅ Hermes docs identified and wired as a real-world scrape target example

## Phase 2: Complete Node Library (10/30 = 33%)

**Completed:**
- [x] FetchNode ✅
- [x] ParseNode ✅
- [x] GenerateAnswerNode ✅
- [x] ConditionalNode ✅ (with factory condition functions)
- [x] ReasoningNode ✅ (chain-of-thought)
- [x] MergeAnswersNode ✅ (multi-source merging)
- [x] SearchInternetNode ✅ (LLM + DuckDuckGo)
- [x] **GraphIteratorNode** ✅ (sub-graph per item, timeout support)
- [x] **SearchLinkNode** ✅ (HTML link extraction, LLM filtering)

**Not Started:**
- [ ] FetchNodeLevelK (depth-aware)
- [ ] FetchScreenNode (screenshots)
- [ ] RobotsNode (robots.txt)
- [ ] ParseNodeDepthK (depth-aware parsing)
- [x] MarkdownifyNode (HTML→MD)
- [ ] HTMLAnalyzerNode (structure analysis)
- [ ] GenerateAnswerNodeKLevel
- [x] GenerateAnswerCSVNode ✅ (initial parity version; currently reuses standard extraction flow)
- [ ] GenerateAnswerFromImageNode
- [ ] GenerateAnswerOmniNode
- [ ] GenerateCodeNode
- [ ] GenerateScraperNode
- [ ] PromptRefinerNode
- [x] ConcatAnswersNode
- [ ] MergeGeneratedScriptsNode
- [ ] GetProbableTagsNode
- [ ] ImageToTextNode
- [ ] TextToSpeechNode
- [ ] RAGNode
- [ ] DescriptionNode
- [ ] SearchNodeWithContext

## Phase 3: All Graph Types (11/26 = 42%)

**Single-Page Extraction:**
- ✅ SmartScraperGraph (with retry logic + Rod/UTLS)
- ✅ SmartScraperLiteGraph
- ✅ **DocumentScraperGraph**
- [ ] ScreenshotScraperGraph
- ✅ **JSONScraperGraph**
- ✅ **XMLScraperGraph**
- ✅ **MarkdownifyGraph**

**Search & Discovery:**
- ✅ SearchGraph (timeout support, per-URL limits)
- ✅ **DepthSearchGraph** (recursive link following, LLM link filtering)
- ✅ **SearchLinkGraph**
- [ ] OmniSearchGraph

**Structured Data:**
- ✅ **CSVScraperGraph**

**Multi-Page / Code Gen / Specialized:**
- [ ] 15 more graph types

**Multi-Page Extraction:**
- ✅ SmartScraperMultiGraph

## Phase 4: Multi-Model Support (1/10 = 10%)

**Completed:**
- ✅ OpenAI (gpt-4o, gpt-4o-mini, JSON mode, streaming-ready)
- ✅ LLM interface (all providers implement same contract)

**Not Started:**
- [ ] Anthropic (Claude)
- [ ] HuggingFace
- [ ] Gemini
- [ ] DeepSeek
- [ ] Ollama
- [ ] Azure OpenAI
- [ ] Bedrock
- [ ] Vertex AI

## Phase 5: Advanced Features (~40%)

**Completed:**
- ✅ Prompt template system (extraction, reasoning, search templates)
- ✅ Error handling & retry logic (SmartScraperGraph)
- ✅ Conditional branching (graph engine + ConditionalNode)
- ✅ **Infinite loop protection** (MaxIterations = 100)
- ✅ **Context timeout support** (SearchGraph overall + per-URL)
- ✅ **Browser automation** (Rod headless + stealth + Cloudflare bypass)
- ✅ **Block-aware fetch escalation** (`FetchStrategy`, block detection heuristics, UTLS→Rod fallback)
- ✅ **Telemetry basics** (HTTP structured logging + timing)
- ✅ **Graph/node timing telemetry baseline**
- ✅ **Document loaders (PDF, DOCX)**

**Not Started:**
 - [ ] Telemetry (metrics, tracing, richer aggregation)
- [ ] Screenshot capture
- [ ] Rate limiting
- [ ] Caching layer
- [ ] Streaming responses
- [ ] Vision model support
- [ ] Speech synthesis

## Overall Completion: ~44%

**What's Working:**
- SmartScraperGraph with retry + Rod headless browser support
- DocumentScraperGraph for local PDF/DOCX extraction
- JSONScraperGraph for local JSON file/directory extraction
- XMLScraperGraph for local XML file/directory extraction
- CSVScraperGraph for local CSV file/directory extraction
- SmartScraperMultiGraph for batch scraping with optional deterministic concatenation
- HTTP `/document-scrape` endpoint for document extraction workflows
- HTTP `/multi-scrape` endpoint for batch scraping workflows
- Hermes docs example for real-world single-page and multi-page validation
- Hermes docs site-wide `ResearchGraph` example with saved output at `examples/hermes_docs/output.json`
- Hermes docs site-wide example now defaults to a named balanced crawl profile (`MaxDepth: 2`, `MaxPages: 20`, `MaxLinksPerPage: 10`) instead of the earlier more conservative settings
- SearchGraph with per-URL timeouts and overall timeout
- DepthSearchGraph for recursive crawling with LLM link filtering
- DepthSearchGraph crawl guardrails: host/domain/path restriction, include/exclude filtering, max pages, URL normalization
- GraphIteratorNode for running sub-graphs on item lists
- SearchLinkNode for link discovery with LLM relevance filtering
- ResearchGraph unified library entrypoint for direct, search-first, or depth-crawl extraction flows
- Live Hermes validation confirms the end-to-end web scraping pipeline works, though output quality remains below desired level with the lightweight validation model
- ConditionalNode for branching logic
- ReasoningNode for chain-of-thought extraction
- LLM interface for multi-provider support
- Prompt template system
- Infinite loop protection in graph executor
- DuckDuckGo web search
- Rod browser fetching (JS-rendered pages, Cloudflare bypass)
- UTLS HTTP fetching (static pages, stealth TLS)
- Shared loader factory for consistent fetch selection across graphs
- Auto fetch strategy (`FetchStrategy: auto`) with UTLS→Rod escalation on likely anti-bot blocks
- Centralized anti-bot block detection heuristics for common challenge/block pages
- HTML→text conversion
- OpenAI extraction (gpt-4o/gpt-4o-mini)
- HTTP API (scrape, search, depth-search, fetch, health)
- HTTP timing/structured logs for endpoint telemetry
- Graph/node timing logs for execution telemetry
- Real PDF extraction support
- Real DOCX extraction support
- Structured-data loaders for JSON/XML/CSV files and directories
- MarkdownifyGraph for direct HTML/page → Markdown conversion
- MarkdownifyNode for reusable HTML→Markdown state transformation
- SearchLinkGraph for fetch-first link discovery and relevance filtering
- SmartScraperLiteGraph for lighter-weight single-page extraction with reduced HTML/chunk defaults

**What's Missing:**
- 68% of the library
- 21 more node types
- 17 more graph types
- 9 model providers
- Telemetry
 - Additional document graph variants and richer document processing
- Better extraction quality tuning for large-site crawls, including testing with stronger models before judging final scrape quality
- Validation rerun of the new Hermes balanced crawl profile to measure whether broader link frontier improves result quality materially
- Proxy rotation, cookie/session persistence, and richer anti-bot escalation policies beyond the current UTLS→Rod fallback

## Next Priority Tasks

**Immediate (Week 3):**
1. Add more canonical graph types (screenshot and structured-data multi families)
2. Add broader graph smoke tests / examples coverage
3. Consider fixing the `examples/` package layout so `go test ./...` passes cleanly
4. Expand telemetry beyond timing into metrics/tracing aggregation
5. Rerun Hermes with the balanced crawl profile only as a quality-validation task, not as a product-scope driver

**Short-term (Week 4):**
1. Implement screenshot parity slice, then continue multi variants
2. Continue remaining parity-native graph/node work before new provider expansion

## Blockers
None — all infrastructure is in place.