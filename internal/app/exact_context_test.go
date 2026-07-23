package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/compiler"
	"github.com/Lokee86/grimoire/internal/queryshape"
)

func TestContextSelectsAutomaticBudgetWhenOmitted(t *testing.T) {
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
		"context", "--root", root, "--query", "Where is ResolveDamage?", "--structure=false",
	}, &output, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	var pkg compiler.Package
	if err := json.Unmarshal(output.Bytes(), &pkg); err != nil {
		t.Fatal(err)
	}
	if pkg.Budget != queryshape.FocusedTargetTokens {
		t.Fatalf("expected focused automatic budget %d, got %d", queryshape.FocusedTargetTokens, pkg.Budget)
	}
	if pkg.Assembly == nil || pkg.Assembly.Scope != queryshape.ScopeFocused || pkg.Assembly.StopReason == "" {
		t.Fatalf("expected focused assembly decision, got %+v", pkg.Assembly)
	}
}

func TestContextUsesExactRecoveryDuringSemanticFallback(t *testing.T) {
	root := t.TempDir()
	content := "package damage\n\nfunc ResolveDamage() int { return 10 }\n"
	if err := os.WriteFile(filepath.Join(root, "damage.go"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Run([]string{"index", "--root", root}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	var output, errors bytes.Buffer
	if err := Run([]string{
		"context", "--root", root, "--query", "ResolveDamage", "--budget", "500",
	}, &output, &errors); err != nil {
		t.Fatal(err)
	}
	var pkg compiler.Package
	if err := json.Unmarshal(output.Bytes(), &pkg); err != nil {
		t.Fatal(err)
	}
	if pkg.Budget != 500 {
		t.Fatalf("expected explicit budget 500, got %d", pkg.Budget)
	}
	if pkg.Assembly != nil {
		t.Fatalf("explicit budget should preserve fixed assembly, got %+v", pkg.Assembly)
	}
	if len(pkg.Selections) != 1 || pkg.Selections[0].RetrievalSource != "exact" {
		t.Fatalf("expected exact selection, got %+v", pkg.Selections)
	}
	if len(pkg.RetrievalSources) != 1 || pkg.RetrievalSources[0] != "exact" {
		t.Fatalf("expected exact package source, got %+v", pkg.RetrievalSources)
	}
	if !strings.Contains(strings.Join(pkg.Selections[0].Reasons, "\n"), "also retrieved by lexical rank 1") {
		t.Fatalf("missing lexical provider evidence: %+v", pkg.Selections[0].Reasons)
	}
	if !strings.Contains(errors.String(), "using lexical fallback") {
		t.Fatalf("expected fallback warning, got %q", errors.String())
	}
}
