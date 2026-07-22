package scan

import (
	"path/filepath"
	"sort"
	"strings"

	lexfiles "github.com/Lokee86/lexicon/internal/files"
	"github.com/Lokee86/lexicon/internal/state"
)

type analysisPlan struct {
	Language     string
	Full         bool
	ChangedFiles []string
	RemovedFiles []string
	ContextFiles []string
}

func (s *Scanner) plansFor(changes []state.Change, drift []string) ([]analysisPlan, error) {
	plans := make(map[string]*analysisPlan)
	for _, language := range drift {
		plans[language] = &analysisPlan{Language: language, Full: true}
	}
	for _, change := range changes {
		paths := []string{change.New}
		if change.Old != "" {
			paths = append(paths, change.Old)
		}
		for _, path := range paths {
			for _, language := range lexfiles.Languages(path) {
				if !s.languageEnabled(language) {
					continue
				}
				plan := plans[language]
				if plan == nil {
					plan = &analysisPlan{Language: language}
					plans[language] = plan
				}
				if structuralChange(change, language, path) {
					plan.Full = true
					continue
				}
				if change.New != "" && languageOwnsSource(language, change.New) {
					plan.ChangedFiles = append(plan.ChangedFiles, change.New)
				}
			}
		}
	}

	result := make([]analysisPlan, 0, len(plans))
	for _, plan := range plans {
		if !plan.Full {
			roots := uniqueSorted(plan.ChangedFiles)
			fullRequired, err := s.Store.DirectChangesRequireFull(plan.Language, roots)
			if err != nil || fullRequired {
				plan.Full = true
			} else {
				impacted, context, err := s.Store.DependencyScope(plan.Language, roots)
				if err != nil {
					plan.Full = true
				} else {
					plan.ChangedFiles = impacted
					plan.RemovedFiles = []string{}
					plan.ContextFiles = context
				}
			}
		}
		if plan.Full {
			plan.ChangedFiles = nil
			plan.RemovedFiles = nil
			plan.ContextFiles = nil
		}
		result = append(result, *plan)
	}
	sort.Slice(result, func(left, right int) bool {
		return result[left].Language < result[right].Language
	})
	return result, nil
}

func structuralChange(change state.Change, language, path string) bool {
	status := strings.TrimSpace(change.Status)
	if status == "" || status[0] != 'M' {
		return true
	}
	return !languageOwnsSource(language, path)
}

func languageOwnsSource(language, path string) bool {
	extension := strings.ToLower(filepath.Ext(path))
	switch language {
	case "go":
		return extension == ".go"
	case "python":
		return extension == ".py"
	case "ruby":
		return extension == ".rb" || extension == ".gemspec"
	case "gdscript":
		return extension == ".gd"
	case "rust":
		return extension == ".rs"
	case "typescript":
		return extension == ".ts" || extension == ".tsx" || extension == ".mts" || extension == ".cts" ||
			extension == ".js" || extension == ".jsx" || extension == ".mjs" || extension == ".cjs"
	default:
		return false
	}
}

func uniqueSorted(paths []string) []string {
	set := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		if path != "" {
			set[filepath.ToSlash(path)] = struct{}{}
		}
	}
	result := make([]string, 0, len(set))
	for path := range set {
		result = append(result, path)
	}
	sort.Strings(result)
	return result
}
