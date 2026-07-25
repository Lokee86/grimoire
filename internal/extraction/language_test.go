package extraction

import (
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/index"
)

func TestLanguageDiscovererPrefersNamedMethodOverContainingType(t *testing.T) {
	content := strings.Join([]string{
		"package service",
		"",
		"type Service struct {}",
		"",
		"func (s *Service) Run() {",
		"    s.prepare()",
		"}",
	}, "\n")
	spans, err := NewLanguageDiscoverer().Discover(DiscoveryRequest{
		Chunk: index.Chunk{Path: "service.go", Text: content},
		Query: "where is Service.Run implemented",
		Terms: queryTerms("where is Service.Run implemented"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) == 0 || spans[0].StartLine != 4 || spans[0].EndLine != 6 {
		t.Fatalf("unexpected language span: %+v", spans)
	}
}

func TestLanguageDiscovererReturnsSeparatedRelevantFunctions(t *testing.T) {
	content := strings.Join([]string{
		"package damage",
		"",
		"func ResolveDamage() {",
		"}",
		"",
		"func unrelated() {",
		"}",
		"",
		"func ApplyShieldGate() {",
		"}",
	}, "\n")
	spans, err := NewLanguageDiscoverer().Discover(DiscoveryRequest{
		Chunk: index.Chunk{Path: "damage.go", Text: content},
		Query: "ResolveDamage ApplyShieldGate",
		Terms: queryTerms("ResolveDamage ApplyShieldGate"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 2 {
		t.Fatalf("unexpected separated spans: %+v", spans)
	}
	starts := map[int]bool{spans[0].StartLine: true, spans[1].StartLine: true}
	if !starts[2] || !starts[8] {
		t.Fatalf("unexpected separated spans: %+v", spans)
	}
}
