package embedding

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientFormatsQueryAndNormalizesTruncatedVectors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/embeddings" {
			t.Fatalf("unexpected path %q", request.URL.Path)
		}
		var body embeddingsRequest
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if len(body.Input) != 1 || !strings.HasPrefix(body.Input[0], "Instruct: "+QueryInstruction+"\nQuery:") {
			t.Fatalf("unexpected query input: %+v", body.Input)
		}
		vector := make([]float64, NativeDimensions)
		vector[0] = 3
		vector[1] = 4
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"data": []any{map[string]any{"index": 0, "embedding": vector}},
		})
	}))
	defer server.Close()

	vector, err := NewClient(server.URL+"/v1").EmbedQuery(context.Background(), "find damage")
	if err != nil {
		t.Fatal(err)
	}
	if len(vector) != Dimensions {
		t.Fatalf("got %d dimensions", len(vector))
	}
	if math.Abs(float64(vector[0])-0.6) > 0.00001 || math.Abs(float64(vector[1])-0.8) > 0.00001 {
		t.Fatalf("unexpected normalized vector prefix: %v", vector[:2])
	}
}

func TestClientPreservesResponseIndexes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		first := make([]float64, NativeDimensions)
		second := make([]float64, NativeDimensions)
		first[0] = 1
		second[1] = 1
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"data": []any{
				map[string]any{"index": 1, "embedding": second},
				map[string]any{"index": 0, "embedding": first},
			},
		})
	}))
	defer server.Close()

	vectors, err := NewClient(server.URL).EmbedDocuments(context.Background(), []string{"first", "second"})
	if err != nil {
		t.Fatal(err)
	}
	if vectors[0][0] != 1 || vectors[1][1] != 1 {
		t.Fatalf("response order was not restored: %v %v", vectors[0][:2], vectors[1][:2])
	}
}

func TestClientRejectsShortVector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"data": []any{map[string]any{"index": 0, "embedding": []float64{1, 2}}},
		})
	}))
	defer server.Close()

	_, err := NewClient(server.URL).EmbedDocuments(context.Background(), []string{"document"})
	if err == nil || !strings.Contains(err.Error(), "need at least 512") {
		t.Fatalf("unexpected error: %v", err)
	}
}
