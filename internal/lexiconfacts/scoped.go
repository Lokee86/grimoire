package lexiconfacts

import (
	"sort"
	"strings"

	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

// SearchDetailedScoped resolves implementation declarations only inside
// lexical discovery ranges. It does not perform global symbol search or graph
// expansion; Arcana may inspect the resolved seeds afterward.
func (corpus *Corpus) SearchDetailedScoped(
	snapshot index.Snapshot,
	query string,
	scopes []retrieve.Candidate,
	limit int,
) Result {
	if corpus == nil || len(scopes) == 0 || limit <= 0 {
		return Result{}
	}
	terms := queryTerms(query)
	degrees := graphDegrees(corpus.facts.edges)
	scored := make(map[string]scoredNode)
	for scopeIndex, scope := range scopes {
		for _, node := range corpus.facts.nodes {
			if !scopedImplementationNode(node, scope.Chunk) {
				continue
			}
			queryScore, reasons := scoreNode(node, strings.ToLower(query), terms)
			scopeScore := float64((len(scopes) - scopeIndex) * 100)
			entry := scoredNode{
				node: node, score: scopeScore + max(queryScore, 0), primary: true,
				reasons: append([]string{"declaration overlaps lexical discovery range"}, reasons...),
				graph: &evidence.GraphSignals{
					Distance: 0, ModuleProximity: 1, SymbolRole: node.Kind,
					Centrality: normalizedCentrality(degrees[node.ID]),
				},
			}
			if existing, exists := scored[node.ID]; !exists || entry.score > existing.score {
				scored[node.ID] = entry
			}
		}
	}
	if len(scored) == 0 {
		return Result{}
	}
	seeds := make([]scoredNode, 0, len(scored))
	for _, entry := range scored {
		seeds = append(seeds, entry)
	}
	sort.Slice(seeds, func(left, right int) bool {
		if seeds[left].score != seeds[right].score {
			return seeds[left].score > seeds[right].score
		}
		leftSize, rightSize := nodeSpanSize(seeds[left].node), nodeSpanSize(seeds[right].node)
		if leftSize != rightSize {
			return leftSize < rightSize
		}
		return seeds[left].node.ID < seeds[right].node.ID
	})
	seedLimit := min(limit, 12)
	if len(seeds) > seedLimit {
		seeds = seeds[:seedLimit]
	}
	selected := make(map[string]scoredNode, len(seeds))
	for _, seed := range seeds {
		selected[seed.node.ID] = seed
	}
	return Result{
		Candidates: chunksForNodes(snapshot, selected, limit),
		Evidence:   evidenceForSeeds(seeds, corpus.facts, min(limit, 8)),
		Seeds:      seedNodes(seeds, min(limit, 8)),
	}
}

func scopedImplementationNode(node Node, chunk index.Chunk) bool {
	if !localNode(node) || node.Span == nil || symbolKindWeight(node.Kind) <= 0 {
		return false
	}
	if strings.ReplaceAll(node.Span.Path, "\\", "/") != strings.ReplaceAll(chunk.Path, "\\", "/") {
		return false
	}
	end := node.Span.EndLine
	if end < node.Span.StartLine {
		end = node.Span.StartLine
	}
	return end >= chunk.StartLine && node.Span.StartLine <= chunk.EndLine
}

func nodeSpanSize(node Node) int {
	if node.Span == nil {
		return int(^uint(0) >> 1)
	}
	return max(node.Span.EndLine-node.Span.StartLine, 0)
}
