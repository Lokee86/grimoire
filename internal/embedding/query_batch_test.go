package embedding

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestQueryBatchesCapSplitRequestsAt64Tokens(t *testing.T) {
	plan := make([]QueryInput, 10)
	for index := range plan {
		plan[index] = QueryInput{Text: fmt.Sprintf("window-%d", index)}
	}
	batches, err := queryBatches(plan, DefaultQueryOptions())
	if err != nil {
		t.Fatal(err)
	}
	if len(batches) != 3 {
		t.Fatalf("got %d batches, want 3", len(batches))
	}
	for index, want := range []struct{ start, size int }{{0, 4}, {4, 4}, {8, 2}} {
		if batches[index].start != want.start || len(batches[index].inputs) != want.size {
			t.Fatalf("batch %d = start %d size %d", index, batches[index].start, len(batches[index].inputs))
		}
	}
}

func TestQueryBatchesKeepQualityFullQuerySeparate(t *testing.T) {
	options := DefaultQueryOptions()
	options.Mode = QueryModeQuality
	plan := []QueryInput{{Text: "full"}, {Text: "one"}, {Text: "two"}, {Text: "three"}, {Text: "four"}}
	batches, err := queryBatches(plan, options)
	if err != nil {
		t.Fatal(err)
	}
	if len(batches) != 2 || len(batches[0].inputs) != 1 || len(batches[1].inputs) != 4 {
		t.Fatalf("unexpected quality batches: %+v", batches)
	}
}

func TestEmbedQueryPlanBatchesRequestsAndPreservesOrder(t *testing.T) {
	var mutex sync.Mutex
	requestSizes := make([]int, 0, 3)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var body embeddingsRequest
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Error(err)
			return
		}
		mutex.Lock()
		requestSizes = append(requestSizes, len(body.Input))
		mutex.Unlock()
		data := make([]map[string]any, len(body.Input))
		for responseIndex, input := range body.Input {
			marker := input[strings.LastIndex(input, "q-")+2:]
			inputIndex, err := strconv.Atoi(marker)
			if err != nil {
				t.Error(err)
				return
			}
			vector := make([]float64, NativeDimensions)
			vector[inputIndex] = 1
			data[responseIndex] = map[string]any{"index": responseIndex, "embedding": vector}
		}
		_ = json.NewEncoder(writer).Encode(map[string]any{"data": data})
	}))
	defer server.Close()

	plan := make([]QueryInput, 10)
	for index := range plan {
		plan[index] = QueryInput{Text: fmt.Sprintf("q-%d", index)}
	}
	vectors, err := NewClient(server.URL).EmbedQueryPlan(context.Background(), plan, DefaultQueryOptions())
	if err != nil {
		t.Fatal(err)
	}
	if len(vectors) != len(plan) {
		t.Fatalf("got %d vectors, want %d", len(vectors), len(plan))
	}
	for index, vector := range vectors {
		if vector[index] != 1 {
			t.Fatalf("vector %d was returned out of order", index)
		}
	}
	sort.Ints(requestSizes)
	if fmt.Sprint(requestSizes) != "[2 4 4]" {
		t.Fatalf("request sizes = %v, want [2 4 4]", requestSizes)
	}
}
