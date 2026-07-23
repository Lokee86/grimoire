package evaluation

import (
	"sort"
	"strings"

	"github.com/Lokee86/grimoire/internal/structure"
)

const (
	FailureEmbeddingMiss               = "embedding miss"
	FailureVectorRankingMiss           = "vector ranking miss"
	FailureExactRecoveryMiss           = "exact recovery miss"
	FailureCandidateMergeLoss          = "candidate merge loss"
	FailureCurationLoss                = "curation loss"
	FailureBudgetFittingLoss           = "budget-fitting loss"
	FailureStaleOrIncompleteIndex      = "stale or incomplete index"
	FailureIncorrectExpectation        = "incorrect evaluation expectation"
	FailureStructuralProviderMiss      = "structural provider miss"
	FailureStructuralCompositionLoss   = "structural composition loss"
	FailureStructuralBudgetFittingLoss = "structural budget-fitting loss"
)

type Stages struct {
	Indexed            []Candidate
	BroadProbe         []Candidate
	Retrieved          []Candidate
	Exact              []Candidate
	Merged             []Candidate
	Curated            []Candidate
	Included           []Candidate
	StructuralProduced []structure.Evidence
	StructuralComposed []structure.Evidence
	StructuralIncluded []structure.Evidence
}

func ScoreCase(entry Case, run *CaseRun, stages Stages) {
	run.Ranking = ScoreRanking(entry, stages.Retrieved)
	run.Required = scoreEvidenceGroup(entry.Query, entry.Required, stages)
	run.Supporting = scoreEvidenceGroup(entry.Query, entry.Supporting, stages)
	run.RequiredEvidenceRecall = recall(run.Required)
	run.SupportingEvidenceRecall = recall(run.Supporting)
	run.RequiredStructural = scoreStructuralGroup(entry.RequiredStructural, stages)
	run.SupportingStructural = scoreStructuralGroup(entry.SupportingStructural, stages)
	run.RequiredStructuralRecall = structuralRecall(run.RequiredStructural)
	run.SupportingStructuralRecall = structuralRecall(run.SupportingStructural)
	run.Pass = run.Error == "" && requiredSatisfied(run.Required) && structuralRequiredSatisfied(run.RequiredStructural)

	relevant := 0
	for index := range run.Selections {
		candidate := Candidate{
			Path:      run.Selections[index].Path,
			StartLine: run.Selections[index].StartLine,
			EndLine:   run.Selections[index].EndLine,
			Symbols:   run.Selections[index].Symbols,
		}
		run.Selections[index].Relevant = matchesAny(entry.Required, candidate) || matchesAny(entry.Supporting, candidate)
		run.Selections[index].Forbidden = matchesAny(entry.Forbidden, candidate)
		if run.Selections[index].Relevant {
			relevant++
		}
	}
	if len(run.Selections) > 0 {
		run.IrrelevantSelectionRate = float64(len(run.Selections)-relevant) / float64(len(run.Selections))
	}
	for _, forbidden := range entry.Forbidden {
		if evidencePresent(forbidden, stages.Included) {
			run.ForbiddenRecovered = append(run.ForbiddenRecovered, forbidden)
		}
	}
	structuralRelevant := 0
	for index := range run.StructuralSelections {
		item := run.StructuralSelections[index].Evidence
		run.StructuralSelections[index].Relevant = structuralMatchesAny(entry.RequiredStructural, item) ||
			structuralMatchesAny(entry.SupportingStructural, item)
		run.StructuralSelections[index].Forbidden = structuralMatchesAny(entry.ForbiddenStructural, item)
		if run.StructuralSelections[index].Relevant {
			structuralRelevant++
		}
	}
	if len(run.StructuralSelections) > 0 {
		run.IrrelevantStructuralRate = float64(len(run.StructuralSelections)-structuralRelevant) /
			float64(len(run.StructuralSelections))
	}
	for _, forbidden := range entry.ForbiddenStructural {
		if structuralEvidencePresent(forbidden, stages.StructuralIncluded) {
			run.ForbiddenStructuralRecovered = append(run.ForbiddenStructuralRecovered, forbidden)
		}
	}

	classifications := make(map[string]struct{})
	for _, status := range run.Required {
		if status.Included {
			continue
		}
		switch status.FailureStage {
		case FailureEmbeddingMiss, FailureVectorRankingMiss, FailureExactRecoveryMiss,
			FailureCandidateMergeLoss, FailureCurationLoss, FailureBudgetFittingLoss,
			FailureStaleOrIncompleteIndex, FailureIncorrectExpectation:
			classifications[status.FailureStage] = struct{}{}
		}
		switch status.FailureStage {
		case FailureEmbeddingMiss, FailureVectorRankingMiss, FailureExactRecoveryMiss:
			run.RequiredNeverRetrieved++
		case FailureCandidateMergeLoss:
			run.RequiredLostDuringMerge++
		case FailureCurationLoss:
			run.RequiredLostDuringCuration++
		case FailureBudgetFittingLoss:
			run.RequiredOmittedForBudget++
		}
	}
	for _, status := range run.RequiredStructural {
		if status.Included {
			continue
		}
		classifications[status.FailureStage] = struct{}{}
		switch status.FailureStage {
		case FailureStructuralProviderMiss:
			run.RequiredStructuralNeverProduced++
		case FailureStructuralCompositionLoss:
			run.RequiredStructuralLostComposition++
		case FailureStructuralBudgetFittingLoss:
			run.RequiredStructuralOmittedBudget++
		}
	}
	for classification := range classifications {
		run.FailureClassifications = append(run.FailureClassifications, classification)
	}
	sort.Strings(run.FailureClassifications)
}

func scoreEvidenceGroup(query string, group []Evidence, stages Stages) []EvidenceStatus {
	result := make([]EvidenceStatus, 0, len(group))
	for _, evidence := range group {
		status := EvidenceStatus{
			Evidence:       evidence,
			Indexed:        evidencePresent(evidence, stages.Indexed),
			BroadProbe:     evidencePresent(evidence, stages.BroadProbe),
			Retrieved:      evidencePresent(evidence, stages.Retrieved),
			RetrievedRank:  evidenceRank(evidence, stages.Retrieved),
			ExactRecovered: evidencePresent(evidence, stages.Exact),
			Merged:         evidencePresent(evidence, stages.Merged),
			Curated:        evidencePresent(evidence, stages.Curated),
			Included:       evidencePresent(evidence, stages.Included),
		}
		status.FailureStage = classifyEvidenceFailure(query, status)
		result = append(result, status)
	}
	return result
}

func classifyEvidenceFailure(query string, status EvidenceStatus) string {
	if status.Included {
		return ""
	}
	if !status.Indexed {
		return FailureStaleOrIncompleteIndex
	}
	if status.Curated {
		return FailureBudgetFittingLoss
	}
	if status.Merged {
		return FailureCurationLoss
	}
	if status.Retrieved || status.ExactRecovered {
		return FailureCandidateMergeLoss
	}
	if expectsExact(query, status.Evidence) {
		return FailureExactRecoveryMiss
	}
	if status.BroadProbe {
		return FailureVectorRankingMiss
	}
	return FailureEmbeddingMiss
}

func expectsExact(query string, evidence Evidence) bool {
	query = strings.ToLower(query)
	path := strings.ToLower(filepathKey(evidence.Path))
	if strings.Contains(query, path) {
		return true
	}
	parts := strings.Split(path, "/")
	if len(parts) > 0 && strings.Contains(query, parts[len(parts)-1]) {
		return true
	}
	for _, symbol := range evidence.Symbols {
		if strings.Contains(query, strings.ToLower(symbol)) {
			return true
		}
	}
	return false
}

func evidencePresent(evidence Evidence, candidates []Candidate) bool {
	path := filepathKey(evidence.Path)
	matchedSymbols := make(map[string]struct{}, len(evidence.Symbols))
	pathPresent := false
	for _, candidate := range candidates {
		if filepathKey(candidate.Path) != path {
			continue
		}
		pathPresent = true
		for _, symbol := range evidence.Symbols {
			if candidateContains(candidate, symbol) {
				matchedSymbols[symbol] = struct{}{}
			}
		}
	}
	return pathPresent && len(matchedSymbols) == len(evidence.Symbols)
}

func matchesAny(group []Evidence, candidate Candidate) bool {
	for _, evidence := range group {
		if filepathKey(evidence.Path) == filepathKey(candidate.Path) {
			return true
		}
	}
	return false
}

func candidateContains(candidate Candidate, symbol string) bool {
	if strings.Contains(candidate.Text, symbol) {
		return true
	}
	for _, known := range candidate.Symbols {
		if known == symbol {
			return true
		}
	}
	return false
}

func filepathKey(path string) string {
	return strings.TrimPrefix(strings.ReplaceAll(strings.TrimSpace(path), "\\", "/"), "./")
}

func recall(statuses []EvidenceStatus) float64 {
	if len(statuses) == 0 {
		return 0
	}
	included := 0
	for _, status := range statuses {
		if status.Included {
			included++
		}
	}
	return float64(included) / float64(len(statuses))
}

func AggregateRuns(group string, runs []CaseRun) Aggregate {
	aggregate := Aggregate{Group: group, Cases: len(runs)}
	if len(runs) == 0 {
		return aggregate
	}
	latencies := make([]float64, 0, len(runs))
	requiredIncluded, requiredTotal := 0, 0
	supportingIncluded, supportingTotal := 0, 0
	selectionTotal, irrelevantTotal := 0, 0
	structuralTotal, structuralIrrelevant := 0, 0
	structuralRequiredIncluded, structuralRequiredTotal := 0, 0
	structuralSupportingIncluded, structuralSupportingTotal := 0, 0
	rankingRequiredAt10, rankingRequiredAt20 := 0.0, 0.0
	rankingReciprocalRank := 0.0
	rankingRelevantAt10, rankingRelevantAt20 := 0.0, 0.0
	for _, run := range runs {
		if run.ExpectedQueryProfile != nil {
			aggregate.ProfileCases++
			if run.QueryProfileMatched {
				aggregate.ProfileMatches++
			}
		}
		if run.Pass {
			aggregate.Passes++
		}
		latencies = append(latencies, run.Timings.TotalMS)
		if len(run.Required) > 0 && run.Error == "" {
			aggregate.RankingCases++
			rankingRequiredAt10 += run.Ranking.RequiredRecallAt10
			rankingRequiredAt20 += run.Ranking.RequiredRecallAt20
			rankingReciprocalRank += run.Ranking.ReciprocalRank
			rankingRelevantAt10 += run.Ranking.RelevantRateAt10
			rankingRelevantAt20 += run.Ranking.RelevantRateAt20
		}
		for _, status := range run.Required {
			requiredTotal++
			if status.Included {
				requiredIncluded++
			}
		}
		for _, status := range run.Supporting {
			supportingTotal++
			if status.Included {
				supportingIncluded++
			}
		}
		for _, status := range run.RequiredStructural {
			structuralRequiredTotal++
			if status.Included {
				structuralRequiredIncluded++
			}
		}
		for _, status := range run.SupportingStructural {
			structuralSupportingTotal++
			if status.Included {
				structuralSupportingIncluded++
			}
		}
		selectionTotal += len(run.Selections)
		for _, selection := range run.Selections {
			if !selection.Relevant {
				irrelevantTotal++
			}
		}
		structuralTotal += len(run.StructuralSelections)
		for _, selected := range run.StructuralSelections {
			if !selected.Relevant {
				structuralIrrelevant++
			}
		}
	}
	aggregate.PassRate = float64(aggregate.Passes) / float64(aggregate.Cases)
	if requiredTotal > 0 {
		aggregate.RequiredEvidenceRecall = float64(requiredIncluded) / float64(requiredTotal)
	}
	if supportingTotal > 0 {
		aggregate.SupportingEvidenceRecall = float64(supportingIncluded) / float64(supportingTotal)
	}
	if structuralRequiredTotal > 0 {
		aggregate.RequiredStructuralRecall = float64(structuralRequiredIncluded) / float64(structuralRequiredTotal)
	}
	if structuralSupportingTotal > 0 {
		aggregate.SupportingStructuralRecall = float64(structuralSupportingIncluded) / float64(structuralSupportingTotal)
	}
	if selectionTotal > 0 {
		aggregate.IrrelevantSelectionRate = float64(irrelevantTotal) / float64(selectionTotal)
	}
	if structuralTotal > 0 {
		aggregate.IrrelevantStructuralRate = float64(structuralIrrelevant) / float64(structuralTotal)
	}
	if aggregate.ProfileCases > 0 {
		aggregate.ProfileMatchRate = float64(aggregate.ProfileMatches) / float64(aggregate.ProfileCases)
	}
	if aggregate.RankingCases > 0 {
		count := float64(aggregate.RankingCases)
		aggregate.RequiredRecallAt10 = rankingRequiredAt10 / count
		aggregate.RequiredRecallAt20 = rankingRequiredAt20 / count
		aggregate.MeanReciprocalRank = rankingReciprocalRank / count
		aggregate.RelevantRateAt10 = rankingRelevantAt10 / count
		aggregate.RelevantRateAt20 = rankingRelevantAt20 / count
	}
	sort.Float64s(latencies)
	aggregate.MedianLatencyMS = percentile(latencies, 0.5)
	aggregate.P95LatencyMS = percentile(latencies, 0.95)
	return aggregate
}

func percentile(sorted []float64, fraction float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}
	position := fraction * float64(len(sorted)-1)
	lower := int(position)
	upper := lower + 1
	if upper >= len(sorted) {
		return sorted[lower]
	}
	weight := position - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}
