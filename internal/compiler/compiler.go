package compiler

import (
	"encoding/json"
	"fmt"

	"github.com/Lokee86/grimoire/internal/assembly"
	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/structure"
	"github.com/Lokee86/grimoire/internal/tokenizer"
)

const (
	PackageVersion      = 6
	maxTokenCountPasses = 8
)

type Package struct {
	Version                    int                       `json:"version"`
	Query                      string                    `json:"query"`
	Budget                     int                       `json:"budget"`
	Tokenizer                  string                    `json:"tokenizer"`
	TokenCount                 int                       `json:"token_count"`
	IndexVersion               int                       `json:"index_version"`
	RetrievalSources           []string                  `json:"retrieval_sources"`
	StructuralSources          []string                  `json:"structural_sources,omitempty"`
	StructuralState            []structure.ProviderState `json:"structural_state,omitempty"`
	StructuralEvidence         []structure.Evidence      `json:"structural_evidence,omitempty"`
	Assembly                   *assembly.Decision        `json:"assembly,omitempty"`
	Selections                 []Selection               `json:"selections"`
	FacetProtection            bool                      `json:"facet_protection,omitempty"`
	FacetCompanionDepth        int                       `json:"facet_companion_depth,omitempty"`
	FacetsAvailable            int                       `json:"facets_available,omitempty"`
	FacetsProtected            int                       `json:"facets_protected,omitempty"`
	FacetsOmittedForBudget     int                       `json:"facets_omitted_for_budget,omitempty"`
	OmittedStructuralForBudget int                       `json:"omitted_structural_for_budget"`
	OmittedForBudget           int                       `json:"omitted_for_budget"`
}

type Selection struct {
	Path            string   `json:"path"`
	StartLine       int      `json:"start_line"`
	EndLine         int      `json:"end_line"`
	Score           float64  `json:"score"`
	RetrievalSource string   `json:"retrieval_source"`
	RetrievalRank   int      `json:"retrieval_rank"`
	Reasons         []string `json:"reasons"`
	FacetIDs        []string `json:"facet_ids,omitempty"`
	ProtectedFacet  string   `json:"protected_facet,omitempty"`
	TokenCount      int      `json:"token_count"`
	Content         string   `json:"content"`
}

func Compile(
	query string,
	budget int,
	indexVersion int,
	indexTokenizer string,
	retrievalSources []string,
	candidates []retrieve.Candidate,
) (Package, error) {
	return CompileWithEvidence(
		query, budget, indexVersion, indexTokenizer,
		retrievalSources, nil, nil, candidates,
	)
}

// CompileWithEvidence fits one structural fact and one source selection first,
// then continues through the remaining structural facts and source candidates.
// This keeps structural data first-class without allowing it to consume every
// token that could carry the underlying implementation evidence.
func CompileWithEvidence(
	query string,
	budget int,
	indexVersion int,
	indexTokenizer string,
	retrievalSources []string,
	providerState []structure.ProviderState,
	evidence []structure.Evidence,
	candidates []retrieve.Candidate,
) (Package, error) {
	return compileWithEvidence(
		query, budget, indexVersion, indexTokenizer, retrievalSources,
		providerState, evidence, nil, candidates, LegacyConfig(),
	)
}

// CompileAdaptiveWithEvidence retains the assembly decision in the versioned
// package so automatic stopping behavior is inspectable by consumers.
func CompileAdaptiveWithEvidence(
	query string,
	budget int,
	indexVersion int,
	indexTokenizer string,
	retrievalSources []string,
	providerState []structure.ProviderState,
	evidence []structure.Evidence,
	decision assembly.Decision,
	candidates []retrieve.Candidate,
) (Package, error) {
	return CompileAdaptiveWithEvidenceConfig(
		query, budget, indexVersion, indexTokenizer, retrievalSources,
		providerState, evidence, decision, candidates, DefaultConfig(),
	)
}

// CompileAdaptiveWithEvidenceConfig supports paired evaluation of final fitting
// behavior without changing production defaults.
func CompileAdaptiveWithEvidenceConfig(
	query string,
	budget int,
	indexVersion int,
	indexTokenizer string,
	retrievalSources []string,
	providerState []structure.ProviderState,
	evidence []structure.Evidence,
	decision assembly.Decision,
	candidates []retrieve.Candidate,
	config Config,
) (Package, error) {
	return compileWithEvidence(
		query, budget, indexVersion, indexTokenizer, retrievalSources,
		providerState, evidence, &decision, candidates, config,
	)
}

func compileWithEvidence(
	query string,
	budget int,
	indexVersion int,
	indexTokenizer string,
	retrievalSources []string,
	providerState []structure.ProviderState,
	evidence []structure.Evidence,
	decision *assembly.Decision,
	candidates []retrieve.Candidate,
	config Config,
) (Package, error) {
	if budget <= 0 {
		return Package{}, fmt.Errorf("token budget must be positive")
	}
	if indexTokenizer != tokenizer.Name {
		return Package{}, fmt.Errorf("index tokenizer %q does not match %q", indexTokenizer, tokenizer.Name)
	}

	config = normalizedConfig(config)
	facetProtection := config.ProtectFacets && decision != nil && decision.CoverageAware
	facetPlans := []facetFitPlan(nil)
	if facetProtection {
		facetPlans = buildFacetFitPlans(candidates)
	}
	result := Package{
		Version:                    PackageVersion,
		Query:                      query,
		Budget:                     budget,
		Tokenizer:                  tokenizer.Name,
		IndexVersion:               indexVersion,
		RetrievalSources:           append([]string(nil), retrievalSources...),
		StructuralEvidence:         make([]structure.Evidence, 0),
		Assembly:                   decision,
		Selections:                 make([]Selection, 0),
		FacetProtection:            facetProtection,
		FacetCompanionDepth:        config.CompanionDepth,
		FacetsAvailable:            len(facetPlans),
		FacetsOmittedForBudget:     len(facetPlans),
		OmittedStructuralForBudget: len(evidence),
		OmittedForBudget:           len(candidates),
	}
	if err := stabilizeTokenCount(&result); err != nil {
		return Package{}, err
	}
	if result.TokenCount > budget {
		return Package{}, fmt.Errorf(
			"token budget %d is smaller than the %d-token package metadata",
			budget,
			result.TokenCount,
		)
	}

	fitEvidence := func(evidenceIndex int) (bool, error) {
		item := evidence[evidenceIndex]
		tentative := result
		tentative.StructuralEvidence = append(
			append([]structure.Evidence(nil), result.StructuralEvidence...), item,
		)
		tentative.StructuralSources = structuralEvidenceSources(tentative.StructuralEvidence)
		tentative.StructuralState = retainedProviderState(providerState, tentative.StructuralSources)
		tentative.OmittedStructuralForBudget = len(evidence) - len(tentative.StructuralEvidence)
		if err := stabilizeTokenCount(&tentative); err != nil {
			return false, err
		}
		if tentative.TokenCount > budget {
			return false, nil
		}
		result = tentative
		return true, nil
	}
	protectedFacetCounts := make(map[string]int, len(facetPlans))
	protectedFacetFiles := make(map[string]map[string]struct{}, len(facetPlans))
	protectedFacetFileOrder := make(map[string][]string, len(facetPlans))
	protectedFacetFileEligible := make(map[string]map[string]bool, len(facetPlans))
	protectedFacetFileTerms := make(map[string]map[string]map[string]struct{}, len(facetPlans))
	fitCandidate := func(candidateIndex int, protectedFacet string) (bool, error) {
		candidate := candidates[candidateIndex]
		selection := Selection{
			Path:            candidate.Chunk.Path,
			StartLine:       candidate.Chunk.StartLine,
			EndLine:         candidate.Chunk.EndLine,
			Score:           candidate.Score,
			RetrievalSource: candidate.Source,
			RetrievalRank:   candidate.Rank,
			Reasons:         append([]string(nil), candidate.Reasons...),
			FacetIDs:        candidateFacetIDs(candidate),
			ProtectedFacet:  protectedFacet,
			TokenCount:      candidate.Chunk.TokenCount,
			Content:         candidate.Chunk.Text,
		}
		tentative := result
		tentative.Selections = append(append([]Selection(nil), result.Selections...), selection)
		tentative.OmittedForBudget = len(candidates) - len(tentative.Selections)
		if protectedFacet != "" && protectedFacetCounts[protectedFacet] == 0 {
			tentative.FacetsProtected++
			tentative.FacetsOmittedForBudget = tentative.FacetsAvailable - tentative.FacetsProtected
		}
		if err := stabilizeTokenCount(&tentative); err != nil {
			return false, err
		}
		if tentative.TokenCount > budget {
			return false, nil
		}
		result = tentative
		if protectedFacet != "" {
			protectedFacetCounts[protectedFacet]++
			files := protectedFacetFiles[protectedFacet]
			if files == nil {
				files = make(map[string]struct{})
				protectedFacetFiles[protectedFacet] = files
			}
			fileKey := candidateFileKey(candidate)
			if _, exists := files[fileKey]; !exists {
				files[fileKey] = struct{}{}
				protectedFacetFileOrder[protectedFacet] = append(protectedFacetFileOrder[protectedFacet], fileKey)
				eligible := protectedFacetFileEligible[protectedFacet]
				if eligible == nil {
					eligible = make(map[string]bool)
					protectedFacetFileEligible[protectedFacet] = eligible
				}
				eligible[fileKey] = candidateSupportsCompanions(candidate)
				facetTerms := protectedFacetFileTerms[protectedFacet]
				if facetTerms == nil {
					facetTerms = make(map[string]map[string]struct{})
					protectedFacetFileTerms[protectedFacet] = facetTerms
				}
				facetTerms[fileKey] = make(map[string]struct{})
			}
			addRankingTerms(candidate, protectedFacetFileTerms[protectedFacet][fileKey])
		}
		return true, nil
	}

	evidenceStart := 0
	if len(evidence) > 0 {
		if _, err := fitEvidence(0); err != nil {
			return Package{}, err
		}
		evidenceStart = 1
	}
	if facetProtection && len(facetPlans) > 0 {
		attempted := make(map[int]struct{}, len(candidates))
		for _, plan := range facetPlans {
			for _, candidateIndex := range plan.candidateIndexes {
				if _, exists := attempted[candidateIndex]; exists {
					continue
				}
				attempted[candidateIndex] = struct{}{}
				fitted, err := fitCandidate(candidateIndex, plan.facet)
				if err != nil {
					return Package{}, err
				}
				if fitted {
					break
				}
			}
		}
		for companion := 0; companion < config.CompanionDepth; companion++ {
			for _, plan := range facetPlans {
				for _, fileKey := range protectedFacetFileOrder[plan.facet] {
					if !protectedFacetFileEligible[plan.facet][fileKey] {
						continue
					}
					covered := protectedFacetFileTerms[plan.facet][fileKey]
					for _, candidateIndex := range plan.candidateIndexes {
						if _, exists := attempted[candidateIndex]; exists {
							continue
						}
						candidate := candidates[candidateIndex]
						if candidateFileKey(candidate) != fileKey || !addsRankingTerm(candidate, covered) {
							continue
						}
						attempted[candidateIndex] = struct{}{}
						fitted, err := fitCandidate(candidateIndex, plan.facet)
						if err != nil {
							return Package{}, err
						}
						if fitted {
							break
						}
					}
				}
			}
		}
		for evidenceIndex := evidenceStart; evidenceIndex < len(evidence); evidenceIndex++ {
			if _, err := fitEvidence(evidenceIndex); err != nil {
				return Package{}, err
			}
		}
		for candidateIndex := range candidates {
			if _, exists := attempted[candidateIndex]; exists {
				continue
			}
			if _, err := fitCandidate(candidateIndex, ""); err != nil {
				return Package{}, err
			}
		}
	} else {
		candidateStart := 0
		if len(candidates) > 0 {
			if _, err := fitCandidate(0, ""); err != nil {
				return Package{}, err
			}
			candidateStart = 1
		}
		for evidenceIndex := evidenceStart; evidenceIndex < len(evidence); evidenceIndex++ {
			if _, err := fitEvidence(evidenceIndex); err != nil {
				return Package{}, err
			}
		}
		for candidateIndex := candidateStart; candidateIndex < len(candidates); candidateIndex++ {
			if _, err := fitCandidate(candidateIndex, ""); err != nil {
				return Package{}, err
			}
		}
	}
	result.OmittedStructuralForBudget = len(evidence) - len(result.StructuralEvidence)
	result.OmittedForBudget = len(candidates) - len(result.Selections)
	result.FacetsOmittedForBudget = result.FacetsAvailable - result.FacetsProtected
	for {
		if err := stabilizeTokenCount(&result); err != nil {
			return Package{}, err
		}
		if result.TokenCount <= budget {
			return result, nil
		}
		if index := lastUnprotectedSelection(result.Selections); index >= 0 {
			result.Selections = append(result.Selections[:index], result.Selections[index+1:]...)
			result.OmittedForBudget = len(candidates) - len(result.Selections)
			continue
		}
		if len(result.StructuralEvidence) > 1 {
			result.StructuralEvidence = result.StructuralEvidence[:len(result.StructuralEvidence)-1]
			result.StructuralSources = structuralEvidenceSources(result.StructuralEvidence)
			result.StructuralState = retainedProviderState(providerState, result.StructuralSources)
			result.OmittedStructuralForBudget = len(evidence) - len(result.StructuralEvidence)
			continue
		}
		if len(result.Selections) > 0 {
			result.Selections = result.Selections[:len(result.Selections)-1]
			result.FacetsProtected = protectedFacetCount(result.Selections)
			result.FacetsOmittedForBudget = result.FacetsAvailable - result.FacetsProtected
			result.OmittedForBudget = len(candidates) - len(result.Selections)
			continue
		}
		if len(result.StructuralEvidence) > 0 {
			result.StructuralEvidence = result.StructuralEvidence[:len(result.StructuralEvidence)-1]
			result.StructuralSources = structuralEvidenceSources(result.StructuralEvidence)
			result.StructuralState = retainedProviderState(providerState, result.StructuralSources)
			result.OmittedStructuralForBudget = len(evidence) - len(result.StructuralEvidence)
			continue
		}
		return Package{}, fmt.Errorf(
			"token budget %d is smaller than the %d-token package metadata",
			budget,
			result.TokenCount,
		)
	}
}

func retainedProviderState(
	states []structure.ProviderState,
	sources []string,
) []structure.ProviderState {
	byProvider := make(map[string]structure.ProviderState, len(states))
	for _, state := range states {
		if state.Provider == "" || state.Snapshot == "" {
			continue
		}
		byProvider[state.Provider] = state
	}
	result := make([]structure.ProviderState, 0, len(sources))
	for _, source := range sources {
		state, exists := byProvider[source]
		if exists {
			result = append(result, state)
		}
	}
	return result
}

func structuralEvidenceSources(evidence []structure.Evidence) []string {
	seen := make(map[string]struct{}, len(evidence))
	result := make([]string, 0, len(evidence))
	for _, item := range evidence {
		if item.Provider == "" {
			continue
		}
		if _, exists := seen[item.Provider]; exists {
			continue
		}
		seen[item.Provider] = struct{}{}
		result = append(result, item.Provider)
	}
	return result
}

func Marshal(pkg Package) ([]byte, error) {
	data, err := marshalPackage(pkg)
	if err != nil {
		return nil, err
	}
	count, err := tokenizer.Count(string(data))
	if err != nil {
		return nil, err
	}
	if count != pkg.TokenCount {
		return nil, fmt.Errorf("package token count changed from %d to %d", pkg.TokenCount, count)
	}
	return data, nil
}

func stabilizeTokenCount(pkg *Package) error {
	for range maxTokenCountPasses {
		data, err := marshalPackage(*pkg)
		if err != nil {
			return err
		}
		count, err := tokenizer.Count(string(data))
		if err != nil {
			return err
		}
		if count == pkg.TokenCount {
			return nil
		}
		pkg.TokenCount = count
	}
	return fmt.Errorf("package token count did not stabilize")
}

func marshalPackage(pkg Package) ([]byte, error) {
	data, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal context package: %w", err)
	}
	return append(data, '\n'), nil
}
