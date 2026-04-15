---
name: hermes-goscrape2
description: Hermes-internal stealth scraping and research skill backed by the GoScrape2 engine. Runs locally via Go, uses Hermes-owned env/config from HERMES_HOME, and supports fetch, scrape, multi-scrape, depth-search, document-scrape, and research workflows without a sidecar service.
version: 0.1.0
author: local
license: MIT
metadata:
  hermes:
    tags: [Research, Scraping, Stealth, Documents, Crawling, Go]
prerequisites:
  commands: [bash, go]
---

# Hermes GoScrape2

This skill runs `GoScrape2` as an internal Hermes skill helper, not as a daemon or extra container.

## Runtime Model

- Hermes invokes the local wrapper script.
- The wrapper builds or reuses a local Go binary.
- The Go binary reads secrets from `HERMES_HOME/.env`.
- No standalone HTTP server is required.

## Hermes Secret Flow

This skill follows Hermes' existing mounted-home secret pattern:

- `HERMES_HOME/.env`
- `HERMES_HOME/config.yaml`

The current CLI reads:

- `HERMES_GOSCRAPE_API_KEY`
- `HERMES_GOSCRAPE_BASE_URL`
- `OPENAI_API_KEY`
- `OPENAI_BASE_URL`
- `HERMES_GOSCRAPE_MODEL`
- falls back to `model.default` from `config.yaml`

Existing environment variables in the current process still win over `.env`.

## Commands

```bash
"$HERMES_HOME/skills/hermes-goscrape2/scripts/hermes_goscrape2.sh" fetch --url https://example.com
"$HERMES_HOME/skills/hermes-goscrape2/scripts/hermes_goscrape2.sh" scrape --url https://example.com --prompt "Extract the main headline as JSON"
"$HERMES_HOME/skills/hermes-goscrape2/scripts/hermes_goscrape2.sh" search --prompt "Find the best official docs for Hermes Agent installation"
"$HERMES_HOME/skills/hermes-goscrape2/scripts/hermes_goscrape2.sh" multi-scrape --urls https://a.com,https://b.com --prompt "Extract title and summary"
"$HERMES_HOME/skills/hermes-goscrape2/scripts/hermes_goscrape2.sh" depth-search --url https://example.com/docs --prompt "Collect all setup steps" --max-depth 2
"$HERMES_HOME/skills/hermes-goscrape2/scripts/hermes_goscrape2.sh" document-scrape --path /workspace/report.pdf --prompt "Extract action items"
"$HERMES_HOME/skills/hermes-goscrape2/scripts/hermes_goscrape2.sh" research --prompt "Research Hermes Agent setup" --search-first
```

## Notes

- `fetch` uses UTLS by default and Rod when `--headless` is supplied.
- `search` currently uses the non-premium search layer already present in GoScrape2.
- `research` chooses direct, depth, or search-led orchestration based on flags.
- Output is JSON on stdout so Hermes can consume it directly.
