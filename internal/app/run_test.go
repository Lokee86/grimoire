package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Lokee86/grimoire/internal/agentquery"
	"github.com/Lokee86/grimoire/internal/compiler"
	"github.com/Lokee86/grimoire/internal/index"
)

func TestRootHelpIsUseful(t *testing.T) {
	for _, args := range [][]string{nil, {"help"}, {"-h"}, {"--help"}} {
		var output bytes.Buffer
		if err := Run(args, &output, &bytes.Buffer{}); err != nil {
			t.Fatalf("Run(%v): %v", args, err)
		}
		for _, expected := range []string{"Usage:", "grimoire model start", "grimoire query orient", "grimoire context", "Lexicon", "Arcana"} {
			if !bytes.Contains(output.Bytes(), []byte(expected)) {
				t.Fatalf("Run(%v) help missing %q:\n%s", args, expected, output.String())
			}
		}
	}
}

func TestUnknownCommandPointsToHelp(t *testing.T) {
	err := Run([]string{"unknown"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !bytes.Contains([]byte(err.Error()), []byte("grimoire help")) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInvestigationLifecycleCommands(t *testing.T) {
	root := t.TempDir()
	var output bytes.Buffer
	if err := Run([]string{"investigation", "create", "--root", root, "--session", "cli-session", "--snapshot", "repo:1", "--provider", "arcana=arc:1"}, &output, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(output.Bytes(), []byte(`"session_id": "cli-session"`)) {
		t.Fatalf("create output missing session: %s", output.String())
	}
	output.Reset()
	if err := Run([]string{"investigation", "close", "--root", root, "--session", "cli-session"}, &output, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(output.Bytes(), []byte(`"closed_at"`)) {
		t.Fatalf("close output missing closed_at: %s", output.String())
	}
}

func TestIndexUsesConfiguredIgnoreFile(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".contextignore"), []byte("ignored.go\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "ignored.go"), []byte("package ignored\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "visible.go"), []byte("package visible\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Run([]string{
		"index", "--root", root, "--ignore-file", ".contextignore",
	}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	snapshot, err := index.Load(filepath.Join(root, ".grimoire"))
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Files) != 1 || snapshot.Files[0].Path != "visible.go" {
		t.Fatalf("unexpected indexed files: %+v", snapshot.Files)
	}
}

func TestIndexIncludeGeneratedFlag(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "vendor"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "vendor", "dependency.go"), []byte("package dependency\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "visible.go"), []byte("package visible\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Run([]string{"index", "--root", root}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	snapshot, err := index.Load(filepath.Join(root, ".grimoire"))
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Files) != 1 || snapshot.Files[0].Path != "visible.go" {
		t.Fatalf("unexpected default files: %+v", snapshot.Files)
	}

	if err := Run([]string{"index", "--root", root, "--include-generated"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	snapshot, err = index.Load(filepath.Join(root, ".grimoire"))
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Files) != 2 || snapshot.Files[0].Path != "vendor/dependency.go" || snapshot.Files[1].Path != "visible.go" {
		t.Fatalf("unexpected included files: %+v", snapshot.Files)
	}
}

func TestIndexThenCompileContext(t *testing.T) {
	root := t.TempDir()
	content := "package damage\n\nfunc ResolveDamage() int { return 10 }\n"
	if err := os.WriteFile(filepath.Join(root, "damage.go"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	var indexOutput bytes.Buffer
	if err := Run([]string{"index", "--root", root}, &indexOutput, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".grimoire", "objects")); err != nil {
		t.Fatal(err)
	}

	var contextOutput bytes.Buffer
	var contextErrors bytes.Buffer
	if err := Run([]string{
		"context", "--root", root,
		"--query", "resolve damage",
		"--budget", "500",
	}, &contextOutput, &contextErrors); err != nil {
		t.Fatal(err)
	}

	var result compiler.Package
	if err := json.Unmarshal(contextOutput.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Selections) != 1 {
		t.Fatalf("expected one selection, got %+v", result.Selections)
	}
	if result.Selections[0].Path != "damage.go" {
		t.Fatalf("unexpected selection: %+v", result.Selections[0])
	}
	if len(result.RetrievalSources) != 2 || result.RetrievalSources[0] != "lexical" || result.RetrievalSources[1] != "lexical-file" {
		t.Fatalf("expected chunk and file lexical discovery, got %+v", result.RetrievalSources)
	}
	if result.Selections[0].RetrievalSource != "lexical" || result.Selections[0].RetrievalRank != 1 {
		t.Fatalf("unexpected fallback provenance: %+v", result.Selections[0])
	}
	if contextErrors.Len() != 0 {
		t.Fatalf("unexpected context warning: %q", contextErrors.String())
	}
}

func TestQueryAcceptsVersionedJSONRequest(t *testing.T) {
	root := t.TempDir()
	content := "package damage\n\nfunc ResolveDamage() int { return 10 }\n"
	if err := os.WriteFile(filepath.Join(root, "damage.go"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Run([]string{"index", "--root", root}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	request, err := json.Marshal(agentquery.Request{
		Schema: agentquery.SchemaVersion, Mode: "search", Root: root,
		Query: "ResolveDamage", Limit: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := Run([]string{"query", "--request", string(request)}, &output, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	var response agentquery.Response
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Schema != agentquery.SchemaVersion || response.Mode != "search" ||
		len(response.Results) == 0 || response.Results[0].Node.Handle.Value == "" {
		t.Fatalf("unexpected query response: %+v", response)
	}
}
