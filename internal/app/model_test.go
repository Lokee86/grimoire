package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Lokee86/grimoire/internal/embedding"
)

func TestModelInfoReportsFixedContract(t *testing.T) {
	var output bytes.Buffer
	if err := Run([]string{"model", "info"}, &output, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	var result struct {
		Model      string `json:"model"`
		Dimensions int    `json:"dimensions"`
	}
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Model != embedding.ModelReference || result.Dimensions != embedding.Dimensions {
		t.Fatalf("unexpected model info: %+v", result)
	}
}

func TestModelProbeUsesEmbeddingEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		vector := make([]float64, embedding.NativeDimensions)
		vector[0] = 1
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"data": []any{map[string]any{"index": 0, "embedding": vector}},
		})
	}))
	defer server.Close()

	var output bytes.Buffer
	if err := Run([]string{"model", "probe", "--endpoint", server.URL}, &output, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	var result struct {
		Dimensions int     `json:"dimensions"`
		Similarity float64 `json:"similarity"`
	}
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Dimensions != embedding.Dimensions || result.Similarity != 1 {
		t.Fatalf("unexpected probe result: %+v", result)
	}
}
