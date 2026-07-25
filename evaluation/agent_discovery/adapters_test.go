package agentdiscovery

import (
	"path/filepath"
	"testing"
)

func TestBuiltInAdaptersImportFixtures(t *testing.T) {
	for _, item := range []struct{ name, file string }{
		{"progressive-jsonl", "progressive.jsonl"}, {"raw", "raw.jsonl"}, {"grimoire-context", "grimoire-context.json"},
	} {
		adapter, ok := AdapterFor(item.name)
		if !ok {
			t.Fatalf("missing %s adapter", item.name)
		}
		transcripts, err := adapter(filepath.Join("testdata", item.file))
		if err != nil || len(transcripts) != 1 || len(transcripts[0].Events) == 0 {
			t.Fatalf("%s import = %+v, %v", item.name, transcripts, err)
		}
	}
}

func TestCBMAdapterCanBeRegisteredExternally(t *testing.T) {
	RegisterAdapter("cbm-fixture", func(string) ([]Transcript, error) { return []Transcript{{Adapter: "cbm", CaseID: "case"}}, nil })
	adapter, ok := AdapterFor("cbm-fixture")
	if !ok {
		t.Fatal("registered CBM adapter unavailable")
	}
	transcripts, err := adapter("ignored")
	if err != nil || transcripts[0].Adapter != "cbm" {
		t.Fatalf("adapter = %+v, %v", transcripts, err)
	}
}
