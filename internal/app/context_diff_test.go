package app

import (
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/diffcontext"
	"github.com/Lokee86/grimoire/internal/index"
)

func TestBuildContextDiffPrioritizesChangedChunksAndAnchorsRetrieval(t *testing.T) {
	snapshot := index.Snapshot{Files: []index.FileRecord{{
		Path: "internal/auth/login.go",
		Chunks: []index.Chunk{{
			ID: "login", Path: "internal/auth/login.go", StartLine: 8, EndLine: 20,
			TokenCount: 20, Text: "func Login(user User) error {",
		}},
	}}}
	result := buildContextDiff(snapshot, "", []diffcontext.Change{{
		Path: "internal/auth/login.go", StartLine: 10, EndLine: 12,
		Summary: "func Login(user User) error",
	}}, 200)

	if result.PackageQuery != diffcontext.DefaultQuery {
		t.Fatalf("package query = %q, want default diff review query", result.PackageQuery)
	}
	if len(result.Candidates) != 1 || result.Candidates[0].Source != "git-diff" {
		t.Fatalf("unexpected changed candidates: %#v", result.Candidates)
	}
	if len(result.Evidence) != 1 || result.Evidence[0].Provider != "git-diff" {
		t.Fatalf("unexpected changed evidence: %#v", result.Evidence)
	}
	for _, anchor := range []string{"internal/auth/login.go", "func Login(user User) error"} {
		if !strings.Contains(result.RetrievalQuery, anchor) {
			t.Fatalf("retrieval query %q does not contain %q", result.RetrievalQuery, anchor)
		}
	}
}

func TestBuildContextDiffKeepsExplicitPackageQuery(t *testing.T) {
	result := buildContextDiff(index.Snapshot{}, "check authorization regressions", []diffcontext.Change{{
		Path: "auth.go", StartLine: 1, EndLine: 1,
	}}, 10)
	if result.PackageQuery != "check authorization regressions" {
		t.Fatalf("explicit query changed: %q", result.PackageQuery)
	}
}
