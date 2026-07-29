package agentquery

import (
	"strings"
	"testing"
)

func TestSearchUsesCompactDefaults(t *testing.T) {
	search := normalizeRequest(Request{Mode: "search"})
	orient := normalizeRequest(Request{Mode: "orient"})
	trace := normalizeRequest(Request{Mode: "trace"})
	if search.Limit != 6 || orient.Limit != 6 || trace.Limit != 8 {
		t.Fatalf("unexpected default limits: search=%d orient=%d trace=%d", search.Limit, orient.Limit, trace.Limit)
	}
}

func TestDiscoveryExcerptIsBounded(t *testing.T) {
	text := strings.Repeat("abcdefghij ", 100)
	excerpt := compactExcerpt(text)
	if len(excerpt) > maxDiscoveryExcerptBytes+len("…") || !strings.HasSuffix(excerpt, "…") {
		t.Fatalf("excerpt was not compacted: bytes=%d value=%q", len(excerpt), excerpt)
	}
}

func TestRelationshipBucketsAreInterleavedBySeed(t *testing.T) {
	node := func(value string) Node {
		return Node{Handle: Handle{Value: value, Provider: "lexicon", NodeIdentity: value}}
	}
	buckets := [][]RelationshipMatch{
		{
			{Subject: node("seed-a"), Direction: "outgoing", Relation: "calls", Object: node("a-1"), SeedLane: "exact_matches"},
			{Subject: node("seed-a"), Direction: "outgoing", Relation: "calls", Object: node("a-2"), SeedLane: "exact_matches"},
		},
		{
			{Subject: node("seed-b"), Direction: "outgoing", Relation: "uses", Object: node("b-1"), SeedLane: "symbol_matches"},
			{Subject: node("seed-b"), Direction: "outgoing", Relation: "uses", Object: node("b-2"), SeedLane: "symbol_matches"},
		},
	}

	matches, truncated := interleaveRelationshipBuckets(buckets, 2)
	if !truncated || len(matches) != 2 {
		t.Fatalf("interleaved matches = %+v truncated=%v", matches, truncated)
	}
	if matches[0].Object.Handle.Value != "a-1" || matches[1].Object.Handle.Value != "b-1" {
		t.Fatalf("first seed monopolized relationship budget: %+v", matches)
	}
	if matches[0].Rank != 1 || matches[1].Rank != 2 {
		t.Fatalf("relationship ranks were not assigned after interleaving: %+v", matches)
	}
}
