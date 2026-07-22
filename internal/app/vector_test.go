package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/vectorstore"
)

func TestVectorBuildReusesObjectsAndSearches(t *testing.T) {
	if _, err := vectorstore.FindLibrary(""); err != nil {
		t.Skipf("Rust vector DLL is not built: %v", err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
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
		Embedded int `json:"embedded"`
		Reused   int `json:"reused"`
	}
	if err := json.Unmarshal(first.Bytes(), &firstBuild); err != nil {
		t.Fatal(err)
	}
	if firstBuild.Embedded != 1 || firstBuild.Reused != 0 {
		t.Fatalf("unexpected first build: %+v", firstBuild)
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
}
