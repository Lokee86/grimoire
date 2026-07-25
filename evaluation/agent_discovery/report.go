package agentdiscovery

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Repeatability struct {
	Runs       int  `json:"runs"`
	Measured   bool `json:"measured"`
	Consistent bool `json:"consistent"`
}

type Aggregate struct {
	Adapter               string        `json:"adapter"`
	Runs                  int           `json:"runs"`
	Correct               int           `json:"correct"`
	AverageInputTokens    float64       `json:"average_input_tokens"`
	AverageOutputTokens   float64       `json:"average_output_tokens"`
	AverageRepeatedInput  float64       `json:"average_repeated_input_tokens"`
	AverageDiscoveryCalls float64       `json:"average_discovery_calls"`
	AverageToolCalls      float64       `json:"average_tool_calls"`
	Repeatability         Repeatability `json:"repeatability"`
}

type Report struct {
	Version    int         `json:"version"`
	Repository string      `json:"repository"`
	Revision   string      `json:"revision"`
	Scores     []Score     `json:"scores"`
	Aggregates []Aggregate `json:"aggregates"`
}

func BuildReport(corpus Corpus, transcripts []Transcript) (Report, error) {
	report := Report{Version: SchemaVersion, Repository: corpus.Repository, Revision: corpus.Revision}
	for _, transcript := range transcripts {
		entry, ok := findCase(corpus, transcript.CaseID)
		if !ok {
			return Report{}, fmt.Errorf("unknown case %q", transcript.CaseID)
		}
		report.Scores = append(report.Scores, ScoreTranscript(entry, transcript))
	}
	sort.Slice(report.Scores, func(i, j int) bool {
		return report.Scores[i].Adapter+"\x00"+report.Scores[i].CaseID+"\x00"+report.Scores[i].RunID < report.Scores[j].Adapter+"\x00"+report.Scores[j].CaseID+"\x00"+report.Scores[j].RunID
	})
	report.Aggregates = aggregates(report.Scores)
	return report, nil
}

func aggregates(scores []Score) []Aggregate {
	byAdapter, byCase := map[string][]Score{}, map[string][]Score{}
	for _, score := range scores {
		byAdapter[score.Adapter] = append(byAdapter[score.Adapter], score)
		byCase[score.Adapter+"\x00"+score.CaseID] = append(byCase[score.Adapter+"\x00"+score.CaseID], score)
	}
	result := make([]Aggregate, 0, len(byAdapter))
	for adapter, values := range byAdapter {
		item := Aggregate{Adapter: adapter, Runs: len(values), Repeatability: Repeatability{Consistent: true}}
		for _, value := range values {
			if value.Correct {
				item.Correct++
			}
			item.AverageInputTokens += float64(value.TotalInputTokens)
			item.AverageOutputTokens += float64(value.TotalOutputTokens)
			item.AverageRepeatedInput += float64(value.RepeatedInputTokens)
			item.AverageDiscoveryCalls += float64(value.DiscoveryCalls)
			item.AverageToolCalls += float64(value.ToolCalls)
		}
		item.AverageInputTokens /= float64(item.Runs)
		item.AverageOutputTokens /= float64(item.Runs)
		item.AverageRepeatedInput /= float64(item.Runs)
		item.AverageDiscoveryCalls /= float64(item.Runs)
		item.AverageToolCalls /= float64(item.Runs)
		for _, runs := range byCase {
			if len(runs) > 1 && runs[0].Adapter == adapter {
				item.Repeatability.Measured = true
				item.Repeatability.Runs += len(runs)
				fingerprint := scoreFingerprint(runs[0])
				for _, run := range runs[1:] {
					if scoreFingerprint(run) != fingerprint {
						item.Repeatability.Consistent = false
					}
				}
			}
		}
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Adapter < result[j].Adapter })
	return result
}

func WriteReport(report Report, directory, name string) (string, string, error) {
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return "", "", err
	}
	jsonPath, markdownPath := filepath.Join(directory, name+".json"), filepath.Join(directory, name+".md")
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", "", err
	}
	if err := os.WriteFile(jsonPath, append(data, '\n'), 0o644); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(markdownPath, []byte(markdown(report)), 0o644); err != nil {
		return "", "", err
	}
	return jsonPath, markdownPath, nil
}

func markdown(report Report) string {
	var out strings.Builder
	fmt.Fprintf(&out, "# Agent discovery benchmark\n\nRepository: %s (`%s`)\n\n", report.Repository, report.Revision)
	out.WriteString("| Adapter | Runs | Correct | Avg input | Repeated input | Avg calls | Repeatable |\n|---|---:|---:|---:|---:|---:|---|\n")
	for _, item := range report.Aggregates {
		fmt.Fprintf(&out, "| %s | %d | %d | %.1f | %.1f | %.1f | %t |\n", item.Adapter, item.Runs, item.Correct, item.AverageInputTokens, item.AverageRepeatedInput, item.AverageDiscoveryCalls, item.Repeatability.Consistent)
	}
	out.WriteString("\n| Adapter | Case | Run | Correct | Evidence | Input / output | Repeated | Irrelevant branches | Unsupported |\n|---|---|---|---|---|---:|---:|---:|---:|\n")
	for _, score := range report.Scores {
		fmt.Fprintf(&out, "| %s | %s | %s | %t | %d/%d + %d/%d | %d / %d | %d | %d | %d |\n", score.Adapter, score.CaseID, score.RunID, score.Correct, score.RequiredRecovered, score.RequiredTotal, score.StructuralRecovered, score.StructuralTotal, score.TotalInputTokens, score.TotalOutputTokens, score.RepeatedInputTokens, len(score.IrrelevantBranches), len(score.UnsupportedConclusions))
	}
	return out.String()
}
