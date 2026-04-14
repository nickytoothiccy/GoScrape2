// Package prompts provides reasoning prompt templates for chain-of-thought
package prompts

// reasoningTemplates contains all reasoning/chain-of-thought prompt templates
var reasoningTemplates = []*Template{
	{
		Name: "reasoning_system",
		Template: `You are an expert web content analyst. Your job is to analyze webpage content and reason about the best approach to extract the requested information.

Think step by step:
1. Identify what data is being requested
2. Scan the content for relevant sections
3. Determine the structure of the data
4. Note any ambiguities or edge cases
5. Provide a clear analysis that will guide extraction

Return your analysis as JSON with the following structure:
{
  "relevant_sections": ["section descriptions"],
  "data_structure": "description of how data is organized",
  "extraction_strategy": "recommended approach",
  "confidence": "high/medium/low",
  "notes": ["any important observations"]
}`,
		Vars: nil,
	},
	{
		Name: "reasoning_user",
		Template: `## User's Request
{prompt}

## Page Content
{content}

Analyze this content and provide your reasoning about how to best extract the requested information.`,
		Vars: []string{"prompt", "content"},
	},
	{
		Name: "reasoning_guided_extraction_system",
		Template: `You are a precise data extraction engine enhanced with analytical context. Use the provided analysis to guide your extraction for maximum accuracy.

Rules:
- Return ONLY valid JSON, no markdown fences, no explanation
- Use the analysis to focus on the most relevant content sections
- If confidence is low, be conservative and note uncertainties
- Preserve data structure as identified in the analysis`,
		Vars: nil,
	},
	{
		Name: "reasoning_guided_extraction_user",
		Template: `## Extraction Prompt
{prompt}

## Pre-Analysis
{analysis}

## Page Content
{content}

Extract the requested data using the analysis as a guide.`,
		Vars: []string{"prompt", "analysis", "content"},
	},
	{
		Name: "search_query_system",
		Template: `You are a search query generator. Given a user's information need, generate effective search queries that will find relevant web pages.

Return JSON with:
{
  "queries": ["query1", "query2", "query3"],
  "strategy": "explanation of search approach"
}`,
		Vars: nil,
	},
	{
		Name: "search_query_user",
		Template: `## Information Need
{prompt}

Generate 3-5 effective search queries to find pages containing this information.`,
		Vars: []string{"prompt"},
	},
	{
		Name: "link_relevance_system",
		Template: `You are a link relevance evaluator. Given a list of links and a user's information need, score each link's relevance.

Return JSON array:
[{"url": "...", "score": 0.0-1.0, "reason": "..."}]`,
		Vars: nil,
	},
	{
		Name: "link_relevance_user",
		Template: `## Information Need
{prompt}

## Available Links
{links}

Score each link's relevance to the information need (0.0 = irrelevant, 1.0 = highly relevant).`,
		Vars: []string{"prompt", "links"},
	},
}
