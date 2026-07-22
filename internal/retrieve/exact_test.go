package retrieve

import (
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/index"
)

func TestExactMatchesEverySignalClass(t *testing.T) {
	tests := []struct {
		name, query, path, text, reason string
	}{
		{"quoted phrase", `"exact recovery"`, "notes.txt", "exact recovery details", "quoted phrase"},
		{"path", "internal/retrieve/exact.go", "internal/retrieve/exact.go", "path fixture", "path"},
		{"filename", "README.md", "docs/README.md", "filename fixture", "filename"},
		{"identifier", "ResolveDamage", "damage.go", "func ResolveDamage() {}", "identifier"},
		{"configuration key", "server.port=8080", "config.txt", "server.port = 8080", "configuration key"},
		{"error code", "ERR-42", "errors.txt", "return ERR-42", "error code"},
		{"version", "1.2.3", "versions.txt", "release 1.2.3", "version string"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			results := Exact(oneChunk(test.path, test.text), test.query, 10)
			if len(results) != 1 || results[0].Source != "exact" || results[0].Rank != 1 {
				t.Fatalf("unexpected result: %+v", results)
			}
			if !hasReason(results[0], test.reason) {
				t.Fatalf("missing %q reason: %+v", test.reason, results[0].Reasons)
			}
		})
	}
}

func TestExactIgnoresNaturalLanguageWithoutSignals(t *testing.T) {
	snapshot := oneChunk("notes.txt", "repository state operation")
	if results := Exact(snapshot, "repository state operation", 10); results != nil {
		t.Fatalf("ordinary lowercase prose activated exact retrieval: %+v", results)
	}
	if results := Exact(oneChunk("where.md", "Where should context be compiled?"), "Where is context compiled?", 10); results != nil {
		t.Fatalf("capitalized question prose activated exact retrieval: %+v", results)
	}
	if results := Exact(oneChunk("config.toml", "max_per_hit = 100"), "maximum damage per hit", 10); results != nil {
		t.Fatalf("ordinary prose activated exact retrieval: %+v", results)
	}
}

func TestExactRecoversDottedConfigurationTerminalKey(t *testing.T) {
	snapshot := oneChunk("config.toml", "[damage]\nmax_per_hit = 100")
	results := Exact(snapshot, "damage.max_per_hit", 10)
	if len(results) != 1 || results[0].Chunk.Path != "config.toml" || results[0].Rank != 1 {
		t.Fatalf("dotted configuration did not recover terminal key: %+v", results)
	}
	if !hasReason(results[0], "configuration key") || !hasReason(results[0], "damage.max_per_hit") {
		t.Fatalf("missing dotted configuration provenance: %+v", results[0].Reasons)
	}
}

func TestExactAggregatesReasonsAndIgnoresNonExactWords(t *testing.T) {
	snapshot := index.Snapshot{Version: index.FormatVersion, Files: []index.FileRecord{
		{Path: "recovery.go", Chunks: []index.Chunk{{ID: "hit", Path: "recovery.go", StartLine: 4, EndLine: 8,
			Text: `exact recovery ResolveDamage server.port = 8080 ERR-42 version 1.2.3`}}},
		{Path: "ordinary.txt", Chunks: []index.Chunk{{ID: "miss", Path: "ordinary.txt", StartLine: 1, Text: "ordinary words only"}}},
	}}
	results := Exact(snapshot, `"exact recovery" ResolveDamage server.port=8080 ERR-42 1.2.3`, 10)
	if len(results) != 1 || results[0].Chunk.ID != "hit" {
		t.Fatalf("unexpected aggregate results: %+v", results)
	}
	for _, reason := range []string{"quoted phrase", "identifier", "configuration key", "error code", "version string"} {
		if !hasReason(results[0], reason) {
			t.Errorf("missing aggregated %q reason: %+v", reason, results[0].Reasons)
		}
	}
}

func TestExactLimitsAndBreaksTiesDeterministically(t *testing.T) {
	snapshot := index.Snapshot{Version: index.FormatVersion, Files: []index.FileRecord{
		{Path: "b.go", Chunks: []index.Chunk{{ID: "b", Path: "b.go", StartLine: 1, Text: "ResolveDamage"}}},
		{Path: "a.go", Chunks: []index.Chunk{{ID: "a", Path: "a.go", StartLine: 1, Text: "ResolveDamage"}}},
	}}
	results := Exact(snapshot, "ResolveDamage", 1)
	if len(results) != 1 || results[0].Chunk.Path != "a.go" || results[0].Rank != 1 {
		t.Fatalf("unexpected limited tie-break results: %+v", results)
	}
	results = Exact(snapshot, "ResolveDamage", 10)
	if len(results) != 2 || results[0].Rank != 1 || results[1].Rank != 2 || results[1].Chunk.Path != "b.go" {
		t.Fatalf("unexpected full tie-break results: %+v", results)
	}
}

func oneChunk(path, text string) index.Snapshot {
	return index.Snapshot{Version: index.FormatVersion, Files: []index.FileRecord{{Path: path, Chunks: []index.Chunk{{ID: "chunk", Path: path, StartLine: 1, Text: text}}}}}
}

func hasReason(candidate Candidate, want string) bool {
	for _, reason := range candidate.Reasons {
		if strings.Contains(reason, want) {
			return true
		}
	}
	return false
}
