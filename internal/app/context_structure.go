package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/Lokee86/grimoire/internal/arcanagraph"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/lexiconfacts"
	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/structure"
)

type arcanaSemanticMode string

const (
	arcanaSemanticAuto arcanaSemanticMode = "auto"
	arcanaSemanticOn   arcanaSemanticMode = "on"
	arcanaSemanticOff  arcanaSemanticMode = "off"
)

type structuralContextOptions struct {
	Enabled           bool
	ArcanaEnabled     bool
	ArcanaSemantic    arcanaSemanticMode
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
	Lexicon          lexiconfacts.Result
	Arcana           []structure.Evidence
	ArcanaCandidates []retrieve.Candidate
	Combined         []structure.Evidence
	ProviderState    []structure.ProviderState
	Warnings         []string
	LexiconTime      time.Duration
	ArcanaTime       time.Duration
	TotalTime        time.Duration
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
			var semanticSeeds []arcanagraph.SemanticSeed
			if shouldUseArcanaSemantic(options.ArcanaSemantic, query, result.Lexicon.Seeds) {
				semanticSeeds, arcanaErr = client.RankedSemanticSeeds(
					ctx,
					filepath.Dir(filepath.Dir(arcanaSnapshot)),
					arcanaSnapshotID,
					options.EmbeddingEndpoint,
					query,
					arcanagraph.SemanticCandidateLimit(6),
				)
				if arcanaErr != nil {
					result.Warnings = append(result.Warnings, fmt.Sprintf("Arcana semantic graph retrieval unavailable: %v", arcanaErr))
				}
			}
			rankedSeeds, rerankErr := client.RerankSeeds(
				ctx, arcanaSnapshot, query, result.Lexicon.Seeds, semanticSeeds, 6,
			)
			if rerankErr != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Arcana graph-proximity reranking unavailable: %v", rerankErr))
			}
			seeds := make([]structure.Node, 0, len(rankedSeeds))
			for _, ranked := range rankedSeeds {
				seeds = append(seeds, ranked.Node)
			}
			if len(seeds) > 0 {
				result.Arcana, arcanaErr = client.Search(ctx, arcanaSnapshot, seeds)
				if arcanaErr != nil {
					result.Warnings = append(result.Warnings, fmt.Sprintf("Arcana structural evidence unavailable: %v", arcanaErr))
					result.Arcana = nil
				} else {
					result.ArcanaCandidates = arcanagraph.SourceCandidates(snapshot, result.Arcana, options.Limit)
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

func parseArcanaSemanticMode(value string) (arcanaSemanticMode, error) {
	switch arcanaSemanticMode(strings.ToLower(strings.TrimSpace(value))) {
	case "", arcanaSemanticAuto:
		return arcanaSemanticAuto, nil
	case arcanaSemanticOn:
		return arcanaSemanticOn, nil
	case arcanaSemanticOff:
		return arcanaSemanticOff, nil
	default:
		return "", fmt.Errorf("unsupported Arcana semantic mode %q; expected auto, on, or off", value)
	}
}

func shouldUseArcanaSemantic(mode arcanaSemanticMode, query string, lexicon []structure.Node) bool {
	switch mode {
	case arcanaSemanticOff:
		return false
	case arcanaSemanticOn:
		return true
	case "", arcanaSemanticAuto:
		return !queryExplicitlyNamesLexiconSeed(query, lexicon)
	default:
		return false
	}
}

func queryExplicitlyNamesLexiconSeed(query string, seeds []structure.Node) bool {
	normalizedQuery := normalizeIdentifierText(query)
	pathQuery := strings.ToLower(filepath.ToSlash(query))
	for _, seed := range seeds {
		name := normalizeIdentifierText(seed.Name)
		terms := strings.Fields(name)
		if len(terms) >= 2 && strings.Contains(normalizedQuery, name) {
			return true
		}
		if len(terms) == 1 && len(terms[0]) >= 12 && strings.Contains(normalizedQuery, terms[0]) {
			return true
		}
		path := strings.ToLower(filepath.ToSlash(strings.TrimSpace(seed.Path)))
		if path != "" && strings.Contains(pathQuery, path) {
			return true
		}
	}
	return false
}

func normalizeIdentifierText(value string) string {
	var output strings.Builder
	var previous rune
	for _, current := range value {
		if unicode.IsUpper(current) && (unicode.IsLower(previous) || unicode.IsDigit(previous)) {
			output.WriteByte(' ')
		}
		if unicode.IsLetter(current) || unicode.IsDigit(current) {
			output.WriteRune(unicode.ToLower(current))
		} else {
			output.WriteByte(' ')
		}
		previous = current
	}
	return strings.Join(strings.Fields(output.String()), " ")
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
