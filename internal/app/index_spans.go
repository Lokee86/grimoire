package app

import (
	"context"

	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/lexiconfacts"
)

func resolveIndexSourceSpans(
	root, grimoireState, explicitFacts, lexiconState, lexiconCommand string,
) ([]index.SourceSpan, string, error) {
	directory, snapshot, err := lexiconfacts.ResolveExport(context.Background(), lexiconfacts.ExportOptions{
		Root: root, GrimoireState: grimoireState,
		ExplicitDirectory: explicitFacts, LexiconState: lexiconState,
		Command: lexiconCommand,
	})
	if err != nil || directory == "" {
		return nil, snapshot, err
	}
	corpus, err := lexiconfacts.Load(directory)
	if err != nil {
		return nil, snapshot, err
	}
	return corpus.SourceSpans(), snapshot, nil
}
