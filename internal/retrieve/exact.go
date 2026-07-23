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
				value := signal.weight + 1
				reason := exactReason(signal, "path")
				candidate.Score += value
				candidate.Reasons = append(candidate.Reasons, reason)
				candidate.ScoreDetails = append(candidate.ScoreDetails, ScoreDetail{
					Name: reason, Value: value,
				})
			}
			if exactContains(chunk.Text, signal.value, signal.kind) {
				reason := exactReason(signal, "content")
				candidate.Score += signal.weight
				candidate.Reasons = append(candidate.Reasons, reason)
				candidate.ScoreDetails = append(candidate.ScoreDetails, ScoreDetail{
					Name: reason, Value: signal.weight,
				})
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
