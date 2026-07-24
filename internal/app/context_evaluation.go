package app

import (
	"context"
	"fmt"
	"time"

	"github.com/Lokee86/grimoire/internal/assembly"
	"github.com/Lokee86/grimoire/internal/compiler"
	"github.com/Lokee86/grimoire/internal/embedding"
	"github.com/Lokee86/grimoire/internal/evaluation"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/queryshape"
	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/selection"
	"github.com/Lokee86/grimoire/internal/structure"
)

type evaluatedContext struct {
	Package         compiler.Package
	Stages          evaluation.Stages
	Timings         evaluation.Timings
	QueryProfile    queryshape.Profile
	RetrievalPolicy queryshape.RetrievalPolicy
	Warnings        []string
}

type evaluatedContextOptions struct {
	Mode            string
	Query           string
	Budget          int
	Adaptive        bool
	Limit           int
	ProbeLimit      int
	StatePath       string
	Endpoint        string
	EnginePath      string
	Structural      structuralContextOptions
	QueryOptions    embedding.QueryOptions
	SelectionConfig *selection.Config
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
	intents := activeRetrievalIntents(options.Query)

	switch options.Mode {
	case "lexical":
		searchStart := time.Now()
		base = intentLexicalCandidates(snapshot, intents, options.Limit)
		result.Timings.LexicalSearchMS = durationMS(time.Since(searchStart))
		probeLimit := options.ProbeLimit
		if probeLimit <= 0 {
			probeLimit = options.Limit
		}
		probeStart := time.Now()
		broad = intentLexicalCandidates(snapshot, intents, probeLimit)
		result.Timings.DiagnosticProbeMS = durationMS(time.Since(probeStart))
	case "hybrid":
		searchStart := time.Now()
		lexical := intentLexicalCandidates(snapshot, intents, options.Limit)
		result.Timings.LexicalSearchMS = durationMS(time.Since(searchStart))
		probeLimit := options.ProbeLimit
		if probeLimit <= 0 {
			probeLimit = options.Limit
		}
		probeStart := time.Now()
		lexicalBroad := intentLexicalCandidates(snapshot, intents, probeLimit)
		result.Timings.DiagnosticProbeMS = durationMS(time.Since(probeStart))

		semantic, err := semanticCandidatesForEvaluation(
			ctx, snapshot, options.StatePath, options.Query, options.Endpoint,
			options.EnginePath, options.Limit, options.ProbeLimit, options.QueryOptions,
		)
		if err != nil {
			return result, err
		}
		mergeStart := time.Now()
		semantic.Candidates = rankCandidatesForIntent(semantic.Candidates, intents[0], false)
		semantic.BroadProbe = rankCandidatesForIntent(semantic.BroadProbe, intents[0], false)
		base = mergeRankedProviders(options.Limit, lexical, semantic.Candidates)
		broad = mergeRankedProviders(probeLimit, lexicalBroad, semantic.BroadProbe)
		result.Timings.SnapshotValidationMS = durationMS(semantic.Metrics.SnapshotValidation)
		result.Timings.EmbeddingMS = durationMS(semantic.Metrics.Embedding)
		result.Timings.VectorSearchMS = durationMS(semantic.Metrics.VectorSearch)
		result.Timings.CandidateMergeMS = durationMS(semantic.Metrics.CandidateMerge) + durationMS(time.Since(mergeStart))
		result.Timings.DiagnosticProbeMS += durationMS(semantic.Metrics.DiagnosticProbe)
	default:
		semantic, err := semanticCandidatesForEvaluation(
			ctx, snapshot, options.StatePath, options.Query, options.Endpoint,
			options.EnginePath, options.Limit, options.ProbeLimit, options.QueryOptions,
		)
		if err != nil {
			return result, err
		}
		base = rankCandidatesForIntent(semantic.Candidates, intents[0], false)
		broad = rankCandidatesForIntent(semantic.BroadProbe, intents[0], false)
		result.Timings.SnapshotValidationMS = durationMS(semantic.Metrics.SnapshotValidation)
		result.Timings.EmbeddingMS = durationMS(semantic.Metrics.Embedding)
		result.Timings.VectorSearchMS = durationMS(semantic.Metrics.VectorSearch)
		result.Timings.CandidateMergeMS = durationMS(semantic.Metrics.CandidateMerge)
		result.Timings.DiagnosticProbeMS = durationMS(semantic.Metrics.DiagnosticProbe)
	}

	structuralIntent := structuralRetrievalIntent(options.Query, intents)
	structural := collectStructuralContext(context.Background(), snapshot, structuralIntent.Query, options.Structural)
	structural = annotateStructuralIntent(structural, structuralIntent)
	result.Warnings = append(result.Warnings, structural.Warnings...)
	result.Timings.LexiconSearchMS = durationMS(structural.LexiconTime)
	result.Timings.ArcanaSearchMS = durationMS(structural.ArcanaTime)
	result.Timings.StructuralProviderMS = durationMS(structural.TotalTime)

	exactStart := time.Now()
	exact := intentExactCandidates(snapshot, intents, min(options.Limit, maxExactCandidates))
	result.Timings.ExactRecoveryMS = durationMS(time.Since(exactStart))

	mergeStart := time.Now()
	merged := mergeContextProviders(options.Limit, exact, base, structural.Lexicon.Candidates)
	result.Timings.CandidateMergeMS += durationMS(time.Since(mergeStart))
	profileBudget := options.Budget
	if options.Adaptive {
		profileBudget = 0
	}
	result.QueryProfile, result.RetrievalPolicy = queryshape.Analyze(queryshape.Input{
		Query: options.Query, RequestedBudget: profileBudget,
		Exact: exact, Ranked: base, Candidates: merged, Structural: structural.Combined,
	})

	curationStart := time.Now()
	var curated []retrieve.Candidate
	if options.SelectionConfig != nil {
		curated = selection.CurateWithConfig(snapshot, merged, *options.SelectionConfig)
	} else {
		curated = selection.Curate(snapshot, merged)
	}
	result.Timings.CurationMS = durationMS(time.Since(curationStart))
	assembledCandidates := curated
	assembledEvidence := structural.Combined
	effectiveBudget := options.Budget
	var decision *assembly.Decision
	if options.Adaptive {
		result.RetrievalPolicy = queryshape.Activate(result.RetrievalPolicy)
		effectiveBudget = result.RetrievalPolicy.TargetTokens
		assemblyStart := time.Now()
		planned := assembly.Plan(result.RetrievalPolicy, curated, structural.Combined)
		result.Timings.AssemblyMS = durationMS(time.Since(assemblyStart))
		assembledCandidates = planned.Candidates
		assembledEvidence = planned.Structural
		decision = &planned.Decision
	}

	compileStart := time.Now()
	var pkg compiler.Package
	var err error
	if decision != nil {
		pkg, err = compiler.CompileAdaptiveWithEvidence(
			options.Query, effectiveBudget, snapshot.Version, snapshot.Tokenizer,
			contextCandidateSources(assembledCandidates), structural.ProviderState,
			assembledEvidence, *decision, assembledCandidates,
		)
	} else {
		pkg, err = compiler.CompileWithEvidence(
			options.Query, effectiveBudget, snapshot.Version, snapshot.Tokenizer,
			contextCandidateSources(assembledCandidates), structural.ProviderState,
			assembledEvidence, assembledCandidates,
		)
	}
	result.Timings.PackageCompilationMS = durationMS(time.Since(compileStart))
	result.Timings.SelectionCompilationMS = result.Timings.CurationMS + result.Timings.AssemblyMS + result.Timings.PackageCompilationMS
	result.Timings.TotalMS = durationMS(time.Since(totalStart)) - result.Timings.DiagnosticProbeMS
	if result.Timings.TotalMS < 0 {
		result.Timings.TotalMS = 0
	}
	if err != nil {
		return result, fmt.Errorf("compile context package: %w", err)
	}
	result.Package = pkg
	result.Stages = evaluation.Stages{
		Indexed:             chunksToEvaluation(snapshot.AllChunks()),
		BroadProbe:          candidatesToEvaluation(broad),
		Retrieved:           candidatesToEvaluation(mergeContextProviders(options.Limit, nil, base, structural.Lexicon.Candidates)),
		Exact:               candidatesToEvaluation(exact),
		Merged:              candidatesToEvaluation(merged),
		Curated:             candidatesToEvaluation(curated),
		Assembled:           candidatesToEvaluation(assembledCandidates),
		Included:            selectionsToEvaluation(pkg.Selections),
		StructuralProduced:  append(append([]structure.Evidence(nil), structural.Lexicon.Evidence...), structural.Arcana...),
		StructuralComposed:  append([]structure.Evidence(nil), structural.Combined...),
		StructuralAssembled: append([]structure.Evidence(nil), assembledEvidence...),
		StructuralIncluded:  append([]structure.Evidence(nil), pkg.StructuralEvidence...),
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
			Score:           candidate.Score,
			ScoreDetails:    scoreDetailsToEvaluation(candidate.ScoreDetails),
			Reasons:         append([]string(nil), candidate.Reasons...),
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
			Score:           selected.Score,
			Reasons:         append([]string(nil), selected.Reasons...),
			TokenCount:      selected.TokenCount,
		})
	}
	return result
}

func scoreDetailsToEvaluation(details []retrieve.ScoreDetail) []evaluation.ScoreDetail {
	result := make([]evaluation.ScoreDetail, 0, len(details))
	for _, detail := range details {
		result = append(result, evaluation.ScoreDetail{Name: detail.Name, Value: detail.Value})
	}
	return result
}

func durationMS(value time.Duration) float64 {
	return float64(value) / float64(time.Millisecond)
}
