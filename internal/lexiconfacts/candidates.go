package lexiconfacts

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/Lokee86/grimoire/internal/evidence"
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
			context := candidateContext(entry, chunk)
			candidate := retrieve.Candidate{
				Chunk: chunk, Score: entry.score, Source: source,
				Reasons: append([]string(nil), entry.reasons...),
				Context: &context,
			}
			key := chunk.ID
			if key == "" {
				key = fmt.Sprintf("%s:%d:%d", chunk.Path, chunk.StartLine, chunk.EndLine)
			}
			if existing, exists := byChunk[key]; exists {
				if existing.Score >= candidate.Score {
					existing.Reasons = uniqueStrings(append(existing.Reasons, candidate.Reasons...))
					existing.Context = mergeContext(existing.Context, candidate.Context)
					byChunk[key] = existing
					continue
				}
				candidate.Reasons = uniqueStrings(append(candidate.Reasons, existing.Reasons...))
				candidate.Context = mergeContext(candidate.Context, existing.Context)
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

func candidateContext(entry scoredNode, chunk index.Chunk) evidence.Descriptor {
	identity := sourceRangeIdentity(entry.node)
	if identity == "" {
		identity = evidence.RangeIdentity(chunk.Path, chunk.StartLine, chunk.EndLine)
	}
	role := evidence.RoleSupporting
	if entry.primary {
		role = evidence.RolePrimary
	}
	return evidence.Descriptor{
		Identity:        identity,
		Roles:           []evidence.Role{role},
		GroupIDs:        []string{nodeGroupID(entry.node)},
		EstimatedTokens: max(chunk.TokenCount, 1),
		RedundancyKey:   nodeRedundancyKey(entry.node, chunk),
	}
}

func mergeContext(left, right *evidence.Descriptor) *evidence.Descriptor {
	if left == nil && right == nil {
		return nil
	}
	var leftValue, rightValue evidence.Descriptor
	if left != nil {
		leftValue = *left
	}
	if right != nil {
		rightValue = *right
	}
	merged := evidence.Merge(leftValue, rightValue)
	return &merged
}

func sourceRangeIdentity(node Node) string {
	if node.Span == nil || node.Span.Path == "" || node.Span.StartLine <= 0 {
		return ""
	}
	end := node.Span.EndLine
	if end < node.Span.StartLine {
		end = node.Span.StartLine
	}
	return evidence.RangeIdentity(node.Span.Path, node.Span.StartLine, end)
}

func nodeGroupID(node Node) string {
	if identity := sourceRangeIdentity(node); identity != "" {
		return evidence.StableID("lexicon-node", identity)
	}
	if node.ID != "" {
		return evidence.StableID("lexicon-node", "node:"+node.ID)
	}
	return evidence.StableID("lexicon-node", nodePath(node), node.Name, node.QualifiedName)
}

func nodeRedundancyKey(node Node, chunk index.Chunk) string {
	path := filepath.ToSlash(nodePath(node))
	if path == "" {
		path = filepath.ToSlash(chunk.Path)
	}
	name := node.QualifiedName
	if name == "" {
		name = node.Name
	}
	if name == "" {
		name = node.ID
	}
	if name == "" {
		return path
	}
	return path + "::" + name
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
