package compiler

import (
	"sort"

	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

type requiredFitPlan struct {
	id               string
	candidateIndexes []int
	intentPriority   int
	firstPosition    int
}

func buildRequiredFitPlans(candidates []retrieve.Candidate) []requiredFitPlan {
	if len(candidates) < 2 {
		return nil
	}
	identityIndexes := compilerCandidateIdentityIndexes(candidates)
	parents := make([]int, len(candidates))
	for index := range parents {
		parents[index] = index
	}
	var find func(int) int
	find = func(index int) int {
		if parents[index] != index {
			parents[index] = find(parents[index])
		}
		return parents[index]
	}
	union := func(left, right int) {
		leftRoot, rightRoot := find(left), find(right)
		if leftRoot == rightRoot {
			return
		}
		if leftRoot < rightRoot {
			parents[rightRoot] = leftRoot
		} else {
			parents[leftRoot] = rightRoot
		}
	}

	edgeMembers := make(map[int]struct{})
	for sourceIndex, candidate := range candidates {
		if candidate.Context == nil {
			continue
		}
		for _, link := range candidate.Context.Links {
			if !link.Required || link.Identity == "" {
				continue
			}
			targetIndex, exists := identityIndexes[link.Identity]
			if !exists || targetIndex == sourceIndex {
				continue
			}
			union(sourceIndex, targetIndex)
			edgeMembers[sourceIndex] = struct{}{}
			edgeMembers[targetIndex] = struct{}{}
		}
	}
	if len(edgeMembers) == 0 {
		return nil
	}

	components := make(map[int][]int)
	for index := range candidates {
		root := find(index)
		components[root] = append(components[root], index)
	}
	plans := make([]requiredFitPlan, 0, len(components))
	for root, indexes := range components {
		hasRequiredEdge := false
		for _, index := range indexes {
			if _, exists := edgeMembers[index]; exists {
				hasRequiredEdge = true
				break
			}
		}
		if !hasRequiredEdge || len(indexes) < 2 {
			continue
		}
		identities := make([]string, 0, len(indexes))
		priority := 4
		for _, index := range indexes {
			identities = append(identities, compilerCandidateIdentity(candidates[index]))
			if candidatePriority := candidateIntentPriority(candidates[index]); candidatePriority < priority {
				priority = candidatePriority
			}
		}
		plans = append(plans, requiredFitPlan{
			id:               evidence.StableID("required-link-group", identities...),
			candidateIndexes: indexes,
			intentPriority:   priority,
			firstPosition:    root,
		})
	}
	sort.Slice(plans, func(left, right int) bool {
		if plans[left].intentPriority != plans[right].intentPriority {
			return plans[left].intentPriority < plans[right].intentPriority
		}
		if plans[left].firstPosition != plans[right].firstPosition {
			return plans[left].firstPosition < plans[right].firstPosition
		}
		return plans[left].id < plans[right].id
	})
	return plans
}

func compilerCandidateIdentityIndexes(candidates []retrieve.Candidate) map[string]int {
	result := make(map[string]int, len(candidates)*2)
	for index, candidate := range candidates {
		for _, identity := range compilerCandidateIdentities(candidate) {
			if _, exists := result[identity]; !exists {
				result[identity] = index
			}
		}
	}
	return result
}

func compilerCandidateIdentities(candidate retrieve.Candidate) []string {
	rangeIdentity := evidence.RangeIdentity(
		candidate.Chunk.Path, candidate.Chunk.StartLine, candidate.Chunk.EndLine,
	)
	if candidate.Context == nil || candidate.Context.Identity == "" || candidate.Context.Identity == rangeIdentity {
		return []string{rangeIdentity}
	}
	return []string{candidate.Context.Identity, rangeIdentity}
}

func compilerCandidateIdentity(candidate retrieve.Candidate) string {
	if candidate.Context != nil && candidate.Context.Identity != "" {
		return candidate.Context.Identity
	}
	return evidence.RangeIdentity(candidate.Chunk.Path, candidate.Chunk.StartLine, candidate.Chunk.EndLine)
}

func requiredFitPlanIndexes(plans []requiredFitPlan) map[int]int {
	result := make(map[int]int)
	for planIndex, plan := range plans {
		for _, candidateIndex := range plan.candidateIndexes {
			result[candidateIndex] = planIndex
		}
	}
	return result
}

func protectedRequiredGroupCount(selections []Selection) int {
	seen := make(map[string]struct{})
	for _, selection := range selections {
		if selection.ProtectedLinkGroup != "" {
			seen[selection.ProtectedLinkGroup] = struct{}{}
		}
	}
	return len(seen)
}

func removeLastProtectedUnit(selections []Selection) []Selection {
	if len(selections) == 0 {
		return selections
	}
	last := selections[len(selections)-1]
	if last.ProtectedLinkGroup == "" {
		return selections[:len(selections)-1]
	}
	group := last.ProtectedLinkGroup
	kept := make([]Selection, 0, len(selections))
	for _, selection := range selections {
		if selection.ProtectedLinkGroup != group {
			kept = append(kept, selection)
		}
	}
	return kept
}
