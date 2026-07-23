package compiler

import (
	"encoding/json"
	"fmt"

	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/structure"
	"github.com/Lokee86/grimoire/internal/tokenizer"
)

const (
	PackageVersion      = 4
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
	Selections                 []Selection               `json:"selections"`
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
	if budget <= 0 {
		return Package{}, fmt.Errorf("token budget must be positive")
	}
	if indexTokenizer != tokenizer.Name {
		return Package{}, fmt.Errorf("index tokenizer %q does not match %q", indexTokenizer, tokenizer.Name)
	}

	result := Package{
		Version:                    PackageVersion,
		Query:                      query,
		Budget:                     budget,
		Tokenizer:                  tokenizer.Name,
		IndexVersion:               indexVersion,
		RetrievalSources:           append([]string(nil), retrievalSources...),
		StructuralEvidence:         make([]structure.Evidence, 0),
		Selections:                 make([]Selection, 0),
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

	omittedEvidence := 0
	omittedSelections := 0
	fitEvidence := func(evidenceIndex int) error {
		item := evidence[evidenceIndex]
		tentative := result
		tentative.StructuralEvidence = append(
			append([]structure.Evidence(nil), result.StructuralEvidence...), item,
		)
		tentative.StructuralSources = structuralEvidenceSources(tentative.StructuralEvidence)
		tentative.StructuralState = retainedProviderState(providerState, tentative.StructuralSources)
		tentative.OmittedStructuralForBudget = omittedEvidence + len(evidence) - evidenceIndex - 1
		if err := stabilizeTokenCount(&tentative); err != nil {
			return err
		}
		if tentative.TokenCount <= budget {
			result = tentative
		} else {
			omittedEvidence++
		}
		return nil
	}
	fitCandidate := func(candidateIndex int) error {
		candidate := candidates[candidateIndex]
		selection := Selection{
			Path:            candidate.Chunk.Path,
			StartLine:       candidate.Chunk.StartLine,
			EndLine:         candidate.Chunk.EndLine,
			Score:           candidate.Score,
			RetrievalSource: candidate.Source,
			RetrievalRank:   candidate.Rank,
			Reasons:         candidate.Reasons,
			TokenCount:      candidate.Chunk.TokenCount,
			Content:         candidate.Chunk.Text,
		}
		tentative := result
		tentative.Selections = append(append([]Selection(nil), result.Selections...), selection)
		tentative.OmittedForBudget = omittedSelections + len(candidates) - candidateIndex - 1
		if err := stabilizeTokenCount(&tentative); err != nil {
			return err
		}
		if tentative.TokenCount <= budget {
			result = tentative
		} else {
			omittedSelections++
		}
		return nil
	}

	evidenceStart := 0
	if len(evidence) > 0 {
		if err := fitEvidence(0); err != nil {
			return Package{}, err
		}
		evidenceStart = 1
	}
	candidateStart := 0
	if len(candidates) > 0 {
		if err := fitCandidate(0); err != nil {
			return Package{}, err
		}
		candidateStart = 1
	}
	for evidenceIndex := evidenceStart; evidenceIndex < len(evidence); evidenceIndex++ {
		if err := fitEvidence(evidenceIndex); err != nil {
			return Package{}, err
		}
	}
	result.OmittedStructuralForBudget = omittedEvidence
	for candidateIndex := candidateStart; candidateIndex < len(candidates); candidateIndex++ {
		if err := fitCandidate(candidateIndex); err != nil {
			return Package{}, err
		}
	}
	result.OmittedForBudget = omittedSelections
	for {
		if err := stabilizeTokenCount(&result); err != nil {
			return Package{}, err
		}
		if result.TokenCount <= budget {
			return result, nil
		}
		if len(result.Selections) > 0 {
			result.Selections = result.Selections[:len(result.Selections)-1]
			result.OmittedForBudget++
			continue
		}
		if len(result.StructuralEvidence) > 0 {
			result.StructuralEvidence = result.StructuralEvidence[:len(result.StructuralEvidence)-1]
			result.StructuralSources = structuralEvidenceSources(result.StructuralEvidence)
			result.StructuralState = retainedProviderState(providerState, result.StructuralSources)
			result.OmittedStructuralForBudget++
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
