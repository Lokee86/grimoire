package agentruntime

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/agentquery"
	"github.com/Lokee86/grimoire/internal/repostate"
)

func pathWithoutProviders(value string) string {
	kept := make([]string, 0)
	for _, directory := range filepath.SplitList(value) {
		containsProvider := false
		for _, provider := range []string{"lexicon", "arcana"} {
			name := provider
			if runtime.GOOS == "windows" {
				name += ".exe"
			}
			if regularFile(filepath.Join(directory, name)) {
				containsProvider = true
				break
			}
		}
		if !containsProvider {
			kept = append(kept, directory)
		}
	}
	return strings.Join(kept, string(os.PathListSeparator))
}

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
	if os.Getenv("GRIMOIRE_SMOKE_CLEAR_PATH") != "" {
		pathValue := pathWithoutProviders(os.Getenv("PATH"))
		t.Setenv("PATH", pathValue)
		if _, err := exec.LookPath("git"); err != nil {
			t.Fatalf("provider-isolation PATH removed required git dependency: %v", err)
		}
		for _, provider := range []string{"lexicon", "arcana"} {
			if path, err := exec.LookPath(provider); err == nil {
				t.Fatalf("provider %s remains discoverable through PATH at %s", provider, path)
			}
		}
	}
	options := Options{GrimoireCommand: command}
	search, err := Execute(context.Background(), Request{
		Request: agentquery.Request{
			Mode: "search", Root: absolute, Query: "player movement realtime snapshot lifecycle", Limit: 8,
		},
		StateMode: repostate.RefreshIfNeeded,
	}, options)
	if err != nil {
		encoded, _ := json.MarshalIndent(search, "", "  ")
		t.Log(string(encoded))
		t.Fatal(err)
	}
	codeResults := append([]agentquery.Result(nil), search.ExactMatches...)
	codeResults = append(codeResults, search.SourceMatches...)
	codeResults = append(codeResults, search.SymbolMatches...)
	if len(codeResults) == 0 {
		t.Fatalf("search returned no code results: %#v", search)
	}
	if len(search.DocumentMatches) == 0 {
		t.Fatalf("search returned no document results: %#v", search)
	}
	for _, result := range codeResults {
		if strings.HasPrefix(result.Node.Path, "docs/") || strings.HasSuffix(result.Node.Path, ".md") {
			t.Fatalf("code lane returned documentation: %+v", result)
		}
	}
	encoded, _ := json.Marshal(search)
	t.Logf("search response bytes=%d code_results=%d document_results=%d", len(encoded), len(codeResults), len(search.DocumentMatches))

	anchor := codeResults[0].Node.Handle.Value
	inspection, err := Execute(context.Background(), Request{
		Request: agentquery.Request{
			Mode: "inspect", Root: absolute, Handles: []string{anchor}, Adjacent: 2,
		},
		StateMode: repostate.RefreshIfNeeded,
	}, options)
	if err != nil {
		t.Fatal(err)
	}
	if len(inspection.Inspections) != 1 || inspection.Inspections[0].Source == "" {
		t.Fatalf("inspect did not return exact code: %#v", inspection)
	}
	inspectionJSON, _ := json.Marshal(inspection)
	t.Logf("inspect response bytes=%d inspections=%d", len(inspectionJSON), len(inspection.Inspections))
}
