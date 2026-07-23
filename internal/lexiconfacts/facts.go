package lexiconfacts

import (
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

const source = "lexicon"

// Search loads every exported Lexicon JSONL library in directory and converts
// matching symbols plus their immediate relationships into prepared chunks.
func Search(snapshot index.Snapshot, query, directory string, limit int) ([]retrieve.Candidate, error) {
	result, err := SearchDetailed(snapshot, query, directory, limit)
	return result.Candidates, err
}

// SearchDetailed returns the source candidates, first-class structural
// evidence, and directly matched symbols from one immutable Lexicon export.
func SearchDetailed(snapshot index.Snapshot, query, directory string, limit int) (Result, error) {
	corpus, err := Load(directory)
	if err != nil || corpus == nil {
		return Result{}, err
	}
	return corpus.SearchDetailed(snapshot, query, limit), nil
}

func (corpus *Corpus) Search(snapshot index.Snapshot, query string, limit int) []retrieve.Candidate {
	return corpus.SearchDetailed(snapshot, query, limit).Candidates
}

func (corpus *Corpus) SearchDetailed(snapshot index.Snapshot, query string, limit int) Result {
	if corpus == nil || limit <= 0 {
		return Result{}
	}
	terms := queryTerms(query)
	if len(terms) == 0 {
		return Result{}
	}
	seeds := rankNodes(corpus.facts.nodes, query, terms)
	if len(seeds) == 0 {
		return Result{}
	}
	seedLimit := min(limit, 48)
	if len(seeds) > seedLimit {
		seeds = seeds[:seedLimit]
	}
	scored := make(map[string]scoredNode, len(seeds)*2)
	for _, seed := range seeds {
		scored[seed.node.ID] = seed
	}
	expandRelationships(scored, seeds, corpus.facts)
	return Result{
		Candidates: chunksForNodes(snapshot, scored, limit),
		Evidence:   evidenceForSeeds(seeds, corpus.facts, min(limit, 8)),
		Seeds:      seedNodes(seeds, min(limit, 8)),
	}
}
