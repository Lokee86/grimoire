package spandiscovery

import (
	"reflect"
	"testing"
)

func TestDiscoverPythonUsesIndentationBoundaries(t *testing.T) {
	content := "class Service:\n    def first(self):\n        if self.ready:\n            return 1\n\n    async def second(self):\n        return 2\n\ndef helper():\n    return 3"
	got := Discover("service.py", content)
	want := []Span{
		{StartLine: 1, EndLine: 8, Kind: KindType, Name: "Service", Language: "python"},
		{StartLine: 2, EndLine: 5, Kind: KindMethod, Name: "first", Language: "python"},
		{StartLine: 6, EndLine: 8, Kind: KindMethod, Name: "second", Language: "python"},
		{StartLine: 9, EndLine: 10, Kind: KindFunction, Name: "helper", Language: "python"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("spans = %#v, want %#v", got, want)
	}
}

func TestDiscoverRubyIncludesMatchingEndLines(t *testing.T) {
	content := "class Service\n  def run\n    if ready?\n      work\n    end\n  end\nend\n\ndef helper\n  true\nend"
	got := Discover("service.rb", content)
	want := []Span{
		{StartLine: 1, EndLine: 7, Kind: KindType, Name: "Service", Language: "ruby"},
		{StartLine: 2, EndLine: 6, Kind: KindMethod, Name: "run", Language: "ruby"},
		{StartLine: 9, EndLine: 11, Kind: KindFunction, Name: "helper", Language: "ruby"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("spans = %#v, want %#v", got, want)
	}
}

func TestDiscoverGDScriptFindsClassMethods(t *testing.T) {
	content := "class Pilot:\n\tfunc steer():\n\t\tpass\n\nfunc launch():\n\tpass"
	got := Discover("pilot.gd", content)
	want := []Span{
		{StartLine: 1, EndLine: 4, Kind: KindType, Name: "Pilot", Language: "gdscript"},
		{StartLine: 2, EndLine: 4, Kind: KindMethod, Name: "steer", Language: "gdscript"},
		{StartLine: 5, EndLine: 6, Kind: KindFunction, Name: "launch", Language: "gdscript"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("spans = %#v, want %#v", got, want)
	}
}
