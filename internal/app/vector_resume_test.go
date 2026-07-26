package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/Lokee86/grimoire/internal/knowledgevector"
	"github.com/Lokee86/grimoire/internal/vectorstore"
)

func TestKnowledgeVectorBuildResumesAfterFailedBatch(t *testing.T) {
	if _, err := vectorstore.FindLibrary(""); err != nil {
		t.Skipf("Rust vector DLL is not built: %v", err)
	}
	var requests atomic.Int64
	var failSecond atomic.Bool
	failSecond.Store(true)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestNumber := requests.Add(1)
		if failSecond.Load() && requestNumber == 2 {
			http.Error(writer, "planned embedding failure", http.StatusInternalServerError)
			return
		}
		var body struct {
			Input []string `json:"input"`
		}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Error(err)
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		data := make([]map[string]any, len(body.Input))
		for index := range body.Input {
			vector := make([]float64, 512)
			vector[index%len(vector)] = 1
			data[index] = map[string]any{"index": index, "embedding": vector}
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(map[string]any{"data": data})
	}))
	defer server.Close()

	root := t.TempDir()
	for name, content := range map[string]string{
		"alpha.md": "# Alpha\n\nAlpha design rationale.\n",
		"beta.md":  "# Beta\n\nBeta design rationale.\n",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := Run([]string{"knowledge", "index", "--root", root}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	args := []string{
		"vector", "build", "--root", root, "--endpoint", server.URL,
		"--batch-size", "1", "--batch-concurrency", "1",
	}
	if err := Run(args, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
		t.Fatal("expected the planned second-batch failure")
	}
	if requests.Load() != 2 {
		t.Fatalf("first build made %d requests, expected 2", requests.Load())
	}

	failSecond.Store(false)
	beforeResume := requests.Load()
	var output bytes.Buffer
	if err := Run(args, &output, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if requests.Load()-beforeResume != 1 {
		t.Fatalf("resume made %d requests, expected 1", requests.Load()-beforeResume)
	}
	var result knowledgevector.BuildResult
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Sections != 2 || result.EmbeddedVectors != 1 || result.ReusedVectors != 1 {
		t.Fatalf("unexpected resumed build: %+v", result)
	}
}
