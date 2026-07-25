package graphrank

import (
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

func TestRerankPinsExactAndPromotesStrongGraphCandidate(t *testing.T) {
	candidates := []retrieve.Candidate{
		{Chunk: index.Chunk{ID: "exact"}, Source: "exact", Rank: 1},
		{Chunk: index.Chunk{ID: "base"}, Source: "lexical", Rank: 1},
		{
			Chunk: index.Chunk{ID: "graph"}, Source: "lexicon", Rank: 5,
			Context: &evidence.Descriptor{Graph: &evidence.GraphSignals{
				Distance: 1, Relations: []string{"outgoing:calls"}, ModuleProximity: 1,
				SymbolRole: "function", Centrality: 1,
			}},
		},
	}

	got := RerankWithConfig(candidates, evidence.IntentCallChain, BoundedConfig())
	if len(got) != 3 || got[0].Chunk.ID != "exact" || got[1].Chunk.ID != "graph" || got[2].Chunk.ID != "base" {
		t.Fatalf("Rerank() order = %+v", candidateIDs(got))
	}
	if len(got[1].ScoreDetails) != 0 || len(got[1].GraphScoreDetails) != 5 {
		t.Fatalf("graph diagnostics polluted retrieval score details: %+v %+v", got[1].ScoreDetails, got[1].GraphScoreDetails)
	}
	for index, candidate := range got {
		if candidate.Rank != index+1 {
			t.Fatalf("candidate %s rank = %d, want %d", candidate.Chunk.ID, candidate.Rank, index+1)
		}
	}
}

func TestScoreUsesIntentSpecificRelationWeights(t *testing.T) {
	descriptor := &evidence.Descriptor{Graph: &evidence.GraphSignals{
		Distance: 1, Relations: []string{"outgoing:calls"},
	}}
	callChain := scoreDetailValue(Score(descriptor, evidence.IntentCallChain), "graph relation")
	architecture := scoreDetailValue(Score(descriptor, evidence.IntentArchitecture), "graph relation")
	if callChain != 6 || architecture != 2 {
		t.Fatalf("relation scores = call-chain %.1f architecture %.1f", callChain, architecture)
	}
}

func TestScoreBoundsCentralityContribution(t *testing.T) {
	descriptor := &evidence.Descriptor{Graph: &evidence.GraphSignals{Centrality: 9}}
	if got := scoreDetailValue(Score(descriptor, evidence.IntentMechanism), "graph weak centrality"); got != 1.5 {
		t.Fatalf("centrality score = %.1f, want 1.5", got)
	}
}

func TestRerankDefaultShadowModeDoesNotPromoteRelationships(t *testing.T) {
	candidates := []retrieve.Candidate{
		{Chunk: index.Chunk{ID: "first"}, Source: "lexical", Rank: 1},
		{
			Chunk: index.Chunk{ID: "related"}, Source: "lexicon", Rank: 2,
			Context: &evidence.Descriptor{Graph: &evidence.GraphSignals{
				Distance: 1, Relations: []string{"outgoing:calls"}, ModuleProximity: 1,
				SymbolRole: "function", Centrality: 1,
			}},
		},
	}
	got := Rerank(candidates, evidence.IntentCallChain)
	if got[0].Chunk.ID != "first" || got[1].Chunk.ID != "related" || len(got[1].GraphScoreDetails) != 5 {
		t.Fatalf("shadow ranking changed order or omitted diagnostics: %+v", got)
	}
}

func TestRerankDirectSeedAnnotatesWithoutReorderingOrChangingScore(t *testing.T) {
	candidates := []retrieve.Candidate{
		{Chunk: index.Chunk{ID: "first"}, Source: "lexical", Rank: 1, Score: 10},
		{
			Chunk: index.Chunk{ID: "direct"}, Source: "lexicon", Rank: 2, Score: 4,
			Context: &evidence.Descriptor{Graph: &evidence.GraphSignals{
				Distance: 0, ModuleProximity: 1, SymbolRole: "function", Centrality: 1,
			}},
		},
	}
	got := Rerank(candidates, evidence.IntentMechanism)
	if got[0].Chunk.ID != "first" || got[1].Chunk.ID != "direct" || got[1].Score != 4 || len(got[1].ScoreDetails) != 0 {
		t.Fatalf("direct seed changed source retrieval: %+v", got)
	}
	if len(got[1].GraphScoreDetails) == 0 {
		t.Fatal("direct seed omitted graph diagnostics")
	}
}

func TestRerankWithoutGraphSignalsPreservesProviderRanks(t *testing.T) {
	candidates := []retrieve.Candidate{
		{Chunk: index.Chunk{ID: "first"}, Source: "lexical", Rank: 1},
		{Chunk: index.Chunk{ID: "second"}, Source: "lexical", Rank: 2},
	}
	got := Rerank(candidates, evidence.IntentMechanism)
	if got[0].Chunk.ID != "first" || got[1].Chunk.ID != "second" {
		t.Fatalf("Rerank() changed non-graph order: %+v", candidateIDs(got))
	}
}

func candidateIDs(candidates []retrieve.Candidate) []string {
	result := make([]string, len(candidates))
	for index, candidate := range candidates {
		result[index] = candidate.Chunk.ID
	}
	return result
}

func scoreDetailValue(details []retrieve.ScoreDetail, prefix string) float64 {
	for _, detail := range details {
		if strings.HasPrefix(detail.Name, prefix) {
			return detail.Value
		}
	}
	return 0
}
