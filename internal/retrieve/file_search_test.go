package retrieve

import (
	"testing"

	"github.com/Lokee86/grimoire/internal/index"
)

func TestSearchFilesRanksWholeFileContent(t *testing.T) {
	snapshot := index.Snapshot{Files: []index.FileRecord{
		{Path: "internal/recovery.go", Chunks: []index.Chunk{
			{ID: "recovery-a", Path: "internal/recovery.go", StartLine: 1, EndLine: 20, Text: "package recovery\nfunc reopenChannel() {}"},
			{ID: "recovery-b", Path: "internal/recovery.go", StartLine: 21, EndLine: 40, Text: "func restoreRealtimeTransport() {}"},
		}},
		{Path: "internal/config.go", Chunks: []index.Chunk{
			{ID: "config", Path: "internal/config.go", StartLine: 1, EndLine: 20, Text: "package config\nfunc loadSettings() {}"},
		}},
	}}

	results := SearchFilesManyWithConfig(
		snapshot, []string{"realtime channel recovery"}, 4, DefaultConfig(),
	)
	if len(results) != 1 || len(results[0]) == 0 {
		t.Fatalf("missing file results: %+v", results)
	}
	if results[0][0].Path != "internal/recovery.go" || results[0][0].Rank != 1 {
		t.Fatalf("unexpected first file: %+v", results[0][0])
	}
}
