package arcanagraph

import (
	"math"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Lokee86/grimoire/internal/structure"
)

func buildHybridSeedCandidates(
	query string,
	lexicon []structure.Node,
	semantic []SemanticSeed,
) []hybridSeedCandidate {
	byKey := make(map[string]*hybridSeedCandidate, len(lexicon)+len(semantic))
	order := make([]string, 0, len(lexicon)+len(semantic))
	for index, node := range lexicon {
		if strings.TrimSpace(node.Name) == "" {
			continue
		}
		key := hybridSeedKey(node)
		candidate := byKey[key]
		if candidate == nil {
			candidate = &hybridSeedCandidate{node: node}
			byKey[key] = candidate
			order = append(order, key)
		}
		if candidate.lexiconRank == 0 {
			candidate.lexiconRank = index + 1
		}
	}
	for index, ranked := range semantic {
		node := ranked.Node
		if strings.TrimSpace(node.Name) == "" {
			continue
		}
		key := hybridSeedKey(node)
		candidate := byKey[key]
		if candidate == nil {
			candidate = &hybridSeedCandidate{node: node}
			byKey[key] = candidate
			order = append(order, key)
		}
		rank := ranked.Rank
		if rank <= 0 {
			rank = index + 1
		}
		if candidate.semanticRank == 0 || rank < candidate.semanticRank {
			candidate.semanticRank = rank
			candidate.semanticScore = ranked.Score
			candidate.node = node
		}
	}

	semanticMin, semanticMax := semanticScoreBounds(semantic)
	semanticPopulation := semanticRankPopulation(semantic)
	queryText := normalizedHybridText(query)
	queryTokens := meaningfulHybridTokens(queryText)
	result := make([]hybridSeedCandidate, 0, len(order))
	seen := make(map[string]struct{}, len(order))
	for _, key := range order {
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		candidate := *byKey[key]
		candidate.identifier = hybridIdentifierScore(queryText, queryTokens, candidate.node)
		candidate.path = hybridPathScore(queryTokens, candidate.node.Path)
		candidate.quality = hybridDeclarationQuality(candidate.node)
		if candidate.lexiconRank > 0 && candidate.semanticRank > 0 {
			candidate.agreement = 1
		}
		semanticSignal := hybridSemanticSignal(candidate, semanticPopulation, semanticMin, semanticMax)
		lexiconSignal := 0.0
		if candidate.lexiconRank > 0 {
			lexiconSignal = 1 / math.Sqrt(float64(candidate.lexiconRank))
		}
		candidate.base = semanticSignal*0.40 +
			candidate.identifier*0.20 +
			candidate.path*0.08 +
			lexiconSignal*0.16 +
			candidate.quality*0.08 +
			candidate.agreement*0.08
		candidate.final = candidate.base
		result = append(result, candidate)
	}
	return result
}

func semanticRankPopulation(seeds []SemanticSeed) int {
	population := len(seeds)
	for _, seed := range seeds {
		population = max(population, seed.Rank)
	}
	return population
}

func semanticScoreBounds(seeds []SemanticSeed) (float64, float64) {
	if len(seeds) == 0 {
		return 0, 0
	}
	minimum, maximum := seeds[0].Score, seeds[0].Score
	for _, seed := range seeds[1:] {
		minimum = min(minimum, seed.Score)
		maximum = max(maximum, seed.Score)
	}
	return minimum, maximum
}

func hybridSemanticSignal(
	candidate hybridSeedCandidate,
	count int,
	minimum, maximum float64,
) float64 {
	if candidate.semanticRank <= 0 || count <= 0 {
		return 0
	}
	percentile := 1.0
	if count > 1 {
		percentile = 1 - float64(candidate.semanticRank-1)/float64(count-1)
	}
	similarity := 1.0
	if maximum > minimum {
		similarity = (candidate.semanticScore - minimum) / (maximum - minimum)
	}
	return clampHybridScore(similarity*0.65 + percentile*0.35)
}

func sortHybridSeedCandidates(candidates []hybridSeedCandidate, graph bool) {
	sort.SliceStable(candidates, func(left, right int) bool {
		leftScore, rightScore := candidates[left].base, candidates[right].base
		if graph {
			leftScore, rightScore = candidates[left].final, candidates[right].final
		}
		if leftScore != rightScore {
			return leftScore > rightScore
		}
		if candidates[left].base != candidates[right].base {
			return candidates[left].base > candidates[right].base
		}
		leftSemantic := rankOrMax(candidates[left].semanticRank)
		rightSemantic := rankOrMax(candidates[right].semanticRank)
		if leftSemantic != rightSemantic {
			return leftSemantic < rightSemantic
		}
		leftLexicon := rankOrMax(candidates[left].lexiconRank)
		rightLexicon := rankOrMax(candidates[right].lexiconRank)
		if leftLexicon != rightLexicon {
			return leftLexicon < rightLexicon
		}
		leftPath := filepath.ToSlash(candidates[left].node.Path)
		rightPath := filepath.ToSlash(candidates[right].node.Path)
		if leftPath != rightPath {
			return leftPath < rightPath
		}
		return candidates[left].node.Name < candidates[right].node.Name
	})
}

func hybridSeedKey(node structure.Node) string {
	path := filepath.ToSlash(filepath.Clean(strings.TrimSpace(node.Path)))
	return strings.ToLower(strings.TrimSpace(node.Name)) + "\x00" + strings.ToLower(path)
}

func rankOrMax(rank int) int {
	if rank > 0 {
		return rank
	}
	return int(^uint(0) >> 1)
}

func clampHybridScore(value float64) float64 {
	return min(max(value, 0), 1)
}
