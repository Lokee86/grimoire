package evaluation

import (
	"strings"
	"testing"
)

func TestBuildCandidateDiagnosticsTracksStageMovement(t *testing.T) {
	entry := Case{
		Required: []Evidence{{Path: "internal/owner.go", Symbols: []string{"Own"}}},
	}
	owner := Candidate{
		Path: "internal/owner.go", StartLine: 10, EndLine: 20, Text: "func Own() {}",
		RetrievalSource: "lexical", ProviderRank: 2, Score: 14,
		ScoreDetails: []ScoreDetail{{Name: "content matches owner", Value: 14}},
	}
	stages := Stages{
		Retrieved: []Candidate{{Path: "README.md", StartLine: 1, EndLine: 4}, owner},
		Exact: []Candidate{{
			Path: "internal/owner.go", StartLine: 10, EndLine: 20, Text: "func Own() {}",
			RetrievalSource: "exact", ProviderRank: 1, Score: 70,
			ScoreDetails: []ScoreDetail{{Name: "identifier Own matches content", Value: 70}},
		}},
		Merged:    []Candidate{owner},
		Curated:   []Candidate{owner},
		Assembled: []Candidate{owner},
		Included:  []Candidate{owner},
	}

	diagnostics := BuildCandidateDiagnostics(entry, stages)
	var found *CandidateDiagnostic
	for index := range diagnostics {
		if diagnostics[index].Path == "internal/owner.go" {
			found = &diagnostics[index]
			break
		}
	}
	if found == nil {
		t.Fatal("required candidate diagnostic was not produced")
	}
	if !found.Required || found.Retrieved == nil || found.Retrieved.Rank != 2 ||
		found.Exact == nil || found.Exact.Rank != 1 || found.Merged == nil ||
		found.Curated == nil || found.Assembled == nil || found.Included == nil {
		t.Fatalf("stage movement was not preserved: %+v", *found)
	}
	if len(found.Retrieved.ScoreDetails) != 1 || found.Retrieved.ScoreDetails[0].Value != 14 {
		t.Fatalf("score attribution was not preserved: %+v", found.Retrieved)
	}
}

func TestMarkdownIncludesCandidateScoreAttribution(t *testing.T) {
	report := Report{
		Repository: "fixture",
		Variant:    "diagnostic",
		Runs: []CaseRun{{
			CaseID: "case-1", Query: "find Own", Mode: "lexical",
			CandidateDiagnostics: []CandidateDiagnostic{{
				Path: "internal/owner.go", StartLine: 10, EndLine: 20, Required: true,
				Retrieved: &CandidateStageDiagnostic{
					Rank: 2, RetrievalSource: "lexical", ProviderRank: 2, Score: 14,
					ScoreDetails: []ScoreDetail{{Name: "content matches owner", Value: 14}},
				},
			}},
		}},
	}
	markdown := Markdown(report)
	for _, expected := range []string{
		"## Candidate score attribution",
		"internal/owner.go:10-20",
		"content matches owner=14.000",
	} {
		if !strings.Contains(markdown, expected) {
			t.Fatalf("missing %q in diagnostic report:\n%s", expected, markdown)
		}
	}
}
