package smartpromptopt

import (
	"strings"
	"unicode/utf8"
	
	"github.com/maximhq/bifrost/core/schemas"
)

func ChunkByRune(s string, chunkSize int) []string {
	if chunkSize <= 0 { return []string{s} }
	var chunks []string
	var b strings.Builder
	b.Grow(chunkSize)
	count := 0
	for _, r := range s {
		b.WriteRune(r)
		count++
		if count >= chunkSize {
			chunks = append(chunks, b.String())
			b.Reset()
			b.Grow(chunkSize)
			count = 0
		}
	}
	if b.Len() > 0 { chunks = append(chunks, b.String()) }
	return chunks
}

func Truncate(s string, max int) string {
	if max <= 0 { return "" }
	if utf8.RuneCountInString(s) <= max { return s }
	var b strings.Builder
	count := 0
	for _, r := range s {
		if count >= max { break }
		b.WriteRune(r)
		count++
	}
	return b.String() + "…"
}

// collectRequestText extracts all text content from a BifrostRequest
func collectRequestText(req *schemas.BifrostRequest) string {
	var parts []string
	
	// Collect text completion input
	if req.Input.TextCompletionInput != nil {
		parts = append(parts, *req.Input.TextCompletionInput)
	}
	
	// Collect chat messages
	if req.Input.ChatCompletionInput != nil {
		for _, m := range *req.Input.ChatCompletionInput {
			// Extract text from string content
			if m.Content.ContentStr != nil {
				parts = append(parts, *m.Content.ContentStr)
			}
			// Extract text from content blocks
			if m.Content.ContentBlocks != nil {
				for _, block := range *m.Content.ContentBlocks {
					if block.Text != nil {
						parts = append(parts, *block.Text)
					}
				}
			}
		}
	}
	
	return strings.Join(parts, "\n")
}
