package diffcontext

import (
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

func TestEffectiveQueryDefaultsOnlyWhenEmpty(t *testing.T) {
	if EffectiveQuery("  ") != DefaultQuery {
		t.Fatalf("empty query did not use default: %q", EffectiveQuery("  "))
	}
	if EffectiveQuery("check auth") != "check auth" {
		t.Fatalf("explicit query changed: %q", EffectiveQuery("check auth"))
	}
}

func TestRetrievalQueryAddsBoundedUniqueAnchors(t *testing.T) {
	query := RetrievalQuery("review auth", []Change{
		{Path: "internal/auth/login.go", StartLine: 10, EndLine: 12, Summary: "func Login(user User) error"},
		{Path: "internal/auth/login.go", StartLine: 20, EndLine: 21, Summary: "func Login(user User) error"},
	}, []retrieve.Candidate{{Chunk: index.Chunk{
		Path: "internal/auth/login.go", Text: "// comment\nfunc Login(user User) error {",
	}}})
	for _, required := range []string{"review auth", "internal/auth/login.go", "func Login(user User) error"} {
		if !strings.Contains(query, required) {
			t.Fatalf("retrieval query %q does not contain %q", query, required)
		}
	}
	if strings.Count(query, "internal/auth/login.go") != 1 {
		t.Fatalf("anchors were not deduplicated: %q", query)
	}
}
