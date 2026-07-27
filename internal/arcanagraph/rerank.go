package arcanagraph

import (
	"context"

	"github.com/Lokee86/grimoire/internal/structure"
)

const (
	maxHybridSemanticCandidates = 2048
	maxHybridGraphCandidates    = 256
)

// RerankedSeed is one bounded Arcana entry point after deterministic fusion.
// Source remains compatible with the existing evaluation labels.
type RerankedSeed struct {
	Node   structure.Node
	Source string
	Score  float64
}

type hybridSeedCandidate struct {
	node          structure.Node
	lexiconRank   int
	semanticRank  int
	semanticScore float64
	identifier    float64
	path          float64
	quality       float64
	agreement     float64
	base          float64
	graph         float64
	final         float64
}

// SemanticCandidateLimit widens vector recall before the final six-seed budget
// is applied. Arcana already scores the full index, so this only retains enough
// ranked metadata for deterministic fusion to rescue conceptually relevant
// declarations that raw vector rank placed outside the production cutoff.
func SemanticCandidateLimit(finalLimit int) int {
	if finalLimit <= 0 {
		return 0
	}
	return min(max(finalLimit*256, 256), maxHybridSemanticCandidates)
}

// RerankSeeds combines identifier, path, declaration quality, Lexicon rank,
// vector similarity, and bounded graph proximity. Graph lookup is advisory: a
// protocol failure returns the deterministic non-graph ranking together with
// the error so callers can preserve retrieval and surface a warning.
func (client Client) RerankSeeds(
	ctx context.Context,
	snapshot string,
	query string,
	lexicon []structure.Node,
	semantic []SemanticSeed,
	limit int,
) ([]RerankedSeed, error) {
	candidates := buildHybridSeedCandidates(query, lexicon, semantic)
	if len(candidates) == 0 || limit <= 0 {
		return nil, nil
	}
	sortHybridSeedCandidates(candidates, false)

	graphPool := candidates
	if len(graphPool) > maxHybridGraphCandidates {
		graphPool = graphPool[:maxHybridGraphCandidates]
	}
	graphScores, graphErr := client.seedGraphProximity(ctx, snapshot, query, graphPool)
	if graphErr == nil {
		for index := range candidates {
			candidates[index].graph = graphScores[hybridSeedKey(candidates[index].node)]
			blended := candidates[index].base*0.58 + candidates[index].graph*0.42
			candidates[index].final = blended * (0.55 + candidates[index].quality*0.45)
		}
		sortHybridSeedCandidates(candidates, true)
	}

	if len(candidates) > limit {
		candidates = candidates[:limit]
	}
	result := make([]RerankedSeed, 0, len(candidates))
	for _, candidate := range candidates {
		source := "lexicon"
		if candidate.semanticRank > 0 {
			source = "vector"
		}
		result = append(result, RerankedSeed{
			Node: candidate.node, Source: source, Score: candidate.final,
		})
	}
	return result, graphErr
}
