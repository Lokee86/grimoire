package spandiscovery

import (
	"reflect"
	"testing"
)

func TestDiscoverGoIgnoresBracesInStringsAndComments(t *testing.T) {
	content := "package service\n\ntype Service struct {\n\tName string\n}\n\nfunc (s *Service) Run() {\n\tpayload := \"}\"\n\t// { ignored\n\tif payload != \"\" {\n\t}\n}\n\nfunc helper() {\n}"
	got := Discover("service.go", content)
	want := []Span{
		{StartLine: 3, EndLine: 5, Kind: KindType, Name: "Service", Language: "go"},
		{StartLine: 7, EndLine: 12, Kind: KindMethod, Name: "Run", Language: "go"},
		{StartLine: 14, EndLine: 15, Kind: KindFunction, Name: "helper", Language: "go"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("spans = %#v, want %#v", got, want)
	}
}

func TestDiscoverRustClassifiesFunctionsInsideImplAsMethods(t *testing.T) {
	content := "impl Service {\n    pub fn run(&self) {\n        if self.ready {\n        }\n    }\n}\n\npub fn helper() {\n}"
	got := Discover("service.rs", content)
	want := []Span{
		{StartLine: 1, EndLine: 6, Kind: KindType, Name: "Service", Language: "rust"},
		{StartLine: 2, EndLine: 5, Kind: KindMethod, Name: "run", Language: "rust"},
		{StartLine: 8, EndLine: 9, Kind: KindFunction, Name: "helper", Language: "rust"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("spans = %#v, want %#v", got, want)
	}
}

func TestDiscoverRustFindsGenericTraitImpls(t *testing.T) {
	content := "impl<T> Render for Service<T> {\n    fn draw(&self) {\n    }\n}"
	got := Discover("render.rs", content)
	want := []Span{
		{StartLine: 1, EndLine: 4, Kind: KindType, Name: "Service", Language: "rust"},
		{StartLine: 2, EndLine: 3, Kind: KindMethod, Name: "draw", Language: "rust"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("spans = %#v, want %#v", got, want)
	}
}

func TestDiscoverCPPFindsQualifiedDefinitions(t *testing.T) {
	content := "class Service {\npublic:\n    void run() {\n    }\n};\n\nvoid Service::stop() {\n}"
	got := Discover("service.cpp", content)
	want := []Span{
		{StartLine: 1, EndLine: 5, Kind: KindType, Name: "Service", Language: "cpp"},
		{StartLine: 3, EndLine: 4, Kind: KindMethod, Name: "run", Language: "cpp"},
		{StartLine: 7, EndLine: 8, Kind: KindFunction, Name: "stop", Language: "cpp"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("spans = %#v, want %#v", got, want)
	}
}

func TestDiscoverTypeScriptFindsClassMethodsAndArrowFunctions(t *testing.T) {
	content := "export class Worker {\n  async run(value: string) {\n    return value\n  }\n}\n\nexport const createWorker = () => {\n  return new Worker()\n}"
	got := Discover("worker.ts", content)
	want := []Span{
		{StartLine: 1, EndLine: 5, Kind: KindType, Name: "Worker", Language: "typescript"},
		{StartLine: 2, EndLine: 4, Kind: KindMethod, Name: "run", Language: "typescript"},
		{StartLine: 7, EndLine: 9, Kind: KindFunction, Name: "createWorker", Language: "typescript"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("spans = %#v, want %#v", got, want)
	}
}
