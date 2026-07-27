package arcanaevaluation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var structuralKinds = map[string]struct{}{
	"symbol": {}, "operational_role": {}, "impact": {}, "call_chain": {}, "unresolved": {},
}

func LoadCorpus(path string) (Corpus, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Corpus{}, fmt.Errorf("read Arcana evaluation corpus: %w", err)
	}
	var corpus Corpus
	if err := json.Unmarshal(data, &corpus); err != nil {
		return Corpus{}, fmt.Errorf("decode Arcana evaluation corpus: %w", err)
	}
	if corpus.Version == 0 {
		corpus.Version = FormatVersion
	}
	if corpus.Version != FormatVersion {
		return Corpus{}, fmt.Errorf("unsupported Arcana evaluation corpus version %d", corpus.Version)
	}
	if strings.TrimSpace(corpus.Repository) == "" {
		return Corpus{}, fmt.Errorf("Arcana evaluation corpus repository is required")
	}
	if len(corpus.Cases) == 0 {
		return Corpus{}, fmt.Errorf("Arcana evaluation corpus contains no cases")
	}
	if corpus.TopK == 0 {
		corpus.TopK = MaxSeedLimit
	}
	if corpus.TopK < 1 || corpus.TopK > MaxSeedLimit {
		return Corpus{}, fmt.Errorf("Arcana evaluation corpus top_k must be between 1 and %d", MaxSeedLimit)
	}
	corpus.RecallAtK = normalizeKs(corpus.RecallAtK)
	if len(corpus.RecallAtK) == 0 {
		corpus.RecallAtK = []int{1, 3, corpus.TopK}
		corpus.RecallAtK = normalizeKs(corpus.RecallAtK)
	}

	seenIDs := make(map[string]struct{}, len(corpus.Cases))
	for index := range corpus.Cases {
		entry := &corpus.Cases[index]
		entry.ID = strings.TrimSpace(entry.ID)
		entry.Query = strings.TrimSpace(entry.Query)
		entry.Category = strings.TrimSpace(entry.Category)
		if entry.ID == "" || entry.Query == "" || entry.Category == "" {
			return Corpus{}, fmt.Errorf("case %d requires id, query, and category", index+1)
		}
		if _, exists := seenIDs[entry.ID]; exists {
			return Corpus{}, fmt.Errorf("duplicate Arcana evaluation case id %q", entry.ID)
		}
		seenIDs[entry.ID] = struct{}{}
		if len(entry.RequiredSeeds) == 0 {
			return Corpus{}, fmt.Errorf("case %q requires at least one required seed", entry.ID)
		}
		if len(entry.RequiredStructural) == 0 {
			return Corpus{}, fmt.Errorf("case %q requires at least one required structural expectation", entry.ID)
		}
		for _, group := range [][]SeedExpectation{entry.RequiredSeeds, entry.SupportingSeeds} {
			for _, expectation := range group {
				if err := validateSeedExpectation(entry.ID, expectation); err != nil {
					return Corpus{}, err
				}
			}
		}
		for _, group := range [][]StructuralExpectation{entry.RequiredStructural, entry.SupportingStructural} {
			for _, expectation := range group {
				if err := validateStructuralExpectation(entry.ID, expectation); err != nil {
					return Corpus{}, err
				}
			}
		}
	}
	return corpus, nil
}

func validateSeedExpectation(caseID string, expectation SeedExpectation) error {
	if strings.TrimSpace(expectation.Name) == "" {
		return fmt.Errorf("case %q seed name is required", caseID)
	}
	if !validRelativePath(expectation.Path) {
		return fmt.Errorf("case %q has invalid seed path %q", caseID, expectation.Path)
	}
	return nil
}

func validateStructuralExpectation(caseID string, expectation StructuralExpectation) error {
	if !strings.EqualFold(strings.TrimSpace(expectation.Provider), "arcana") {
		return fmt.Errorf("case %q structural provider must be Arcana, got %q", caseID, expectation.Provider)
	}
	kind := strings.ToLower(strings.TrimSpace(expectation.Kind))
	if _, valid := structuralKinds[kind]; !valid {
		return fmt.Errorf("case %q has invalid structural kind %q", caseID, expectation.Kind)
	}
	for label, path := range map[string]string{"path": expectation.Path, "target_path": expectation.TargetPath} {
		if strings.TrimSpace(path) != "" && !validRelativePath(path) {
			return fmt.Errorf("case %q has invalid structural %s %q", caseID, label, path)
		}
	}
	if expectation.Direction != "" && expectation.Direction != "incoming" && expectation.Direction != "outgoing" {
		return fmt.Errorf("case %q has invalid structural direction %q", caseID, expectation.Direction)
	}
	if expectation.Certainty != "" && expectation.Certainty != "definite" && expectation.Certainty != "possible" {
		return fmt.Errorf("case %q has invalid structural certainty %q", caseID, expectation.Certainty)
	}
	if kind == "call_chain" && len(expectation.Chain) > 0 && len(expectation.Chain) < 2 {
		return fmt.Errorf("case %q call_chain expectation requires at least two chain symbols", caseID)
	}
	return nil
}

func validRelativePath(path string) bool {
	path = filepath.ToSlash(strings.TrimSpace(path))
	return path != "" && !filepath.IsAbs(path) && path != "." && path != ".." && !strings.HasPrefix(path, "../")
}

func normalizeKs(values []int) []int {
	seen := make(map[int]struct{}, len(values))
	result := make([]int, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Ints(result)
	return result
}
