package index

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

const fallbackChunkLines = 48

func chunkFile(path string, content string) []Chunk {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.TrimSuffix(content, "\n")
	if strings.TrimSpace(content) == "" {
		return nil
	}

	lines := strings.Split(content, "\n")
	chunks := make([]Chunk, 0, (len(lines)/fallbackChunkLines)+1)
	start := 0
	lastBoundary := -1

	for current := 0; current < len(lines); current++ {
		if strings.TrimSpace(lines[current]) == "" {
			lastBoundary = current + 1
		}
		if current-start+1 < fallbackChunkLines {
			continue
		}

		end := current + 1
		if lastBoundary > start+8 {
			end = lastBoundary
		}
		chunks = appendChunk(chunks, path, lines, start, end)
		start = end
		lastBoundary = -1
	}

	return appendChunk(chunks, path, lines, start, len(lines))
}

func appendChunk(chunks []Chunk, path string, lines []string, start, end int) []Chunk {
	for start < end && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	if start >= end {
		return chunks
	}

	text := strings.Join(lines[start:end], "\n")
	startLine := start + 1
	endLine := end
	return append(chunks, Chunk{
		ID:              chunkID(path, startLine, endLine, text),
		Path:            path,
		StartLine:       startLine,
		EndLine:         endLine,
		EstimatedTokens: estimateTokens(text),
		Text:            text,
	})
}

func chunkID(path string, startLine, endLine int, text string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s\x00%d\x00%d\x00%s", path, startLine, endLine, text)))
	return hex.EncodeToString(sum[:16])
}

func estimateTokens(text string) int {
	if text == "" {
		return 0
	}
	return max(1, (len(text)+2)/3)
}
