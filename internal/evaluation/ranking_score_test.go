package evaluation

import (
	"math"
	"testing"
)

func TestScoreRankingMeasuresRequiredEvidenceOrder(t *testing.T) {
	required := []Evidence{
		{Path: "first.go", Symbols: []string{"First"}},
		{Path: "second.go", Symbols: []string{"Second"}},
		{Path: "third.go", Symbols: []string{"Third"}},
	}
	supporting := []Evidence{{Path: "support.go"}}
	candidates := make([]Candidate, 25)
	for index := range candidates {
		candidates[index] = Candidate{Path: "noise.go"}
	}
	candidates[1] = Candidate{Path: "support.go"}
	candidates[4] = Candidate{Path: "first.go", Text: "func First() {}"}
	candidates[14] = Candidate{Path: "second.go", Text: "func Second() {}"}
	candidates[24] = Candidate{Path: "third.go", Text: "func Third() {}"}

	metrics := ScoreRanking(Case{Required: required, Supporting: supporting}, candidates)
	assertClose(t, metrics.RequiredRecallAt10, 1.0/3.0)
	assertClose(t, metrics.RequiredRecallAt20, 2.0/3.0)
	assertClose(t, metrics.SupportingRecallAt10, 1)
	assertClose(t, metrics.ReciprocalRank, 0.2)
	assertClose(t, metrics.RelevantRateAt10, 0.2)
	assertClose(t, metrics.RelevantRateAt20, 0.15)
	if metrics.FirstRequiredRank != 5 {
		t.Fatalf("first required rank = %d, want 5", metrics.FirstRequiredRank)
	}
}

func TestEvidenceRankAllowsSymbolsAcrossChunks(t *testing.T) {
	evidence := Evidence{Path: "owner.go", Symbols: []string{"Alpha", "Beta"}}
	candidates := []Candidate{
		{Path: "owner.go", Text: "func Alpha() {}"},
		{Path: "owner.go", Text: "func Beta() {}"},
	}
	if got := evidenceRank(evidence, candidates); got != 2 {
		t.Fatalf("evidence rank = %d, want 2", got)
	}
}

func assertClose(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.000001 {
		t.Fatalf("got %v, want %v", got, want)
	}
}
