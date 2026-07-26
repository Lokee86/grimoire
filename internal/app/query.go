package app

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Lokee86/grimoire/internal/agentquery"
)

type queryListFlag []string

func (values *queryListFlag) String() string {
	return strings.Join(*values, ",")
}

func (values *queryListFlag) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("value must not be empty")
	}
	*values = append(*values, value)
	return nil
}

func runQuery(args []string, stdout, stderr io.Writer) error {
	positionalMode := ""
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		positionalMode = args[0]
		args = args[1:]
	}
	flags := flag.NewFlagSet("query", flag.ContinueOnError)
	flags.SetOutput(stderr)
	mode := flags.String("mode", positionalMode, "query mode: orient, search, trace, impact, or inspect")
	root := flags.String("root", ".", "repository root")
	state := flags.String("state", "", "prepared index repository path")
	query := flags.String("query", "", "literal, symbol, or behavior query")
	anchor := flags.String("anchor", "", "name, query anchor, or stable returned handle")
	target := flags.String("target", "", "optional trace target name or handle")
	limit := flags.Int("limit", 0, "maximum returned results; defaults to 8 for trace and 12 otherwise")
	depth := flags.Int("depth", 3, "maximum graph traversal depth")
	direction := flags.String("direction", "", "graph direction: incoming, outgoing, or both")
	adjacent := flags.Int("adjacent-context", 0, "source lines adjacent to an inspected declaration")
	codeOnly := flags.Bool("code-only", false, "exclude documentation from source and structural result lanes")
	detail := flags.String("detail", "", "trace detail: summary or full; defaults to summary")
	requestJSON := flags.String("request", "", "complete "+agentquery.SchemaVersion+" JSON request object")
	lexiconFacts := flags.String("lexicon-facts", "", "explicit directory containing exported Lexicon JSONL libraries")
	lexiconState := flags.String("lexicon-state", "", "Lexicon state directory; defaults to <root>/.lexicon")
	lexiconCommand := flags.String("lexicon-command", "lexicon", "Lexicon executable used for immutable snapshot export")
	arcanaState := flags.String("arcana-state", "", "Arcana state directory; defaults to <root>/.arcana")
	arcanaCommand := flags.String("arcana-command", "arcana", "Arcana executable used for graph queries")
	timeout := flags.Duration("timeout", 30*time.Second, "complete query timeout")
	var handles queryListFlag
	var relations queryListFlag
	flags.Var(&handles, "handle", "stable handle to inspect; may be repeated")
	flags.Var(&relations, "relation", "graph relation filter; may be repeated")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected query arguments: %s", strings.Join(flags.Args(), " "))
	}
	if *timeout <= 0 {
		return errors.New("positive --timeout is required")
	}

	var request agentquery.Request
	if strings.TrimSpace(*requestJSON) != "" {
		if err := json.Unmarshal([]byte(*requestJSON), &request); err != nil {
			return fmt.Errorf("decode --request: %w", err)
		}
		if positionalMode != "" {
			request.Mode = positionalMode
		}
	} else {
		request = agentquery.Request{
			Schema: agentquery.SchemaVersion, Mode: *mode, Root: *root, State: *state,
			Query: *query, Anchor: *anchor, Target: *target, Handles: handles,
			Limit: *limit, Depth: *depth, Direction: *direction,
			Relations: relations, Adjacent: *adjacent, CodeOnly: *codeOnly, Detail: *detail,
			LexiconFacts: *lexiconFacts, LexiconState: *lexiconState,
			LexiconCmd: *lexiconCommand, ArcanaState: *arcanaState, ArcanaCmd: *arcanaCommand,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	response, err := agentquery.Execute(ctx, request)
	if err != nil {
		return err
	}
	return writeJSON(stdout, response)
}
