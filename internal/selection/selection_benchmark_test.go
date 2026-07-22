package selection

import (
	"fmt"
	"testing"

	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

func BenchmarkCurateTwoHundredCandidates(b *testing.B) {
	files := make([]index.FileRecord, 0, 100)
	candidates := make([]retrieve.Candidate, 0, 200)
	for fileNumber := 0; fileNumber < 100; fileNumber++ {
		path := fmt.Sprintf("internal/subsystem%02d/file%03d.go", fileNumber%10, fileNumber)
		chunks := make([]index.Chunk, 0, 4)
		for chunkNumber := 0; chunkNumber < 4; chunkNumber++ {
			chunk := index.Chunk{
				ID: fmt.Sprintf("%03d-%d", fileNumber, chunkNumber), Path: path,
				StartLine: chunkNumber*20 + 1, EndLine: chunkNumber*20 + 20,
				TokenCount: 80, Text: fmt.Sprintf("chunk %d for file %d", chunkNumber, fileNumber),
			}
			chunks = append(chunks, chunk)
			if chunkNumber == 1 || chunkNumber == 2 {
				candidates = append(candidates, retrieve.Candidate{
					Chunk: chunk, Score: float64(1000 - len(candidates)),
					Source: "vector", Rank: len(candidates) + 1,
					Reasons: []string{"benchmark candidate"},
				})
			}
		}
		files = append(files, index.FileRecord{Path: path, Chunks: chunks})
	}
	snapshot := index.Snapshot{Version: index.FormatVersion, Files: files}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		curated := Curate(snapshot, candidates)
		if len(curated) < len(candidates) {
			b.Fatalf("curation unexpectedly discarded unique candidates: %d", len(curated))
		}
	}
}
