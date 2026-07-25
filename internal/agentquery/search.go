package agentquery

import (
	"context"

	"github.com/Lokee86/grimoire/internal/retrieve"
)

func (engine *Engine) search(ctx context.Context, request Request, response *Response) error {
	seen := make(map[string]bool)
	add := func(result Result) bool {
		key := handleKey(result.Node.Handle)
		if seen[key] || len(response.Results) >= request.Limit {
			return false
		}
		seen[key] = true
		result.Rank = len(response.Results) + 1
		response.Results = append(response.Results, result)
		return true
	}

	for _, candidate := range retrieve.Exact(engine.source, request.Query, request.Limit) {
		add(Result{
			Provider: "exact", Kind: sourceKind(candidate.Chunk),
			Node: engine.sourceNode(candidate.Chunk), Score: candidate.Score,
			Reasons: append([]string{"literal source match"}, candidate.Reasons...),
		})
	}

	var lexiconSeeds []string
	if engine.lexicon != nil && len(response.Results) < request.Limit {
		for _, match := range engine.lexicon.Find(request.Query, request.Limit) {
			node := engine.node("lexicon", engine.lexiconSnapshot, match.Node)
			add(Result{
				Provider: "lexicon", Kind: match.Node.Kind, Node: node,
				Score: match.Score, Reasons: match.Reasons,
			})
			lexiconSeeds = append(lexiconSeeds, match.Node.Name)
		}
	}

	if engine.arcanaSnapshot != "" && len(response.Results) < request.Limit {
		if len(lexiconSeeds) == 0 {
			lexiconSeeds = []string{request.Query}
		}
		for _, seed := range unique(lexiconSeeds) {
			nodes, err := engine.arcana.Resolve(ctx, engine.arcanaSnapshot, seed, "", request.Limit)
			if err != nil {
				response.Warnings = append(response.Warnings, "Arcana search unavailable: "+err.Error())
				break
			}
			for _, value := range nodes {
				add(Result{
					Provider: "arcana", Kind: value.Kind,
					Node:    engine.node("arcana", engine.arcanaSnapshotID, value),
					Reasons: []string{"Arcana exact graph node for structural anchor " + seed},
				})
			}
		}
	}

	if len(response.Results) < request.Limit {
		for _, candidate := range retrieve.Search(engine.source, request.Query, request.Limit) {
			add(Result{
				Provider: "lexical", Kind: sourceKind(candidate.Chunk),
				Node: engine.sourceNode(candidate.Chunk), Score: candidate.Score,
				Reasons: append([]string{"prepared source BM25 fallback"}, candidate.Reasons...),
			})
		}
	}
	response.Truncated = len(response.Results) == request.Limit
	return nil
}
