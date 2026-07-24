package arcanagraph

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSemanticSeedsConvertsRankedHits(t *testing.T) {
	client := Client{
		Command: "arcana-test",
		RunSemantic: func(
			_ context.Context,
			command string,
			state string,
			endpoint string,
			query string,
			limit int,
		) ([]semanticHit, error) {
			if command != "arcana-test" || state != "state" || endpoint != "endpoint" {
				t.Fatalf("unexpected invocation: command=%q state=%q endpoint=%q", command, state, endpoint)
			}
			if query != "profile persistence" || limit != 4 {
				t.Fatalf("unexpected query: query=%q limit=%d", query, limit)
			}
			return []semanticHit{
				{Score: 0.9, NodeKey: "1", Kind: "function", Path: "profile.go", Name: "CreateProfile"},
				{Score: 0.8, NodeKey: "2", Kind: "function", Path: "repository.go", Name: "InsertProfile"},
				{Score: 0.7, NodeKey: "3", Kind: "function", Path: "profile.go", Name: "CreateProfile"},
			}, nil
		},
	}

	seeds, err := client.SemanticSeeds(
		context.Background(), "state", "endpoint", "profile persistence", 4,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(seeds) != 2 {
		t.Fatalf("expected two unique seeds, got %+v", seeds)
	}
	if seeds[0].Name != "CreateProfile" || seeds[0].Path != "profile.go" || seeds[0].Identity != "1" {
		t.Fatalf("unexpected first seed: %+v", seeds[0])
	}
	if seeds[1].Name != "InsertProfile" || seeds[1].Path != "repository.go" {
		t.Fatalf("unexpected second seed: %+v", seeds[1])
	}
}

func TestSemanticSeedsSilentlySkipsMissingIndex(t *testing.T) {
	state := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(state, "CURRENT"),
		[]byte("sha256:"+strings.Repeat("0", 64)+"\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	client := Client{Command: filepath.Join(state, "missing-arcana")}
	seeds, err := client.SemanticSeeds(
		context.Background(), state, "endpoint", "profile persistence", 4,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(seeds) != 0 {
		t.Fatalf("expected no seeds, got %+v", seeds)
	}
}

func TestSemanticSeedsSkipsEmptyInputs(t *testing.T) {
	client := Client{
		RunSemantic: func(context.Context, string, string, string, string, int) ([]semanticHit, error) {
			t.Fatal("semantic runner should not be called")
			return nil, nil
		},
	}
	for _, test := range []struct {
		state string
		query string
		limit int
	}{
		{state: "", query: "query", limit: 1},
		{state: "state", query: "", limit: 1},
		{state: "state", query: "query", limit: 0},
	} {
		seeds, err := client.SemanticSeeds(context.Background(), test.state, "", test.query, test.limit)
		if err != nil || len(seeds) != 0 {
			t.Fatalf("expected empty result, got seeds=%+v err=%v", seeds, err)
		}
	}
}
