package arcanaevaluation

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedInSpaceRocksCorpusHasSemanticCrossLanguageCoverage(t *testing.T) {
	corpus, err := LoadCorpus(filepath.Join("..", "..", "evaluation", "arcana", "space-rocks.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(corpus.Cases) < 8 || len(corpus.Cases) > 12 {
		t.Fatalf("Space Rocks corpus must contain 8-12 cases, got %d", len(corpus.Cases))
	}

	languages := map[string]bool{".go": false, ".gd": false, ".rb": false}
	domains := map[string]bool{
		"cross-language": false,
		"lifecycle":      false,
		"networking":     false,
		"observability":  false,
	}
	crossLanguageCases := 0
	for _, entry := range corpus.Cases {
		category := strings.ToLower(entry.Category)
		for domain := range domains {
			if strings.Contains(category, domain) {
				domains[domain] = true
			}
		}

		caseLanguages := make(map[string]struct{})
		for _, expected := range entry.RequiredSeeds {
			extension := strings.ToLower(filepath.Ext(expected.Path))
			if _, tracked := languages[extension]; tracked {
				languages[extension] = true
				caseLanguages[extension] = struct{}{}
			}
			if strings.Contains(strings.ToLower(entry.Query), strings.ToLower(expected.Name)) {
				t.Fatalf("case %q repeats required seed %q in its semantic query", entry.ID, expected.Name)
			}
			if !hasOperationalRoleExpectation(entry, expected) {
				t.Fatalf("case %q seed %s at %s lacks a matching required operational role", entry.ID, expected.Name, expected.Path)
			}
		}
		if len(caseLanguages) > 1 {
			crossLanguageCases++
		}
	}

	for language, covered := range languages {
		if !covered {
			t.Errorf("Space Rocks corpus lacks required %s seed coverage", language)
		}
	}
	for domain, covered := range domains {
		if !covered {
			t.Errorf("Space Rocks corpus lacks %s category coverage", domain)
		}
	}
	if crossLanguageCases < 2 {
		t.Errorf("Space Rocks corpus requires at least two cases with required seeds in multiple languages, got %d", crossLanguageCases)
	}
}

func hasOperationalRoleExpectation(entry Case, seed SeedExpectation) bool {
	for _, expected := range entry.RequiredStructural {
		if strings.EqualFold(expected.Provider, "arcana") &&
			strings.EqualFold(expected.Kind, "operational_role") &&
			strings.EqualFold(expected.Symbol, seed.Name) &&
			filepath.ToSlash(expected.Path) == filepath.ToSlash(seed.Path) {
			return true
		}
	}
	return false
}
