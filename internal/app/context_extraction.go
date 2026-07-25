package app

import (
	"github.com/Lokee86/grimoire/internal/extraction"
	"github.com/Lokee86/grimoire/internal/queryshape"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

func extractContextCandidates(
	query string,
	intents []queryshape.RetrievalIntent,
	candidates []retrieve.Candidate,
) ([]retrieve.Candidate, error) {
	facetQueries := make(map[string]string)
	for _, planned := range intents {
		if planned.FacetID == "" || planned.Query == "" {
			continue
		}
		facetQueries[planned.FacetID] = planned.Query
	}
	config := extraction.DefaultConfig()
	config.MinSpans = 2
	extractor := extraction.New(
		config,
		extraction.NewLanguageDiscoverer(),
	)
	return extractor.Refine(extraction.Request{
		Query:        query,
		FacetQueries: facetQueries,
	}, candidates)
}
