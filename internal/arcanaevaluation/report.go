package arcanaevaluation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func Write(report Report, jsonPath, markdownPath string) error {
	BuildAggregates(&report)
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("encode Arcana evaluation report: %w", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(jsonPath), 0o755); err != nil {
		return fmt.Errorf("create Arcana evaluation result directory: %w", err)
	}
	if err := os.WriteFile(jsonPath, data, 0o644); err != nil {
		return fmt.Errorf("write Arcana evaluation JSON: %w", err)
	}
	if err := os.WriteFile(markdownPath, []byte(Markdown(report)), 0o644); err != nil {
		return fmt.Errorf("write Arcana evaluation Markdown: %w", err)
	}
	return nil
}

func Markdown(report Report) string {
	BuildAggregates(&report)
	var output bytes.Buffer
	fmt.Fprintf(&output, "# Arcana semantic graph retrieval evaluation: %s\n\n", report.Repository)
	fmt.Fprintf(&output, "Generated: %s  \n", report.GeneratedAt.Format("2006-01-02 15:04:05Z07:00"))
	fmt.Fprintf(&output, "Variant: `%s`  \n", report.Variant)
	fmt.Fprintf(&output, "Arcana snapshot: `%s`  \n", report.ArcanaSnapshot)
	fmt.Fprintf(&output, "Embedding identity: `%s`  \n", report.EmbeddingIdentity)
	fmt.Fprintf(&output, "Corpus cases: `%d`\n\n", len(report.Cases)/2)

	output.WriteString("The paired modes use the same prepared source, Lexicon export, and Arcana graph snapshot. `lexicon-seeds` bypasses semantic lookup; `lexicon-plus-vector` adds the existing Arcana vector index before the same deterministic graph expansion.\n\n")
	output.WriteString("## Aggregate comparison\n\n")
	output.WriteString("| Mode | Pass | Seed recall | MRR | Structural recall | Median latency | p95 latency | Median payload | p95 payload | Provider calls |\n")
	output.WriteString("| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |\n")
	for _, aggregate := range report.Aggregates {
		fmt.Fprintf(&output, "| %s | %.1f%% | %.1f%% | %.3f | %.1f%% | %.1f ms | %.1f ms | %.0f B | %.0f B | %.2f |\n",
			aggregate.Mode, aggregate.PassRate*100, aggregate.RequiredSeedRecall*100, aggregate.MRR,
			aggregate.RequiredStructuralRecall*100, aggregate.MedianLatencyMS, aggregate.P95LatencyMS,
			aggregate.MedianPayloadBytes, aggregate.P95PayloadBytes, aggregate.MeanProviderCalls)
	}

	output.WriteString("\n## Seed recall at k\n\n")
	output.WriteString("| Mode |")
	for _, k := range report.RecallAtK {
		fmt.Fprintf(&output, " R@%d |", k)
	}
	output.WriteString("\n| --- |")
	for range report.RecallAtK {
		output.WriteString(" ---: |")
	}
	output.WriteByte('\n')
	for _, aggregate := range report.Aggregates {
		fmt.Fprintf(&output, "| %s |", aggregate.Mode)
		for _, metric := range aggregate.RecallAtK {
			fmt.Fprintf(&output, " %.1f%% |", metric.Value*100)
		}
		output.WriteByte('\n')
	}

	comparison := report.Comparison
	output.WriteString("\n## Vector-minus-baseline deltas\n\n")
	output.WriteString("| Metric | Delta |\n| --- | ---: |\n")
	fmt.Fprintf(&output, "| Pass rate | %+.1f pp |\n", comparison.PassRateDelta*100)
	fmt.Fprintf(&output, "| Required seed recall | %+.1f pp |\n", comparison.RequiredSeedRecallDelta*100)
	for _, metric := range comparison.RecallAtKDelta {
		fmt.Fprintf(&output, "| Seed recall@%d | %+.1f pp |\n", metric.K, metric.Value*100)
	}
	fmt.Fprintf(&output, "| MRR | %+.3f |\n", comparison.MRRDelta)
	fmt.Fprintf(&output, "| Required structural recall | %+.1f pp |\n", comparison.RequiredStructuralRecallDelta*100)
	fmt.Fprintf(&output, "| Median latency | %+.1f ms |\n", comparison.MedianLatencyMSDelta)
	fmt.Fprintf(&output, "| Median payload | %+.0f B |\n", comparison.MedianPayloadBytesDelta)
	fmt.Fprintf(&output, "| Mean provider calls | %+.2f |\n", comparison.MeanProviderCallsDelta)

	output.WriteString("\n## Cases\n\n")
	output.WriteString("| Case | Mode | Pass | Seed recall | MRR | Structural recall | Latency | Payload | Calls | Error |\n")
	output.WriteString("| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |\n")
	for _, result := range report.Cases {
		fmt.Fprintf(&output, "| `%s` | %s | %t | %.1f%% | %.3f | %.1f%% | %.1f ms | %d B | %d | %s |\n",
			result.CaseID, result.Mode, result.Pass, result.RequiredSeedRecall*100, result.MRR,
			result.RequiredStructuralRecall*100, result.Timings.TotalMS, result.PayloadBytes,
			result.ProviderCalls, result.Error)
	}

	output.WriteString("\n## Per-case seed rankings\n\n")
	for _, result := range report.Cases {
		fmt.Fprintf(&output, "### `%s` / `%s`\n\n", result.CaseID, result.Mode)
		fmt.Fprintf(&output, "%s\n\n", result.Query)
		output.WriteString("| Rank | Source | Node | Path | Required | Supporting |\n| ---: | --- | --- | --- | ---: | ---: |\n")
		for _, seed := range result.Seeds {
			fmt.Fprintf(&output, "| %d | %s | `%s` | `%s` | %t | %t |\n",
				seed.Rank, seed.Source, seed.Node.Name, seed.Node.Path, seed.Required, seed.Supporting)
		}
		output.WriteByte('\n')
	}
	return output.String()
}
