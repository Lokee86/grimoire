package retrieve

import (
	"sort"

	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/lexical"
)

// FileCandidate is one whole-file lexical match. It is intentionally separate
// from Candidate because files are discovery scopes, not context-package units.
type FileCandidate struct {
	Path         string
	Score        float64
	Rank         int
	Reasons      []string
	ScoreDetails []ScoreDetail
}

func SearchFilesManyWithConfig(
	snapshot index.Snapshot,
	queries []string,
	limit int,
	config Config,
) [][]FileCandidate {
	config = normalizedConfig(config)
	specs, terms := compileLexicalQueries(queries)
	results := make([][]FileCandidate, len(queries))
	if len(terms) == 0 || limit <= 0 {
		return results
	}

	lexicalIndex := snapshot.FileLexicalIndex()
	corpus := newBM25Corpus(lexicalIndex, terms)
	for queryIndex, spec := range specs {
		if len(spec.termIndexes) == 0 {
			continue
		}
		candidateTerms := make([]string, 0, len(spec.termIndexes))
		for _, termIndex := range spec.termIndexes {
			candidateTerms = append(candidateTerms, corpus.terms[termIndex].text)
		}
		candidates := make([]FileCandidate, 0)
		for _, documentIndex := range lexicalIndex.CandidateDocuments(candidateTerms) {
			if documentIndex < 0 || documentIndex >= len(snapshot.Files) {
				continue
			}
			document := lexicalIndex.Document(documentIndex)
			candidate := scoreFile(snapshot.Files[documentIndex].Path, documentIndex, corpus, document, spec)
			if candidate.Score > 0 {
				candidates = append(candidates, candidate)
			}
		}
		sort.Slice(candidates, func(left, right int) bool {
			if candidates[left].Score != candidates[right].Score {
				return candidates[left].Score > candidates[right].Score
			}
			return candidates[left].Path < candidates[right].Path
		})
		if len(candidates) > limit {
			candidates = candidates[:limit]
		}
		for index := range candidates {
			candidates[index].Rank = index + 1
		}
		results[queryIndex] = candidates
	}
	return results
}

func scoreFile(
	path string,
	documentIndex int,
	corpus bm25Corpus,
	document lexical.Document,
	spec lexicalQuerySpec,
) FileCandidate {
	candidate := FileCandidate{Path: path}
	for _, termIndex := range spec.termIndexes {
		term := corpus.terms[termIndex]
		if lexical.Contains(document.BaseTokens, term.text) {
			candidate.addScore("filename matches "+term.text, 8)
		} else if lexical.Contains(document.PathTokens, term.text) {
			candidate.addScore("path matches "+term.text, 4)
		}
		if value := corpus.score(documentIndex, termIndex); value > 0 {
			candidate.addScore("BM25 file content matches "+term.text, value)
		}
	}
	return candidate
}

func (candidate *FileCandidate) addScore(name string, value float64) {
	candidate.Score += value
	candidate.Reasons = append(candidate.Reasons, name)
	candidate.ScoreDetails = append(candidate.ScoreDetails, ScoreDetail{Name: name, Value: value})
}
