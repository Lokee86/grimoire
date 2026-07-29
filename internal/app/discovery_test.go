package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Lokee86/grimoire/internal/agentquery"
	"github.com/Lokee86/grimoire/internal/agentruntime"
)

func TestSearchUsesDefaultLimitPerLane(t *testing.T) {
	root := t.TempDir()
	content := "package damage\n\nfunc ResolveDamage() int { return 10 }\n"
	if err := os.WriteFile(filepath.Join(root, "damage.go"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Run([]string{"index", "--root", root}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	if err := Run([]string{
		"search", "--root", root, "--query", "Where is ResolveDamage?", "--code-only",
	}, &output, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	var response agentruntime.Response
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Schema != agentquery.SchemaVersion || response.Mode != "search" {
		t.Fatalf("unexpected discovery envelope: %+v", response)
	}
	if len(response.ExactMatches)+len(response.SourceMatches) == 0 {
		t.Fatalf("default discovery returned no source evidence: %+v", response)
	}
	for _, laneSize := range []int{
		len(response.ExactMatches),
		len(response.SourceMatches),
		len(response.SymbolMatches),
		len(response.RelationshipMatches),
	} {
		if laneSize > 6 {
			t.Fatalf("default per-lane limit exceeded: %+v", response)
		}
	}
}

func TestSearchDoesNotSpendBudgetOnDuplicateLexicalEvidence(t *testing.T) {
	root := t.TempDir()
	content := "package damage\n\nfunc ResolveDamage() int { return 10 }\n"
	if err := os.WriteFile(filepath.Join(root, "damage.go"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Run([]string{"index", "--root", root}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	if err := Run([]string{
		"search", "--root", root, "--query", "ResolveDamage", "--limit", "1", "--code-only",
	}, &output, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	var response agentruntime.Response
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.ExactMatches) != 1 || response.ExactMatches[0].Node.Path != "damage.go" {
		t.Fatalf("expected exact source evidence, got %+v", response.ExactMatches)
	}
	if len(response.SourceMatches) != 0 {
		t.Fatalf("duplicate lexical evidence consumed the global budget: %+v", response.SourceMatches)
	}
	if response.ExactMatches[0].Provider != "exact" {
		t.Fatalf("exact evidence provenance changed: %+v", response.ExactMatches)
	}
}
