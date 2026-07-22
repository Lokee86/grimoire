package compiler

import (
	"encoding/json"
	"fmt"

	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/tokenizer"
)

const (
	PackageVersion      = 2
	maxTokenCountPasses = 8
)

type Package struct {
	Version          int         `json:"version"`
	Query            string      `json:"query"`
	Budget           int         `json:"budget"`
	Tokenizer        string      `json:"tokenizer"`
	TokenCount       int         `json:"token_count"`
	IndexVersion     int         `json:"index_version"`
	RetrievalSources []string    `json:"retrieval_sources"`
	Selections       []Selection `json:"selections"`
	OmittedForBudget int         `json:"omitted_for_budget"`
}

type Selection struct {
	Path       string   `json:"path"`
	StartLine  int      `json:"start_line"`
	EndLine    int      `json:"end_line"`
	Score      int      `json:"score"`
	Reasons    []string `json:"reasons"`
	TokenCount int      `json:"token_count"`
	Content    string   `json:"content"`
}

func Compile(
	query string,
	budget int,
	indexVersion int,
	indexTokenizer string,
	candidates []retrieve.Candidate,
) (Package, error) {
	if budget <= 0 {
		return Package{}, fmt.Errorf("token budget must be positive")
	}
	if indexTokenizer != tokenizer.Name {
		return Package{}, fmt.Errorf("index tokenizer %q does not match %q", indexTokenizer, tokenizer.Name)
	}

	result := Package{
		Version:          PackageVersion,
		Query:            query,
		Budget:           budget,
		Tokenizer:        tokenizer.Name,
		IndexVersion:     indexVersion,
		RetrievalSources: []string{"lexical"},
		Selections:       make([]Selection, 0),
		OmittedForBudget: len(candidates),
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

	omitted := 0
	for candidateIndex, candidate := range candidates {
		selection := Selection{
			Path:       candidate.Chunk.Path,
			StartLine:  candidate.Chunk.StartLine,
			EndLine:    candidate.Chunk.EndLine,
			Score:      candidate.Score,
			Reasons:    candidate.Reasons,
			TokenCount: candidate.Chunk.TokenCount,
			Content:    candidate.Chunk.Text,
		}
		tentative := result
		tentative.Selections = append(append([]Selection(nil), result.Selections...), selection)
		tentative.OmittedForBudget = omitted + len(candidates) - candidateIndex - 1
		if err := stabilizeTokenCount(&tentative); err != nil {
			return Package{}, err
		}
		if tentative.TokenCount <= budget {
			result = tentative
			continue
		}
		omitted++
	}

	result.OmittedForBudget = omitted
	for {
		if err := stabilizeTokenCount(&result); err != nil {
			return Package{}, err
		}
		if result.TokenCount <= budget {
			return result, nil
		}
		if len(result.Selections) == 0 {
			return Package{}, fmt.Errorf(
				"token budget %d is smaller than the %d-token package metadata",
				budget,
				result.TokenCount,
			)
		}
		result.Selections = result.Selections[:len(result.Selections)-1]
		result.OmittedForBudget++
	}
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
