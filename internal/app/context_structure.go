package app

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/Lokee86/grimoire/internal/arcanagraph"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/lexiconfacts"
	"github.com/Lokee86/grimoire/internal/structure"
)

type structuralContextOptions struct {
	Enabled           bool
	ArcanaEnabled     bool
	EmitLexicon       bool
	Root              string
	GrimoireState     string
	LexiconFacts      string
	LexiconState      string
	LexiconCommand    string
	ArcanaState       string
	ArcanaCommand     string
	EmbeddingEndpoint string
	Limit             int
	Timeout           time.Duration
}

type structuralContextResult struct {
	Lexicon       lexiconfacts.Result
	Arcana        []structure.Evidence
	Combined      []structure.Evidence
	ProviderState []structure.ProviderState
	Warnings      []string
	LexiconTime   time.Duration
	ArcanaTime    time.Duration
	TotalTime     time.Duration
}

func collectStructuralContext(
	ctx context.Context,
	snapshot index.Snapshot,
	query string,
	options structuralContextOptions,
) structuralContextResult {
	var result structuralContextResult
	if !options.Enabled {
		return result
	}
	started := time.Now()
	if options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	lexiconStarted := time.Now()
	exportDirectory, lexiconSnapshot, err := lexiconfacts.ResolveExport(ctx, lexiconfacts.ExportOptions{
		Root: options.Root, GrimoireState: options.GrimoireState,
		ExplicitDirectory: options.LexiconFacts, LexiconState: options.LexiconState,
		Command: options.LexiconCommand,
	})
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Lexicon structural evidence unavailable: %v", err))
	} else if exportDirectory != "" {
		if lexiconSnapshot != "" {
			result.ProviderState = append(result.ProviderState, structure.ProviderState{
				Provider: "lexicon", Snapshot: lexiconSnapshot,
			})
		}
		result.Lexicon, err = lexiconfacts.SearchDetailed(snapshot, query, exportDirectory, options.Limit)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Lexicon structural evidence unavailable: %v", err))
			result.Lexicon = lexiconfacts.Result{}
		}
	}
	result.LexiconTime = time.Since(lexiconStarted)

	if options.ArcanaEnabled {
		arcanaStarted := time.Now()
		arcanaSnapshot, arcanaSnapshotID, arcanaErr := arcanagraph.ResolveSnapshot(ctx, arcanagraph.StateOptions{
			Root: options.Root, State: options.ArcanaState, LexiconState: options.LexiconState,
			ExpectedLexiconSnapshot: lexiconSnapshot, Command: options.ArcanaCommand,
		})
		if arcanaErr != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Arcana structural evidence unavailable: %v", arcanaErr))
		} else if arcanaSnapshot != "" {
			if arcanaSnapshotID != "" {
				result.ProviderState = append(result.ProviderState, structure.ProviderState{
					Provider: "arcana", Snapshot: arcanaSnapshotID,
				})
			}
			client := arcanagraph.Client{Command: options.ArcanaCommand}
			semanticSeeds, semanticErr := client.SemanticSeeds(
				ctx,
				filepath.Dir(filepath.Dir(arcanaSnapshot)),
				options.EmbeddingEndpoint,
				query,
				min(options.Limit, 6),
			)
			if semanticErr != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Arcana semantic graph retrieval unavailable: %v", semanticErr))
			}
			seeds := mergeArcanaSeeds(result.Lexicon.Seeds, semanticSeeds, 6)
			if len(seeds) > 0 {
				result.Arcana, arcanaErr = client.Search(ctx, arcanaSnapshot, seeds)
				if arcanaErr != nil {
					result.Warnings = append(result.Warnings, fmt.Sprintf("Arcana structural evidence unavailable: %v", arcanaErr))
					result.Arcana = nil
				}
			}
		}
		result.ArcanaTime = time.Since(arcanaStarted)
	}
	if options.EmitLexicon {
		result.Combined = interleaveStructuralEvidence(result.Lexicon.Evidence, result.Arcana)
	} else {
		result.Combined = append([]structure.Evidence(nil), result.Arcana...)
	}
	result.TotalTime = time.Since(started)
	return result
}

func mergeArcanaSeeds(lexicon, semantic []structure.Node, limit int) []structure.Node {
	if limit <= 0 {
		return nil
	}
	groups := [][]structure.Node{semantic, lexicon}
	result := make([]structure.Node, 0, min(limit, len(lexicon)+len(semantic)))
	seen := make(map[string]struct{}, cap(result))
	for index := 0; len(result) < limit; index++ {
		added := false
		for _, group := range groups {
			if index >= len(group) {
				continue
			}
			seed := group[index]
			if seed.Name == "" {
				continue
			}
			key := seed.Name + "\x00" + seed.Path
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			result = append(result, seed)
			added = true
			if len(result) == limit {
				break
			}
		}
		if !added && index >= len(semantic) && index >= len(lexicon) {
			break
		}
	}
	return result
}

// interleaveStructuralEvidence preserves provider-local rank while ensuring one
// provider cannot consume the complete structural portion of a tight package.
func interleaveStructuralEvidence(groups ...[]structure.Evidence) []structure.Evidence {
	total := 0
	for _, group := range groups {
		total += len(group)
	}
	result := make([]structure.Evidence, 0, total)
	for index := 0; len(result) < total; index++ {
		added := false
		for _, group := range groups {
			if index >= len(group) {
				continue
			}
			result = append(result, group[index])
			added = true
		}
		if !added {
			break
		}
	}
	return result
}
