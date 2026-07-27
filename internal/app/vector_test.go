package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/Lokee86/grimoire/internal/agentruntime"
	"github.com/Lokee86/grimoire/internal/knowledge"
	"github.com/Lokee86/grimoire/internal/knowledgevector"
	"github.com/Lokee86/grimoire/internal/vectorstore"
)

func TestKnowledgeVectorBuildReusesObjectsAndSearches(t *testing.T) {
	if _, err := vectorstore.FindLibrary(""); err != nil {
		t.Skipf("Rust vector DLL is not built: %v", err)
	}
	var embeddingRequests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		embeddingRequests.Add(1)
		var body struct {
			Input []string `json:"input"`
		}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Error(err)
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		data := make([]map[string]any, len(body.Input))
		for index, input := range body.Input {
			vector := make([]float64, 512)
			if strings.Contains(strings.ToLower(input), "damage") {
				vector[0] = 1
			} else {
				vector[1] = 1
			}
			data[index] = map[string]any{"index": index, "embedding": vector}
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(map[string]any{"data": data})
	}))
	defer server.Close()

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "damage.go"), []byte("package damage\n\nfunc ResolveDamage() int { return 10 }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "damage.md"), []byte("# Damage resolution\n\nDamage is resolved by ResolveDamage before health is updated.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "architecture.md"), []byte("# Architecture\n\nThe repository separates source analysis from documentation knowledge.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Run([]string{"index", "--root", root}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if err := Run([]string{"knowledge", "index", "--root", root}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	var first bytes.Buffer
	if err := Run([]string{"vector", "build", "--root", root, "--endpoint", server.URL}, &first, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	var firstBuild knowledgevector.BuildResult
	if err := json.Unmarshal(first.Bytes(), &firstBuild); err != nil {
		t.Fatal(err)
	}
	if firstBuild.KnowledgeIdentity == "" || firstBuild.Sections != 2 || firstBuild.EmbeddedVectors != 2 || firstBuild.ReusedVectors != 0 {
		t.Fatalf("unexpected first build: %+v", firstBuild)
	}
	knowledgeState, err := knowledge.DefaultState(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(knowledgevector.ResolvePaths(knowledgeState).Manifest); err != nil {
		t.Fatalf("knowledge vector manifest was not published: %v", err)
	}

	var second bytes.Buffer
	if err := Run([]string{"vector", "build", "--root", root, "--endpoint", server.URL}, &second, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	var secondBuild knowledgevector.BuildResult
	if err := json.Unmarshal(second.Bytes(), &secondBuild); err != nil {
		t.Fatal(err)
	}
	if secondBuild.EmbeddedVectors != 0 || secondBuild.ReusedVectors != 2 || !secondBuild.CachedSnapshot {
		t.Fatalf("unexpected second build: %+v", secondBuild)
	}

	var output bytes.Buffer
	if err := Run([]string{
		"knowledge", "search", "--root", root, "--endpoint", server.URL, "--vectors=true",
		"--query", "how is damage resolved", "--top-k", "1",
	}, &output, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	var result knowledge.SearchResponse
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if !result.VectorUsed || result.VectorError != "" || len(result.Results) != 1 || result.Results[0].Path != "docs/damage.md" {
		t.Fatalf("unexpected knowledge vector search: %+v", result)
	}

	var discoveryOutput bytes.Buffer
	if err := Run([]string{
		"search", "--root", root, "--query", "where is damage resolved",
		"--limit", "1", "--code-only",
	}, &discoveryOutput, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	var discovery agentruntime.Response
	if err := json.Unmarshal(discoveryOutput.Bytes(), &discovery); err != nil {
		t.Fatal(err)
	}
	if len(discovery.SourceMatches) != 1 || discovery.SourceMatches[0].Provider != "lexical" {
		t.Fatalf("source discovery should remain independently lexical: %+v", discovery.SourceMatches)
	}
	if len(discovery.DocumentMatches) != 0 {
		t.Fatalf("code-only discovery returned documentation: %+v", discovery.DocumentMatches)
	}

	if err := os.WriteFile(filepath.Join(root, "docs", "damage.md"), []byte("# Damage resolution\n\nDamage now resolves through ApplyDamage.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Run([]string{"knowledge", "index", "--root", root}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	requestsBeforeStaleSearch := embeddingRequests.Load()
	output.Reset()
	if err := Run([]string{
		"knowledge", "search", "--root", root, "--endpoint", server.URL, "--vectors=true",
		"--query", "how is damage resolved", "--top-k", "1",
	}, &output, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.VectorUsed || !strings.Contains(result.VectorError, "knowledge vector snapshot was built from") {
		t.Fatalf("expected stale knowledge vectors to fall back to BM25: %+v", result)
	}
	if embeddingRequests.Load() != requestsBeforeStaleSearch {
		t.Fatal("stale knowledge vector search embedded the query before validating freshness")
	}
}
