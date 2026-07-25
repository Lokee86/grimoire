package compiler

import (
	"sort"
	"strings"

	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

const facetPrefix = "facet:"

type facetFitPlan struct {
	facet            string
	candidateIndexes []int
	intentPriority   int
	firstPosition    int
}

func buildFacetFitPlans(candidates []retrieve.Candidate) []facetFitPlan {
	plansByFacet := make(map[string]*facetFitPlan)
	for index, candidate := range candidates {
		for _, facet := range candidateFacetIDs(candidate) {
			plan := plansByFacet[facet]
			if plan == nil {
				plan = &facetFitPlan{
					facet: facet, intentPriority: candidateIntentPriority(candidate),
					firstPosition: index,
				}
				plansByFacet[facet] = plan
			} else if priority := candidateIntentPriority(candidate); priority < plan.intentPriority {
				plan.intentPriority = priority
			}
			plan.candidateIndexes = append(plan.candidateIndexes, index)
		}
	}
	plans := make([]facetFitPlan, 0, len(plansByFacet))
	for _, plan := range plansByFacet {
		lexicalPlan := false
		for _, candidateIndex := range plan.candidateIndexes {
			if len(candidateRankingTerms(candidates[candidateIndex])) > 0 {
				lexicalPlan = true
				break
			}
		}
		if lexicalPlan {
			sort.SliceStable(plan.candidateIndexes, func(left, right int) bool {
				leftPriority := candidateRolePriority(candidates[plan.candidateIndexes[left]])
				rightPriority := candidateRolePriority(candidates[plan.candidateIndexes[right]])
				if leftPriority != rightPriority {
					return leftPriority < rightPriority
				}
				return plan.candidateIndexes[left] < plan.candidateIndexes[right]
			})
			plans = append(plans, *plan)
			continue
		}

		type filePlan struct {
			key              string
			candidateIndexes []int
			rolePriority     int
			firstPosition    int
		}
		filesByKey := make(map[string]*filePlan)
		files := make([]*filePlan, 0)
		for _, candidateIndex := range plan.candidateIndexes {
			candidate := candidates[candidateIndex]
			key := candidateFileKey(candidate)
			file := filesByKey[key]
			if file == nil {
				file = &filePlan{
					key: key, rolePriority: candidateRolePriority(candidate),
					firstPosition: candidateIndex,
				}
				filesByKey[key] = file
				files = append(files, file)
			} else if priority := candidateRolePriority(candidate); priority < file.rolePriority {
				file.rolePriority = priority
			}
			file.candidateIndexes = append(file.candidateIndexes, candidateIndex)
		}
		sort.SliceStable(files, func(left, right int) bool {
			if files[left].rolePriority != files[right].rolePriority {
				return files[left].rolePriority < files[right].rolePriority
			}
			if files[left].firstPosition != files[right].firstPosition {
				return files[left].firstPosition < files[right].firstPosition
			}
			return files[left].key < files[right].key
		})
		plan.candidateIndexes = plan.candidateIndexes[:0]
		for _, file := range files {
			sort.SliceStable(file.candidateIndexes, func(left, right int) bool {
				leftCandidate := candidates[file.candidateIndexes[left]]
				rightCandidate := candidates[file.candidateIndexes[right]]
				leftPriority := candidateRolePriority(leftCandidate)
				rightPriority := candidateRolePriority(rightCandidate)
				if leftPriority != rightPriority {
					return leftPriority < rightPriority
				}
				leftPriority = candidateSemanticImplementationPriority(leftCandidate)
				rightPriority = candidateSemanticImplementationPriority(rightCandidate)
				if leftPriority != rightPriority {
					return leftPriority < rightPriority
				}
				return file.candidateIndexes[left] < file.candidateIndexes[right]
			})
			plan.candidateIndexes = append(plan.candidateIndexes, file.candidateIndexes...)
		}
		plans = append(plans, *plan)
	}
	sort.Slice(plans, func(left, right int) bool {
		if plans[left].intentPriority != plans[right].intentPriority {
			return plans[left].intentPriority < plans[right].intentPriority
		}
		if plans[left].firstPosition != plans[right].firstPosition {
			return plans[left].firstPosition < plans[right].firstPosition
		}
		return plans[left].facet < plans[right].facet
	})
	return plans
}

func candidateFacetIDs(candidate retrieve.Candidate) []string {
	if candidate.Context == nil {
		return nil
	}
	result := make([]string, 0, len(candidate.Context.GroupIDs))
	for _, groupID := range candidate.Context.GroupIDs {
		if strings.HasPrefix(groupID, facetPrefix) {
			result = append(result, groupID)
		}
	}
	return result
}

func lastUnprotectedSelection(selections []Selection) int {
	for index := len(selections) - 1; index >= 0; index-- {
		if selections[index].ProtectedFacet == "" && selections[index].ProtectedLinkGroup == "" {
			return index
		}
	}
	return -1
}

func candidateFileKey(candidate retrieve.Candidate) string {
	return strings.ToLower(strings.ReplaceAll(candidate.Chunk.Path, "\\", "/"))
}

func candidateSupportsAdditionalMechanismFiles(candidate retrieve.Candidate) bool {
	if candidate.Context == nil {
		return false
	}
	for _, intent := range candidate.Context.Intents {
		if intent == evidence.IntentMixed {
			return false
		}
		if intent == evidence.IntentMechanism {
			return true
		}
	}
	return false
}

func candidateSupportsCompanions(candidate retrieve.Candidate) bool {
	if candidate.Context == nil {
		return false
	}
	eligible := false
	for _, intent := range candidate.Context.Intents {
		if intent == evidence.IntentMixed {
			return false
		}
		switch intent {
		case evidence.IntentMechanism, evidence.IntentCallChain, evidence.IntentDirectLocation:
			eligible = true
		}
	}
	return eligible
}

func candidateRankingTerms(candidate retrieve.Candidate) []string {
	terms := make([]string, 0, len(candidate.ScoreDetails))
	seen := make(map[string]struct{})
	for _, detail := range candidate.ScoreDetails {
		name := detail.Name
		term := ""
		switch {
		case strings.HasPrefix(name, "BM25 content matches "):
			term = strings.TrimPrefix(name, "BM25 content matches ")
		case strings.HasPrefix(name, "declaration alias "):
			term = strings.TrimPrefix(name, "declaration alias ")
			if separator := strings.Index(term, " -> "); separator >= 0 {
				term = term[:separator]
			}
		case strings.HasPrefix(name, "leading line matches "):
			term = strings.TrimPrefix(name, "leading line matches ")
		}
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}
		if _, exists := seen[term]; exists {
			continue
		}
		seen[term] = struct{}{}
		terms = append(terms, term)
	}
	return terms
}

func candidateSemanticImplementationPriority(candidate retrieve.Candidate) int {
	if len(candidateRankingTerms(candidate)) > 0 {
		return 0
	}
	if !candidateSupportsCompanions(candidate) {
		return 2
	}
	if candidateContainsDeclaration(candidate) {
		return 0
	}
	return 1
}

func candidateContainsDeclaration(candidate retrieve.Candidate) bool {
	text := strings.ToLower(candidate.Chunk.Text)
	for _, marker := range []string{"func ", "function ", "def ", "class ", "type ", "interface ", "fn "} {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}

func addsRankingTerm(candidate retrieve.Candidate, covered map[string]struct{}) bool {
	for _, term := range candidateRankingTerms(candidate) {
		if _, exists := covered[term]; !exists {
			return true
		}
	}
	return false
}

func addRankingTerms(candidate retrieve.Candidate, covered map[string]struct{}) {
	for _, term := range candidateRankingTerms(candidate) {
		covered[term] = struct{}{}
	}
}

func packageProtectedFacetCount(pkg Package) int {
	return min(pkg.FacetsAvailable, protectedFacetCount(pkg.Selections))
}

func protectedFacetCount(selections []Selection) int {
	seen := make(map[string]struct{})
	for _, selection := range selections {
		if selection.ProtectedFacet != "" {
			seen[selection.ProtectedFacet] = struct{}{}
		}
		if selection.ProtectedLinkGroup == "" {
			continue
		}
		for _, facet := range selection.FacetIDs {
			seen[facet] = struct{}{}
		}
	}
	return len(seen)
}

func candidateRolePriority(candidate retrieve.Candidate) int {
	if candidate.Context == nil {
		return 4
	}
	priority := 4
	for _, role := range candidate.Context.Roles {
		switch role {
		case evidence.RolePrimary:
			return 0
		case evidence.RoleStructural:
			if priority > 1 {
				priority = 1
			}
		case evidence.RoleSupporting:
			if priority > 2 {
				priority = 2
			}
		case evidence.RoleContext:
			if priority > 3 {
				priority = 3
			}
		}
	}
	return priority
}

func candidateIntentPriority(candidate retrieve.Candidate) int {
	if candidate.Context == nil {
		return 4
	}
	priority := 4
	for _, intent := range candidate.Context.Intents {
		switch intent {
		case evidence.IntentMechanism, evidence.IntentCallChain:
			return 0
		case evidence.IntentArchitecture:
			if priority > 1 {
				priority = 1
			}
		case evidence.IntentDirectLocation:
			if priority > 2 {
				priority = 2
			}
		case evidence.IntentMixed:
			if priority > 3 {
				priority = 3
			}
		}
	}
	return priority
}
