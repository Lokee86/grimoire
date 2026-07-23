package index

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/Lokee86/grimoire/internal/tokenizer"
)

const (
	fallbackChunkLines    = 48
	defaultMaxChunkTokens = 1536
)

func chunkFile(path string, content string) ([]Chunk, error) {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.TrimSuffix(content, "\n")
	if strings.TrimSpace(content) == "" {
		return nil, nil
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
		var err error
		chunks, err = appendChunk(chunks, path, lines, start, end)
		if err != nil {
			return nil, err
		}
		start = end
		lastBoundary = -1
	}

	return appendChunk(chunks, path, lines, start, len(lines))
}

func appendChunk(chunks []Chunk, path string, lines []string, start, end int) ([]Chunk, error) {
	for start < end && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	if start >= end {
		return chunks, nil
	}

	text := strings.Join(lines[start:end], "\n")
	tokenCount, err := tokenizer.Count(text)
	if err != nil {
		return nil, err
	}
	if tokenCount <= defaultMaxChunkTokens {
		return appendPreparedChunk(chunks, path, start+1, end, text, tokenCount, -1), nil
	}
	if end-start == 1 {
		return appendTokenSlices(chunks, path, start+1, text)
	}

	split, err := largestLinePrefix(lines, start, end)
	if err != nil {
		return nil, err
	}
	if split == start {
		chunks, err = appendTokenSlices(chunks, path, start+1, lines[start])
		if err != nil {
			return nil, err
		}
		return appendChunk(chunks, path, lines, start+1, end)
	}
	chunks, err = appendChunk(chunks, path, lines, start, split)
	if err != nil {
		return nil, err
	}
	return appendChunk(chunks, path, lines, split, end)
}

func largestLinePrefix(lines []string, start, end int) (int, error) {
	low, high := start+1, end-1
	best := start
	for low <= high {
		middle := low + (high-low)/2
		count, err := tokenizer.Count(strings.Join(lines[start:middle], "\n"))
		if err != nil {
			return 0, err
		}
		if count <= defaultMaxChunkTokens {
			best = middle
			low = middle + 1
		} else {
			high = middle - 1
		}
	}
	return best, nil
}

func appendTokenSlices(chunks []Chunk, path string, line int, text string) ([]Chunk, error) {
	tokens, err := tokenizer.Encode(text)
	if err != nil {
		return nil, err
	}
	for start := 0; start < len(tokens); start += defaultMaxChunkTokens {
		end := min(start+defaultMaxChunkTokens, len(tokens))
		part, err := tokenizer.Decode(tokens[start:end])
		if err != nil {
			return nil, err
		}
		count, err := tokenizer.Count(part)
		if err != nil {
			return nil, err
		}
		chunks = appendPreparedChunk(chunks, path, line, line, part, count, start/defaultMaxChunkTokens)
	}
	return chunks, nil
}

func appendPreparedChunk(chunks []Chunk, path string, startLine, endLine int, text string, tokenCount, part int) []Chunk {
	return append(chunks, Chunk{
		ID:         chunkID(path, startLine, endLine, text, part),
		Path:       path,
		StartLine:  startLine,
		EndLine:    endLine,
		TokenCount: tokenCount,
		Text:       text,
	})
}

func chunkID(path string, startLine, endLine int, text string, part int) string {
	identity := fmt.Sprintf("%s\x00%d\x00%d\x00%s", path, startLine, endLine, text)
	if part >= 0 {
		identity += fmt.Sprintf("\x00token-part:%d", part)
	}
	sum := sha256.Sum256([]byte(identity))
	return hex.EncodeToString(sum[:16])
}
