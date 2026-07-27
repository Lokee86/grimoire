package app

import (
	"context"
	"fmt"
	"time"

	"github.com/Lokee86/grimoire/internal/assembly"
	"github.com/Lokee86/grimoire/internal/compiler"
	"github.com/Lokee86/grimoire/internal/evaluation"
	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/graphrank"
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
	Structural      structuralContextOptions
	SelectionConfig *selection.Config
	AssemblyConfig  *assembly.Config
	CompilerConfig  *compiler.Config
	LexicalConfig   *retrieve.Config
}

func evaluateContext(
	ctx context.Context,
	snapshot index.Snapshot,
	options evaluatedContextOptions,
) (evaluatedContext, error) {
	var result evaluatedContext
	totalStart := time.Now()
	intents := activeRetrievalIntents(options.Query)
	lexicalConfig := retrieve.DefaultConfig()
	if options.LexicalConfig != nil {
		lexicalConfig = *options.LexicalConfig
	}
	searchStart := time.Now()
	discovery := discoverLexically(snapshot, intents, options.Limit, lexicalConfig)
	base := discovery.Candidates
	result.Timings.LexicalSearchMS = durationMS(time.Since(searchStart))
	probeLimit := options.ProbeLimit
	if probeLimit <= 0 {
		probeLimit = options.Limit
	}
	probeStart := time.Now()
	broad := discoverLexically(snapshot, intents, probeLimit, lexicalConfig).Candidates
	result.Timings.DiagnosticProbeMS = durationMS(time.Since(probeStart))

	structuralIntent := structuralRetrievalIntent(options.Query, intents)
	structuralOptions := options.Structural
	if structuralOptions.Scoped {
		structuralOptions.Scopes = discovery.Scopes
	}
	structural := collectStructuralContext(ctx, snapshot, structuralIntent.Query, structuralOptions)
	structural = annotateStructuralIntent(structural, structuralIntent)
	result.Warnings = append(result.Warnings, structural.Warnings...)
	result.Timings.LexiconSearchMS = durationMS(structural.LexiconTime)
	result.Timings.ArcanaSearchMS = durationMS(structural.ArcanaTime)
	result.Timings.StructuralProviderMS = durationMS(structural.TotalTime)

	exactStart := time.Now()
	exact := intentExactCandidates(snapshot, intents, min(options.Limit, maxExactCandidates))
	result.Timings.ExactRecoveryMS = durationMS(time.Since(exactStart))

	mergeStart := time.Now()
	var retrieved, merged []retrieve.Candidate
	if structuralOptions.Scoped {
		retrieved = mergeScopedContextProviders(options.Limit, nil, base, structural.Lexicon.Candidates, structural.ArcanaCandidates)
		merged = mergeScopedContextProviders(options.Limit, exact, base, structural.Lexicon.Candidates, structural.ArcanaCandidates)
	} else {
		retrieved = mergeContextProviders(options.Limit, nil, base, structural.Lexicon.Candidates, structural.ArcanaCandidates)
		merged = mergeContextProviders(options.Limit, exact, base, structural.Lexicon.Candidates, structural.ArcanaCandidates)
	}
	if !structuralOptions.Scoped {
		retrieved = graphrank.Rerank(retrieved, structuralIntent.Intent)
		merged = graphrank.Rerank(merged, structuralIntent.Intent)
	}
	result.Timings.CandidateMergeMS += durationMS(time.Since(mergeStart))
	profileBudget := options.Budget
	if options.Adaptive {
		profileBudget = 0
	}
	policyEvidence := structural.Combined
	if structuralOptions.Scoped {
		policyEvidence = nil
	}
	result.QueryProfile, result.RetrievalPolicy = queryshape.Analyze(queryshape.Input{
		Query: options.Query, RequestedBudget: profileBudget,
		Exact: exact, Ranked: base, Candidates: merged, Structural: policyEvidence,
	})

	curationStart := time.Now()
	var curated []retrieve.Candidate
	if options.SelectionConfig != nil {
		curated = selection.CurateWithConfig(snapshot, merged, *options.SelectionConfig)
	} else {
		curated = selection.Curate(snapshot, merged)
	}
	result.Timings.CurationMS = durationMS(time.Since(curationStart))
	assemblyInput := curated
	assembledCandidates := assemblyInput
	assembledEvidence := structural.Combined
	effectiveBudget := options.Budget
	var decision *assembly.Decision
	if options.Adaptive {
		result.RetrievalPolicy = queryshape.Activate(result.RetrievalPolicy)
		effectiveBudget = result.RetrievalPolicy.TargetTokens
		assemblyStart := time.Now()
		planConfig := assembly.DefaultConfig()
		if options.AssemblyConfig != nil {
			planConfig = *options.AssemblyConfig
		}
		planned := assembly.PlanWithConfig(result.RetrievalPolicy, assemblyInput, structural.Combined, planConfig)
		result.Timings.AssemblyMS = durationMS(time.Since(assemblyStart))
		assembledCandidates = planned.Candidates
		assembledEvidence = planned.Structural
		decision = &planned.Decision
	}

	compileStart := time.Now()
	var pkg compiler.Package
	var err error
	if decision != nil {
		compileConfig := compiler.DefaultConfig()
		if options.CompilerConfig != nil {
			compileConfig = *options.CompilerConfig
		}
		compileConfig.SourceFirstEvidence = structuralOptions.Scoped
		pkg, err = compiler.CompileAdaptiveWithEvidenceConfig(
			options.Query, effectiveBudget, snapshot.Version, snapshot.Tokenizer,
			contextCandidateSources(assembledCandidates), structural.ProviderState,
			assembledEvidence, *decision, assembledCandidates, compileConfig,
		)
	} else {
		compileConfig := compiler.LegacyConfig()
		compileConfig.SourceFirstEvidence = structuralOptions.Scoped
		pkg, err = compiler.CompileWithEvidenceConfig(
			options.Query, effectiveBudget, snapshot.Version, snapshot.Tokenizer,
			contextCandidateSources(assembledCandidates), structural.ProviderState,
			assembledEvidence, assembledCandidates, compileConfig,
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
		Retrieved:           candidatesToEvaluation(retrieved),
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

func candidateGraphSignals(candidate retrieve.Candidate) *evidence.GraphSignals {
	if candidate.Context == nil {
		return nil
	}
	return candidate.Context.Graph
}

func candidatesToEvaluation(candidates []retrieve.Candidate) []evaluation.Candidate {
	result := make([]evaluation.Candidate, 0, len(candidates))
	for _, candidate := range candidates {
		result = append(result, evaluation.Candidate{
			Path:              candidate.Chunk.Path,
			StartLine:         candidate.Chunk.StartLine,
			EndLine:           candidate.Chunk.EndLine,
			Text:              candidate.Chunk.Text,
			RetrievalSource:   candidate.Source,
			ProviderRank:      candidate.Rank,
			Score:             candidate.Score,
			ScoreDetails:      scoreDetailsToEvaluation(candidate.ScoreDetails),
			GraphScoreDetails: scoreDetailsToEvaluation(candidate.GraphScoreDetails),
			Reasons:           append([]string(nil), candidate.Reasons...),
			Graph:             evidence.CloneGraphSignals(candidateGraphSignals(candidate)),
			TokenCount:        candidate.Chunk.TokenCount,
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
