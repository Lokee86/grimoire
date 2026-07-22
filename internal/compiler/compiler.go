package compiler

import "github.com/Lokee86/grimoire/internal/retrieve"

const PackageVersion = 1

type Package struct {
	Version          int         `json:"version"`
	Query            string      `json:"query"`
	Budget           int         `json:"budget"`
	EstimatedTokens  int         `json:"estimated_tokens"`
	IndexVersion     int         `json:"index_version"`
	RetrievalSources []string    `json:"retrieval_sources"`
	Selections       []Selection `json:"selections"`
	OmittedForBudget int         `json:"omitted_for_budget"`
}

type Selection struct {
	Path            string   `json:"path"`
	StartLine       int      `json:"start_line"`
	EndLine         int      `json:"end_line"`
	Score           int      `json:"score"`
	Reasons         []string `json:"reasons"`
	EstimatedTokens int      `json:"estimated_tokens"`
	Content         string   `json:"content"`
}

func Compile(query string, budget int, indexVersion int, candidates []retrieve.Candidate) Package {
	result := Package{
		Version:          PackageVersion,
		Query:            query,
		Budget:           budget,
		IndexVersion:     indexVersion,
		RetrievalSources: []string{"lexical"},
		Selections:       make([]Selection, 0),
	}

	for _, candidate := range candidates {
		cost := candidate.Chunk.EstimatedTokens
		if cost > budget-result.EstimatedTokens {
			result.OmittedForBudget++
			continue
		}
		result.Selections = append(result.Selections, Selection{
			Path:            candidate.Chunk.Path,
			StartLine:       candidate.Chunk.StartLine,
			EndLine:         candidate.Chunk.EndLine,
			Score:           candidate.Score,
			Reasons:         candidate.Reasons,
			EstimatedTokens: cost,
			Content:         candidate.Chunk.Text,
		})
		result.EstimatedTokens += cost
	}
	return result
}
