package retrieve

import (
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/Lokee86/grimoire/internal/index"
)

type Candidate struct {
	Chunk   index.Chunk
	Score   float64
	Source  string
	Rank    int
	Reasons []string
}

func Search(snapshot index.Snapshot, query string, limit int) []Candidate {
	phrase := strings.ToLower(strings.TrimSpace(query))
	terms := queryTerms(phrase)
	if len(terms) == 0 {
		return nil
	}

	candidates := make([]Candidate, 0)
	for _, chunk := range snapshot.AllChunks() {
		candidate := scoreChunk(chunk, phrase, terms)
		if candidate.Score > 0 {
			candidates = append(candidates, candidate)
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Score != candidates[j].Score {
			return candidates[i].Score > candidates[j].Score
		}
		if candidates[i].Chunk.Path != candidates[j].Chunk.Path {
			return candidates[i].Chunk.Path < candidates[j].Chunk.Path
		}
		return candidates[i].Chunk.StartLine < candidates[j].Chunk.StartLine
	})
	if limit > 0 && len(candidates) > limit {
		candidates = candidates[:limit]
	}
	for index := range candidates {
		candidates[index].Rank = index + 1
	}
	return candidates
}

func scoreChunk(chunk index.Chunk, phrase string, terms []string) Candidate {
	text := strings.ToLower(chunk.Text)
	path := strings.ToLower(chunk.Path)
	base := strings.ToLower(filepath.Base(chunk.Path))
	firstLine := text
	if newline := strings.IndexByte(firstLine, '\n'); newline >= 0 {
		firstLine = firstLine[:newline]
	}

	candidate := Candidate{Chunk: chunk, Source: "lexical"}
	if len(phrase) > 2 && strings.Contains(text, phrase) {
		candidate.Score += 12
		candidate.Reasons = append(candidate.Reasons, "exact query phrase in content")
	}

	for _, term := range terms {
		if strings.Contains(base, term) {
			candidate.Score += 8
			candidate.Reasons = append(candidate.Reasons, "filename matches "+term)
		} else if strings.Contains(path, term) {
			candidate.Score += 4
			candidate.Reasons = append(candidate.Reasons, "path matches "+term)
		}
		if strings.Contains(firstLine, term) {
			candidate.Score += 4
			candidate.Reasons = append(candidate.Reasons, "leading line matches "+term)
		}
		occurrences := min(strings.Count(text, term), 5)
		if occurrences > 0 {
			candidate.Score += float64(occurrences * 2)
			candidate.Reasons = append(candidate.Reasons, "content matches "+term)
		}
	}
	return candidate
}

func queryTerms(query string) []string {
	fields := strings.FieldsFunc(query, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_'
	})
	seen := make(map[string]struct{}, len(fields))
	terms := make([]string, 0, len(fields))
	for _, field := range fields {
		if len(field) < 2 {
			continue
		}
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		terms = append(terms, field)
	}
	return terms
}
