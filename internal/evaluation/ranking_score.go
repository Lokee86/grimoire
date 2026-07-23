package evaluation

const (
	rankingCutoffShort = 10
	rankingCutoffLong  = 20
)

// ScoreRanking measures the retrieved order before exact-result merging,
// curation, and package fitting can hide or introduce failures.
func ScoreRanking(entry Case, candidates []Candidate) RankingMetrics {
	firstRank := firstRequiredRank(entry.Required, candidates)
	metrics := RankingMetrics{
		CandidateCount:       len(candidates),
		RequiredRecallAt10:   evidenceRecallAt(entry.Required, candidates, rankingCutoffShort),
		RequiredRecallAt20:   evidenceRecallAt(entry.Required, candidates, rankingCutoffLong),
		SupportingRecallAt10: evidenceRecallAt(entry.Supporting, candidates, rankingCutoffShort),
		SupportingRecallAt20: evidenceRecallAt(entry.Supporting, candidates, rankingCutoffLong),
		FirstRequiredRank:    firstRank,
		RelevantRateAt10:     relevantRateAt(entry, candidates, rankingCutoffShort),
		RelevantRateAt20:     relevantRateAt(entry, candidates, rankingCutoffLong),
	}
	if firstRank > 0 {
		metrics.ReciprocalRank = 1 / float64(firstRank)
	}
	return metrics
}

func firstRequiredRank(required []Evidence, candidates []Candidate) int {
	first := 0
	for _, evidence := range required {
		rank := evidenceRank(evidence, candidates)
		if rank > 0 && (first == 0 || rank < first) {
			first = rank
		}
	}
	return first
}

func evidenceRank(evidence Evidence, candidates []Candidate) int {
	for rank := 1; rank <= len(candidates); rank++ {
		if evidencePresent(evidence, candidates[:rank]) {
			return rank
		}
	}
	return 0
}

func evidenceRecallAt(group []Evidence, candidates []Candidate, cutoff int) float64 {
	if len(group) == 0 {
		return 0
	}
	candidates = candidatesAt(candidates, cutoff)
	recovered := 0
	for _, evidence := range group {
		if evidencePresent(evidence, candidates) {
			recovered++
		}
	}
	return float64(recovered) / float64(len(group))
}

func relevantRateAt(entry Case, candidates []Candidate, cutoff int) float64 {
	candidates = candidatesAt(candidates, cutoff)
	if len(candidates) == 0 {
		return 0
	}
	relevant := 0
	for _, candidate := range candidates {
		if matchesAny(entry.Required, candidate) || matchesAny(entry.Supporting, candidate) {
			relevant++
		}
	}
	return float64(relevant) / float64(len(candidates))
}

func candidatesAt(candidates []Candidate, cutoff int) []Candidate {
	if cutoff <= 0 || len(candidates) <= cutoff {
		return candidates
	}
	return candidates[:cutoff]
}
