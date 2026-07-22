package retrieve

import (
	"fmt"
	"testing"

	"github.com/Lokee86/grimoire/internal/index"
)

func BenchmarkSearchTenThousandChunks(b *testing.B) {
	files := make([]index.FileRecord, 0, 1000)
	for fileNumber := 0; fileNumber < 1000; fileNumber++ {
		path := fmt.Sprintf("internal/package%04d/file.go", fileNumber)
		chunks := make([]index.Chunk, 0, 10)
		for chunkNumber := 0; chunkNumber < 10; chunkNumber++ {
			text := fmt.Sprintf("package package%04d\nfunc Operation%02d() { processRepositoryState() }", fileNumber, chunkNumber)
			chunks = append(chunks, index.Chunk{
				ID:              fmt.Sprintf("%04d-%02d", fileNumber, chunkNumber),
				Path:            path,
				StartLine:       chunkNumber*10 + 1,
				EndLine:         chunkNumber*10 + 2,
				EstimatedTokens: 20,
				Text:            text,
			})
		}
		files = append(files, index.FileRecord{Path: path, Chunks: chunks})
	}
	snapshot := index.Snapshot{Version: index.FormatVersion, Files: files}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		results := Search(snapshot, "repository state operation", 200)
		if len(results) != 200 {
			b.Fatalf("expected 200 results, got %d", len(results))
		}
	}
}
