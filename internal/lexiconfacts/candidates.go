package lexiconfacts

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

func chunksForNodes(snapshot index.Snapshot, nodes map[string]scoredNode, limit int) []retrieve.Candidate {
	byPath := make(map[string][]index.Chunk, len(snapshot.Files))
	for _, file := range snapshot.Files {
		byPath[filepath.ToSlash(file.Path)] = file.Chunks
	}
	byChunk := make(map[string]retrieve.Candidate)
	for _, entry := range nodes {
		path := filepath.ToSlash(nodePath(entry.node))
		chunks := byPath[path]
		if len(chunks) == 0 {
			continue
		}
		matched := overlappingChunks(chunks, entry.node.Span)
		for _, chunk := range matched {
			candidate := retrieve.Candidate{
				Chunk: chunk, Score: entry.score, Source: source,
				Reasons: append([]string(nil), entry.reasons...),
			}
			key := chunk.ID
			if key == "" {
				key = fmt.Sprintf("%s:%d:%d", chunk.Path, chunk.StartLine, chunk.EndLine)
			}
			if existing, exists := byChunk[key]; exists {
				if existing.Score >= candidate.Score {
					existing.Reasons = uniqueStrings(append(existing.Reasons, candidate.Reasons...))
					byChunk[key] = existing
					continue
				}
				candidate.Reasons = uniqueStrings(append(candidate.Reasons, existing.Reasons...))
			}
			byChunk[key] = candidate
		}
	}
	result := make([]retrieve.Candidate, 0, len(byChunk))
	for _, candidate := range byChunk {
		result = append(result, candidate)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Score != result[j].Score {
			return result[i].Score > result[j].Score
		}
		if result[i].Chunk.Path != result[j].Chunk.Path {
			return result[i].Chunk.Path < result[j].Chunk.Path
		}
		return result[i].Chunk.StartLine < result[j].Chunk.StartLine
	})
	if len(result) > limit {
		result = result[:limit]
	}
	for index := range result {
		result[index].Rank = index + 1
	}
	return result
}

func overlappingChunks(chunks []index.Chunk, span *Span) []index.Chunk {
	if span == nil || span.StartLine <= 0 {
		return chunks[:1]
	}
	end := span.EndLine
	if end < span.StartLine {
		end = span.StartLine
	}
	result := make([]index.Chunk, 0, 2)
	for _, chunk := range chunks {
		if chunk.EndLine < span.StartLine || chunk.StartLine > end {
			continue
		}
		result = append(result, chunk)
	}
	if len(result) == 0 {
		return chunks[:1]
	}
	return result
}
