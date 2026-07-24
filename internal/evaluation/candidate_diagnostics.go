package evaluation

import (
	"fmt"
	"sort"
)

const (
	diagnosticRetrievedLimit = 50
	diagnosticStageLimit     = 20
)

func BuildCandidateDiagnostics(entry Case, stages Stages) []CandidateDiagnostic {
	byKey := make(map[string]*CandidateDiagnostic)
	addStage := func(name string, candidates []Candidate) {
		for index, candidate := range candidates {
			key := diagnosticCandidateKey(candidate)
			diagnostic, exists := byKey[key]
			if !exists {
				diagnostic = &CandidateDiagnostic{
					Path:       candidate.Path,
					StartLine:  candidate.StartLine,
					EndLine:    candidate.EndLine,
					TokenCount: candidate.TokenCount,
				}
				byKey[key] = diagnostic
			}
			diagnostic.Required = diagnostic.Required || diagnosticMatchesAny(entry.Required, candidate)
			diagnostic.Supporting = diagnostic.Supporting || diagnosticMatchesAny(entry.Supporting, candidate)
			diagnostic.Forbidden = diagnostic.Forbidden || diagnosticMatchesAny(entry.Forbidden, candidate)
			stage := candidateStageDiagnostic(candidate, index+1)
			switch name {
			case "retrieved":
				diagnostic.Retrieved = stage
			case "exact":
				diagnostic.Exact = stage
			case "merged":
				diagnostic.Merged = stage
			case "curated":
				diagnostic.Curated = stage
			case "assembled":
				diagnostic.Assembled = stage
			case "included":
				diagnostic.Included = stage
			}
		}
	}

	addStage("retrieved", stages.Retrieved)
	addStage("exact", stages.Exact)
	addStage("merged", stages.Merged)
	addStage("curated", stages.Curated)
	addStage("assembled", stages.Assembled)
	addStage("included", stages.Included)

	result := make([]CandidateDiagnostic, 0, len(byKey))
	for _, diagnostic := range byKey {
		if !keepCandidateDiagnostic(*diagnostic) {
			continue
		}
		result = append(result, *diagnostic)
	}
	sort.Slice(result, func(left, right int) bool {
		leftRank := diagnosticSortRank(result[left])
		rightRank := diagnosticSortRank(result[right])
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		if result[left].Path != result[right].Path {
			return result[left].Path < result[right].Path
		}
		return result[left].StartLine < result[right].StartLine
	})
	return result
}

func diagnosticMatchesAny(group []Evidence, candidate Candidate) bool {
	for _, evidence := range group {
		if filepathKey(evidence.Path) != filepathKey(candidate.Path) {
			continue
		}
		if len(evidence.Symbols) == 0 {
			return true
		}
		for _, symbol := range evidence.Symbols {
			if candidateContains(candidate, symbol) {
				return true
			}
		}
	}
	return false
}

func candidateStageDiagnostic(candidate Candidate, rank int) *CandidateStageDiagnostic {
	return &CandidateStageDiagnostic{
		Rank:            rank,
		RetrievalSource: candidate.RetrievalSource,
		ProviderRank:    candidate.ProviderRank,
		Score:           candidate.Score,
		ScoreDetails:    append([]ScoreDetail(nil), candidate.ScoreDetails...),
		Reasons:         append([]string(nil), candidate.Reasons...),
	}
}

func diagnosticCandidateKey(candidate Candidate) string {
	return fmt.Sprintf("%s:%d:%d", filepathKey(candidate.Path), candidate.StartLine, candidate.EndLine)
}

func keepCandidateDiagnostic(candidate CandidateDiagnostic) bool {
	return candidate.Required || candidate.Supporting || candidate.Forbidden ||
		candidate.Retrieved != nil && candidate.Retrieved.Rank <= diagnosticRetrievedLimit ||
		candidate.Exact != nil && candidate.Exact.Rank <= diagnosticStageLimit ||
		candidate.Curated != nil && candidate.Curated.Rank <= diagnosticStageLimit ||
		candidate.Assembled != nil && candidate.Assembled.Rank <= diagnosticStageLimit
}

func diagnosticSortRank(candidate CandidateDiagnostic) int {
	const absent = 1_000_000
	if candidate.Retrieved != nil {
		return candidate.Retrieved.Rank
	}
	if candidate.Exact != nil {
		return 100_000 + candidate.Exact.Rank
	}
	if candidate.Merged != nil {
		return 200_000 + candidate.Merged.Rank
	}
	if candidate.Curated != nil {
		return 300_000 + candidate.Curated.Rank
	}
	if candidate.Assembled != nil {
		return 400_000 + candidate.Assembled.Rank
	}
	if candidate.Included != nil {
		return 500_000 + candidate.Included.Rank
	}
	return absent
}
