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

	"github.com/Lokee86/grimoire/internal/compiler"
	"github.com/Lokee86/grimoire/internal/vectorstore"
)

func TestVectorBuildReusesObjectsAndSearches(t *testing.T) {
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
	content := "package damage\n\nfunc ResolveDamage() int { return 10 }\n"
	if err := os.WriteFile(filepath.Join(root, "damage.go"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Run([]string{"index", "--root", root}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	var first bytes.Buffer
	if err := Run([]string{"vector", "build", "--root", root, "--endpoint", server.URL}, &first, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	var firstBuild struct {
		PreparedIdentity string `json:"prepared_identity"`
		Embedded         int    `json:"embedded"`
		Reused           int    `json:"reused"`
	}
	if err := json.Unmarshal(first.Bytes(), &firstBuild); err != nil {
		t.Fatal(err)
	}
	if firstBuild.PreparedIdentity == "" || firstBuild.Embedded != 1 || firstBuild.Reused != 0 {
		t.Fatalf("unexpected first build: %+v", firstBuild)
	}
	if _, err := os.Stat(resolveVectorPaths(filepath.Join(root, ".grimoire")).Manifest); err != nil {
		t.Fatalf("vector snapshot manifest was not published: %v", err)
	}

	var second bytes.Buffer
	if err := Run([]string{"vector", "build", "--root", root, "--endpoint", server.URL}, &second, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	var secondBuild struct {
		Embedded int `json:"embedded"`
		Reused   int `json:"reused"`
	}
	if err := json.Unmarshal(second.Bytes(), &secondBuild); err != nil {
		t.Fatal(err)
	}
	if secondBuild.Embedded != 0 || secondBuild.Reused != 1 {
		t.Fatalf("unexpected second build: %+v", secondBuild)
	}

	var output bytes.Buffer
	if err := Run([]string{
		"vector", "search", "--root", root,
		"--endpoint", server.URL, "--query", "where is damage resolved", "--top-k", "1",
	}, &output, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	var result struct {
		Results []struct {
			Path  string  `json:"path"`
			Score float32 `json:"score"`
		} `json:"results"`
	}
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Results) != 1 || result.Results[0].Path != "damage.go" || result.Results[0].Score < 0.99 {
		t.Fatalf("unexpected vector search: %+v", result.Results)
	}

	var contextOutput bytes.Buffer
	var contextErrors bytes.Buffer
	if err := Run([]string{
		"context", "--root", root, "--endpoint", server.URL,
		"--query", "where is damage resolved", "--candidate-limit", "1", "--budget", "500",
	}, &contextOutput, &contextErrors); err != nil {
		t.Fatal(err)
	}
	var contextPackage compiler.Package
	if err := json.Unmarshal(contextOutput.Bytes(), &contextPackage); err != nil {
		t.Fatal(err)
	}
	if contextErrors.Len() != 0 {
		t.Fatalf("unexpected context warning: %s", contextErrors.String())
	}
	if len(contextPackage.RetrievalSources) != 1 || contextPackage.RetrievalSources[0] != "vector" {
		t.Fatalf("expected vector retrieval, got %+v", contextPackage.RetrievalSources)
	}
	if len(contextPackage.Selections) != 1 {
		t.Fatalf("expected one vector selection, got %+v", contextPackage.Selections)
	}
	selection := contextPackage.Selections[0]
	if selection.Path != "damage.go" || selection.RetrievalSource != "vector" || selection.RetrievalRank != 1 || selection.Score < 0.99 {
		t.Fatalf("unexpected vector context selection: %+v", selection)
	}

	changed := "package damage\n\nfunc ResolveDamage() int { return 20 }\n"
	if err := os.WriteFile(filepath.Join(root, "damage.go"), []byte(changed), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Run([]string{"index", "--root", root}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	requestsBeforeStaleSearch := embeddingRequests.Load()
	if err := Run([]string{
		"vector", "search", "--root", root,
		"--endpoint", server.URL, "--query", "where is damage resolved", "--top-k", "1",
	}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "vector snapshot was built from prepared index") {
		t.Fatalf("expected exact stale-vector rejection, got %v", err)
	}
	if embeddingRequests.Load() != requestsBeforeStaleSearch {
		t.Fatal("stale vector search embedded the query before validating freshness")
	}

	contextOutput.Reset()
	contextErrors.Reset()
	if err := Run([]string{
		"context", "--root", root, "--endpoint", server.URL,
		"--query", "resolve damage", "--candidate-limit", "1", "--budget", "500",
	}, &contextOutput, &contextErrors); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(contextOutput.Bytes(), &contextPackage); err != nil {
		t.Fatal(err)
	}
	if len(contextPackage.RetrievalSources) != 1 || contextPackage.RetrievalSources[0] != "lexical" {
		t.Fatalf("expected stale-vector fallback, got %+v", contextPackage.RetrievalSources)
	}
	if !strings.Contains(contextErrors.String(), "vector snapshot was built from prepared index") {
		t.Fatalf("expected exact stale-vector warning, got %q", contextErrors.String())
	}
	if embeddingRequests.Load() != requestsBeforeStaleSearch {
		t.Fatal("stale context retrieval embedded the query before falling back")
	}
}
