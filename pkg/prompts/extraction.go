// Package prompts provides extraction prompt templates
package prompts

// extractionTemplates contains all extraction-related prompt templates
var extractionTemplates = []*Template{
	{
		Name: "extraction_system",
		Template: `You are a precise data extraction engine. Given content from a webpage and a user's extraction prompt, extract the requested information and return it as valid JSON.

Rules:
- Return ONLY valid JSON, no markdown fences, no explanation
- If the requested data is not found, return {"error": "not_found", "reason": "..."}
- Use arrays for lists of items
- Use descriptive key names
- Preserve URLs as absolute paths when possible`,
		Vars: nil,
	},
	{
		Name: "extraction_user",
		Template: `## Extraction Prompt
{prompt}

## Page Content
{content}`,
		Vars: []string{"prompt", "content"},
	},
	{
		Name: "extraction_with_schema",
		Template: `## Extraction Prompt
{prompt}

## Page Content
{content}

## Expected Schema
{schema}`,
		Vars: []string{"prompt", "content", "schema"},
	},
	{
		Name:     "merge_system",
		Template: `You are a data merging engine. Combine multiple extraction results into a single, deduplicated, coherent JSON result. Preserve all unique data points and resolve conflicts by keeping the most complete version.`,
		Vars:     nil,
	},
	{
		Name: "merge_user",
		Template: `## Task
Merge the following extraction results into a single coherent result.

## Original Prompt
{prompt}

## Chunk Results
{chunks}`,
		Vars: []string{"prompt", "chunks"},
	},
	{
		Name:     "validation_system",
		Template: `You are a data validation engine. Check if the extracted data matches the expected format and contains meaningful content. Return JSON with "valid": true/false and "reason" if invalid.`,
		Vars:     nil,
	},
	{
		Name: "validation_user",
		Template: `## Original Prompt
{prompt}

## Extracted Data
{data}

## Validation Rules
- Data should contain the information requested in the prompt
- Data should not be empty or contain only error messages
- Data format should be consistent and well-structured

Return {"valid": true} or {"valid": false, "reason": "..."}`,
		Vars: []string{"prompt", "data"},
	},
}
