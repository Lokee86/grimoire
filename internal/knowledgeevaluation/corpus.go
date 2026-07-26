package knowledgeevaluation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func LoadCorpus(path string) (Corpus, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Corpus{}, fmt.Errorf("read knowledge evaluation corpus: %w", err)
	}
	var corpus Corpus
	if err := json.Unmarshal(data, &corpus); err != nil {
		return Corpus{}, fmt.Errorf("decode knowledge evaluation corpus: %w", err)
	}
	if corpus.Version == 0 {
		corpus.Version = FormatVersion
	}
	if corpus.Version != FormatVersion {
		return Corpus{}, fmt.Errorf("unsupported knowledge evaluation corpus version %d", corpus.Version)
	}
	if strings.TrimSpace(corpus.Repository) == "" {
		return Corpus{}, fmt.Errorf("knowledge evaluation corpus repository is required")
	}
	if len(corpus.Cases) == 0 {
		return Corpus{}, fmt.Errorf("knowledge evaluation corpus contains no cases")
	}
	if corpus.TopK < 0 {
		return Corpus{}, fmt.Errorf("knowledge evaluation corpus top_k must not be negative")
	}
	corpus.RecallAtK = normalizeKs(corpus.RecallAtK)
	if len(corpus.RecallAtK) == 0 {
		corpus.RecallAtK = effectiveRecallAtK(corpus)
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
			return Corpus{}, fmt.Errorf("duplicate knowledge evaluation case id %q", entry.ID)
		}
		seenIDs[entry.ID] = struct{}{}
		if len(entry.Required) == 0 {
			return Corpus{}, fmt.Errorf("case %q requires at least one required section", entry.ID)
		}
		for _, group := range [][]SectionExpectation{entry.Required, entry.Supporting, entry.Forbidden} {
			for _, expectation := range group {
				if err := validateExpectation(entry.ID, expectation); err != nil {
					return Corpus{}, err
				}
			}
		}
	}
	return corpus, nil
}

func validateExpectation(caseID string, expectation SectionExpectation) error {
	path := filepath.ToSlash(strings.TrimSpace(expectation.Path))
	if path == "" || filepath.IsAbs(path) || path == "." || path == ".." || strings.HasPrefix(path, "../") {
		return fmt.Errorf("case %q has invalid section path %q", caseID, expectation.Path)
	}
	if strings.TrimSpace(expectation.SectionID) == "" && strings.TrimSpace(expectation.Heading) == "" {
		return fmt.Errorf("case %q section %q requires heading or section_id", caseID, path)
	}
	if len(expectation.HeadingPath) > 0 {
		for _, heading := range expectation.HeadingPath {
			if strings.TrimSpace(heading) == "" {
				return fmt.Errorf("case %q section %q has an empty heading_path item", caseID, path)
			}
		}
	}
	return nil
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
