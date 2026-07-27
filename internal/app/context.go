package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Lokee86/grimoire/internal/agentruntime"
	"github.com/Lokee86/grimoire/internal/assembly"
	"github.com/Lokee86/grimoire/internal/compiler"
	"github.com/Lokee86/grimoire/internal/embedding"
	"github.com/Lokee86/grimoire/internal/graphrank"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/queryshape"
	"github.com/Lokee86/grimoire/internal/selection"
)

func runContext(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("context", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "repository root")
	state := flags.String("state", "", "prepared index repository path")
	query := flags.String("query", "", "task or retrieval query; optional when --diff is set")
	diffSpec := flags.String("diff", "", "Git diff scope: working-tree, staged, unstaged, or one revision/range")
	diffTimeout := flags.Duration("diff-timeout", 10*time.Second, "Git diff collection timeout")
	budget := flags.Int("budget", 0, "maximum o200k_base tokens in the emitted package; zero selects automatically")
	limit := flags.Int("candidate-limit", 200, "maximum ranked candidates")
	endpoint := flags.String("endpoint", embedding.DefaultEndpoint, "OpenAI-compatible embeddings endpoint")
	structureEnabled := flags.Bool("structure", true, "include available structural evidence")
	structuralProviders := flags.String("structural-providers", "lexicon,arcana", "structural evidence providers: none, lexicon, arcana, or lexicon,arcana")
	lexiconFacts := flags.String("lexicon-facts", "", "explicit directory containing exported Lexicon JSONL libraries")
	lexiconState := flags.String("lexicon-state", "", "Lexicon state directory; defaults to <root>/.lexicon")
	lexiconCommand := flags.String("lexicon-command", "", "Lexicon executable override; discovered when omitted")
	arcanaState := flags.String("arcana-state", "", "Arcana state directory; defaults to <root>/.arcana")
	arcanaCommand := flags.String("arcana-command", "", "Arcana executable override; discovered when omitted")
	arcanaSemanticValue := flags.String("arcana-semantic", "auto", "Arcana semantic seed expansion: auto, on, or off")
	structureTimeout := flags.Duration("structure-timeout", 30*time.Second, "complete structural-provider timeout")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*query) == "" && strings.TrimSpace(*diffSpec) == "" {
		return errors.New("--query is required unless --diff is set")
	}
	if *budget < 0 || *limit <= 0 || *structureTimeout <= 0 || *diffTimeout <= 0 {
		return errors.New("non-negative --budget and positive --candidate-limit, --structure-timeout, and --diff-timeout are required")
	}
	emitLexicon, arcanaEnabled, err := parseContextStructuralProviders(*structuralProviders)
	if err != nil {
		return err
	}
	arcanaSemantic, err := parseArcanaSemanticMode(*arcanaSemanticValue)
	if err != nil {
		return err
	}
	if !*structureEnabled {
		emitLexicon = false
		arcanaEnabled = false
		arcanaSemantic = arcanaSemanticOff
	}
	statePath, err := resolveState(*root, *state)
	if err != nil {
		return err
	}
	snapshot, err := index.Load(statePath)
	if err != nil {
		return fmt.Errorf("load prepared index: %w", err)
	}
	resolvedLexiconCommand := agentruntime.ResolveProviderCommand(*root, *lexiconCommand, "lexicon")
	resolvedArcanaCommand := agentruntime.ResolveProviderCommand(*root, *arcanaCommand, "arcana")

	diffContext, cancelDiff := context.WithTimeout(context.Background(), *diffTimeout)
	diffResult, err := prepareContextDiff(diffContext, snapshot, *root, *diffSpec, *query, *limit)
	cancelDiff()
	if err != nil {
		return err
	}
	packageQuery := diffResult.PackageQuery
	retrievalQuery := diffResult.RetrievalQuery

	intents := activeRetrievalIntents(retrievalQuery)
	baseCandidates := intentLexicalCandidates(snapshot, intents, *limit)

	structuralIntent := structuralRetrievalIntent(retrievalQuery, intents)
	structural := collectStructuralContext(context.Background(), snapshot, structuralIntent.Query, structuralContextOptions{
		Enabled: emitLexicon || arcanaEnabled, ArcanaEnabled: arcanaEnabled, ArcanaSemantic: arcanaSemantic, EmitLexicon: emitLexicon,
		Root: *root, GrimoireState: statePath, LexiconFacts: *lexiconFacts,
		LexiconState: *lexiconState, LexiconCommand: resolvedLexiconCommand,
		ArcanaState: *arcanaState, ArcanaCommand: resolvedArcanaCommand,
		EmbeddingEndpoint: *endpoint,
		Limit:             *limit, Timeout: *structureTimeout,
	})
	structural = annotateStructuralIntent(structural, structuralIntent)
	for _, warning := range structural.Warnings {
		_, _ = fmt.Fprintf(stderr, "warning: %s\n", warning)
	}

	exact := intentExactCandidates(snapshot, intents, min(*limit, maxExactCandidates))
	lexiconCandidates := structural.Lexicon.Candidates
	if !emitLexicon {
		lexiconCandidates = nil
	}
	merged := mergeContextProviders(*limit, exact, baseCandidates, lexiconCandidates, structural.ArcanaCandidates)
	merged = graphrank.Rerank(merged, structuralIntent.Intent)
	merged = mergeContextProvidersWithPriority(*limit, diffResult.Candidates, nil, merged, nil, nil)
	evidence := interleaveStructuralEvidence(diffResult.Evidence, structural.Combined)
	_, policy := queryshape.Analyze(queryshape.Input{
		Query: packageQuery, RequestedBudget: *budget,
		Exact: exact, Ranked: baseCandidates, Candidates: merged, Structural: evidence,
	})
	effectiveBudget := *budget
	automatic := effectiveBudget == 0
	if automatic {
		policy = queryshape.Activate(policy)
		effectiveBudget = policy.TargetTokens
	}
	candidates := selection.Curate(snapshot, merged)

	var result compiler.Package
	if automatic {
		planned := assembly.Plan(policy, candidates, evidence)
		candidates = planned.Candidates
		evidence = planned.Structural
		result, err = compiler.CompileAdaptiveWithEvidence(
			packageQuery, effectiveBudget, snapshot.Version, snapshot.Tokenizer,
			contextCandidateSources(candidates), structural.ProviderState, evidence,
			planned.Decision, candidates,
		)
	} else {
		result, err = compiler.CompileWithEvidence(
			packageQuery, effectiveBudget, snapshot.Version, snapshot.Tokenizer,
			contextCandidateSources(candidates), structural.ProviderState, evidence, candidates,
		)
	}
	if err != nil {
		return err
	}
	data, err := compiler.Marshal(result)
	if err != nil {
		return err
	}
	_, err = stdout.Write(data)
	return err
}

func parseContextStructuralProviders(value string) (emitLexicon, arcanaEnabled bool, err error) {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" || value == "none" {
		return false, false, nil
	}
	seen := make(map[string]struct{})
	for _, provider := range strings.Split(value, ",") {
		provider = strings.TrimSpace(provider)
		if provider == "" {
			continue
		}
		if _, exists := seen[provider]; exists {
			continue
		}
		seen[provider] = struct{}{}
		switch provider {
		case "lexicon":
			emitLexicon = true
		case "arcana":
			arcanaEnabled = true
		default:
			return false, false, fmt.Errorf("unsupported structural provider %q", provider)
		}
	}
	return emitLexicon, arcanaEnabled, nil
}
