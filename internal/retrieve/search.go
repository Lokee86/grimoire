package retrieve

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/Lokee86/grimoire/internal/evidence"
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
	Context      *evidence.Descriptor
}

type lexicalQuerySpec struct {
	phrase      string
	termIndexes []int
}

type lexicalFields struct {
	text           string
	baseMatches    []bool
	pathMatches    []bool
	leadingMatches []bool
}

func Search(snapshot index.Snapshot, query string, limit int) []Candidate {
	results := SearchMany(snapshot, []string{query}, limit)
	if len(results) == 0 {
		return nil
	}
	return results[0]
}

// SearchMany scores a bounded set of queries against one shared BM25 corpus.
// Repository content and field tokens are scanned once per request rather than
// once per retrieval intent.
func SearchMany(snapshot index.Snapshot, queries []string, limit int) [][]Candidate {
	specs, terms := compileLexicalQueries(queries)
	results := make([][]Candidate, len(queries))
	if len(terms) == 0 {
		return results
	}

	chunks := snapshot.AllChunks()
	texts := make([]string, len(chunks))
	for chunkIndex, chunk := range chunks {
		texts[chunkIndex] = chunk.Text
	}
	corpus := newBM25Corpus(texts, terms)
	fields := make([]lexicalFields, len(chunks))
	for chunkIndex, chunk := range chunks {
		text := strings.ToLower(chunk.Text)
		firstLine := text
		if newline := strings.IndexByte(firstLine, '\n'); newline >= 0 {
			firstLine = firstLine[:newline]
		}
		fields[chunkIndex] = lexicalFields{
			text:           text,
			baseMatches:    corpus.matches(filepath.Base(chunk.Path)),
			pathMatches:    corpus.matches(chunk.Path),
			leadingMatches: corpus.matches(firstLine),
		}
	}

	for queryIndex, spec := range specs {
		if len(spec.termIndexes) == 0 {
			continue
		}
		candidates := make([]Candidate, 0)
		for chunkIndex, chunk := range chunks {
			candidate := scoreChunk(chunk, chunkIndex, corpus, fields[chunkIndex], spec)
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
		results[queryIndex] = candidates
	}
	return results
}

func compileLexicalQueries(queries []string) ([]lexicalQuerySpec, []string) {
	specs := make([]lexicalQuerySpec, len(queries))
	termPositions := make(map[string]int)
	var terms []string
	for queryIndex, query := range queries {
		query = strings.TrimSpace(query)
		specs[queryIndex].phrase = strings.ToLower(query)
		for _, term := range queryTerms(query) {
			position, exists := termPositions[term]
			if !exists {
				position = len(terms)
				termPositions[term] = position
				terms = append(terms, term)
			}
			specs[queryIndex].termIndexes = append(specs[queryIndex].termIndexes, position)
		}
	}
	return specs, terms
}

func scoreChunk(
	chunk index.Chunk,
	documentIndex int,
	corpus bm25Corpus,
	fields lexicalFields,
	spec lexicalQuerySpec,
) Candidate {
	candidate := Candidate{Chunk: chunk, Source: "lexical"}
	if len(spec.termIndexes) > 1 && len(spec.phrase) > 2 && strings.Contains(fields.text, spec.phrase) {
		candidate.addScore("exact query phrase in content", 12)
	}

	for _, termIndex := range spec.termIndexes {
		term := corpus.terms[termIndex]
		if fields.baseMatches[termIndex] {
			candidate.addScore("filename matches "+term.text, 8)
		} else if fields.pathMatches[termIndex] {
			candidate.addScore("path matches "+term.text, 4)
		}
		if fields.leadingMatches[termIndex] {
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
