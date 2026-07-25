package extraction

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/tokenizer"
)

func (extractor Extractor) Refine(request Request, candidates []retrieve.Candidate) ([]retrieve.Candidate, error) {
	if len(candidates) == 0 || len(extractor.discoverers) == 0 {
		return append([]retrieve.Candidate(nil), candidates...), nil
	}
	refined := make([]retrieve.Candidate, 0, len(candidates))
	for _, candidate := range candidates {
		parts, err := extractor.refineCandidate(request, candidate)
		if err != nil {
			return nil, err
		}
		refined = append(refined, parts...)
	}
	return refined, nil
}

func (extractor Extractor) refineCandidate(request Request, candidate retrieve.Candidate) ([]retrieve.Candidate, error) {
	chunk := candidate.Chunk
	lineCount := chunkLineCount(chunk)
	originalTokens := chunk.TokenCount
	if originalTokens <= 0 {
		var err error
		originalTokens, err = tokenizer.Count(chunk.Text)
		if err != nil {
			return nil, fmt.Errorf("count original candidate tokens: %w", err)
		}
	}
	if lineCount < extractor.config.MinChunkLines || originalTokens < extractor.config.MinChunkTokens {
		return []retrieve.Candidate{candidate}, nil
	}
	query := candidateQuery(request, candidate)
	terms := queryTerms(query)
	if len(terms) == 0 {
		return []retrieve.Candidate{candidate}, nil
	}

	discoveryRequest := DiscoveryRequest{Chunk: chunk, Query: query, Terms: terms}
	var spans []Span
	for _, discoverer := range extractor.discoverers {
		discovered, err := discoverer.Discover(discoveryRequest)
		if err != nil {
			return nil, fmt.Errorf("discover spans in %s:%d-%d: %w", chunk.Path, chunk.StartLine, chunk.EndLine, err)
		}
		discovered = normalizeSpans(discovered, lineCount)
		if len(discovered) == 0 {
			continue
		}
		spans = discovered
		break
	}
	if len(spans) == 0 {
		return []retrieve.Candidate{candidate}, nil
	}
	if len(spans) > extractor.config.MaxSpans {
		spans = spans[:extractor.config.MaxSpans]
	}
	sort.Slice(spans, func(left, right int) bool {
		if spans[left].StartLine != spans[right].StartLine {
			return spans[left].StartLine < spans[right].StartLine
		}
		return spans[left].EndLine < spans[right].EndLine
	})
	parts, totalTokens, err := buildCandidates(candidate, spans)
	if err != nil {
		return nil, err
	}
	if totalTokens >= originalTokens-extractor.config.MinTokenSavings ||
		float64(totalTokens) > float64(originalTokens)*extractor.config.MaxRetainedRatio {
		return []retrieve.Candidate{candidate}, nil
	}
	return parts, nil
}

func candidateQuery(request Request, candidate retrieve.Candidate) string {
	if candidate.Context == nil || len(candidate.Context.FacetRanks) == 0 || len(request.FacetQueries) == 0 {
		return request.Query
	}
	type rankedFacet struct {
		id   string
		rank int
	}
	facets := make([]rankedFacet, 0, len(candidate.Context.FacetRanks))
	for facet, rank := range candidate.Context.FacetRanks {
		if _, exists := request.FacetQueries[facet]; exists {
			facets = append(facets, rankedFacet{id: facet, rank: rank})
		}
	}
	sort.Slice(facets, func(left, right int) bool {
		if facets[left].rank != facets[right].rank {
			return facets[left].rank < facets[right].rank
		}
		return facets[left].id < facets[right].id
	})
	queries := make([]string, 0, len(facets))
	seen := make(map[string]struct{}, len(facets))
	for _, facet := range facets {
		query := strings.TrimSpace(request.FacetQueries[facet.id])
		if query == "" {
			continue
		}
		if _, exists := seen[query]; exists {
			continue
		}
		seen[query] = struct{}{}
		queries = append(queries, query)
	}
	if len(queries) == 0 {
		return request.Query
	}
	return strings.Join(queries, "\n")
}

func normalizeSpans(spans []Span, lineCount int) []Span {
	if lineCount <= 0 {
		return nil
	}
	normalized := make([]Span, 0, len(spans))
	for _, span := range spans {
		span.StartLine = max(0, span.StartLine)
		span.EndLine = min(lineCount-1, span.EndLine)
		if span.StartLine > span.EndLine {
			continue
		}
		normalized = append(normalized, span)
	}
	merged := make([]Span, 0, len(normalized))
	for _, span := range normalized {
		target := -1
		for index := 0; index < len(merged); {
			if span.StartLine > merged[index].EndLine+1 || span.EndLine+1 < merged[index].StartLine {
				index++
				continue
			}
			if target < 0 {
				target = index
				merged[target].StartLine = min(merged[target].StartLine, span.StartLine)
				merged[target].EndLine = max(merged[target].EndLine, span.EndLine)
				if merged[target].Reason == "" {
					merged[target].Reason = span.Reason
				}
				index++
				continue
			}
			merged[target].StartLine = min(merged[target].StartLine, merged[index].StartLine)
			merged[target].EndLine = max(merged[target].EndLine, merged[index].EndLine)
			merged = append(merged[:index], merged[index+1:]...)
		}
		if target < 0 {
			merged = append(merged, span)
		}
	}
	return merged
}

func buildCandidates(candidate retrieve.Candidate, spans []Span) ([]retrieve.Candidate, int, error) {
	lines := strings.Split(candidate.Chunk.Text, "\n")
	parts := make([]retrieve.Candidate, 0, len(spans))
	totalTokens := 0
	for spanIndex, span := range spans {
		text := strings.Join(lines[span.StartLine:span.EndLine+1], "\n")
		tokenCount, err := tokenizer.Count(text)
		if err != nil {
			return nil, 0, fmt.Errorf("count extracted candidate tokens: %w", err)
		}
		startLine := candidate.Chunk.StartLine + span.StartLine
		endLine := candidate.Chunk.StartLine + span.EndLine
		part := cloneCandidate(candidate)
		part.Chunk = index.Chunk{
			ID: evidence.StableID(
				"extracted-range", candidate.Chunk.ID, candidate.Chunk.Path,
				strconv.Itoa(startLine), strconv.Itoa(endLine), text,
			),
			Path:       candidate.Chunk.Path,
			StartLine:  startLine,
			EndLine:    endLine,
			TokenCount: tokenCount,
			Text:       text,
		}
		reason := fmt.Sprintf(
			"query-focused extraction %d/%d from %s:%d-%d",
			spanIndex+1, len(spans), candidate.Chunk.Path,
			candidate.Chunk.StartLine, candidate.Chunk.EndLine,
		)
		if span.Reason != "" {
			reason += " (" + span.Reason + ")"
		}
		part.Reasons = appendUnique(part.Reasons, reason)
		part.Context = refinedDescriptor(candidate.Context, part.Chunk)
		parts = append(parts, part)
		totalTokens += tokenCount
	}
	return parts, totalTokens, nil
}

func refinedDescriptor(original *evidence.Descriptor, chunk index.Chunk) *evidence.Descriptor {
	cloned := evidence.Descriptor{}
	if original != nil {
		cloned = evidence.Merge(*original, evidence.Descriptor{})
	}
	previousIdentity := cloned.Identity
	identity := evidence.RangeIdentity(chunk.Path, chunk.StartLine, chunk.EndLine)
	cloned.Identity = identity
	cloned.EstimatedTokens = chunk.TokenCount
	if cloned.RedundancyKey == "" || cloned.RedundancyKey == previousIdentity {
		cloned.RedundancyKey = identity
	}
	if previousIdentity != "" && previousIdentity != identity {
		cloned = evidence.Merge(cloned, evidence.Descriptor{Links: []evidence.Link{{
			Identity: previousIdentity,
			Relation: "extracted_from",
		}}})
	}
	return &cloned
}

func cloneCandidate(candidate retrieve.Candidate) retrieve.Candidate {
	candidate.Reasons = append([]string(nil), candidate.Reasons...)
	candidate.ScoreDetails = append([]retrieve.ScoreDetail(nil), candidate.ScoreDetails...)
	if candidate.Context != nil {
		cloned := evidence.Merge(*candidate.Context, evidence.Descriptor{})
		candidate.Context = &cloned
	}
	return candidate
}

func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func chunkLineCount(chunk index.Chunk) int {
	if chunk.Text == "" {
		return 0
	}
	return strings.Count(chunk.Text, "\n") + 1
}
