package agentruntime

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/agentquery"
	"github.com/Lokee86/grimoire/internal/repostate"
)

func TestManualRepositorySmoke(t *testing.T) {
	root := os.Getenv("GRIMOIRE_SMOKE_ROOT")
	if root == "" {
		t.Skip("set GRIMOIRE_SMOKE_ROOT for a real repository smoke test")
	}
	command := os.Getenv("GRIMOIRE_SMOKE_COMMAND")
	if command == "" {
		t.Skip("set GRIMOIRE_SMOKE_COMMAND to the Grimoire executable")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}
	providerDir := filepath.Dir(command)
	lexiconCommand := filepath.Join(providerDir, "lexicon.exe")
	arcanaCommand := filepath.Join(providerDir, "arcana.exe")
	options := Options{GrimoireCommand: command}
	search, err := Execute(context.Background(), Request{
		Request: agentquery.Request{
			Mode: "search", Root: absolute, Query: "player movement realtime snapshot lifecycle", Limit: 8,
			LexiconCmd: lexiconCommand, ArcanaCmd: arcanaCommand,
		},
		StateMode: repostate.RefreshIfNeeded,
	}, options)
	if err != nil {
		t.Fatal(err)
	}
	if search.Query == nil || len(search.Query.Results) == 0 {
		t.Fatalf("search returned no code results: %#v", search)
	}
	if len(search.Knowledge) == 0 {
		t.Fatalf("search returned no knowledge results: %#v", search)
	}
	for _, result := range search.Query.Results {
		if strings.HasPrefix(result.Node.Path, "docs/") || strings.HasSuffix(result.Node.Path, ".md") {
			t.Fatalf("code lane returned documentation: %+v", result)
		}
	}
	encoded, _ := json.MarshalIndent(search, "", "  ")
	t.Log(string(encoded))

	anchor := search.Query.Results[0].Node.Handle.Value
	inspection, err := Execute(context.Background(), Request{
		Request: agentquery.Request{
			Mode: "inspect", Root: absolute, Handles: []string{anchor}, Adjacent: 2,
			LexiconCmd: lexiconCommand, ArcanaCmd: arcanaCommand,
		},
		StateMode: repostate.RefreshIfNeeded,
	}, options)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Query == nil || len(inspection.Query.Inspections) != 1 || inspection.Query.Inspections[0].Source == "" {
		t.Fatalf("inspect did not return exact code: %#v", inspection)
	}
	inspectionJSON, _ := json.MarshalIndent(inspection, "", "  ")
	t.Log(string(inspectionJSON))
}
