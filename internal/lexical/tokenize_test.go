package lexical

import (
	"slices"
	"testing"
)

func TestAnalyzeIndexesIdentifierComponentsAndFields(t *testing.T) {
	document := Analyze(Input{
		Key:  "chunk-1",
		Path: "internal/damage/ResolveDamage.go",
		Text: "func ResolveDamagePacket() { return ERR_42 }",
	})
	for _, term := range []string{"resolvedamagepacket", "resolve", "damage", "packet", "err", "42"} {
		if !hasTerm(document.Terms, term) {
			t.Errorf("missing content term %q: %+v", term, document.Terms)
		}
	}
	for _, term := range []string{"resolvedamage", "resolve", "damage", "go"} {
		if !slices.Contains(document.BaseTokens, term) {
			t.Errorf("missing basename term %q: %v", term, document.BaseTokens)
		}
	}
	if document.Length == 0 {
		t.Fatal("document length was not recorded")
	}
}

func hasTerm(terms []TermFrequency, want string) bool {
	for _, term := range terms {
		if term.Term == want && term.Frequency > 0 {
			return true
		}
	}
	return false
}
