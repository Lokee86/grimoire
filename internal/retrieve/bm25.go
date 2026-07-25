package retrieve

import (
	"math"
	"sort"

	"github.com/Lokee86/grimoire/internal/lexical"
)

const (
	bm25K1 = 1.2
	bm25B  = 0.75
)

var queryStopwords = map[string]struct{}{
	"a": {}, "an": {}, "and": {}, "are": {}, "do": {}, "does": {},
	"find": {}, "for": {}, "how": {}, "in": {}, "is": {}, "of": {},
	"or": {}, "show": {}, "the": {}, "to": {}, "what": {}, "where": {},
	"which": {}, "with": {},
}

type compiledQueryTerm struct {
	text string
}

type bm25Corpus struct {
	terms       []compiledQueryTerm
	termIndexes map[string]int
	index       *lexical.Index
}

func newBM25Corpus(lexicalIndex *lexical.Index, queryTerms []string) bm25Corpus {
	corpus := bm25Corpus{
		terms:       compileQueryTerms(queryTerms),
		termIndexes: make(map[string]int, len(queryTerms)),
		index:       lexicalIndex,
	}
	for index, term := range corpus.terms {
		corpus.termIndexes[term.text] = index
	}
	return corpus
}

func (corpus bm25Corpus) score(documentIndex, termIndex int) float64 {
	if corpus.index == nil || termIndex < 0 || termIndex >= len(corpus.terms) {
		return 0
	}
	term := corpus.terms[termIndex].text
	postings := corpus.index.Posting(term)
	position := sort.Search(len(postings), func(index int) bool {
		return postings[index].Document >= documentIndex
	})
	if position >= len(postings) || postings[position].Document != documentIndex {
		return 0
	}
	frequency := postings[position].Frequency
	documentFrequency := len(postings)
	documentCount := corpus.index.DocumentCount()
	idf := math.Log(1 + (float64(documentCount-documentFrequency)+0.5)/(float64(documentFrequency)+0.5))
	document := corpus.index.Document(documentIndex)
	lengthRatio := float64(document.Length) / corpus.index.AverageLength()
	denominator := float64(frequency) + bm25K1*(1-bm25B+bm25B*lengthRatio)
	return idf * (float64(frequency) * (bm25K1 + 1)) / denominator
}

func compileQueryTerms(terms []string) []compiledQueryTerm {
	compiled := make([]compiledQueryTerm, 0, len(terms))
	for _, term := range terms {
		compiled = append(compiled, compiledQueryTerm{text: term})
	}
	return compiled
}

func queryTerms(query string) []string {
	tokens := lexical.Tokens(query)
	seen := make(map[string]struct{}, len(tokens))
	allTerms := make([]string, 0, len(tokens))
	filtered := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if len(token) < 2 {
			continue
		}
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		allTerms = append(allTerms, token)
		if _, stopword := queryStopwords[token]; !stopword {
			filtered = append(filtered, token)
		}
	}
	if len(filtered) > 0 {
		return filtered
	}
	return allTerms
}
