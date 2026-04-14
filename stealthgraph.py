"""
StealthGraph Python Client
Thin wrapper around the Go stealth scraping service.
Go handles: stealth fetch (utls) + HTML cleaning + OpenAI extraction.

Usage:
    from stealthgraph import StealthGraph
    sg = StealthGraph()
    result = sg.scrape("https://example.com", "Extract all product names and prices")
"""

import json
from dataclasses import dataclass

import httpx


@dataclass
class StealthGraphConfig:
    base_url: str = "http://localhost:8899"
    timeout: float = 120.0


class StealthGraph:
    def __init__(self, base_url: str = "http://localhost:8899", timeout: float = 120.0):
        self.base_url = base_url.rstrip("/")
        self.client = httpx.Client(timeout=timeout)

    # ---- Full pipeline: fetch + extract ----

    def scrape(
        self,
        url: str,
        prompt: str,
        model: str = "",
        schema_hint: str = "",
        structured: bool = False,
        profile: str = "random",
        proxy_url: str = "",
        timeout_ms: int = 30000,
        headers: dict = None,
    ) -> dict:
        """One-shot: stealth fetch + LLM extract."""
        resp = self.client.post(
            f"{self.base_url}/scrape",
            json={
                "url": url,
                "prompt": prompt,
                "model": model or None,
                "schema_hint": schema_hint or None,
                "structured": structured,
                "profile": profile,
                "proxy_url": proxy_url,
                "timeout_ms": timeout_ms,
                "headers": headers or {},
            },
        )
        resp.raise_for_status()
        return resp.json()

    # ---- Fetch only (get raw HTML) ----

    def fetch(
        self,
        url: str,
        profile: str = "random",
        proxy_url: str = "",
        timeout_ms: int = 30000,
        headers: dict = None,
    ) -> dict:
        resp = self.client.post(
            f"{self.base_url}/fetch",
            json={
                "url": url,
                "profile": profile,
                "proxy_url": proxy_url,
                "timeout_ms": timeout_ms,
                "headers": headers or {},
            },
        )
        resp.raise_for_status()
        return resp.json()

    # ---- Extract only (from pre-fetched HTML) ----

    def extract(
        self,
        html: str,
        prompt: str,
        model: str = "",
        schema_hint: str = "",
        structured: bool = False,
    ) -> dict:
        resp = self.client.post(
            f"{self.base_url}/extract",
            json={
                "html": html,
                "prompt": prompt,
                "model": model or None,
                "schema_hint": schema_hint or None,
                "structured": structured,
            },
        )
        resp.raise_for_status()
        return resp.json()

    # ---- Batch ----

    def scrape_many(self, urls: list[str], prompt: str, **kwargs) -> list[dict]:
        return [self.scrape(url, prompt, **kwargs) for url in urls]

    # ---- Health check ----

    def health(self) -> dict:
        return self.client.get(f"{self.base_url}/health").json()


# ---- CLI ----

if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(description="StealthGraph CLI")
    parser.add_argument("url", help="URL to scrape")
    parser.add_argument("prompt", help="What to extract")
    parser.add_argument("--model", default="")
    parser.add_argument("--profile", default="random")
    parser.add_argument("--proxy", default="")
    parser.add_argument("--structured", action="store_true")
    parser.add_argument("--base-url", default="http://localhost:8899")
    args = parser.parse_args()

    sg = StealthGraph(base_url=args.base_url)
    result = sg.scrape(
        args.url,
        args.prompt,
        model=args.model,
        profile=args.profile,
        proxy_url=args.proxy,
        structured=args.structured,
    )
    print(json.dumps(result, indent=2))
