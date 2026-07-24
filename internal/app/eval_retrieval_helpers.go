package app

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Lokee86/grimoire/internal/compiler"
	"github.com/Lokee86/grimoire/internal/evaluation"
	"github.com/Lokee86/grimoire/internal/structure"
)

func validateEvaluationCase(root string, entry evaluation.Case) error {
	for _, group := range [][]evaluation.Evidence{entry.Required, entry.Supporting, entry.Forbidden} {
		for _, evidence := range group {
			if err := validateExpectedSymbol(root, entry.ID, evidence.Path, evidence.Symbols...); err != nil {
				return err
			}
		}
	}
	for _, group := range [][]evaluation.StructuralExpectation{
		entry.RequiredStructural, entry.SupportingStructural, entry.ForbiddenStructural,
	} {
		for _, expected := range group {
			if expected.Path != "" && expected.Symbol != "" {
				if err := validateExpectedSymbol(root, entry.ID, expected.Path, expected.Symbol); err != nil {
					return err
				}
			} else if expected.Path != "" {
				if err := validateExpectedSymbol(root, entry.ID, expected.Path); err != nil {
					return err
				}
			}
			if expected.TargetPath != "" && expected.TargetSymbol != "" {
				if err := validateExpectedSymbol(root, entry.ID, expected.TargetPath, expected.TargetSymbol); err != nil {
					return err
				}
			} else if expected.TargetPath != "" {
				if err := validateExpectedSymbol(root, entry.ID, expected.TargetPath); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func validateExpectedSymbol(root, caseID, relativePath string, symbols ...string) error {
	path := filepath.Join(root, filepath.FromSlash(relativePath))
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("incorrect evaluation expectation for %s: read %s: %w", caseID, relativePath, err)
	}
	for _, symbol := range symbols {
		if !strings.Contains(string(data), symbol) {
			return fmt.Errorf("incorrect evaluation expectation for %s: symbol %s is absent from %s", caseID, symbol, relativePath)
		}
	}
	return nil
}

func applyExpectationError(run *evaluation.CaseRun) {
	run.FailureClassifications = []string{evaluation.FailureIncorrectExpectation}
	run.RequiredNeverRetrieved = len(run.Required)
	run.RequiredStructuralNeverProduced = len(run.RequiredStructural)
	for index := range run.Required {
		run.Required[index].FailureStage = evaluation.FailureIncorrectExpectation
	}
	for index := range run.RequiredStructural {
		run.RequiredStructural[index].FailureStage = evaluation.FailureIncorrectExpectation
	}
}

func parseEvaluationModes(value string) ([]string, error) {
	allowed := make(map[string]struct{}, len(allowedEvaluationModes))
	for _, mode := range allowedEvaluationModes {
		allowed[mode] = struct{}{}
	}
	seen := make(map[string]struct{})
	var result []string
	for _, raw := range strings.Split(value, ",") {
		mode := strings.ToLower(strings.TrimSpace(raw))
		if mode == "" {
			continue
		}
		if _, valid := allowed[mode]; !valid {
			return nil, fmt.Errorf("unknown evaluation mode %q", mode)
		}
		if _, duplicate := seen[mode]; duplicate {
			continue
		}
		seen[mode] = struct{}{}
		result = append(result, mode)
	}
	if len(result) == 0 {
		return nil, errors.New("at least one evaluation mode is required")
	}
	return result, nil
}

func parseStructuralProviders(value string) ([]string, bool, bool, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || value == "none" {
		return nil, false, false, nil
	}
	seen := make(map[string]struct{})
	var providers []string
	for _, raw := range strings.Split(value, ",") {
		provider := strings.TrimSpace(raw)
		if provider == "" {
			continue
		}
		if provider != "lexicon" && provider != "arcana" {
			return nil, false, false, fmt.Errorf("unknown structural provider %q", provider)
		}
		if _, duplicate := seen[provider]; duplicate {
			continue
		}
		seen[provider] = struct{}{}
		providers = append(providers, provider)
	}
	if _, arcana := seen["arcana"]; arcana {
		if _, lexicon := seen["lexicon"]; !lexicon {
			return nil, false, false, errors.New("Arcana evaluation requires Lexicon")
		}
	}
	_, lexicon := seen["lexicon"]
	_, arcana := seen["arcana"]
	return providers, lexicon, arcana, nil
}

func packageStructuralSelections(evidence []structure.Evidence) []evaluation.StructuralSelection {
	result := make([]evaluation.StructuralSelection, 0, len(evidence))
	for _, item := range evidence {
		result = append(result, evaluation.StructuralSelection{Evidence: item})
	}
	return result
}

func packageSelections(entry evaluation.Case, selections []compiler.Selection) []evaluation.Selection {
	result := make([]evaluation.Selection, 0, len(selections))
	for _, selected := range selections {
		result = append(result, evaluation.Selection{
			Path:            selected.Path,
			StartLine:       selected.StartLine,
			EndLine:         selected.EndLine,
			Symbols:         detectedSymbols(entry, selected.Path, selected.Content),
			RetrievalSource: selected.RetrievalSource,
			ProviderRank:    selected.RetrievalRank,
			TokenCount:      selected.TokenCount,
		})
	}
	return result
}

func detectedSymbols(entry evaluation.Case, path, content string) []string {
	seen := make(map[string]struct{})
	var symbols []string
	for _, group := range [][]evaluation.Evidence{entry.Required, entry.Supporting, entry.Forbidden} {
		for _, evidence := range group {
			if filepath.ToSlash(evidence.Path) != filepath.ToSlash(path) {
				continue
			}
			for _, symbol := range evidence.Symbols {
				if !strings.Contains(content, symbol) {
					continue
				}
				if _, exists := seen[symbol]; exists {
					continue
				}
				seen[symbol] = struct{}{}
				symbols = append(symbols, symbol)
			}
		}
	}
	sort.Strings(symbols)
	return symbols
}

func selectedPaths(selections []evaluation.Selection) []string {
	seen := make(map[string]struct{})
	var paths []string
	for _, selection := range selections {
		if _, exists := seen[selection.Path]; exists {
			continue
		}
		seen[selection.Path] = struct{}{}
		paths = append(paths, selection.Path)
	}
	return paths
}

func applyEvaluationErrorClassification(run *evaluation.CaseRun) {
	classification := evaluation.FailureEmbeddingMiss
	message := strings.ToLower(run.Error)
	if strings.Contains(message, "manifest") || strings.Contains(message, "snapshot") ||
		strings.Contains(message, "prepared index") || strings.Contains(message, "vector result") {
		classification = evaluation.FailureStaleOrIncompleteIndex
	}
	run.FailureClassifications = []string{classification}
	run.RequiredNeverRetrieved = len(run.Required)
	for index := range run.Required {
		if !run.Required[index].Included {
			run.Required[index].FailureStage = classification
		}
	}
}

func defaultEvaluationPrefix(repository, variant string, generated time.Time) string {
	name := strings.ToLower(repository)
	name = strings.NewReplacer(" ", "-", "_", "-", "/", "-").Replace(name)
	variant = strings.ToLower(strings.TrimSpace(variant))
	variant = strings.NewReplacer(" ", "-", "_", "-", "/", "-").Replace(variant)
	return fmt.Sprintf("%s-%s-%s", name, variant, generated.Format("2006-01-02-150405"))
}

func writeEvaluationSummary(stdout io.Writer, report evaluation.Report, jsonPath, markdownPath string) error {
	if _, err := fmt.Fprintln(stdout, "mode\tpass\tsource_required\tr_at_10\tr_at_20\tmrr\tsource_irrelevant\tmedian_ms\tp95_ms"); err != nil {
		return err
	}
	for _, aggregate := range report.ByMode {
		if _, err := fmt.Fprintf(stdout, "%s\t%.1f%%\t%.1f%%\t%.1f%%\t%.1f%%\t%.3f\t%.1f%%\t%.1f\t%.1f\n",
			aggregate.Group, aggregate.PassRate*100, aggregate.RequiredEvidenceRecall*100,
			aggregate.RequiredRecallAt10*100, aggregate.RequiredRecallAt20*100,
			aggregate.MeanReciprocalRank, aggregate.IrrelevantSelectionRate*100,
			aggregate.MedianLatencyMS, aggregate.P95LatencyMS); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(stdout, "json: %s\nmarkdown: %s\n", jsonPath, markdownPath); err != nil {
		return err
	}
	return nil
}
