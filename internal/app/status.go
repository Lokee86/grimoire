package app

import (
	"context"
	"flag"
	"io"

	"github.com/Lokee86/grimoire/internal/repostate"
)

func runStatus(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("status", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "repository root")
	refresh := flags.Bool("refresh", false, "refresh missing or stale repository state")
	force := flags.Bool("force", false, "force-refresh Lexicon, Arcana, and Grimoire state")
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
	status, err := repostate.Ensure(context.Background(), repostate.Options{Root: *root, Mode: mode})
	if status.Version != 0 {
		if writeErr := writeJSON(stdout, status); writeErr != nil && err == nil {
			return writeErr
		}
	}
	return err
}
