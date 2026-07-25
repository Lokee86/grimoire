package retrieve

import (
	"sort"
	"strings"

	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/lexical"
)

func Exact(snapshot index.Snapshot, query string, limit int) []Candidate {
	signals := exactSignals(query)
	if len(signals) == 0 {
		return nil
	}
	chunks := snapshot.AllChunks()
	lexicalIndex := snapshot.LexicalIndex()
	documents, indexed := exactCandidateDocuments(lexicalIndex, signals)
	if !indexed {
		documents = make([]int, len(chunks))
		for document := range chunks {
			documents[document] = document
		}
	}

	candidates := make([]Candidate, 0)
	for _, document := range documents {
		if document < 0 || document >= len(chunks) {
			continue
		}
		chunk := chunks[document]
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

func exactCandidateDocuments(lexicalIndex *lexical.Index, signals []exactSignal) ([]int, bool) {
	seenDocuments := make(map[int]struct{})
	for _, signal := range signals {
		terms := uniqueLexicalTerms(signal.value)
		if len(terms) == 0 {
			return nil, false
		}
		for _, document := range lexicalIndex.CandidateDocumentsAll(terms) {
			seenDocuments[document] = struct{}{}
		}
	}
	result := make([]int, 0, len(seenDocuments))
	for document := range seenDocuments {
		result = append(result, document)
	}
	sort.Ints(result)
	return result, true
}

func uniqueLexicalTerms(value string) []string {
	seen := make(map[string]struct{})
	terms := make([]string, 0)
	for _, term := range lexical.Tokens(value) {
		if _, exists := seen[term]; exists {
			continue
		}
		seen[term] = struct{}{}
		terms = append(terms, term)
	}
	return terms
}
