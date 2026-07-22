package embedding

import (
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/tokenizer"
)

func TestPlanQueryFastOptionalCapAndWindows(t *testing.T) {
	query := strings.Repeat("alpha beta gamma delta ", 80)
	options := testQueryOptions(QueryModeFast, 40)
	inputs, err := PlanQuery(query, options)
	if err != nil {
		t.Fatal(err)
	}
	if len(inputs) != 3 {
		t.Fatalf("got %d windows, want 3", len(inputs))
	}
	total := 0
	for index, input := range inputs {
		count, err := tokenizer.Count(input.Text)
		if err != nil {
			t.Fatal(err)
		}
		if count > options.WindowTokens {
			t.Fatalf("window %d has %d tokens", index, count)
		}
		total += count
	}
	if total != 40 {
		t.Fatalf("planned %d tokens, want 40", total)
	}
}

func TestPlanQueryFastIsUnlimitedByDefault(t *testing.T) {
	query := strings.Repeat("alpha beta gamma delta ", 80)
	inputs, err := PlanQuery(query, DefaultQueryOptions())
	if err != nil {
		t.Fatal(err)
	}
	want, err := tokenizer.Count(strings.TrimSpace(query))
	if err != nil {
		t.Fatal(err)
	}
	got := 0
	for _, input := range inputs {
		count, countErr := tokenizer.Count(input.Text)
		if countErr != nil {
			t.Fatal(countErr)
		}
		got += count
	}
	if got != want {
		t.Fatalf("planned %d tokens, want complete query with %d", got, want)
	}
}

func TestPlanQueryModes(t *testing.T) {
	query := strings.Repeat("alpha beta gamma delta ", 20)
	fast, err := PlanQuery(query, testQueryOptions(QueryModeFast, 64))
	if err != nil {
		t.Fatal(err)
	}
	full, err := PlanQuery(query, testQueryOptions(QueryModeFull, 64))
	if err != nil {
		t.Fatal(err)
	}
	quality, err := PlanQuery(query, testQueryOptions(QueryModeQuality, 64))
	if err != nil {
		t.Fatal(err)
	}
	if len(fast) != 4 || len(full) != 1 || len(quality) != 5 {
		t.Fatalf("unexpected plan sizes: fast=%d full=%d quality=%d", len(fast), len(full), len(quality))
	}
	if full[0].Label != "full query" || quality[0].Label != "full query" {
		t.Fatalf("full-query labels missing: full=%+v quality=%+v", full, quality)
	}
}

func TestPlanQueryQualityDoesNotDuplicateShortQuery(t *testing.T) {
	inputs, err := PlanQuery("find damage", testQueryOptions(QueryModeQuality, 0))
	if err != nil {
		t.Fatal(err)
	}
	if len(inputs) != 1 || inputs[0].Label != "full query" {
		t.Fatalf("unexpected short quality plan: %+v", inputs)
	}
}

func TestQueryOptionsRejectInvalidValues(t *testing.T) {
	for _, options := range []QueryOptions{
		{Mode: "unknown", WindowTokens: 16, BatchTokens: 64, BatchConcurrency: 2},
		{Mode: QueryModeFast, WindowTokens: 0, BatchTokens: 64, BatchConcurrency: 2},
		{Mode: QueryModeFast, WindowTokens: 32, BatchTokens: 16, BatchConcurrency: 2},
		{Mode: QueryModeFast, WindowTokens: 16, BatchTokens: 64, BatchConcurrency: 0},
		{Mode: QueryModeFast, WindowTokens: 16, BatchTokens: 64, BatchConcurrency: 2, MaxTokens: -1},
	} {
		if err := options.Validate(); err == nil {
			t.Fatalf("expected invalid options: %+v", options)
		}
	}
}

func testQueryOptions(mode QueryMode, maxTokens int) QueryOptions {
	options := DefaultQueryOptions()
	options.Mode = mode
	options.MaxTokens = maxTokens
	return options
}
