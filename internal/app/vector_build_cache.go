package app

import (
	"os"
	"sort"

	"github.com/Lokee86/grimoire/internal/embedding"
	"github.com/Lokee86/grimoire/internal/index"
)

func reusableVectorSnapshot(paths vectorStatePaths, preparedIdentity string, count int) (vectorSnapshotManifest, os.FileInfo, bool) {
	manifest, err := readVectorSnapshotManifest(paths.Manifest)
	if err != nil || manifest.PreparedIdentity != preparedIdentity || manifest.Model != embedding.Identity() ||
		manifest.Dimensions != embedding.Dimensions || (count > 0 && manifest.Count != count) {
		return vectorSnapshotManifest{}, nil, false
	}
	info, err := os.Stat(paths.Snapshot)
	if err != nil || !info.Mode().IsRegular() || info.Size() == 0 {
		return vectorSnapshotManifest{}, nil, false
	}
	return manifest, info, true
}

func reusableVectorSources(paths vectorStatePaths) map[string]struct{} {
	manifest, err := readVectorSnapshotManifest(paths.Manifest)
	if err != nil || manifest.Model != embedding.Identity() || manifest.Dimensions != embedding.Dimensions || len(manifest.Sources) == 0 {
		return nil
	}
	info, err := os.Stat(paths.Snapshot)
	if err != nil || !info.Mode().IsRegular() || info.Size() == 0 {
		return nil
	}
	sources := make(map[string]struct{}, len(manifest.Sources))
	for _, source := range manifest.Sources {
		if source != "" {
			sources[source] = struct{}{}
		}
	}
	return sources
}

func vectorEntries(chunks []index.Chunk) ([]vectorChunk, []vectorChunk, map[string]int) {
	all := make([]vectorChunk, 0, len(chunks))
	unique := make([]vectorChunk, 0, len(chunks))
	counts := make(map[string]int, len(chunks))
	seen := make(map[string]struct{}, len(chunks))
	for _, chunk := range chunks {
		entry := vectorChunk{Chunk: chunk, Source: vectorSource(chunk.Text)}
		all = append(all, entry)
		counts[entry.Source]++
		if _, exists := seen[entry.Source]; exists {
			continue
		}
		seen[entry.Source] = struct{}{}
		unique = append(unique, entry)
	}
	return all, unique, counts
}

func vectorSources(entries []vectorChunk) []string {
	sources := make([]string, 0, len(entries))
	for _, entry := range entries {
		sources = append(sources, entry.Source)
	}
	sort.Strings(sources)
	return sources
}
