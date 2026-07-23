package app

import (
	"context"
	"fmt"
	"time"

	"github.com/Lokee86/grimoire/internal/compiler"
	"github.com/Lokee86/grimoire/internal/embedding"
	"github.com/Lokee86/grimoire/internal/evaluation"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/selection"
	"github.com/Lokee86/grimoire/internal/structure"
)

type evaluatedContext struct {
	Package  compiler.Package
	Stages   evaluation.Stages
	Timings  evaluation.Timings
	Warnings []string
}

type evaluatedContextOptions struct {
	Mode         string
	Query        string
	Budget       int
	Limit        int
	ProbeLimit   int
	StatePath    string
	Endpoint     string
	EnginePath   string
	Structural   structuralContextOptions
	QueryOptions embedding.QueryOptions
}

func evaluateContext(
	ctx context.Context,
	snapshot index.Snapshot,
	options evaluatedContextOptions,
) (evaluatedContext, error) {
	var result evaluatedContext
	totalStart := time.Now()
	var base []retrieve.Candidate
	var broad []retrieve.Candidate

	if options.Mode == "lexical" {
		searchStart := time.Now()
		base = retrieve.Search(snapshot, options.Query, options.Limit)
		result.Timings.LexicalSearchMS = durationMS(time.Since(searchStart))
		probeLimit := options.ProbeLimit
		if probeLimit <= 0 {
			probeLimit = options.Limit
		}
		probeStart := time.Now()
		broad = retrieve.Search(snapshot, options.Query, probeLimit)
		result.Timings.DiagnosticProbeMS = durationMS(time.Since(probeStart))
	} else {
		semantic, err := semanticCandidatesForEvaluation(
			ctx, snapshot, options.StatePath, options.Query, options.Endpoint,
			options.EnginePath, options.Limit, options.ProbeLimit, options.QueryOptions,
		)
		if err != nil {
			return result, err
		}
		base = semantic.Candidates
		broad = semantic.BroadProbe
		result.Timings.SnapshotValidationMS = durationMS(semantic.Metrics.SnapshotValidation)
		result.Timings.EmbeddingMS = durationMS(semantic.Metrics.Embedding)
		result.Timings.VectorSearchMS = durationMS(semantic.Metrics.VectorSearch)
		result.Timings.CandidateMergeMS = durationMS(semantic.Metrics.CandidateMerge)
		result.Timings.DiagnosticProbeMS = durationMS(semantic.Metrics.DiagnosticProbe)
	}

	structural := collectStructuralContext(context.Background(), snapshot, options.Query, options.Structural)
	result.Warnings = append(result.Warnings, structural.Warnings...)
	result.Timings.LexiconSearchMS = durationMS(structural.LexiconTime)
	result.Timings.ArcanaSearchMS = durationMS(structural.ArcanaTime)
	result.Timings.StructuralProviderMS = durationMS(structural.TotalTime)

	exactStart := time.Now()
	exact := retrieve.Exact(snapshot, options.Query, min(options.Limit, maxExactCandidates))
	result.Timings.ExactRecoveryMS = durationMS(time.Since(exactStart))

	mergeStart := time.Now()
	merged := mergeContextProviders(options.Limit, exact, base, structural.Lexicon.Candidates)
	result.Timings.CandidateMergeMS += durationMS(time.Since(mergeStart))

	curationStart := time.Now()
	curated := selection.Curate(snapshot, merged)
	result.Timings.CurationMS = durationMS(time.Since(curationStart))

	compileStart := time.Now()
	pkg, err := compiler.CompileWithEvidence(
		options.Query, options.Budget, snapshot.Version, snapshot.Tokenizer,
		contextCandidateSources(curated), structural.ProviderState, structural.Combined, curated,
	)
	result.Timings.PackageCompilationMS = durationMS(time.Since(compileStart))
	result.Timings.SelectionCompilationMS = result.Timings.CurationMS + result.Timings.PackageCompilationMS
	result.Timings.TotalMS = durationMS(time.Since(totalStart)) - result.Timings.DiagnosticProbeMS
	if result.Timings.TotalMS < 0 {
		result.Timings.TotalMS = 0
	}
	if err != nil {
		return result, fmt.Errorf("compile context package: %w", err)
	}
	result.Package = pkg
	result.Stages = evaluation.Stages{
		Indexed:            chunksToEvaluation(snapshot.AllChunks()),
		BroadProbe:         candidatesToEvaluation(broad),
		Retrieved:          candidatesToEvaluation(mergeContextProviders(options.Limit, nil, base, structural.Lexicon.Candidates)),
		Exact:              candidatesToEvaluation(exact),
		Merged:             candidatesToEvaluation(merged),
		Curated:            candidatesToEvaluation(curated),
		Included:           selectionsToEvaluation(pkg.Selections),
		StructuralProduced: append(append([]structure.Evidence(nil), structural.Lexicon.Evidence...), structural.Arcana...),
		StructuralComposed: append([]structure.Evidence(nil), structural.Combined...),
		StructuralIncluded: append([]structure.Evidence(nil), pkg.StructuralEvidence...),
	}
	return result, nil
}

func chunksToEvaluation(chunks []index.Chunk) []evaluation.Candidate {
	result := make([]evaluation.Candidate, 0, len(chunks))
	for _, chunk := range chunks {
		result = append(result, evaluation.Candidate{
			Path:       chunk.Path,
			StartLine:  chunk.StartLine,
			EndLine:    chunk.EndLine,
			Text:       chunk.Text,
			TokenCount: chunk.TokenCount,
		})
	}
	return result
}

func candidatesToEvaluation(candidates []retrieve.Candidate) []evaluation.Candidate {
	result := make([]evaluation.Candidate, 0, len(candidates))
	for _, candidate := range candidates {
		result = append(result, evaluation.Candidate{
			Path:            candidate.Chunk.Path,
			StartLine:       candidate.Chunk.StartLine,
			EndLine:         candidate.Chunk.EndLine,
			Text:            candidate.Chunk.Text,
			RetrievalSource: candidate.Source,
			ProviderRank:    candidate.Rank,
			TokenCount:      candidate.Chunk.TokenCount,
		})
	}
	return result
}

func selectionsToEvaluation(selections []compiler.Selection) []evaluation.Candidate {
	result := make([]evaluation.Candidate, 0, len(selections))
	for _, selected := range selections {
		result = append(result, evaluation.Candidate{
			Path:            selected.Path,
			StartLine:       selected.StartLine,
			EndLine:         selected.EndLine,
			Text:            selected.Content,
			RetrievalSource: selected.RetrievalSource,
			ProviderRank:    selected.RetrievalRank,
			TokenCount:      selected.TokenCount,
		})
	}
	return result
}

func durationMS(value time.Duration) float64 {
	return float64(value) / float64(time.Millisecond)
}
