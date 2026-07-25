package evaluation

import (
	"strings"
	"testing"
)

func TestAggregateRunsTracksKnownFinalPackageCollapse(t *testing.T) {
	statuses := make([]EvidenceStatus, 45)
	for index := range statuses {
		statuses[index].Indexed = true
		statuses[index].Retrieved = index < 35
		statuses[index].Merged = index < 34
		statuses[index].Curated = index < 34
		statuses[index].Assembled = index < 34
		statuses[index].Included = index == 0
	}

	aggregate := AggregateRuns("vector", []CaseRun{{Required: statuses}})
	funnel := aggregate.RequiredEvidenceFunnel
	if funnel.Total != 45 || funnel.Retrieved != 35 || funnel.Merged != 34 ||
		funnel.Curated != 34 || funnel.Assembled != 34 || funnel.Included != 1 {
		t.Fatalf("unexpected required evidence funnel: %+v", funnel)
	}
	if aggregate.RequiredEvidenceRecall != 1.0/45.0 {
		t.Fatalf("required evidence recall = %v, want %v", aggregate.RequiredEvidenceRecall, 1.0/45.0)
	}
}

func TestMarkdownReportsRequiredEvidenceFunnelAndAssembledDiagnostics(t *testing.T) {
	statuses := make([]EvidenceStatus, 45)
	for index := range statuses {
		statuses[index].Indexed = true
		statuses[index].Retrieved = index < 35
		statuses[index].Merged = index < 34
		statuses[index].Curated = index < 34
		statuses[index].Assembled = index < 34
		statuses[index].Included = index == 0
	}
	report := Report{
		Repository: "fixture",
		Runs: []CaseRun{{
			CaseID:   "package-collapse",
			Mode:     "vector",
			Required: statuses,
			CandidateDiagnostics: []CandidateDiagnostic{{
				Path:      "internal/target.go",
				Required:  true,
				Curated:   &CandidateStageDiagnostic{Rank: 2},
				Assembled: &CandidateStageDiagnostic{Rank: 3},
			}},
		}},
	}

	markdown := Markdown(report)
	for _, expected := range []string{
		"## Required evidence funnel",
		"| vector | 45 | 35 (77.8%) | 34 (75.6%) | 34 (75.6%) | 34 (75.6%) | 1 (2.2%) |",
		"| Evidence | Candidate | Retrieved attribution | Exact attribution | Merged | Curated | Assembled | Included |",
	} {
		if !strings.Contains(markdown, expected) {
			t.Fatalf("Markdown missing %q:\n%s", expected, markdown)
		}
	}
}
