package knowledgeevaluation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func Write(report Report, jsonPath, markdownPath string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("encode knowledge evaluation report: %w", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(jsonPath), 0o755); err != nil {
		return fmt.Errorf("create knowledge evaluation result directory: %w", err)
	}
	if err := os.WriteFile(jsonPath, data, 0o644); err != nil {
		return fmt.Errorf("write knowledge evaluation JSON: %w", err)
	}
	if err := os.WriteFile(markdownPath, []byte(Markdown(report)), 0o644); err != nil {
		return fmt.Errorf("write knowledge evaluation Markdown: %w", err)
	}
	return nil
}

func Markdown(report Report) string {
	var output bytes.Buffer
	fmt.Fprintf(&output, "# Knowledge retrieval evaluation: %s\n\n", report.Repository)
	fmt.Fprintf(&output, "Generated: %s  \n", report.GeneratedAt.Format("2006-01-02 15:04:05Z07:00"))
	fmt.Fprintf(&output, "Variant: `%s`  \n", report.Variant)
	fmt.Fprintf(&output, "Vectors requested: `%t`  \n", report.Vectors)
	fmt.Fprintf(&output, "Corpus cases: `%d`  \n\n", len(report.Cases))

	output.WriteString("## Aggregate\n\n")
	output.WriteString("| Metric | Value |\n| --- | ---: |\n")
	fmt.Fprintf(&output, "| Pass rate | %.1f%% |\n", report.Aggregate.PassRate*100)
	fmt.Fprintf(&output, "| Required-section recall | %.1f%% |\n", report.Aggregate.RequiredSectionRecall*100)
	for _, metric := range report.Aggregate.RecallAtK {
		fmt.Fprintf(&output, "| Recall@%d | %.1f%% |\n", metric.K, metric.Value*100)
	}
	fmt.Fprintf(&output, "| MRR | %.3f |\n", report.Aggregate.MRR)
	fmt.Fprintf(&output, "| Irrelevant selections | %d (%.1f%%) |\n", report.Aggregate.IrrelevantSelections, report.Aggregate.IrrelevantSelectionRate*100)
	fmt.Fprintf(&output, "| Vector usage | %d/%d (%.1f%%) |\n", report.Aggregate.VectorUsedCases, report.Aggregate.Cases, report.Aggregate.VectorUsageRate*100)
	fmt.Fprintf(&output, "| Vector errors | %d |\n", report.Aggregate.VectorErrorCases)
	fmt.Fprintf(&output, "| Median latency | %.1f ms |\n", report.Aggregate.MedianLatencyMS)
	fmt.Fprintf(&output, "| p95 latency | %.1f ms |\n\n", report.Aggregate.P95LatencyMS)

	output.WriteString("## Cases\n\n")
	output.WriteString("| Case | Category | Pass | Required recall | MRR | Irrelevant | Vector | Latency |\n")
	output.WriteString("| --- | --- | ---: | ---: | ---: | ---: | --- | ---: |\n")
	for _, result := range report.Cases {
		vector := "no"
		if result.VectorUsed {
			vector = "yes"
		} else if result.VectorError != "" {
			vector = "error"
		}
		fmt.Fprintf(&output, "| `%s` | %s | %t | %.1f%% | %.3f | %d (%.1f%%) | %s | %.1f ms |\n",
			result.CaseID, result.Category, result.Pass, result.RequiredSectionRecall*100,
			result.MRR, result.IrrelevantSelections, result.IrrelevantSelectionRate*100, vector, result.LatencyMS)
	}
	output.WriteString("\n## Per-case rankings\n\n")
	for _, result := range report.Cases {
		fmt.Fprintf(&output, "### `%s`\n\n", result.CaseID)
		fmt.Fprintf(&output, "%s\n\n", result.Query)
		if result.VectorError != "" {
			fmt.Fprintf(&output, "Vector error: `%s`\n\n", result.VectorError)
		}
		output.WriteString("| Rank | Path | Heading | Score | Relevant |\n| ---: | --- | --- | ---: | ---: |\n")
		for _, hit := range result.Results {
			fmt.Fprintf(&output, "| %d | `%s` | %s | %.4f | %t |\n", hit.Rank, hit.Path, hit.Heading, hit.Score, hit.Relevant)
		}
		output.WriteString("\n")
	}
	return output.String()
}
