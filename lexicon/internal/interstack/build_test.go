package interstack

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/lexicon/internal/objectstore"
)

func TestBuildStoresAndExportsSyntheticInterstackLibrary(t *testing.T) {
	sourceRoot := t.TempDir()
	writeFixture(t, sourceRoot, "client/api.gd", `func auth_me_path():
	return "%s/api/auth/me" % API_BASE
`)
	writeFixture(t, sourceRoot, "services/api-server/config/routes.rb", `Rails.application.routes.draw do
  namespace :api do
    namespace :auth do
      get "me", to: "me#show"
    end
  end
end
`)
	writeFixture(t, sourceRoot, "services/api-server/app/controllers/api/auth/me_controller.rb", `module Api
  module Auth
    class MeController
      def show
      end
    end
  end
end
`)

	store := objectstore.Store{Root: filepath.Join(t.TempDir(), "store")}
	manifest := objectstore.Manifest{Version: objectstore.SnapshotVersion, StateCommit: testID('9')}
	gdscriptOutput := writeAnalysisFixture(t, "gdscript", "space-rocks", []Node{
		fileNode(testID('a'), "client/api.gd"),
		callableNode(testID('b'), "method", "auth_me_path", "client/api.gd", 1),
	})
	rubyOutput := writeAnalysisFixture(t, "ruby", "space-rocks", []Node{
		fileNode(testID('c'), "services/api-server/config/routes.rb"),
		fileNode(testID('d'), "services/api-server/app/controllers/api/auth/me_controller.rb"),
		{
			ID: testID('e'), Kind: "method", Name: "show",
			Path:          "services/api-server/app/controllers/api/auth/me_controller.rb",
			QualifiedName: "Api::Auth::MeController#show",
			Span:          &Span{Path: "services/api-server/app/controllers/api/auth/me_controller.rb", StartLine: 4, StartColumn: 11, EndLine: 4, EndColumn: 15},
		},
	})
	for _, input := range []struct {
		language string
		path     string
	}{{"gdscript", gdscriptOutput}, {"ruby", rubyOutput}} {
		entry, err := store.IngestLanguage(input.path, sourceRoot, input.language, testID('8'))
		if err != nil {
			t.Fatalf("ingest %s: %v", input.language, err)
		}
		manifest = manifest.WithLanguage(entry)
	}

	analysis, summary, err := Build(sourceRoot, store, manifest, filepath.Join(t.TempDir(), "interstack.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if summary.HTTPContracts != 1 || summary.HTTPLinks != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	entry, err := store.BuildSharedLanguage(analysis, testID('8'), AdapterVersion)
	if err != nil {
		t.Fatal(err)
	}
	if entry.Language != Language || entry.SharedObjectID == "" || len(entry.Files) != 0 {
		t.Fatalf("unexpected synthetic language entry: %+v", entry)
	}
	data, err := store.ExportLanguage(entry)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, expected := range []string{`"language":"interstack"`, `"relation":"calls-endpoint"`, `"relation":"handled-by"`} {
		if !strings.Contains(text, expected) {
			t.Fatalf("export missing %s:\n%s", expected, text)
		}
	}
}

func writeAnalysisFixture(t *testing.T, language, repository string, nodes []Node) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), language+".jsonl")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(map[string]any{
		"adapter_version": "test", "language": language, "mode": "full",
		"record": "lexicon", "repository": repository, "schema_version": 1,
	}); err != nil {
		t.Fatal(err)
	}
	for _, node := range nodes {
		record := map[string]any{
			"id": node.ID, "kind": node.Kind, "name": node.Name,
			"path": node.Path, "qualified_name": node.QualifiedName, "record": "node",
		}
		if node.Span != nil {
			record["span"] = node.Span
		}
		if err := encoder.Encode(record); err != nil {
			t.Fatal(err)
		}
	}
	return path
}
