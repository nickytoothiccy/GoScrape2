# Project Brief

## Project Name
GoScrape2 — Complete Go Port of ScrapeGraphAI

## Purpose
A complete, production-grade Go reimplementation of the Python ScrapeGraphAI library. This is NOT a simplified clone or "80%" implementation - this is a full 1:1 functional port with Go's type safety, concurrency, and performance benefits.

## Core Goal
Reverse-engineer and reimplement 100% of ScrapeGraphAI's functionality in Go:
- All 26 graph types
- All 27+ node types  
- Full prompt system
- Complete model integration layer (OpenAI, Anthropic, HuggingFace, Groq, etc.)
- Telemetry and observability
- Document loaders (PDF, DOCX, etc.)
- Browser integration (Playwright, Selenium, Rod)

## Architecture Philosophy
ScrapeGraphAI's architecture:
```
AbstractGraph → BaseGraph → Specific Graphs (SmartScraperGraph, SearchGraph, etc.)
    ↓
  Nodes (FetchNode, ParseNode, GenerateAnswerNode, etc.)
    ↓
  Models (OpenAI, Anthropic, HuggingFace, etc.)
    ↓
  Loaders (Browser, HTTP, Document)
```

Our Go architecture:
```
pkg/
├── graph/          # Core engine (AbstractGraph, BaseGraph)
├── graphs/         # All 26 graph implementations
├── nodes/          # All 27+ node implementations
├── models/         # Model abstraction layer
├── loaders/        # Fetch backends
├── prompts/        # Prompt template system
├── telemetry/      # Observability
└── utils/          # Shared utilities
```

## Success Criteria
✅ Phase 1: Core Infrastructure (DONE)
- Graph execution engine
- State management
- Node interface
- Basic SmartScraperGraph

⬜ Phase 2: Complete Node Library (0/27)
- All extraction nodes
- All processing nodes
- Conditional/reasoning nodes

⬜ Phase 3: All Graph Types (1/26)
- SmartScraperGraph ✅
- SearchGraph
- DepthSearchGraph
- Document graphs
- JSON/XML/CSV graphs
- etc.

⬜ Phase 4: Multi-Model Support (0/10+)
- OpenAI ✅ (partial)
- Anthropic
- HuggingFace
- Groq
- Gemini
- DeepSeek
- etc.

⬜ Phase 5: Advanced Features
- Prompt templating system
- Telemetry
- Error handling & retry
- Document loaders
- Screenshot support
- Speech synthesis

## Constraints
- Windows 11 environment
- Must maintain <200 lines per file
- Must preserve type safety
- Must support both library and HTTP API usage
- Working directory: `c:\Users\Nicky\GoScrape2`

## Current State
**Implemented:** ~4% of full ScrapeGraphAI
- Basic graph engine
- 3 nodes (Fetch, Parse, GenerateAnswer)
- 1 graph (SmartScraperGraph)
- 2 loaders (Local, UTLS)
- 1 model (OpenAI - partial)

**Remaining:** 96% of functionality
- 24 more node types
- 25 more graph types
- 9+ more model integrations
- Prompt system
- Telemetry
- Document loaders
- etc.