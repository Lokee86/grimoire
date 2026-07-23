package queryshape

import "strings"

// Analyze deterministically profiles a query and returns a shadow retrieval
// policy. Neither result changes candidate order or context assembly.
func Analyze(input Input) (Profile, RetrievalPolicy) {
	query := strings.ToLower(strings.TrimSpace(input.Query))
	tasks := recognizedTasks(query)
	symbols, paths, errors, configs, quoted := exactCounts(input.Exact, query)
	candidates := input.Candidates
	if len(candidates) == 0 {
		candidates = input.Ranked
	}
	profile := Profile{
		ExactSymbolMatches:  symbols,
		ExactPathMatches:    paths,
		ExactErrorMatches:   errors,
		RecognizedTaskTerms: tasks,
		MatchedSubsystems:   candidateRegions(candidates),
		MatchedGraphRegions: structuralRegions(input.Structural),
		TopScoreGap:         topScoreGap(input.Ranked),
		CandidateDispersion: candidateDispersion(candidates),
	}
	profile.Specificity = specificityLevel(profile, configs, quoted, query)
	profile.Breadth = breadthLevel(profile, query)
	profile.Ambiguity = ambiguityLevel(profile, len(input.Ranked))
	return profile, policyFor(profile, input.RequestedBudget)
}

func specificityLevel(profile Profile, configs, quoted int, query string) Level {
	score := 0
	if profile.ExactSymbolMatches > 0 {
		score += 3
	}
	if profile.ExactPathMatches > 0 {
		score += 3
	}
	if profile.ExactErrorMatches > 0 {
		score += 2
	}
	if configs > 0 {
		score += 2
	}
	if quoted > 0 {
		score++
	}
	for _, task := range profile.RecognizedTaskTerms {
		switch task {
		case "location", "execution-flow", "architecture", "mechanism":
			score += 2
		case "debugging", "modification", "verification":
			score++
		}
	}
	if profile.TopScoreGap >= 0.35 {
		score++
	}
	if len(queryWords(query)) <= 8 {
		score++
	}
	return scoreLevel(score, 2, 5)
}

func breadthLevel(profile Profile, query string) Level {
	score := 0
	if containsTask(profile.RecognizedTaskTerms, "architecture") {
		score += 3
	}
	if containsTask(profile.RecognizedTaskTerms, "execution-flow") {
		score += 3
	}
	if containsTask(profile.RecognizedTaskTerms, "mechanism") {
		score++
	}
	words := len(queryWords(query))
	if words >= 12 {
		score++
	}
	if words >= 30 {
		score++
	}
	if len(profile.RecognizedTaskTerms) >= 2 {
		score++
	}
	if len(profile.MatchedGraphRegions) >= 3 {
		score += 2
	} else if len(profile.MatchedGraphRegions) == 2 {
		score++
	}
	confidentRetrieval := profile.Specificity == LevelHigh ||
		profile.ExactSymbolMatches+profile.ExactPathMatches+profile.ExactErrorMatches > 0 ||
		profile.TopScoreGap >= 0.35
	if confidentRetrieval {
		if len(profile.MatchedSubsystems) >= 3 {
			score += 2
		} else if len(profile.MatchedSubsystems) == 2 {
			score++
		}
		if profile.CandidateDispersion >= 0.45 {
			score += 2
		} else if profile.CandidateDispersion >= 0.20 {
			score++
		}
	}
	if containsTask(profile.RecognizedTaskTerms, "location") {
		score -= 3
	}
	if profile.Specificity == LevelHigh && len(profile.MatchedSubsystems) <= 1 &&
		!containsTask(profile.RecognizedTaskTerms, "architecture") &&
		!containsTask(profile.RecognizedTaskTerms, "execution-flow") {
		score -= 2
	}
	return scoreLevel(score, 2, 4)
}

func ambiguityLevel(profile Profile, rankedCount int) Level {
	score := 0
	if profile.ExactSymbolMatches+profile.ExactPathMatches+profile.ExactErrorMatches == 0 {
		score++
	}
	if rankedCount == 0 {
		score += 3
	} else if profile.TopScoreGap < 0.10 {
		score += 2
	} else if profile.TopScoreGap < 0.25 {
		score++
	}
	if profile.CandidateDispersion >= 0.45 {
		score += 2
	} else if profile.CandidateDispersion >= 0.20 {
		score++
	}
	if len(profile.MatchedGraphRegions) >= 3 {
		score++
	}
	if len(profile.RecognizedTaskTerms) == 0 {
		score++
	}
	if len(profile.RecognizedTaskTerms) >= 3 {
		score++
	}
	if containsTask(profile.RecognizedTaskTerms, "location") {
		score--
	}
	if containsTask(profile.RecognizedTaskTerms, "mechanism") ||
		containsTask(profile.RecognizedTaskTerms, "architecture") ||
		containsTask(profile.RecognizedTaskTerms, "execution-flow") {
		score--
	}
	if profile.Specificity == LevelHigh {
		score -= 2
	}
	return scoreLevel(score, 2, 4)
}

func scoreLevel(score, medium, high int) Level {
	if score >= high {
		return LevelHigh
	}
	if score >= medium {
		return LevelMedium
	}
	return LevelLow
}

func policyFor(profile Profile, budget int) RetrievalPolicy {
	scope := ScopeBounded
	if profile.Specificity == LevelHigh && profile.Breadth == LevelLow && profile.Ambiguity != LevelHigh {
		scope = ScopeFocused
	} else if profile.Breadth == LevelHigh ||
		(profile.Breadth == LevelMedium && profile.Ambiguity == LevelHigh) {
		scope = ScopeExploratory
	}
	policy := RetrievalPolicy{
		Shadow: true, Scope: scope, BudgetMode: "fixed",
		TargetTokens: budget, MaximumTokens: budget,
	}
	if budget <= 0 {
		policy.BudgetMode = "automatic-shadow"
		applyAutomaticBudget(&policy)
	}
	applyScopePolicy(&policy)
	return policy
}

func applyScopePolicy(policy *RetrievalPolicy) {
	switch policy.Scope {
	case ScopeFocused:
		policy.ExpansionRadius = 1
		policy.RequiredEvidence = []string{"exact-match", "implementation", "direct-relationships"}
		policy.DiversityRequirement = 1
		policy.StopConditions = []string{"exact evidence covered", "direct relationships represented", "remaining candidates are redundant"}
	case ScopeExploratory:
		policy.ExpansionRadius = 3
		policy.RequiredEvidence = []string{"subsystem-entry-points", "architecture-context", "representative-implementations", "cross-region-relationships"}
		policy.DiversityRequirement = 3
		policy.StopConditions = []string{"major graph regions represented", "cross-region relationships covered", "configured maximum reached"}
	default:
		policy.ExpansionRadius = 2
		policy.RequiredEvidence = []string{"implementation", "relationships", "tests-or-contracts", "subsystem-context"}
		policy.DiversityRequirement = 2
		policy.StopConditions = []string{"primary subsystem covered", "verification evidence represented", "additional candidates add no new evidence type"}
	}
}
