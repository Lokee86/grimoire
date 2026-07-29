package agentquery

import (
	"fmt"
	"strings"

	"github.com/Lokee86/grimoire/internal/structure"
)

type relationshipSeed struct {
	Node     structure.Node
	Provider string
	Snapshot string
	Lane     string
	Rank     int
	Score    float64
	Reasons  []string
}

func (engine *Engine) searchRelationshipSeeds(
	exactMatches []Result,
	sourceMatches []Result,
	symbolSeeds []relationshipSeed,
	limit int,
) []relationshipSeed {
	if limit <= 0 {
		return nil
	}

	lanes := [][]relationshipSeed{
		engine.sourceRelationshipSeeds(exactMatches, "exact_matches"),
		symbolSeeds,
		engine.sourceRelationshipSeeds(sourceMatches, "source_matches"),
	}
	indices := make([]int, len(lanes))
	seen := make(map[string]bool)
	seeds := make([]relationshipSeed, 0, limit)
	for len(seeds) < limit {
		progressed := false
		for laneIndex, lane := range lanes {
			for indices[laneIndex] < len(lane) {
				seed := lane[indices[laneIndex]]
				indices[laneIndex]++
				key := relationshipSeedKey(seed.Node)
				if key == "" || seen[key] {
					continue
				}
				seen[key] = true
				seeds = append(seeds, seed)
				progressed = true
				break
			}
			if len(seeds) == limit {
				break
			}
		}
		if !progressed {
			break
		}
	}
	return seeds
}

func (engine *Engine) sourceRelationshipSeeds(results []Result, lane string) []relationshipSeed {
	if engine.lexicon == nil {
		return nil
	}
	seeds := make([]relationshipSeed, 0, len(results))
	for _, result := range results {
		span := result.Node.Span
		if span == nil {
			continue
		}
		resolved := engine.lexicon.ResolveSource(span.Path, span.StartLine, span.EndLine, 1)
		if len(resolved) == 0 {
			continue
		}
		reasons := append([]string(nil), result.Reasons...)
		reasons = append(reasons, fmt.Sprintf("mapped %s source range to containing Lexicon symbol", lane))
		seeds = append(seeds, relationshipSeed{
			Node: resolved[0], Provider: "lexicon", Snapshot: engine.lexiconSnapshot,
			Lane: lane, Rank: result.Rank, Score: result.Score, Reasons: reasons,
		})
	}
	return seeds
}

func relationshipSeedKey(node structure.Node) string {
	if node.Identity != "" {
		return node.Identity
	}
	return strings.Join([]string{node.Path, node.QualifiedName, node.Name, node.Kind}, "\x00")
}
