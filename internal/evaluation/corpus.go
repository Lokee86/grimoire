package evaluation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lokee86/grimoire/internal/queryshape"
)

var structuralKinds = map[string]struct{}{
	"symbol": {}, "operational_role": {}, "impact": {}, "call_chain": {}, "unresolved": {},
}

func LoadCorpus(path string) (Corpus, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Corpus{}, fmt.Errorf("read evaluation corpus: %w", err)
	}
	var corpus Corpus
	if err := json.Unmarshal(data, &corpus); err != nil {
		return Corpus{}, fmt.Errorf("decode evaluation corpus: %w", err)
	}
	if corpus.Version == 0 {
		corpus.Version = FormatVersion
	}
	if corpus.Version != FormatVersion {
		return Corpus{}, fmt.Errorf("unsupported evaluation corpus version %d", corpus.Version)
	}
	if strings.TrimSpace(corpus.Repository) == "" {
		return Corpus{}, fmt.Errorf("evaluation corpus repository is required")
	}
	if len(corpus.Cases) == 0 {
		return Corpus{}, fmt.Errorf("evaluation corpus contains no cases")
	}
	ids := make(map[string]struct{}, len(corpus.Cases))
	validCategories := make(map[Category]struct{}, len(Categories))
	for _, category := range Categories {
		validCategories[category] = struct{}{}
	}
	for index := range corpus.Cases {
		entry := &corpus.Cases[index]
		entry.ID = strings.TrimSpace(entry.ID)
		entry.Query = strings.TrimSpace(entry.Query)
		if entry.ID == "" || entry.Query == "" {
			return Corpus{}, fmt.Errorf("case %d requires id and query", index+1)
		}
		if _, exists := ids[entry.ID]; exists {
			return Corpus{}, fmt.Errorf("duplicate evaluation case id %q", entry.ID)
		}
		ids[entry.ID] = struct{}{}
		if _, valid := validCategories[entry.Category]; !valid {
			return Corpus{}, fmt.Errorf("case %q has invalid category %q", entry.ID, entry.Category)
		}
		if entry.Budget <= 0 {
			return Corpus{}, fmt.Errorf("case %q requires a positive budget", entry.ID)
		}
		if err := validateQueryProfileExpectation(entry.ID, entry.ExpectedQueryProfile); err != nil {
			return Corpus{}, err
		}
		if len(entry.Required) == 0 && len(entry.RequiredStructural) == 0 {
			return Corpus{}, fmt.Errorf("case %q requires explicit source or structural evidence", entry.ID)
		}
		for _, group := range [][]Evidence{entry.Required, entry.Supporting, entry.Forbidden} {
			for _, evidence := range group {
				if err := validateEvidence(entry.ID, evidence); err != nil {
					return Corpus{}, err
				}
			}
		}
		for _, group := range [][]StructuralExpectation{
			entry.RequiredStructural, entry.SupportingStructural, entry.ForbiddenStructural,
		} {
			for _, evidence := range group {
				if err := validateStructuralExpectation(entry.ID, evidence); err != nil {
					return Corpus{}, err
				}
			}
		}
	}
	return corpus, nil
}

func validateQueryProfileExpectation(caseID string, expected *QueryProfileExpectation) error {
	if expected == nil {
		return nil
	}
	if !queryshape.ValidScope(expected.Scope) {
		return fmt.Errorf("case %q has invalid expected query scope %q", caseID, expected.Scope)
	}
	for label, level := range map[string]queryshape.Level{
		"specificity": expected.Specificity,
		"breadth":     expected.Breadth,
		"ambiguity":   expected.Ambiguity,
	} {
		if !queryshape.ValidLevel(level) {
			return fmt.Errorf("case %q has invalid expected query %s %q", caseID, label, level)
		}
	}
	return nil
}

func validateEvidence(caseID string, evidence Evidence) error {
	path := filepath.ToSlash(strings.TrimSpace(evidence.Path))
	if !validRelativePath(path) {
		return fmt.Errorf("case %q has invalid evidence path %q", caseID, evidence.Path)
	}
	return nil
}

func validateStructuralExpectation(caseID string, evidence StructuralExpectation) error {
	provider := strings.ToLower(strings.TrimSpace(evidence.Provider))
	if provider != "lexicon" && provider != "arcana" {
		return fmt.Errorf("case %q has invalid structural provider %q", caseID, evidence.Provider)
	}
	kind := strings.ToLower(strings.TrimSpace(evidence.Kind))
	if _, valid := structuralKinds[kind]; !valid {
		return fmt.Errorf("case %q has invalid structural kind %q", caseID, evidence.Kind)
	}
	for label, path := range map[string]string{"path": evidence.Path, "target_path": evidence.TargetPath} {
		path = filepath.ToSlash(strings.TrimSpace(path))
		if path != "" && !validRelativePath(path) {
			return fmt.Errorf("case %q has invalid structural %s %q", caseID, label, path)
		}
	}
	if evidence.Direction != "" && evidence.Direction != "incoming" && evidence.Direction != "outgoing" {
		return fmt.Errorf("case %q has invalid structural direction %q", caseID, evidence.Direction)
	}
	if evidence.Certainty != "" && evidence.Certainty != "definite" && evidence.Certainty != "possible" {
		return fmt.Errorf("case %q has invalid structural certainty %q", caseID, evidence.Certainty)
	}
	if kind == "call_chain" && len(evidence.Chain) > 0 && len(evidence.Chain) < 2 {
		return fmt.Errorf("case %q call_chain expectation requires at least two chain symbols", caseID)
	}
	return nil
}

func validRelativePath(path string) bool {
	return path != "" && !filepath.IsAbs(path) && path != ".." && !strings.HasPrefix(path, "../")
}
