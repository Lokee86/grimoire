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
