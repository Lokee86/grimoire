package retrieve

import (
	"math"
	"strings"
	"unicode"
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

type lexicalDocument struct {
	frequencies []int
	length      int
}

type bm25Corpus struct {
	terms               []compiledQueryTerm
	termIndexes         map[string]int
	documents           []lexicalDocument
	documentCount       int
	averageLength       float64
	documentFrequencies []int
}

func newBM25Corpus(texts []string, queryTerms []string) bm25Corpus {
	corpus := bm25Corpus{
		terms:               compileQueryTerms(queryTerms),
		termIndexes:         make(map[string]int, len(queryTerms)),
		documents:           make([]lexicalDocument, 0, len(texts)),
		documentCount:       len(texts),
		documentFrequencies: make([]int, len(queryTerms)),
	}
	for index, term := range corpus.terms {
		corpus.termIndexes[term.text] = index
	}

	totalLength := 0
	for _, text := range texts {
		document := lexicalDocument{frequencies: make([]int, len(corpus.terms))}
		scanLexicalTokens(text, func(token []rune) {
			document.length++
			termIndex, exists := corpus.termIndexes[strings.ToLower(string(token))]
			if exists {
				document.frequencies[termIndex]++
			}
		})
		for termIndex, frequency := range document.frequencies {
			if frequency > 0 {
				corpus.documentFrequencies[termIndex]++
			}
		}
		corpus.documents = append(corpus.documents, document)
		totalLength += document.length
	}
	if corpus.documentCount > 0 {
		corpus.averageLength = float64(totalLength) / float64(corpus.documentCount)
	}
	if corpus.averageLength == 0 {
		corpus.averageLength = 1
	}
	return corpus
}

func (corpus bm25Corpus) score(documentIndex, termIndex int) float64 {
	if documentIndex < 0 || documentIndex >= len(corpus.documents) {
		return 0
	}
	if termIndex < 0 || termIndex >= len(corpus.terms) {
		return 0
	}
	document := corpus.documents[documentIndex]
	frequency := document.frequencies[termIndex]
	if frequency == 0 {
		return 0
	}
	documentFrequency := corpus.documentFrequencies[termIndex]
	idf := math.Log(1 + (float64(corpus.documentCount-documentFrequency)+0.5)/(float64(documentFrequency)+0.5))
	lengthRatio := float64(document.length) / corpus.averageLength
	denominator := float64(frequency) + bm25K1*(1-bm25B+bm25B*lengthRatio)
	return idf * (float64(frequency) * (bm25K1 + 1)) / denominator
}

func (corpus bm25Corpus) matches(value string) []bool {
	matches := make([]bool, len(corpus.terms))
	scanLexicalTokens(value, func(token []rune) {
		termIndex, exists := corpus.termIndexes[strings.ToLower(string(token))]
		if exists {
			matches[termIndex] = true
		}
	})
	return matches
}

func compileQueryTerms(terms []string) []compiledQueryTerm {
	compiled := make([]compiledQueryTerm, 0, len(terms))
	for _, term := range terms {
		compiled = append(compiled, compiledQueryTerm{text: term})
	}
	return compiled
}

func queryTerms(query string) []string {
	tokens := lexicalTokens(query)
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

func lexicalTokens(value string) []string {
	var tokens []string
	scanLexicalTokens(value, func(token []rune) {
		normalized := strings.ToLower(string(token))
		if normalized != "" {
			tokens = append(tokens, normalized)
		}
	})
	return tokens
}

func scanLexicalTokens(value string, yield func([]rune)) {
	segment := make([]rune, 0, 32)
	flush := func() {
		if len(segment) == 0 {
			return
		}
		yieldIdentifierTokens(segment, yield)
		segment = segment[:0]
	}
	for _, current := range value {
		if unicode.IsLetter(current) || unicode.IsDigit(current) {
			segment = append(segment, current)
			continue
		}
		flush()
	}
	flush()
}

func yieldIdentifierTokens(identifier []rune, yield func([]rune)) {
	yield(identifier)
	start := 0
	for index := 1; index < len(identifier); index++ {
		previous := identifier[index-1]
		current := identifier[index]
		nextIsLower := index+1 < len(identifier) && unicode.IsLower(identifier[index+1])
		boundary := unicode.IsLower(previous) && unicode.IsUpper(current)
		boundary = boundary || unicode.IsLetter(previous) && unicode.IsDigit(current)
		boundary = boundary || unicode.IsDigit(previous) && unicode.IsLetter(current)
		boundary = boundary || unicode.IsUpper(previous) && unicode.IsUpper(current) && nextIsLower
		if !boundary {
			continue
		}
		yield(identifier[start:index])
		start = index
	}
	if start > 0 {
		yield(identifier[start:])
	}
}
