package selection

import (
	"strconv"
	"strings"

	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

const (
	adjacentSource       = "adjacent"
	maxAdjacentPrimaries = 4
)

// Curate removes redundant candidates, applies stable soft diversity, and
// appends neighbors around the first four diversified primary candidates.
func Curate(snapshot index.Snapshot, candidates []retrieve.Candidate) []retrieve.Candidate {
	primaries := uniqueNonOverlapping(candidates)
	primaries = diversify(primaries)

	frontCount := min(maxAdjacentPrimaries, len(primaries))
	front := primaries[:frontCount]
	adjacent, promoted := adjacentCandidates(snapshot, front, primaries)
	curated := append([]retrieve.Candidate(nil), front...)
	curated = append(curated, adjacent...)
	for _, primary := range primaries[frontCount:] {
		if _, exists := promoted[candidateKey(primary)]; exists {
			continue
		}
		curated = append(curated, primary)
	}
	return curated
}

func uniqueNonOverlapping(candidates []retrieve.Candidate) []retrieve.Candidate {
	ordered := append([]retrieve.Candidate(nil), candidates...)

	seen := make(map[string]struct{}, len(ordered))
	kept := make([]retrieve.Candidate, 0, len(ordered))
	for _, candidate := range ordered {
		key := candidateKey(candidate)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		overlaps := false
		for _, prior := range kept {
			if overlappingChunks(prior.Chunk, candidate.Chunk) {
				overlaps = true
				break
			}
		}
		if overlaps {
			continue
		}
		kept = append(kept, candidate)
	}
	return kept
}

func diversify(candidates []retrieve.Candidate) []retrieve.Candidate {
	remaining := append([]retrieve.Candidate(nil), candidates...)
	ordered := make([]retrieve.Candidate, 0, len(remaining))
	fileCounts := make(map[string]int)
	subsystemCounts := make(map[string]int)

	for len(remaining) > 0 {
		best := 0
		bestPenalty := diversityPenalty(remaining[0], 0, fileCounts, subsystemCounts)
		for candidateIndex := 1; candidateIndex < len(remaining); candidateIndex++ {
			penalty := diversityPenalty(remaining[candidateIndex], candidateIndex, fileCounts, subsystemCounts)
			if penalty < bestPenalty {
				best, bestPenalty = candidateIndex, penalty
			}
		}
		candidate := remaining[best]
		remaining = append(remaining[:best], remaining[best+1:]...)
		ordered = append(ordered, candidate)
		fileCounts[fileKey(candidate.Chunk.Path)]++
		subsystemCounts[subsystemKey(candidate.Chunk.Path)]++
	}
	return ordered
}

func diversityPenalty(
	candidate retrieve.Candidate,
	position int,
	files, subsystems map[string]int,
) int {
	return position + files[fileKey(candidate.Chunk.Path)]*4 +
		subsystems[subsystemKey(candidate.Chunk.Path)]*2
}

func adjacentCandidates(
	snapshot index.Snapshot,
	primaries, existing []retrieve.Candidate,
) ([]retrieve.Candidate, map[string]struct{}) {
	files := make(map[string]index.FileRecord, len(snapshot.Files))
	for _, file := range snapshot.Files {
		files[file.Path] = file
	}
	existingByKey := make(map[string]retrieve.Candidate, len(existing))
	for _, candidate := range existing {
		existingByKey[candidateKey(candidate)] = candidate
	}
	seen := make(map[string]struct{}, len(primaries)*3)
	for _, primary := range primaries {
		seen[candidateKey(primary)] = struct{}{}
	}
	promoted := make(map[string]struct{})

	adjacent := make([]retrieve.Candidate, 0, len(primaries)*2)
	for _, primary := range primaries {
		file, exists := files[primary.Chunk.Path]
		if !exists {
			continue
		}
		chunkIndex := preparedChunkIndex(file.Chunks, primary.Chunk)
		if chunkIndex < 0 {
			continue
		}
		for _, neighbor := range []struct {
			offset int
			label  string
		}{{-1, "previous"}, {1, "next"}} {
			neighborIndex := chunkIndex + neighbor.offset
			if neighborIndex < 0 || neighborIndex >= len(file.Chunks) {
				continue
			}
			chunk := file.Chunks[neighborIndex]
			key := candidateKey(retrieve.Candidate{Chunk: chunk})
			if _, exists := seen[key]; exists {
				continue
			}
			candidate, exists := existingByKey[key]
			if exists {
				promoted[key] = struct{}{}
			} else {
				candidate = retrieve.Candidate{
					Chunk:  chunk,
					Source: adjacentSource,
					Reasons: []string{
						"immediate " + neighbor.label + " prepared chunk adjacent to ranked candidate",
					},
				}
			}
			if overlapsAny(candidate, primaries) && !exists || overlapsAny(candidate, adjacent) {
				continue
			}
			seen[key] = struct{}{}
			adjacent = append(adjacent, candidate)
		}
	}
	return adjacent, promoted
}

func preparedChunkIndex(chunks []index.Chunk, target index.Chunk) int {
	for chunkIndex, chunk := range chunks {
		if target.ID != "" && chunk.ID == target.ID {
			return chunkIndex
		}
		if target.ID == "" && chunk.Path == target.Path &&
			chunk.StartLine == target.StartLine && chunk.EndLine == target.EndLine {
			return chunkIndex
		}
	}
	return -1
}

func overlapsAny(candidate retrieve.Candidate, existing []retrieve.Candidate) bool {
	for _, prior := range existing {
		if overlappingChunks(prior.Chunk, candidate.Chunk) {
			return true
		}
	}
	return false
}

func overlappingChunks(left, right index.Chunk) bool {
	return left.Path == right.Path && left.StartLine <= right.EndLine && right.StartLine <= left.EndLine
}

func candidateKey(candidate retrieve.Candidate) string {
	if candidate.Chunk.ID != "" {
		return "id:" + candidate.Chunk.ID
	}
	return "range:" + candidate.Chunk.Path + ":" +
		strconv.Itoa(candidate.Chunk.StartLine) + ":" + strconv.Itoa(candidate.Chunk.EndLine)
}

func fileKey(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

func subsystemKey(path string) string {
	parts := strings.Split(fileKey(path), "/")
	if len(parts) > 1 && parts[0] == "internal" {
		return parts[1]
	}
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
