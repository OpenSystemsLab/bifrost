package smartpromptopt

import (
	"strings"
	"unicode/utf8"
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
