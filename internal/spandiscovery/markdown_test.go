package spandiscovery

import (
	"reflect"
	"testing"
)

func TestDiscoverMarkdownBuildsNestedSectionsAndIgnoresFences(t *testing.T) {
	content := "# Top\nintro\n## Child\ntext\n```go\n# not a heading\n```\n## Next\nmore\n# Tail"
	got := Discover("docs/guide.md", content)
	want := []Span{
		{StartLine: 1, EndLine: 9, Kind: KindSection, Name: "Top", Language: "markdown"},
		{StartLine: 3, EndLine: 7, Kind: KindSection, Name: "Child", Language: "markdown"},
		{StartLine: 8, EndLine: 9, Kind: KindSection, Name: "Next", Language: "markdown"},
		{StartLine: 10, EndLine: 10, Kind: KindSection, Name: "Tail", Language: "markdown"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("spans = %#v, want %#v", got, want)
	}
}

func TestDiscoverTOMLBuildsTableSections(t *testing.T) {
	content := "title = \"demo\"\n\n[server]\nport = 8080\n\n[[workers]]\nname = \"one\""
	got := Discover("config/app.toml", content)
	want := []Span{
		{StartLine: 3, EndLine: 5, Kind: KindSection, Name: "server", Language: "toml"},
		{StartLine: 6, EndLine: 7, Kind: KindSection, Name: "workers", Language: "toml"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("spans = %#v, want %#v", got, want)
	}
}
