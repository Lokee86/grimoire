package app

import (
	"context"
	"fmt"
	"time"

	"github.com/Lokee86/grimoire/internal/arcanagraph"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/lexiconfacts"
	"github.com/Lokee86/grimoire/internal/structure"
)

type structuralContextOptions struct {
	Enabled        bool
	ArcanaEnabled  bool
	Root           string
	GrimoireState  string
	LexiconFacts   string
	LexiconState   string
	LexiconCommand string
	ArcanaState    string
	ArcanaCommand  string
	Limit          int
	Timeout        time.Duration
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

	if options.ArcanaEnabled && len(result.Lexicon.Seeds) > 0 {
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
			result.Arcana, arcanaErr = (arcanagraph.Client{Command: options.ArcanaCommand}).Search(
				ctx, arcanaSnapshot, result.Lexicon.Seeds,
			)
			if arcanaErr != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Arcana structural evidence unavailable: %v", arcanaErr))
				result.Arcana = nil
			}
		}
		result.ArcanaTime = time.Since(arcanaStarted)
	}
	result.Combined = interleaveStructuralEvidence(result.Lexicon.Evidence, result.Arcana)
	result.TotalTime = time.Since(started)
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
