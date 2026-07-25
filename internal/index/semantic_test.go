package index

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChunkFileWithSourceSpansUsesDeclarationBoundariesAndFallbackGaps(t *testing.T) {
	content := "package sample\n\nvar before = 1\n\nfunc First() {\n\tprintln(before)\n}\n\ntype Holder struct {\n\tValue int\n}\n\nfunc (Holder) Second() {\n\tprintln(\"second\")\n}\n\nvar after = 2\n"
	chunks, err := chunkFileWithSourceSpans("sample.go", content, []SourceSpan{
		{Path: "sample.go", StartLine: 5, EndLine: 7, Kind: "function", Name: "First"},
		{Path: "sample.go", StartLine: 9, EndLine: 15, Kind: "type", Name: "Holder"},
		{Path: "sample.go", StartLine: 13, EndLine: 15, Kind: "method", Name: "Second"},
	})
	if err != nil {
		t.Fatal(err)
	}

	semantic := make(map[string]Chunk)
	previousEnd := 0
	for _, chunk := range chunks {
		assertChunkTokenLimit(t, chunk)
		if chunk.StartLine <= previousEnd {
			t.Fatalf("chunk ranges overlap: previous end %d, current %+v", previousEnd, chunk)
		}
		previousEnd = chunk.EndLine
		if chunk.SemanticName != "" {
			semantic[chunk.SemanticName] = chunk
		}
	}
	if first := semantic["First"]; first.StartLine != 5 || first.EndLine != 7 || first.SemanticKind != "function" {
		t.Fatalf("unexpected First chunk: %+v", first)
	}
	if second := semantic["Second"]; second.StartLine != 13 || second.EndLine != 15 || second.SemanticKind != "method" {
		t.Fatalf("unexpected Second chunk: %+v", second)
	}
	if _, exists := semantic["Holder"]; exists {
		t.Fatal("containing type span should not duplicate its nested method source")
	}
	if chunks[0].StartLine != 1 || chunks[len(chunks)-1].EndLine != 17 {
		t.Fatalf("fallback gaps did not retain complete nonblank source coverage: %+v", chunks)
	}
}

func TestBuildRebuildsWhenSemanticBoundariesChange(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "sample.go")
	content := []byte("package sample\n\nfunc Value() int {\n\treturn 1\n}\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	first, firstStats, err := Build(root, nil, BuildOptions{SourceSpans: []SourceSpan{{
		Path: "sample.go", StartLine: 3, EndLine: 5, Kind: "function", Name: "Value",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	if firstStats.SemanticFiles != 1 || firstStats.SemanticChunks != 1 {
		t.Fatalf("unexpected semantic stats: %+v", firstStats)
	}

	second, secondStats, err := Build(root, &first, BuildOptions{SourceSpans: []SourceSpan{{
		Path: "sample.go", StartLine: 3, EndLine: 4, Kind: "function", Name: "Value",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	if secondStats.Updated != 1 || secondStats.Reused != 0 {
		t.Fatalf("changed semantic boundary was reused: %+v", secondStats)
	}
	if second.Files[0].PreparationHash == first.Files[0].PreparationHash {
		t.Fatal("preparation hash did not change with the semantic boundary")
	}
}
