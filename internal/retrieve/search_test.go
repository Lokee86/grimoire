package retrieve

import (
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/index"
)

func TestSearchRanksFilenameAndContentDeterministically(t *testing.T) {
	snapshot := index.Snapshot{
		Version: index.FormatVersion,
		Files: []index.FileRecord{
			{Path: "internal/damage/resolver.go", Chunks: []index.Chunk{{
				ID: "a", Path: "internal/damage/resolver.go", StartLine: 1, EndLine: 2,
				TokenCount: 5, Text: "package damage\nfunc ResolveDamage() {}",
			}}},
			{Path: "internal/player/state.go", Chunks: []index.Chunk{{
				ID: "b", Path: "internal/player/state.go", StartLine: 1, EndLine: 2,
				TokenCount: 5, Text: "package player\n// damage is recorded here",
			}}},
		},
	}

	results := Search(snapshot, "damage resolver", 10)
	if len(results) != 2 {
		t.Fatalf("expected two results, got %d", len(results))
	}
	if results[0].Chunk.Path != "internal/damage/resolver.go" {
		t.Fatalf("unexpected first result: %+v", results[0])
	}
	if results[0].Score <= results[1].Score {
		t.Fatalf("expected strict score ordering: %+v", results)
	}
	if results[0].Source != "lexical" || results[0].Rank != 1 || results[1].Rank != 2 {
		t.Fatalf("unexpected lexical provenance: %+v", results)
	}
	var attributed float64
	for _, detail := range results[0].ScoreDetails {
		attributed += detail.Value
	}
	if len(results[0].ScoreDetails) == 0 || attributed != results[0].Score {
		t.Fatalf("lexical score is not fully attributed: %+v", results[0])
	}
}

func TestSearchUsesStablePathTieBreak(t *testing.T) {
	snapshot := index.Snapshot{Version: index.FormatVersion, Files: []index.FileRecord{
		{Path: "b.go", Chunks: []index.Chunk{{Path: "b.go", StartLine: 1, Text: "needle", TokenCount: 1}}},
		{Path: "a.go", Chunks: []index.Chunk{{Path: "a.go", StartLine: 1, Text: "needle", TokenCount: 1}}},
	}}
	results := Search(snapshot, "needle", 10)
	if results[0].Chunk.Path != "a.go" {
		t.Fatalf("tie break was not stable: %+v", results)
	}
}

func TestSearchDoesNotSubstringMatchSingleTerm(t *testing.T) {
	snapshot := index.Snapshot{Version: index.FormatVersion, Files: []index.FileRecord{{
		Path:   "command.go",
		Chunks: []index.Chunk{{Path: "command.go", StartLine: 1, Text: "package command", TokenCount: 2}},
	}}}
	if results := Search(snapshot, "and", 10); len(results) != 0 {
		t.Fatalf("single term matched inside a larger token: %+v", results)
	}
}

func TestSearchBM25RewardsRareMultiTermEvidence(t *testing.T) {
	snapshot := index.Snapshot{Version: index.FormatVersion, Files: []index.FileRecord{
		{Path: "alpha/common.go", Chunks: []index.Chunk{{Path: "alpha/common.go", StartLine: 1, Text: "package alpha\ncache cache cache cache", TokenCount: 5}}},
		{Path: "worker/state.go", Chunks: []index.Chunk{{Path: "worker/state.go", StartLine: 1, Text: "package worker\ncache invalidation sentinel", TokenCount: 4}}},
	}}

	results := Search(snapshot, "cache sentinel", 10)
	if len(results) != 2 {
		t.Fatalf("expected two results, got %d", len(results))
	}
	if results[0].Chunk.Path != "worker/state.go" {
		t.Fatalf("rare multi-term evidence did not lead ranking: %+v", results)
	}
}

func TestSearchBM25NormalizesDocumentLength(t *testing.T) {
	noise := strings.Repeat(" unrelated", 100)
	snapshot := index.Snapshot{Version: index.FormatVersion, Files: []index.FileRecord{
		{Path: "short.go", Chunks: []index.Chunk{{Path: "short.go", StartLine: 1, Text: "critical path", TokenCount: 2}}},
		{Path: "long.go", Chunks: []index.Chunk{{Path: "long.go", StartLine: 1, Text: "critical path" + noise, TokenCount: 102}}},
	}}

	results := Search(snapshot, "critical path", 10)
	if len(results) != 2 || results[0].Chunk.Path != "short.go" {
		t.Fatalf("focused document did not outrank long noisy document: %+v", results)
	}
}

func TestSearchSplitsCodeIdentifiers(t *testing.T) {
	snapshot := index.Snapshot{Version: index.FormatVersion, Files: []index.FileRecord{{
		Path:   "network/handler.go",
		Chunks: []index.Chunk{{Path: "network/handler.go", StartLine: 1, Text: "func ResolveDamagePacket() {}", TokenCount: 3}},
	}}}

	results := Search(snapshot, "resolve damage packet", 10)
	if len(results) != 1 {
		t.Fatalf("expected identifier components to match: %+v", results)
	}
	for _, term := range []string{"resolve", "damage", "packet"} {
		found := false
		for _, detail := range results[0].ScoreDetails {
			if detail.Name == "BM25 content matches "+term {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing BM25 attribution for %q: %+v", term, results[0].ScoreDetails)
		}
	}
}

func TestNearestDeclarationAliasFindsCodeFacingIdentifier(t *testing.T) {
	vocabulary := map[string]declarationVocabularyEntry{
		"validate": {documentFrequency: 2},
		"value":    {documentFrequency: 4},
		"visitor":  {documentFrequency: 1},
	}
	alias, ok := nearestDeclarationAlias("validation", vocabulary)
	if !ok || alias.token != "validate" {
		t.Fatalf("nearest alias = %+v, %v; want validate", alias, ok)
	}
}

func TestNearestDeclarationAliasRejectsWeakPrefixMatch(t *testing.T) {
	vocabulary := map[string]declarationVocabularyEntry{
		"server": {documentFrequency: 1},
		"settle": {documentFrequency: 1},
	}
	if alias, ok := nearestDeclarationAlias("selection", vocabulary); ok {
		t.Fatalf("weak alias unexpectedly accepted: %+v", alias)
	}
}

func TestSearchDeclarationAliasPromotesMatchingDeclaration(t *testing.T) {
	snapshot := index.Snapshot{Version: index.FormatVersion, Files: []index.FileRecord{
		{Path: "owner.go", Chunks: []index.Chunk{{
			Path: "owner.go", StartLine: 1, Text: "func ValidateSnapshot() {}", TokenCount: 3,
		}}},
		{Path: "notes.go", Chunks: []index.Chunk{{
			Path: "notes.go", StartLine: 1, Text: "// validation behavior", TokenCount: 3,
		}}},
	}}
	results := SearchWithConfig(snapshot, "snapshot validation", 10, Config{DeclarationAliasBonus: 8})
	if len(results) != 2 || results[0].Chunk.Path != "owner.go" {
		t.Fatalf("declaration alias did not promote ValidateSnapshot: %+v", results)
	}
	found := false
	for _, detail := range results[0].ScoreDetails {
		if detail.Name == "declaration alias validation -> validate" && detail.Value > 0 {
			found = true
		}
	}
	if !found {
		t.Fatalf("declaration alias score missing: %+v", results[0].ScoreDetails)
	}
}

func TestSearchLegacyConfigDoesNotUseDeclarationAliases(t *testing.T) {
	snapshot := index.Snapshot{Version: index.FormatVersion, Files: []index.FileRecord{{
		Path: "owner.go", Chunks: []index.Chunk{{
			Path: "owner.go", StartLine: 1, Text: "func ValidateSnapshot() {}", TokenCount: 3,
		}},
	}}}
	if results := SearchWithConfig(snapshot, "validation", 10, LegacyConfig()); len(results) != 0 {
		t.Fatalf("legacy search unexpectedly used a declaration alias: %+v", results)
	}
}

func TestSearchLegacyConfigKeepsFixedFieldBonus(t *testing.T) {
	snapshot := index.Snapshot{Version: index.FormatVersion, Files: []index.FileRecord{{
		Path: "server.go", Chunks: []index.Chunk{{Path: "server.go", StartLine: 1, Text: "server sentinel", TokenCount: 2}},
	}}}
	results := SearchWithConfig(snapshot, "server sentinel", 10, LegacyConfig())
	if len(results) != 1 {
		t.Fatalf("expected one result, got %+v", results)
	}
	for _, detail := range results[0].ScoreDetails {
		if detail.Name == "filename matches server" && detail.Value != 8 {
			t.Fatalf("legacy filename bonus = %.3f, want 8", detail.Value)
		}
	}
}

func TestQueryTermsSuppressesPromptScaffolding(t *testing.T) {
	terms := queryTerms("Find where the damage resolver is")
	if strings.Join(terms, ",") != "damage,resolver" {
		t.Fatalf("unexpected normalized query terms: %v", terms)
	}
}

func TestQueryTermsSplitCamelCase(t *testing.T) {
	terms := queryTerms("ResolveDamagePacket")
	if strings.Join(terms, ",") != "resolvedamagepacket,resolve,damage,packet" {
		t.Fatalf("unexpected identifier terms: %v", terms)
	}
}

func TestQueryTermsRetainAllStopwordIdentifier(t *testing.T) {
	terms := queryTerms("Show")
	if strings.Join(terms, ",") != "show" {
		t.Fatalf("single identifier was suppressed as prompt scaffolding: %v", terms)
	}
}
