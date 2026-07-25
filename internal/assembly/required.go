package assembly

import (
	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

type requiredCandidateLink struct {
	from string
	to   string
}

// prioritizeRequiredLinks keeps a linked source span immediately behind the
// first candidate that requires it. This preserves the existing candidate
// order for unrelated evidence while making required extraction companions
// reachable before normal coverage stopping and candidate caps.
func prioritizeRequiredLinks(candidates []retrieve.Candidate) ([]retrieve.Candidate, []requiredCandidateLink) {
	ordered := append([]retrieve.Candidate(nil), candidates...)
	identityIndexes := candidateIdentityIndexes(ordered)
	links := resolvableRequiredLinks(ordered, identityIndexes)
	if len(links) == 0 {
		return ordered, nil
	}

	bySource := make(map[string][]string)
	for _, link := range links {
		bySource[link.from] = append(bySource[link.from], link.to)
	}
	emitted := make(map[int]struct{}, len(ordered))
	result := make([]retrieve.Candidate, 0, len(ordered))
	var emit func(int)
	emit = func(index int) {
		if _, exists := emitted[index]; exists {
			return
		}
		emitted[index] = struct{}{}
		candidate := ordered[index]
		result = append(result, candidate)
		identity := requiredCandidateIdentity(candidate)
		for _, linkedIdentity := range bySource[identity] {
			if linkedIndex, exists := identityIndexes[linkedIdentity]; exists {
				emit(linkedIndex)
			}
		}
	}
	for index := range ordered {
		emit(index)
	}
	return result, links
}

func resolvableRequiredLinks(
	candidates []retrieve.Candidate,
	identityIndexes map[string]int,
) []requiredCandidateLink {
	seen := make(map[requiredCandidateLink]struct{})
	links := make([]requiredCandidateLink, 0)
	for _, candidate := range candidates {
		if candidate.Context == nil {
			continue
		}
		from := requiredCandidateIdentity(candidate)
		if from == "" {
			continue
		}
		for _, link := range candidate.Context.Links {
			if !link.Required || link.Identity == "" || link.Identity == from {
				continue
			}
			if _, exists := identityIndexes[link.Identity]; !exists {
				continue
			}
			edge := requiredCandidateLink{from: from, to: link.Identity}
			if _, exists := seen[edge]; exists {
				continue
			}
			seen[edge] = struct{}{}
			links = append(links, edge)
		}
	}
	return links
}

func candidateIdentityIndexes(candidates []retrieve.Candidate) map[string]int {
	result := make(map[string]int, len(candidates)*2)
	for index, candidate := range candidates {
		for _, identity := range candidateIdentities(candidate) {
			if _, exists := result[identity]; !exists {
				result[identity] = index
			}
		}
	}
	return result
}

func candidateIdentities(candidate retrieve.Candidate) []string {
	rangeIdentity := evidence.RangeIdentity(
		candidate.Chunk.Path, candidate.Chunk.StartLine, candidate.Chunk.EndLine,
	)
	if candidate.Context == nil || candidate.Context.Identity == "" || candidate.Context.Identity == rangeIdentity {
		return []string{rangeIdentity}
	}
	return []string{candidate.Context.Identity, rangeIdentity}
}

func requiredCandidateIdentity(candidate retrieve.Candidate) string {
	if candidate.Context != nil && candidate.Context.Identity != "" {
		return candidate.Context.Identity
	}
	return evidence.RangeIdentity(candidate.Chunk.Path, candidate.Chunk.StartLine, candidate.Chunk.EndLine)
}

func representedRequiredLinks(
	selected []retrieve.Candidate,
	links []requiredCandidateLink,
) int {
	selectedIdentities := make(map[string]struct{}, len(selected)*2)
	for _, candidate := range selected {
		for _, identity := range candidateIdentities(candidate) {
			selectedIdentities[identity] = struct{}{}
		}
	}
	represented := 0
	for _, link := range links {
		if _, sourceSelected := selectedIdentities[link.from]; !sourceSelected {
			continue
		}
		if _, targetSelected := selectedIdentities[link.to]; targetSelected {
			represented++
		}
	}
	return represented
}

func selectedRequiredLinksSatisfied(
	selected []retrieve.Candidate,
	links []requiredCandidateLink,
) bool {
	selectedIdentities := make(map[string]struct{}, len(selected)*2)
	for _, candidate := range selected {
		for _, identity := range candidateIdentities(candidate) {
			selectedIdentities[identity] = struct{}{}
		}
	}
	for _, link := range links {
		if _, sourceSelected := selectedIdentities[link.from]; !sourceSelected {
			continue
		}
		if _, targetSelected := selectedIdentities[link.to]; !targetSelected {
			return false
		}
	}
	return true
}
