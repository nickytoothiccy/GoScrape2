// Package chunking handles text splitting for LLM processing
package chunking

import (
	"strings"
	"unicode/utf8"
)

// Chunker splits text into token-sized chunks
type Chunker struct {
	chunkSize    int
	chunkOverlap int
}

// NewChunker creates a text chunker
func NewChunker(chunkSize, chunkOverlap int) *Chunker {
	if chunkSize == 0 {
		chunkSize = 8000
	}
	if chunkOverlap == 0 {
		chunkOverlap = 200
	}
	return &Chunker{
		chunkSize:    chunkSize,
		chunkOverlap: chunkOverlap,
	}
}

// Split divides text into overlapping chunks
func (c *Chunker) Split(text string) []string {
	// Approximate token count as chars / 4
	approxTokens := utf8.RuneCountInString(text) / 4

	// If small enough, return as single chunk
	if approxTokens <= c.chunkSize {
		return []string{text}
	}

	// Split by paragraphs
	paragraphs := strings.Split(text, "\n\n")

	var chunks []string
	var currentChunk strings.Builder
	currentSize := 0

	for _, para := range paragraphs {
		paraSize := utf8.RuneCountInString(para) / 4

		// If adding this paragraph exceeds chunk size, finalize current chunk
		if currentSize+paraSize > c.chunkSize && currentSize > 0 {
			chunks = append(chunks, currentChunk.String())

			// Start new chunk with overlap from previous
			currentChunk.Reset()
			currentSize = 0

			// Add overlap from end of previous chunk
			if len(chunks) > 0 {
				overlap := getOverlap(chunks[len(chunks)-1], c.chunkOverlap)
				currentChunk.WriteString(overlap)
				currentSize = utf8.RuneCountInString(overlap) / 4
			}
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n\n")
		}
		currentChunk.WriteString(para)
		currentSize += paraSize
	}

	// Add final chunk
	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}

func getOverlap(text string, overlapChars int) string {
	runes := []rune(text)
	if len(runes) <= overlapChars {
		return text
	}
	return string(runes[len(runes)-overlapChars:])
}
