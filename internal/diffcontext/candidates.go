package diffcontext

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/structure"
)

const (
	gitDiffSource   = "git-diff"
	gitDiffPriority = 1_000_000

	DefaultCandidateLimit = 64
	DefaultEvidenceLimit  = 64
)

// Candidates returns prepared-index chunks that overlap current-file changes.
// The optional limit is applied after deterministic ordering; an omitted limit
// retains every matching chunk.
func Candidates(snapshot index.Snapshot, changes []Change, limits ...int) []retrieve.Candidate {
	if len(limits) > 0 && limits[0] <= 0 {
		return nil
	}
	limit := 0
	if len(limits) > 0 {
		limit = limits[0]
	}

	normalized := normalizedChanges(changes)
	byPath := make(map[string][]index.Chunk, len(snapshot.Files))
	for _, file := range snapshot.Files {
		path := normalizePath(file.Path)
		byPath[path] = append(byPath[path], file.Chunks...)
	}

	type matched struct {
		candidate retrieve.Candidate
		changes   []Change
	}
	byChunk := make(map[string]*matched)
	for _, change := range normalized {
		if change.Deleted {
			continue
		}
		for _, chunk := range byPath[change.Path] {
			if normalizePath(chunk.Path) != change.Path {
				continue
			}
			if !overlaps(chunk.StartLine, chunk.EndLine, change.StartLine, change.EndLine) {
				continue
			}
			key := chunkKey(chunk)
			entry := byChunk[key]
			if entry == nil {
				chunk.Path = normalizePath(chunk.Path)
				candidate := retrieve.Candidate{Chunk: chunk, Source: gitDiffSource}
				entry = &matched{candidate: candidate}
				byChunk[key] = entry
			}
			entry.changes = append(entry.changes, change)
		}
	}

	result := make([]retrieve.Candidate, 0, len(byChunk))
	for _, entry := range byChunk {
		candidate := entry.candidate
		candidate.Reasons = make([]string, 0, len(entry.changes)+1)
		for _, change := range entry.changes {
			candidate.Reasons = append(candidate.Reasons,
				fmt.Sprintf("overlaps git diff %s:%d-%d", change.Path, change.StartLine, change.EndLine))
			if change.OldPath != "" {
				candidate.Reasons = append(candidate.Reasons,
					fmt.Sprintf("renamed from %s", change.OldPath))
			}
		}
		candidate.Reasons = uniqueReasons(candidate.Reasons)
		context := evidence.Descriptor{
			Identity:        evidence.RangeIdentity(candidate.Chunk.Path, candidate.Chunk.StartLine, candidate.Chunk.EndLine),
			Roles:           []evidence.Role{evidence.RolePrimary},
			EstimatedTokens: max(candidate.Chunk.TokenCount, 1),
			RedundancyKey:   normalizePath(candidate.Chunk.Path),
		}
		candidate.Context = &context
		result = append(result, candidate)
	}

	sort.Slice(result, func(left, right int) bool {
		a, b := result[left], result[right]
		if normalizePath(a.Chunk.Path) != normalizePath(b.Chunk.Path) {
			return normalizePath(a.Chunk.Path) < normalizePath(b.Chunk.Path)
		}
		if a.Chunk.StartLine != b.Chunk.StartLine {
			return a.Chunk.StartLine < b.Chunk.StartLine
		}
		if a.Chunk.EndLine != b.Chunk.EndLine {
			return a.Chunk.EndLine < b.Chunk.EndLine
		}
		return chunkKey(a.Chunk) < chunkKey(b.Chunk)
	})
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	for rank := range result {
		result[rank].Rank = rank + 1
		result[rank].Score = float64(gitDiffPriority - rank)
	}
	return result
}

// SourceCandidates is a descriptive alias for Candidates.
func SourceCandidates(snapshot index.Snapshot, changes []Change, limits ...int) []retrieve.Candidate {
	return Candidates(snapshot, changes, limits...)
}

// Evidence returns bounded structural change evidence, one item per distinct
// changed file/span. Its provider and kind make the git origin explicit.
func Evidence(changes []Change, limits ...int) []structure.Evidence {
	limit := DefaultEvidenceLimit
	if len(limits) > 0 {
		if limits[0] <= 0 {
			return nil
		}
		limit = limits[0]
	}

	normalized := normalizedChanges(changes)
	result := make([]structure.Evidence, 0, min(limit, len(normalized)))
	for _, change := range normalized {
		if len(result) >= limit {
			break
		}
		node := structure.Node{
			Kind: "change",
			Path: change.Path,
			Span: &structure.Span{
				Path: change.Path, StartLine: change.StartLine, EndLine: change.EndLine,
			},
		}
		reasons := []string{fmt.Sprintf("git diff changed %s:%d-%d", change.Path, change.StartLine, change.EndLine)}
		if change.Deleted {
			reasons = append(reasons, "file content deleted in git diff")
		}
		if change.OldPath != "" {
			reasons = append(reasons, "renamed from "+change.OldPath)
		}
		context := evidence.Descriptor{
			Identity: evidence.RangeIdentity(change.Path, change.StartLine, change.EndLine),
			Roles:    []evidence.Role{evidence.RoleStructural},
		}
		result = append(result, structure.Evidence{
			Provider: gitDiffSource,
			Kind:     "change",
			Rank:     len(result) + 1,
			Score:    float64(gitDiffPriority - len(result)),
			Reasons:  reasons,
			Node:     &node,
			Summary:  change.Summary,
			Context:  &context,
		})
	}
	return result
}

// ChangeEvidence is a descriptive alias for Evidence.
func ChangeEvidence(changes []Change, limits ...int) []structure.Evidence {
	return Evidence(changes, limits...)
}

func normalizedChanges(changes []Change) []Change {
	result := make([]Change, 0, len(changes))
	for _, change := range changes {
		result = append(result, normalizeChange(change))
	}
	return deduplicateChanges(result)
}

func overlaps(leftStart, leftEnd, rightStart, rightEnd int) bool {
	if leftEnd < leftStart {
		leftEnd = leftStart
	}
	if rightEnd < rightStart {
		rightEnd = rightStart
	}
	return leftStart <= rightEnd && rightStart <= leftEnd
}

func chunkKey(chunk index.Chunk) string {
	if chunk.ID != "" {
		return "id:" + chunk.ID
	}
	return "range:" + normalizePath(chunk.Path) + ":" +
		strconv.Itoa(chunk.StartLine) + ":" + strconv.Itoa(chunk.EndLine) + "\x00" + chunk.Text
}

func uniqueReasons(reasons []string) []string {
	seen := make(map[string]struct{}, len(reasons))
	result := make([]string, 0, len(reasons))
	for _, reason := range reasons {
		if strings.TrimSpace(reason) == "" {
			continue
		}
		if _, exists := seen[reason]; exists {
			continue
		}
		seen[reason] = struct{}{}
		result = append(result, reason)
	}
	return result
}

func min(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func max(left, right int) int {
	if left > right {
		return left
	}
	return right
}
