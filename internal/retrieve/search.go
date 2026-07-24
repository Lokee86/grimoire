package retrieve

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/Lokee86/grimoire/internal/index"
)

type ScoreDetail struct {
	Name  string
	Value float64
}

type Candidate struct {
	Chunk        index.Chunk
	Score        float64
	Source       string
	Rank         int
	Reasons      []string
	ScoreDetails []ScoreDetail
}

func Search(snapshot index.Snapshot, query string, limit int) []Candidate {
	query = strings.TrimSpace(query)
	phrase := strings.ToLower(query)
	terms := queryTerms(query)
	if len(terms) == 0 {
		return nil
	}

	chunks := snapshot.AllChunks()
	texts := make([]string, len(chunks))
	for chunkIndex, chunk := range chunks {
		texts[chunkIndex] = chunk.Text
	}
	corpus := newBM25Corpus(texts, terms)

	candidates := make([]Candidate, 0)
	for chunkIndex, chunk := range chunks {
		candidate := scoreChunk(chunk, chunkIndex, corpus, phrase)
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

func scoreChunk(chunk index.Chunk, documentIndex int, corpus bm25Corpus, phrase string) Candidate {
	text := strings.ToLower(chunk.Text)
	firstLine := text
	if newline := strings.IndexByte(firstLine, '\n'); newline >= 0 {
		firstLine = firstLine[:newline]
	}
	baseMatches := corpus.matches(filepath.Base(chunk.Path))
	pathMatches := corpus.matches(chunk.Path)
	leadingMatches := corpus.matches(firstLine)

	candidate := Candidate{Chunk: chunk, Source: "lexical"}
	if len(corpus.terms) > 1 && len(phrase) > 2 && strings.Contains(text, phrase) {
		candidate.addScore("exact query phrase in content", 12)
	}

	for termIndex, term := range corpus.terms {
		if baseMatches[termIndex] {
			candidate.addScore("filename matches "+term.text, 8)
		} else if pathMatches[termIndex] {
			candidate.addScore("path matches "+term.text, 4)
		}
		if leadingMatches[termIndex] {
			candidate.addScore("leading line matches "+term.text, 4)
		}
		if value := corpus.score(documentIndex, termIndex); value > 0 {
			candidate.addScore("BM25 content matches "+term.text, value)
		}
	}
	return candidate
}

func (candidate *Candidate) addScore(name string, value float64) {
	candidate.Score += value
	candidate.Reasons = append(candidate.Reasons, name)
	candidate.ScoreDetails = append(candidate.ScoreDetails, ScoreDetail{Name: name, Value: value})
}
