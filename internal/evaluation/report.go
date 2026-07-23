package evaluation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func BuildAggregates(report *Report) {
	modeGroups := make(map[string][]CaseRun)
	categoryGroups := make(map[string][]CaseRun)
	modeCategoryGroups := make(map[string][]CaseRun)
	for _, run := range report.Runs {
		modeGroups[run.Mode] = append(modeGroups[run.Mode], run)
		categoryGroups[string(run.Category)] = append(categoryGroups[string(run.Category)], run)
		key := run.Mode + "/" + string(run.Category)
		modeCategoryGroups[key] = append(modeCategoryGroups[key], run)
	}
	report.ByMode = aggregateGroups(modeGroups)
	report.ByCategory = aggregateGroups(categoryGroups)
	report.ByModeCategory = aggregateGroups(modeCategoryGroups)
}

func aggregateGroups(groups map[string][]CaseRun) []Aggregate {
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]Aggregate, 0, len(keys))
	for _, key := range keys {
		result = append(result, AggregateRuns(key, groups[key]))
	}
	return result
}

func Write(report Report, jsonPath, markdownPath string) error {
	BuildAggregates(&report)
	if err := os.MkdirAll(filepath.Dir(jsonPath), 0o755); err != nil {
		return fmt.Errorf("create evaluation result directory: %w", err)
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("encode evaluation results: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(jsonPath, data, 0o644); err != nil {
		return fmt.Errorf("write evaluation JSON: %w", err)
	}
	if err := os.WriteFile(markdownPath, []byte(Markdown(report)), 0o644); err != nil {
		return fmt.Errorf("write evaluation Markdown: %w", err)
	}
	return nil
}

func Markdown(report Report) string {
	BuildAggregates(&report)
	var output bytes.Buffer
	fmt.Fprintf(&output, "# Retrieval evaluation: %s\n\n", report.Repository)
	fmt.Fprintf(&output, "Generated: %s  \n", report.GeneratedAt.Format("2006-01-02 15:04:05Z07:00"))
	fmt.Fprintf(&output, "Variant: `%s`  \n", report.Variant)
	if report.SourceURL != "" {
		fmt.Fprintf(&output, "Source: `%s`  \n", report.SourceURL)
	}
	if report.Revision != "" {
		fmt.Fprintf(&output, "Revision: `%s`  \n", report.Revision)
	}
	if report.Scope != "" {
		fmt.Fprintf(&output, "Scope: `%s`  \n", report.Scope)
	}
	if report.JudgedAt != "" {
		fmt.Fprintf(&output, "Judged: `%s`  \n", report.JudgedAt)
	}
	fmt.Fprintf(&output, "Cases: %d  \n", uniqueCaseCount(report.Runs))
	fmt.Fprintf(&output, "Runs: %d  \n", len(report.Runs))
	fmt.Fprintf(&output, "Structural providers: `%s`\n\n", strings.Join(report.StructuralProviders, ","))

	output.WriteString("## Mode comparison\n\n")
	output.WriteString("| Mode | Pass | Source req. | Structural req. | Source irrelevant | Structural irrelevant | Median | p95 |\n")
	output.WriteString("| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |\n")
	for _, aggregate := range report.ByMode {
		fmt.Fprintf(&output, "| %s | %.1f%% | %.1f%% | %.1f%% | %.1f%% | %.1f%% | %.1f ms | %.1f ms |\n",
			aggregate.Group, aggregate.PassRate*100, aggregate.RequiredEvidenceRecall*100,
			aggregate.RequiredStructuralRecall*100, aggregate.IrrelevantSelectionRate*100,
			aggregate.IrrelevantStructuralRate*100, aggregate.MedianLatencyMS, aggregate.P95LatencyMS)
	}

	output.WriteString("\n## Pre-curation source ranking\n\n")
	output.WriteString("These metrics score the retrieved order before exact-result merging, curation, and package fitting.\n\n")
	output.WriteString("| Mode | Queries | Required R@10 | Required R@20 | MRR | Relevant @10 | Relevant @20 |\n")
	output.WriteString("| --- | ---: | ---: | ---: | ---: | ---: | ---: |\n")
	for _, aggregate := range report.ByMode {
		fmt.Fprintf(&output, "| %s | %d | %.1f%% | %.1f%% | %.3f | %.1f%% | %.1f%% |\n",
			aggregate.Group, aggregate.RankingCases, aggregate.RequiredRecallAt10*100,
			aggregate.RequiredRecallAt20*100, aggregate.MeanReciprocalRank,
			aggregate.RelevantRateAt10*100, aggregate.RelevantRateAt20*100)
	}

	output.WriteString("\n## Category comparison\n\n")
	output.WriteString("| Category | Pass | Source req. | Structural req. | Source irrelevant | Structural irrelevant | Median | p95 |\n")
	output.WriteString("| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |\n")
	for _, aggregate := range report.ByCategory {
		fmt.Fprintf(&output, "| %s | %.1f%% | %.1f%% | %.1f%% | %.1f%% | %.1f%% | %.1f ms | %.1f ms |\n",
			aggregate.Group, aggregate.PassRate*100, aggregate.RequiredEvidenceRecall*100,
			aggregate.RequiredStructuralRecall*100, aggregate.IrrelevantSelectionRate*100,
			aggregate.IrrelevantStructuralRate*100, aggregate.MedianLatencyMS, aggregate.P95LatencyMS)
	}

	output.WriteString("\n## Mode by category\n\n")
	output.WriteString("| Mode/category | Pass | Source req. | Structural req. | Source irrelevant | Structural irrelevant | Median | p95 |\n")
	output.WriteString("| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |\n")
	for _, aggregate := range report.ByModeCategory {
		fmt.Fprintf(&output, "| %s | %.1f%% | %.1f%% | %.1f%% | %.1f%% | %.1f%% | %.1f ms | %.1f ms |\n",
			aggregate.Group, aggregate.PassRate*100, aggregate.RequiredEvidenceRecall*100,
			aggregate.RequiredStructuralRecall*100, aggregate.IrrelevantSelectionRate*100,
			aggregate.IrrelevantStructuralRate*100, aggregate.MedianLatencyMS, aggregate.P95LatencyMS)
	}

	output.WriteString("\n## Per-case results\n\n")
	output.WriteString("| Case | Mode | Pass | Source req. | Structural req. | Source irrelevant | Structural irrelevant | Tokens | Latency | Failure |\n")
	output.WriteString("| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |\n")
	for _, run := range report.Runs {
		failure := strings.Join(run.FailureClassifications, ", ")
		if run.Error != "" {
			failure = run.Error
		}
		fmt.Fprintf(&output, "| %s | %s | %t | %.1f%% | %.1f%% | %.1f%% | %.1f%% | %d | %.1f ms | %s |\n",
			run.CaseID, run.Mode, run.Pass, run.RequiredEvidenceRecall*100,
			run.RequiredStructuralRecall*100, run.IrrelevantSelectionRate*100,
			run.IrrelevantStructuralRate*100, run.FinalPackageTokens, run.Timings.TotalMS, escapeCell(failure))
	}

	failures := failedRuns(report.Runs)
	if len(failures) > 0 {
		output.WriteString("\n## Concrete failures\n\n")
		for _, run := range failures {
			fmt.Fprintf(&output, "- `%s` / `%s`: %s", run.CaseID, run.Mode, strings.Join(run.FailureClassifications, ", "))
			if run.Error != "" {
				fmt.Fprintf(&output, "error: %s", run.Error)
			}
			output.WriteByte('\n')
			for _, status := range run.Required {
				if status.Included {
					continue
				}
				fmt.Fprintf(&output, "  - `%s`", status.Evidence.Path)
				if len(status.Evidence.Symbols) > 0 {
					fmt.Fprintf(&output, " symbols `%s`", strings.Join(status.Evidence.Symbols, "`, `"))
				}
				fmt.Fprintf(&output, ": %s\n", status.FailureStage)
			}
			for _, status := range run.RequiredStructural {
				if status.Included {
					continue
				}
				fmt.Fprintf(&output, "  - `%s:%s`", status.Evidence.Provider, status.Evidence.Kind)
				if status.Evidence.Symbol != "" {
					fmt.Fprintf(&output, " symbol `%s`", status.Evidence.Symbol)
				}
				if len(status.Evidence.Chain) > 0 {
					fmt.Fprintf(&output, " chain `%s`", strings.Join(status.Evidence.Chain, " -> "))
				}
				fmt.Fprintf(&output, ": %s\n", status.FailureStage)
			}
		}
	}
	return output.String()
}

func failedRuns(runs []CaseRun) []CaseRun {
	result := make([]CaseRun, 0)
	for _, run := range runs {
		if !run.Pass {
			result = append(result, run)
		}
	}
	return result
}

func uniqueCaseCount(runs []CaseRun) int {
	ids := make(map[string]struct{})
	for _, run := range runs {
		ids[run.CaseID] = struct{}{}
	}
	return len(ids)
}

func escapeCell(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "|", "\\|"), "\n", " ")
}
