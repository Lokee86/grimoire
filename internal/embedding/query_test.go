package embedding

import (
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/tokenizer"
)

func TestPlanQueryFastCapsAndGroupsTokens(t *testing.T) {
	query := strings.Repeat("alpha beta gamma delta ", 80)
	options := QueryOptions{Mode: QueryModeFast, WindowTokens: 16, MaxTokens: 40}
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
		if count > 16 {
			t.Fatalf("window %d has %d tokens", index, count)
		}
		total += count
	}
	if total != 40 {
		t.Fatalf("planned %d tokens, want 40", total)
	}
}

func TestPlanQueryModes(t *testing.T) {
	query := strings.Repeat("alpha beta gamma delta ", 20)
	fast, err := PlanQuery(query, QueryOptions{Mode: QueryModeFast, WindowTokens: 16, MaxTokens: 64})
	if err != nil {
		t.Fatal(err)
	}
	full, err := PlanQuery(query, QueryOptions{Mode: QueryModeFull, WindowTokens: 16, MaxTokens: 64})
	if err != nil {
		t.Fatal(err)
	}
	quality, err := PlanQuery(query, QueryOptions{Mode: QueryModeQuality, WindowTokens: 16, MaxTokens: 64})
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
	inputs, err := PlanQuery("find damage", QueryOptions{
		Mode: QueryModeQuality, WindowTokens: 16, MaxTokens: 128,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(inputs) != 1 || inputs[0].Label != "full query" {
		t.Fatalf("unexpected short quality plan: %+v", inputs)
	}
}

func TestQueryOptionsRejectInvalidValues(t *testing.T) {
	for _, options := range []QueryOptions{
		{Mode: "unknown", WindowTokens: 16, MaxTokens: 128},
		{Mode: QueryModeFast, WindowTokens: 0, MaxTokens: 128},
		{Mode: QueryModeFast, WindowTokens: 32, MaxTokens: 16},
	} {
		if err := options.Validate(); err == nil {
			t.Fatalf("expected invalid options: %+v", options)
		}
	}
}
