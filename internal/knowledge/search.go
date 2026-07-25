package knowledge

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/Lokee86/grimoire/internal/lexical"
)

const (
	bm25K1 = 1.2
	bm25B  = 0.75
)

var stopwords = map[string]struct{}{
	"a": {}, "an": {}, "and": {}, "are": {}, "do": {}, "does": {}, "find": {},
	"for": {}, "how": {}, "in": {}, "is": {}, "of": {}, "or": {}, "show": {},
	"the": {}, "to": {}, "what": {}, "where": {}, "which": {}, "with": {},
}

type searchableSection struct {
	document Document
	section  Section
	terms    map[string]int
	length   int
}

func Search(ctx context.Context, index Index, query string, options SearchOptions) (SearchResponse, error) {
	if strings.TrimSpace(query) == "" {
		return SearchResponse{}, fmt.Errorf("knowledge query must not be empty")
	}
	if options.TopK <= 0 {
		options.TopK = 20
	}
	terms := queryTerms(query)
	candidates := make([]searchableSection, 0)
	for _, document := range index.Documents {
		if options.Path != "" && !strings.HasPrefix(document.Path, options.Path) {
			continue
		}
		if options.Kind != "" && document.Kind != options.Kind {
			continue
		}
		if options.CommitID != "" && document.CommitID != options.CommitID {
			continue
		}
		if !options.Since.IsZero() && (document.CommitTime == nil || document.CommitTime.Before(options.Since)) {
			continue
		}
		if !options.Until.IsZero() && (document.CommitTime == nil || document.CommitTime.After(options.Until)) {
			continue
		}
		for _, section := range document.Sections {
			if options.Heading != "" && !headingMatches(section, options.Heading) {
				continue
			}
			termsMap := make(map[string]int, len(section.Terms))
			length := 0
			for _, term := range section.Terms {
				termsMap[term.Term], length = term.Frequency, length+term.Frequency
			}
			candidates = append(candidates, searchableSection{document: document, section: section, terms: termsMap, length: length})
		}
	}
	if len(candidates) == 0 {
		return SearchResponse{Results: []Result{}}, nil
	}
	averageLength := 0.0
	for _, candidate := range candidates {
		averageLength += float64(candidate.length)
	}
	averageLength /= float64(len(candidates))
	if averageLength == 0 {
		averageLength = 1
	}
	documentFrequency := make(map[string]int)
	for _, candidate := range candidates {
		for term := range candidate.terms {
			documentFrequency[term]++
		}
	}
	vectorSections := make([]Section, len(candidates))
	for i := range candidates {
		vectorSections[i] = candidates[i].section
	}
	vectorScores := map[string]float64(nil)
	response := SearchResponse{}
	if options.Vector != nil {
		var err error
		vectorScores, err = options.Vector.Rank(ctx, query, vectorSections)
		if err != nil {
			response.VectorError = err.Error()
		} else {
			response.VectorUsed = true
		}
	}
	results := make([]Result, 0, len(candidates))
	for _, candidate := range candidates {
		score, reasons := bm25Score(candidate, terms, documentFrequency, len(candidates), averageLength)
		if vectorScore, ok := vectorScores[candidate.section.ID]; ok {
			score += vectorScore * 0.25
			reasons = append(reasons, fmt.Sprintf("supplemental vector score %.4f", vectorScore))
		}
		if headingMatches(candidate.section, query) {
			reasons = append(reasons, "heading matches query")
		}
		if strings.Contains(strings.ToLower(candidate.document.Path), strings.ToLower(query)) {
			reasons = append(reasons, "path matches query")
		}
		results = append(results, makeResult(candidate.document, candidate.section, score, reasons))
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		if results[i].Path != results[j].Path {
			return results[i].Path < results[j].Path
		}
		return results[i].SectionID < results[j].SectionID
	})
	if len(results) > options.TopK {
		results = results[:options.TopK]
	}
	response.Results = results
	return response, nil
}

func bm25Score(candidate searchableSection, query []string, dfs map[string]int, corpusSize int, averageLength float64) (float64, []string) {
	score := 0.0
	reasons := make([]string, 0, len(query))
	for _, term := range query {
		frequency := candidate.terms[term]
		if frequency == 0 {
			continue
		}
		df := dfs[term]
		idf := math.Log(1 + (float64(corpusSize-df)+0.5)/(float64(df)+0.5))
		lengthRatio := float64(candidate.length) / averageLength
		denominator := float64(frequency) + bm25K1*(1-bm25B+bm25B*lengthRatio)
		score += idf * (float64(frequency) * (bm25K1 + 1)) / denominator
		reasons = append(reasons, fmt.Sprintf("BM25 term %q", term))
	}
	return score, reasons
}

func makeResult(document Document, section Section, score float64, reasons []string) Result {
	return Result{
		Handle: document.Handle(section), DocumentID: document.ID, SectionID: section.ID,
		Path: document.Path, Kind: document.Kind, Heading: section.Heading, HeadingPath: section.HeadingPath,
		StartByte: section.StartByte, EndByte: section.EndByte, StartLine: section.StartLine, EndLine: section.EndLine,
		Hash: section.Hash, Text: section.Text, CommitID: document.CommitID, CommitTime: document.CommitTime,
		Score: score, Reasons: reasons, CodeLinks: section.CodeLinks,
	}
}

func queryTerms(query string) []string {
	seen := make(map[string]struct{})
	terms := make([]string, 0)
	all := make([]string, 0)
	for _, term := range lexical.Tokens(query) {
		if len(term) < 2 {
			continue
		}
		if _, ok := seen[term]; ok {
			continue
		}
		seen[term] = struct{}{}
		all = append(all, term)
		if _, stop := stopwords[term]; !stop {
			terms = append(terms, term)
		}
	}
	if len(terms) > 0 {
		return terms
	}
	return all
}

func headingMatches(section Section, query string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return false
	}
	for _, heading := range append([]string{section.Heading}, section.HeadingPath...) {
		if strings.Contains(strings.ToLower(heading), query) {
			return true
		}
	}
	return false
}

// Compile-time use keeps the lexical package's token contract explicit here.
var _ = lexical.Contains
