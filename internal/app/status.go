package app

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Lokee86/grimoire/internal/agentruntime"
	"github.com/Lokee86/grimoire/internal/repostate"
)

func runStatus(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("status", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "repository root")
	refresh := flags.Bool("refresh", false, "refresh missing or stale repository state")
	force := flags.Bool("force", false, "force-refresh Lexicon, Arcana, Grimoire source, and knowledge state")
	lexiconCommand := flags.String("lexicon-command", "", "Lexicon executable override")
	arcanaCommand := flags.String("arcana-command", "", "Arcana executable override")
	if err := flags.Parse(args); err != nil {
		return err
	}
	mode := repostate.CurrentOnly
	if *refresh {
		mode = repostate.RefreshIfNeeded
	}
	if *force {
		mode = repostate.ForceRefresh
	}
	options := repostate.Options{
		Root: *root, Mode: mode,
		LexiconCommand: agentruntime.ResolveProviderCommand(*root, *lexiconCommand, "lexicon"),
		ArcanaCommand:  agentruntime.ResolveProviderCommand(*root, *arcanaCommand, "arcana"),
	}
	if mode != repostate.CurrentOnly {
		current, executableErr := os.Executable()
		if executableErr != nil {
			return fmt.Errorf("resolve current Grimoire executable: %w", executableErr)
		}
		options.GrimoireCommand = current
	}
	status, err := repostate.Ensure(context.Background(), options)
	if status.Version != 0 {
		if writeErr := writeJSON(stdout, status); writeErr != nil && err == nil {
			return writeErr
		}
	}
	return err
}
