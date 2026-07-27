package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/Lokee86/grimoire/internal/agentruntime"
	"github.com/Lokee86/grimoire/internal/assembly"
	"github.com/Lokee86/grimoire/internal/compiler"
	"github.com/Lokee86/grimoire/internal/embedding"
	"github.com/Lokee86/grimoire/internal/evaluation"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/selection"
	"github.com/Lokee86/grimoire/internal/structure"
)

var evaluationModes = []string{"lexical"}
var allowedEvaluationModes = []string{"lexical"}

func runEval(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 || args[0] != "retrieval" {
		return errors.New("expected evaluation command: retrieval")
	}
	flags := flag.NewFlagSet("eval retrieval", flag.ContinueOnError)
	flags.SetOutput(stderr)
	casesPath := flags.String("cases", "", "judged retrieval corpus JSON")
	root := flags.String("root", ".", "repository root")
	state := flags.String("state", "", "prepared index repository path")
	modesValue := flags.String("modes", strings.Join(evaluationModes, ","), "comma-separated modes: lexical")
	variant := flags.String("variant", "standalone", "evaluation variant label")
	budgetOverride := flags.Int("budget", 0, "override every case token budget")
	adaptive := flags.Bool("adaptive", false, "use automatic query-shape budgets and evidence-coverage assembly")
	limit := flags.Int("candidate-limit", 200, "maximum ranked candidates")
	probeLimit := flags.Int("probe-limit", 800, "broader diagnostic ranking probe")
	selectionDefaults := selection.DefaultConfig()
	selectionFilePenalty := flags.Int("selection-file-penalty", selectionDefaults.FileRepeatPenalty, "curation penalty for each previously selected chunk from the same file")
	selectionSubsystemPenalty := flags.Int("selection-subsystem-penalty", selectionDefaults.SubsystemRepeatPenalty, "curation penalty for each previously selected chunk from the same subsystem")
	selectionAdjacentPrimaries := flags.Int("selection-adjacent-primaries", selectionDefaults.AdjacentPrimaryLimit, "number of diversified primaries whose immediate prepared neighbors are promoted")
	assemblyStrategy := flags.String("assembly-strategy", "coverage", "adaptive assembly strategy: legacy or coverage")
	assemblyFacetDepth := flags.Int("assembly-facet-depth", assembly.DefaultConfig().FacetDepth, "candidate depth reserved for each query facet")
	compilerFacetProtection := flags.Bool("compiler-facet-protection", compiler.DefaultConfig().ProtectFacets, "protect source candidates for each query facet during final token fitting")
	compilerFacetFileDepth := flags.Int("compiler-facet-file-depth", compiler.DefaultConfig().FacetFileDepth, "distinct mechanism source files protected per eligible query facet")
	compilerCompanionDepth := flags.Int("compiler-companion-depth", compiler.DefaultConfig().CompanionDepth, "additional protected chunks per selected facet source file")
	compilerRequiredLinkProtection := flags.Bool("compiler-required-link-protection", compiler.DefaultConfig().ProtectRequiredLinks, "protect complete required source-link groups during final token fitting")
	lexicalDeclarationAliasBonus := flags.Float64("lexical-declaration-alias-bonus", retrieve.DefaultConfig().DeclarationAliasBonus, "score for one repository-derived high-similarity declaration alias per absent query term")
	endpoint := flags.String("endpoint", embedding.DefaultEndpoint, "OpenAI-compatible embeddings endpoint")
	structuralProvidersValue := flags.String("structural-providers", "none", "structural providers: none, lexicon, or lexicon,arcana")
	lexiconFacts := flags.String("lexicon-facts", "", "explicit directory containing exported Lexicon JSONL libraries")
	lexiconState := flags.String("lexicon-state", "", "Lexicon state directory; defaults to <root>/.lexicon")
	lexiconCommand := flags.String("lexicon-command", "", "Lexicon executable override; discovered when omitted")
	arcanaState := flags.String("arcana-state", "", "Arcana state directory; defaults to <root>/.arcana")
	arcanaCommand := flags.String("arcana-command", "", "Arcana executable override; discovered when omitted")
	structuralScopeValue := flags.String("structural-scope", "lexical", "structural discovery scope: lexical or global")
	arcanaSemanticValue := flags.String("arcana-semantic", "auto", "Arcana semantic seed expansion for global structural scope: auto, on, or off")
	structureTimeout := flags.Duration("structure-timeout", 30*time.Second, "per-case structural-provider timeout")
	timeout := flags.Duration("timeout", 10*time.Second, "per-case retrieval timeout")
	outputDir := flags.String("output-dir", "evaluation/results", "result directory")
	outputPrefix := flags.String("output-prefix", "", "result filename prefix")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if strings.TrimSpace(*casesPath) == "" || *limit <= 0 || *probeLimit <= 0 || *timeout <= 0 || *structureTimeout <= 0 ||
		*selectionFilePenalty < 0 || *selectionSubsystemPenalty < 0 || *selectionAdjacentPrimaries < 0 || *assemblyFacetDepth < 0 ||
		*compilerFacetFileDepth < 0 || *compilerCompanionDepth < 0 || *lexicalDeclarationAliasBonus < 0 {
		return errors.New("--cases, positive limits and timeouts, and non-negative ranking, selection, and assembly calibration values are required")
	}
	if *adaptive && *budgetOverride > 0 {
		return errors.New("--adaptive cannot be combined with a fixed --budget override")
	}
	lexicalConfig := retrieve.Config{DeclarationAliasBonus: *lexicalDeclarationAliasBonus}
	compilerConfig := compiler.Config{
		ProtectFacets:        *compilerFacetProtection,
		FacetFileDepth:       *compilerFacetFileDepth,
		CompanionDepth:       *compilerCompanionDepth,
		ProtectRequiredLinks: *compilerRequiredLinkProtection,
	}
	assemblyConfig := assembly.DefaultConfig()
	switch strings.ToLower(strings.TrimSpace(*assemblyStrategy)) {
	case "coverage":
		assemblyConfig.CoverageAware = true
		assemblyConfig.FacetDepth = *assemblyFacetDepth
	case "legacy":
		assemblyConfig = assembly.LegacyConfig()
	default:
		return fmt.Errorf("unsupported --assembly-strategy %q", *assemblyStrategy)
	}
	modes, err := parseEvaluationModes(*modesValue)
	if err != nil {
		return err
	}
	structuralProviders, structureEnabled, arcanaEnabled, err := parseStructuralProviders(*structuralProvidersValue)
	if err != nil {
		return err
	}
	structuralScope, err := parseStructuralScopeMode(*structuralScopeValue)
	if err != nil {
		return err
	}
	arcanaSemantic, err := parseArcanaSemanticMode(*arcanaSemanticValue)
	if err != nil {
		return err
	}
	corpus, err := evaluation.LoadCorpus(*casesPath)
	if err != nil {
		return err
	}
	absoluteRoot, err := filepath.Abs(*root)
	if err != nil {
		return fmt.Errorf("resolve evaluation root: %w", err)
	}
	statePath, err := resolveState(absoluteRoot, *state)
	if err != nil {
		return err
	}
	snapshot, err := index.Load(statePath)
	if err != nil {
		return fmt.Errorf("load prepared index: %w", err)
	}
	resolvedLexiconCommand := agentruntime.ResolveProviderCommand(absoluteRoot, *lexiconCommand, "lexicon")
	resolvedArcanaCommand := agentruntime.ResolveProviderCommand(absoluteRoot, *arcanaCommand, "arcana")
	actualProbeLimit := min(*probeLimit, len(snapshot.AllChunks()))
	if actualProbeLimit < *limit {
		actualProbeLimit = min(*limit, len(snapshot.AllChunks()))
	}

	report := evaluation.Report{
		Version:             evaluation.FormatVersion,
		GeneratedAt:         time.Now(),
		Repository:          corpus.Repository,
		SourceURL:           corpus.SourceURL,
		Revision:            corpus.Revision,
		Scope:               corpus.Scope,
		JudgedAt:            corpus.JudgedAt,
		Root:                absoluteRoot,
		CasesFile:           *casesPath,
		State:               statePath,
		Variant:             *variant,
		Modes:               modes,
		StructuralProviders: structuralProviders,
	}
	for _, entry := range corpus.Cases {
		expectationErr := validateEvaluationCase(absoluteRoot, entry)
		budget := entry.Budget
		if *budgetOverride > 0 {
			budget = *budgetOverride
		}
		for _, mode := range modes {
			run := evaluation.CaseRun{
				CaseID:   entry.ID,
				Query:    entry.Query,
				Category: entry.Category,
				Mode:     mode,
				Variant:  *variant,
				Budget:   budget,
			}
			if expectationErr != nil {
				run.Error = expectationErr.Error()
				evaluation.ScoreCase(entry, &run, evaluation.Stages{})
				applyExpectationError(&run)
				report.Runs = append(report.Runs, run)
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), *timeout)
			runStart := time.Now()
			executed, executeErr := evaluateContext(ctx, snapshot, evaluatedContextOptions{
				Mode:       mode,
				Query:      entry.Query,
				Budget:     budget,
				Adaptive:   *adaptive,
				Limit:      *limit,
				ProbeLimit: actualProbeLimit,
				Structural: structuralContextOptions{
					Enabled: structureEnabled, ArcanaEnabled: arcanaEnabled, ArcanaSemantic: arcanaSemantic,
					Root: absoluteRoot, GrimoireState: statePath, LexiconFacts: *lexiconFacts,
					LexiconState: *lexiconState, LexiconCommand: resolvedLexiconCommand,
					ArcanaState: *arcanaState, ArcanaCommand: resolvedArcanaCommand,
					EmbeddingEndpoint: *endpoint,
					Scoped:            structuralScope == structuralScopeLexical,
					Limit:             *limit, Timeout: *structureTimeout,
				},
				SelectionConfig: &selection.Config{
					FileRepeatPenalty:      *selectionFilePenalty,
					SubsystemRepeatPenalty: *selectionSubsystemPenalty,
					AdjacentPrimaryLimit:   *selectionAdjacentPrimaries,
				},
				AssemblyConfig: &assemblyConfig,
				CompilerConfig: &compilerConfig,
				LexicalConfig:  &lexicalConfig,
			})
			cancel()
			run.Timings = executed.Timings
			run.QueryProfile = executed.QueryProfile
			run.RetrievalPolicy = executed.RetrievalPolicy
			run.Assembly = executed.Package.Assembly
			evaluation.ScoreQueryProfile(entry, &run)
			if run.Timings.TotalMS <= 0 {
				run.Timings.TotalMS = max(durationMS(time.Since(runStart)), 0.001)
			}
			if executeErr != nil {
				run.Warnings = append([]string(nil), executed.Warnings...)
				run.Error = executeErr.Error()
				evaluation.ScoreCase(entry, &run, executed.Stages)
				applyEvaluationErrorClassification(&run)
				report.Runs = append(report.Runs, run)
				continue
			}
			run.Warnings = append([]string(nil), executed.Warnings...)
			run.Budget = executed.Package.Budget
			run.Assembly = executed.Package.Assembly
			run.RetrievalSources = append([]string(nil), executed.Package.RetrievalSources...)
			run.StructuralSources = append([]string(nil), executed.Package.StructuralSources...)
			run.StructuralState = append([]structure.ProviderState(nil), executed.Package.StructuralState...)
			run.FinalPackageTokens = executed.Package.TokenCount
			run.CandidateCount = len(executed.Stages.Merged)
			run.CuratedCount = len(executed.Stages.Curated)
			run.AssembledCount = len(executed.Stages.Assembled)
			run.FacetProtection = executed.Package.FacetProtection
			run.FacetFileDepth = executed.Package.FacetFileDepth
			run.FacetCompanionDepth = executed.Package.FacetCompanionDepth
			run.FacetsAvailable = executed.Package.FacetsAvailable
			run.FacetsProtected = executed.Package.FacetsProtected
			run.FacetsOmittedForBudget = executed.Package.FacetsOmittedForBudget
			run.RequiredLinkProtection = executed.Package.RequiredLinkProtection
			run.RequiredLinkGroupsAvailable = executed.Package.RequiredLinkGroupsAvailable
			run.RequiredLinkGroupsProtected = executed.Package.RequiredLinkGroupsProtected
			run.RequiredLinkGroupsOmitted = executed.Package.RequiredLinkGroupsOmitted
			run.OmittedForBudget = executed.Package.OmittedForBudget
			run.OmittedStructuralForBudget = executed.Package.OmittedStructuralForBudget
			run.Selections = packageSelections(entry, executed.Package.Selections)
			run.StructuralSelections = packageStructuralSelections(executed.Package.StructuralEvidence)
			run.SelectedPaths = selectedPaths(run.Selections)
			evaluation.ScoreCase(entry, &run, executed.Stages)
			report.Runs = append(report.Runs, run)
		}
	}
	evaluation.BuildAggregates(&report)

	prefix := strings.TrimSpace(*outputPrefix)
	if prefix == "" {
		prefix = defaultEvaluationPrefix(corpus.Repository, *variant, report.GeneratedAt)
	}
	resolvedOutputDir := *outputDir
	if !filepath.IsAbs(resolvedOutputDir) {
		resolvedOutputDir = filepath.Join(absoluteRoot, resolvedOutputDir)
	}
	jsonPath := filepath.Join(resolvedOutputDir, prefix+".json")
	markdownPath := filepath.Join(resolvedOutputDir, prefix+".md")
	if err := evaluation.Write(report, jsonPath, markdownPath); err != nil {
		return err
	}
	if err := writeEvaluationSummary(stdout, report, jsonPath, markdownPath); err != nil {
		return err
	}
	return nil
}
