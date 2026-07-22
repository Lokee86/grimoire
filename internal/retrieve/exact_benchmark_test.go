package retrieve

import (
	"fmt"
	"testing"

	"github.com/Lokee86/grimoire/internal/index"
)

func BenchmarkExactTenThousandChunks(b *testing.B) {
	files := make([]index.FileRecord, 0, 1000)
	for fileNumber := 0; fileNumber < 1000; fileNumber++ {
		path := fmt.Sprintf("internal/package%04d/config.go", fileNumber)
		chunks := make([]index.Chunk, 0, 10)
		for chunkNumber := 0; chunkNumber < 10; chunkNumber++ {
			identifier := fmt.Sprintf("ResolveOperation%04d_%02d", fileNumber, chunkNumber)
			text := fmt.Sprintf("package package%04d\nconst ErrorCode = \"ERR_PACKAGE_%04d\"\nfunc %s() {}", fileNumber, fileNumber, identifier)
			chunks = append(chunks, index.Chunk{
				ID: fmt.Sprintf("%04d-%02d", fileNumber, chunkNumber), Path: path,
				StartLine: chunkNumber*10 + 1, EndLine: chunkNumber*10 + 3,
				TokenCount: 24, Text: text,
			})
		}
		files = append(files, index.FileRecord{Path: path, Chunks: chunks})
	}
	snapshot := index.Snapshot{Version: index.FormatVersion, Files: files}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		results := Exact(snapshot, "Where is ResolveOperation0420_07?", 20)
		if len(results) == 0 || results[0].Chunk.ID != "0420-07" {
			b.Fatalf("unexpected exact results: %+v", results)
		}
	}
}
