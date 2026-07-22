package retrieve

import (
	"sort"
	"strings"

	"github.com/Lokee86/grimoire/internal/index"
)

func Exact(snapshot index.Snapshot, query string, limit int) []Candidate {
	signals := exactSignals(query)
	if len(signals) == 0 {
		return nil
	}
	candidates := make([]Candidate, 0)
	for _, chunk := range snapshot.AllChunks() {
		candidate := Candidate{Chunk: chunk, Source: "exact"}
		for _, signal := range signals {
			if strings.Contains(chunk.Path, signal.value) {
				candidate.Score += signal.weight + 1
				candidate.Reasons = append(candidate.Reasons, exactReason(signal, "path"))
			}
			if exactContains(chunk.Text, signal.value, signal.kind) {
				candidate.Score += signal.weight
				candidate.Reasons = append(candidate.Reasons, exactReason(signal, "content"))
			}
		}
		if candidate.Score > 0 {
			candidates = append(candidates, candidate)
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if a.Score != b.Score {
			return a.Score > b.Score
		}
		if a.Chunk.Path != b.Chunk.Path {
			return a.Chunk.Path < b.Chunk.Path
		}
		if a.Chunk.StartLine != b.Chunk.StartLine {
			return a.Chunk.StartLine < b.Chunk.StartLine
		}
		return a.Chunk.ID < b.Chunk.ID
	})
	if limit > 0 && len(candidates) > limit {
		candidates = candidates[:limit]
	}
	for i := range candidates {
		candidates[i].Rank = i + 1
	}
	return candidates
}
