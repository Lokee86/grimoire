package spandiscovery

import (
	"reflect"
	"testing"
)

func TestSmallestContainingPrefersNarrowestRange(t *testing.T) {
	spans := []Span{
		{StartLine: 1, EndLine: 20, Kind: KindType, Name: "Service"},
		{StartLine: 5, EndLine: 10, Kind: KindMethod, Name: "Run"},
	}
	got, ok := SmallestContaining(spans, 7, 8)
	if !ok || got.Name != "Run" {
		t.Fatalf("span = %#v, ok = %v", got, ok)
	}
}

func TestOverlappingOrdersNarrowSpansFirst(t *testing.T) {
	spans := []Span{
		{StartLine: 1, EndLine: 20, Kind: KindType, Name: "Service"},
		{StartLine: 5, EndLine: 10, Kind: KindMethod, Name: "Run"},
		{StartLine: 30, EndLine: 40, Kind: KindFunction, Name: "Other"},
	}
	got := Overlapping(spans, 8, 12)
	want := []Span{spans[1], spans[0]}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("spans = %#v, want %#v", got, want)
	}
}

func TestDiscoverUnsupportedFileReturnsNoSpans(t *testing.T) {
	if got := Discover("notes.txt", "plain text"); got != nil {
		t.Fatalf("spans = %#v, want nil", got)
	}
}
