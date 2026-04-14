# System Patterns

## ScrapeGraphAI Architecture Analysis

### Graph Hierarchy
```
AbstractGraph (base config, model setup)
    ↓
BaseGraph (execution engine, node orchestration)
    ↓
Specific Graphs (26 types)
```

### Complete Graph Inventory

**Single-Page Extraction (6 graphs):**
1. ✅ `SmartScraperGraph` - basic URL→data extraction
2. `SmartScraperLiteGraph` - lightweight version
3. `DocumentScraperGraph` - PDF/DOCX extraction
4. `ScreenshotScraperGraph` - visual scraping
5. `JSONScraperGraph` - JSON data extraction
6. `XMLScraperGraph` - XML data extraction

**Multi-Page Extraction (7 graphs):**
7. ✅ `SmartScraperMultiGraph` - multiple URLs, separate results
8. `SmartScraperMultiConcatGraph` - multiple URLs, merged result
9. `SmartScraperMultiLiteGraph` - lightweight multi-page
10. `DocumentScraperMultiGraph` - batch document processing
11. `JSONScraperMultiGraph` - batch JSON processing
12. `XMLScraperMultiGraph` - batch XML processing
13. `CSVScraperMultiGraph` - CSV data processing

**Search & Discovery (4 graphs):**
14. `SearchGraph` - search engines → extract results
15. `SearchLinkGraph` - find links matching criteria
16. `OmniSearchGraph` - multi-source search
17. `DepthSearchGraph` - recursive link following

**Code Generation (4 graphs):**
18. `ScriptCreatorGraph` - generate scraping scripts
19. `ScriptCreatorMultiGraph` - batch script generation
20. `CodeGeneratorGraph` - generate extraction code
21. `OmniScraperGraph` - intelligent scraper generation

**Specialized (5 graphs):**
22. `CSVScraperGraph` - CSV to structured data
23. `SpeechGraph` - text-to-speech output
24. `MarkdownifyGraph` - HTML→Markdown conversion
25. `SearchNodeWithContext` - contextual search
26. `GraphIteratorNode` - graph composition

### Complete Node Inventory

**Fetching Nodes (4):**
1. ✅ `FetchNode` - basic HTTP/browser fetch
2. `FetchNodeLevelK` - depth-aware fetching
3. `FetchScreenNode` - screenshot capture
4. `RobotsNode` - robots.txt handling

**Parsing Nodes (4):**
5. ✅ `ParseNode` - HTML→text + chunking
6. `ParseNodeDepthKNode` - depth-aware parsing
7. `MarkdownifyNode` - HTML→Markdown
8. `HTMLAnalyzerNode` - structure analysis

**Generation Nodes (8):**
9. ✅ `GenerateAnswerNode` - LLM extraction
10. `GenerateAnswerNodeKLevel` - depth-aware extraction
11. `GenerateAnswerCSVNode` - CSV-specific extraction
12. `GenerateAnswerFromImageNode` - image→data
13. `GenerateAnswerOmniNode` - multi-modal extraction
14. `GenerateCodeNode` - code generation
15. `GenerateScraperNode` - scraper generation
16. `PromptRefinerNode` - prompt optimization

**Processing Nodes (6):**
17. `ConditionalNode` - if/else branching
18. `ReasoningNode` - chain-of-thought
19. `MergeAnswersNode` - result merging
20. `ConcatAnswersNode` - result concatenation
21. `MergeGeneratedScriptsNode` - script merging
22. `GraphIteratorNode` - sub-graph execution

**Search Nodes (4):**
23. `SearchInternetNode` - search engine queries
24. `SearchLinkNode` - link discovery
25. `SearchNodeWithContext` - contextual search
26. `GetProbableTagsNode` - tag extraction

**Specialized Nodes (3):**
27. `ImageToTextNode` - OCR/image analysis
28. `TextToSpeechNode` - speech synthesis
29. `RAGNode` - retrieval augmented generation
30. `DescriptionNode` - metadata extraction

## Go Package Structure

```
GoScrape2/
├── internal/
│   ├── models/         # Shared types
│   ├── config/         # Configuration
│   └── errors/         # Error types
├── pkg/
│   ├── graph/          # Core engine
│   │   ├── abstract.go     # AbstractGraph base
│   │   ├── base.go         # BaseGraph executor
│   │   ├── state.go        # State management
│   │   └── node.go         # Node interface
│   ├── graphs/         # 26 graph implementations
│   │   ├── smart_scraper.go
│   │   ├── search.go
│   │   ├── depth_search.go
│   │   ├── document_scraper.go
│   │   ├── json_scraper.go
│   │   └── ... (21 more)
│   ├── nodes/          # 30+ node implementations
│   │   ├── fetch.go
│   │   ├── parse.go
│   │   ├── generate.go
│   │   ├── conditional.go
│   │   ├── reasoning.go
│   │   └── ... (25 more)
│   ├── models/         # Model providers
│   │   ├── interface.go
│   │   ├── openai.go
│   │   ├── anthropic.go
│   │   ├── huggingface.go
│   │   └── ... (7 more)
│   ├── loaders/        # Fetch backends
│   │   ├── http.go
│   │   ├── browser.go      # Playwright/Rod
│   │   ├── document.go     # PDF/DOCX
│   │   └── screenshot.go
│   ├── prompts/        # Prompt templates
│   │   ├── templates.go
│   │   ├── extraction.go
│   │   ├── reasoning.go
│   │   └── generation.go
│   ├── telemetry/      # Observability
│   │   ├── metrics.go
│   │   ├── logging.go
│   │   └── tracing.go
│   ├── chunking/       # Text splitting ✅
│   ├── markdown/       # HTML conversion ✅
│   └── utils/          # Shared utilities
├── cmd/
│   └── server/         # HTTP API ✅
└── examples/           # Usage examples
```

## State Flow Pattern

Every graph follows this pattern:
```
1. Initialize state with inputs
2. Execute nodes in sequence
3. Each node:
   - Reads from state (inputs)
   - Processes
   - Writes to state (outputs)
4. Final node outputs to user
```

State keys are typed and validated.

## Multi-Graph Pattern

`SmartScraperMultiGraph` uses a composition-first pattern:

```go
urls -> GraphIteratorNode -> []json.RawMessage answers
                           -> ConcatAnswersNode (optional)
```

- `GraphIteratorNode` owns per-item timeout and partial-failure handling
- `SmartScraperGraph` remains the reusable single-page primitive
- `ConcatAnswersNode` provides deterministic merging without an LLM call
- Failed URLs are tracked separately for caller inspection

## Prompt System Pattern

ScrapeGraphAI uses template-based prompts:
```python
TEMPLATE_EXTRACTION = """
You are extracting {data_type} from a webpage.
Context: {context}
Schema: {schema}
Return JSON only.
"""
```

Go implementation:
```go
type PromptTemplate struct {
    Name     string
    Template string
    Vars     []string
}

func (t *PromptTemplate) Render(vars map[string]string) string
```

## Model Abstraction Pattern

All models implement a common interface:
```go
type LLM interface {
    Generate(ctx context.Context, prompt string, config GenerationConfig) (string, error)
    GenerateJSON(ctx context.Context, prompt string, schema string) (json.RawMessage, error)
    Tokenize(text string) []int
    CountTokens(text string) int
}
```

Then specific implementations:
- OpenAIModel
- AnthropicModel
- HuggingFaceModel
- GroqModel
- GeminiModel
- etc.