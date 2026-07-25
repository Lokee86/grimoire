package agentdiscovery

import (
	"encoding/json"
	"strings"
	"testing"
)

func syntheticCase() Case {
	return Case{
		ID: "case", Task: "task", Ownership: "owner", OwnershipEvidence: []Evidence{{Path: "src/owner.go", Symbols: []string{"Owner"}}},
		Required: []Evidence{{Path: "src/owner.go", Symbols: []string{"Owner"}}}, Structural: []Evidence{{Path: "docs/contract.md", Symbols: []string{"Contract"}}},
		Forbidden: []Forbidden{{ID: "invented", Pattern: "client owns server authority"}}, RelevantBranches: []string{"src", "docs"},
	}
}

func TestScoreMetricsAndRepeatedEvidence(t *testing.T) {
	entry := syntheticCase()
	transcript := Transcript{Adapter: "progressive-jsonl", RunID: "one", CaseID: entry.ID, Events: []Event{
		{TimeMS: 5, Kind: "source_open", Path: "src/owner.go", Symbol: "Owner", InputTokens: 100, InputText: "open owner"},
		{TimeMS: 6, Kind: "source_open", Path: "src/owner.go", Symbol: "Owner", InputTokens: 100, InputText: "open owner"},
		{TimeMS: 8, Kind: "source_open", Path: "docs/contract.md", Symbol: "Contract", InputTokens: 20},
		{TimeMS: 9, Kind: "source_open", Path: "noise/guess.go", InputTokens: 5},
	}}
	score := ScoreTranscript(entry, transcript)
	if score.TotalInputTokens != 225 || score.RepeatedInputTokens != 100 {
		t.Fatalf("token accounting = %+v", score)
	}
	if score.DiscoveryCalls != 4 || score.ToolCalls != 4 || len(score.Opened) != 3 {
		t.Fatalf("call or range accounting = %+v", score)
	}
	if score.FirstOwnershipMS != 5 || score.EvidenceCompleteMS != 8 || !score.Correct {
		t.Fatalf("evidence scoring = %+v", score)
	}
	if got := strings.Join(score.IrrelevantBranches, ","); got != "noise" {
		t.Fatalf("irrelevant branches = %q", got)
	}
}

func TestUnsupportedConclusionBlocksCorrectness(t *testing.T) {
	entry := syntheticCase()
	score := ScoreTranscript(entry, Transcript{CaseID: entry.ID, Events: []Event{
		{Kind: "source_open", Path: "src/owner.go", Symbol: "Owner"},
		{Kind: "source_open", Path: "docs/contract.md", Symbol: "Contract"},
		{Kind: "claim", Claim: "The client owns server authority."},
	}})
	if score.Correct || len(score.UnsupportedConclusions) != 1 || score.UnsupportedConclusions[0] != "invented" {
		t.Fatalf("unsupported conclusion was not scored: %+v", score)
	}
}

func TestReportsAreDeterministicAndMeasureRepeatability(t *testing.T) {
	entry := syntheticCase()
	corpus := Corpus{Version: SchemaVersion, Repository: "fixture", Cases: []Case{entry}}
	transcripts := []Transcript{
		{Adapter: "raw", RunID: "a", CaseID: entry.ID, Events: []Event{{Kind: "source_open", Path: "src/owner.go", Symbol: "Owner"}, {Kind: "source_open", Path: "docs/contract.md", Symbol: "Contract"}}},
		{Adapter: "raw", RunID: "b", CaseID: entry.ID, Events: []Event{{Kind: "source_open", Path: "src/owner.go", Symbol: "Owner"}, {Kind: "source_open", Path: "docs/contract.md", Symbol: "Contract"}}},
	}
	first, err := BuildReport(corpus, transcripts)
	if err != nil {
		t.Fatal(err)
	}
	second, err := BuildReport(corpus, transcripts)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := json.Marshal(first)
	b, _ := json.Marshal(second)
	if string(a) != string(b) || markdown(first) != markdown(second) {
		t.Fatal("reports differ for identical inputs")
	}
	if len(first.Aggregates) != 1 || !first.Aggregates[0].Repeatability.Measured || !first.Aggregates[0].Repeatability.Consistent {
		t.Fatalf("repeatability = %+v", first.Aggregates)
	}
	if first.Aggregates[0].Correct != 2 {
		t.Fatalf("zero-time ownership evidence should be correct: %+v", first.Aggregates[0])
	}
}
