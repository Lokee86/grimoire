package app

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/queryshape"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

func TestProviderRetrievalIntentsDropsLowWeightLongContextPass(t *testing.T) {
	intents := []queryshape.RetrievalIntent{
		{Intent: evidence.IntentMixed, Query: "very long original prompt", Weight: 0.15},
		{Intent: evidence.IntentMechanism, Query: "score retrieval evidence", Weight: 1},
	}
	got := providerRetrievalIntents(intents)
	if len(got) != 1 || got[0].Query != "score retrieval evidence" {
		t.Fatalf("provider intents = %+v", got)
	}
}

func TestProviderRetrievalIntentsKeepsShortMixedContextPass(t *testing.T) {
	intents := []queryshape.RetrievalIntent{
		{Intent: evidence.IntentMixed, Query: "short mixed prompt", Weight: 0.4},
		{Intent: evidence.IntentMechanism, Query: "explain retrieval", Weight: 1},
	}
	got := providerRetrievalIntents(intents)
	if len(got) != 2 {
		t.Fatalf("provider intents = %+v", got)
	}
}

func TestIntentLexicalCandidatesPrioritizeDirectLocationPaths(t *testing.T) {
	snapshot := index.Snapshot{Version: index.FormatVersion, Files: []index.FileRecord{
		{Path: "noise/content.go", Chunks: []index.Chunk{{
			ID: "noise", Path: "noise/content.go", StartLine: 1, EndLine: 1,
			Text: strings.Repeat("resolver ", 50), TokenCount: 50,
		}}},
		{Path: "internal/profile/resolver.go", Chunks: []index.Chunk{{
			ID: "resolver", Path: "internal/profile/resolver.go", StartLine: 1, EndLine: 2,
			Text: "package profile\nfunc Build() {}", TokenCount: 6,
		}}},
	}}
	intents := []queryshape.RetrievalIntent{{
		Intent: evidence.IntentDirectLocation, Query: "Where is resolver?", Weight: 1,
	}}

	candidates := intentLexicalCandidates(snapshot, intents, 10)
	if len(candidates) < 2 || candidates[0].Chunk.Path != "internal/profile/resolver.go" {
		t.Fatalf("direct-location path did not lead ranking: %+v", candidates)
	}
	if candidates[0].Context == nil || !containsEvidenceIntent(candidates[0].Context.Intents, evidence.IntentDirectLocation) {
		t.Fatalf("direct-location intent metadata missing: %+v", candidates[0].Context)
	}
}

func TestRankCandidatesForIntentPrefersImplementationForDirectLocation(t *testing.T) {
	planned := queryshape.RetrievalIntent{
		Intent: evidence.IntentDirectLocation,
		Query:  "Where is snapshot freshness validated?",
		Weight: 1,
	}
	candidates := []retrieve.Candidate{
		{
			Chunk:  index.Chunk{ID: "docs", Path: "docs/architecture/prepared-index.md", Text: "snapshot freshness validated", TokenCount: 5},
			Source: "lexical", Score: 55, Rank: 1,
		},
		{
			Chunk:  index.Chunk{ID: "source", Path: "internal/app/vector_manifest.go", Text: "func validateVectorSnapshotManifest() {}", TokenCount: 6},
			Source: "lexical", Score: 38, Rank: 2,
		},
	}
	got := rankCandidatesForIntent(candidates, planned, true)
	if got[0].Chunk.Path != "internal/app/vector_manifest.go" {
		t.Fatalf("implementation did not outrank documentation: %+v", got)
	}
}

func TestRankCandidatesForIntentPenalizesGeneratedEvaluationResults(t *testing.T) {
	planned := queryshape.RetrievalIntent{
		Intent: evidence.IntentMechanism,
		Query:  "explain evaluation reporting",
		Weight: 1,
	}
	candidates := []retrieve.Candidate{
		{
			Chunk:  index.Chunk{ID: "result", Path: "evaluation/results/old-report.json", Text: "evaluation reporting", TokenCount: 5},
			Source: "lexical", Score: 50, Rank: 1,
		},
		{
			Chunk:  index.Chunk{ID: "source", Path: "internal/evaluation/report.go", Text: "func Write() {}", TokenCount: 5},
			Source: "lexical", Score: 35, Rank: 2,
		},
	}
	got := rankCandidatesForIntent(candidates, planned, true)
	if got[0].Chunk.Path != "internal/evaluation/report.go" {
		t.Fatalf("generated evaluation result remained ahead of implementation: %+v", got)
	}
}

func TestMergeIntentCandidateGroupsReservesSpecificCoverage(t *testing.T) {
	mixed := queryshape.RetrievalIntent{Intent: evidence.IntentMixed, Query: "mixed", Weight: 1}
	callChain := queryshape.RetrievalIntent{Intent: evidence.IntentCallChain, Query: "trace calls", Weight: 0.5}
	architecture := queryshape.RetrievalIntent{Intent: evidence.IntentArchitecture, Query: "ownership", Weight: 0.5}

	mixedCandidates := make([]retrieve.Candidate, 10)
	for index := range mixedCandidates {
		mixedCandidates[index] = annotateCandidateIntent(intentTestCandidate(fmt.Sprintf("mixed/%02d.go", index), index+1), mixed)
	}
	callCandidates := []retrieve.Candidate{
		annotateCandidateIntent(intentTestCandidate("call/entry.go", 1), callChain),
		annotateCandidateIntent(intentTestCandidate("call/target.go", 2), callChain),
	}
	architectureCandidates := []retrieve.Candidate{
		annotateCandidateIntent(intentTestCandidate("architecture/owner.go", 1), architecture),
		annotateCandidateIntent(intentTestCandidate("architecture/boundary.go", 2), architecture),
	}

	result := mergeIntentCandidateGroups(12, []intentCandidateGroup{
		{Intent: mixed, Candidates: mixedCandidates},
		{Intent: callChain, Candidates: callCandidates},
		{Intent: architecture, Candidates: architectureCandidates},
	})
	if !pathAppearsWithin(result, "call/entry.go", 8) || !pathAppearsWithin(result, "architecture/owner.go", 8) {
		t.Fatalf("specific intent coverage was not reserved near the front: %+v", result)
	}
}

func TestMergeIntentCandidateGroupsUsesUnseenAnchorPerPass(t *testing.T) {
	shared := intentTestCandidate("docs/shared.md", 1)
	first := queryshape.RetrievalIntent{Intent: evidence.IntentCallChain, Query: "first phase", Weight: 1}
	second := queryshape.RetrievalIntent{Intent: evidence.IntentCallChain, Query: "second phase", Weight: 1}
	result := mergeIntentCandidateGroups(10, []intentCandidateGroup{
		{Intent: first, Candidates: []retrieve.Candidate{
			annotateCandidateIntent(shared, first),
			annotateCandidateIntent(intentTestCandidate("internal/first.go", 2), first),
		}},
		{Intent: second, Candidates: []retrieve.Candidate{
			annotateCandidateIntent(shared, second),
			annotateCandidateIntent(intentTestCandidate("internal/second.go", 2), second),
		}},
	})
	if !pathAppearsWithin(result, "internal/second.go", 3) {
		t.Fatalf("duplicate shared candidate prevented a replacement phase anchor: %+v", result)
	}
}

func TestStructuralRetrievalIntentPrefersCallChainClause(t *testing.T) {
	intents := []queryshape.RetrievalIntent{
		{Intent: evidence.IntentMixed, Query: "Explain ownership and trace calls", Weight: 1},
		{Intent: evidence.IntentArchitecture, Query: "Explain ownership", Weight: 0.5},
		{Intent: evidence.IntentCallChain, Query: "trace calls", Weight: 0.33},
	}
	selected := structuralRetrievalIntent(intents[0].Query, intents)
	if selected.Intent != evidence.IntentCallChain || selected.Query != "trace calls" {
		t.Fatalf("unexpected structural retrieval intent: %+v", selected)
	}
}

func TestAnnotateCandidateIntentMarksSupportingEvidence(t *testing.T) {
	planned := queryshape.RetrievalIntent{Intent: evidence.IntentMechanism, Query: "explain", Weight: 1}
	candidate := annotateCandidateIntent(intentTestCandidate("internal/store/store_test.go", 1), planned)
	if candidate.Context == nil || !containsEvidenceRole(candidate.Context.Roles, evidence.RoleSupporting) {
		t.Fatalf("test candidate was not marked supporting: %+v", candidate.Context)
	}
}

func intentTestCandidate(path string, rank int) retrieve.Candidate {
	return retrieve.Candidate{
		Chunk: index.Chunk{
			ID: path, Path: path, StartLine: rank, EndLine: rank,
			Text: "package test", TokenCount: 10,
		},
		Source: "lexical", Score: float64(100 - rank), Rank: rank,
	}
}

func pathAppearsWithin(candidates []retrieve.Candidate, path string, limit int) bool {
	for index, candidate := range candidates {
		if index >= limit {
			return false
		}
		if candidate.Chunk.Path == path {
			return true
		}
	}
	return false
}

func containsEvidenceIntent(intents []evidence.Intent, target evidence.Intent) bool {
	for _, intent := range intents {
		if intent == target {
			return true
		}
	}
	return false
}

func containsEvidenceRole(roles []evidence.Role, target evidence.Role) bool {
	for _, role := range roles {
		if role == target {
			return true
		}
	}
	return false
}
