package knowledgevector

import (
	"context"
	"fmt"
	"os"

	"github.com/Lokee86/grimoire/internal/embedding"
	"github.com/Lokee86/grimoire/internal/knowledge"
	"github.com/Lokee86/grimoire/internal/vectorstore"
)

type Ranker struct {
	State      string
	Index      knowledge.Index
	Endpoint   string
	EnginePath string
}

func (ranker Ranker) Rank(ctx context.Context, query string, sections []knowledge.Section) (map[string]float64, error) {
	paths := ResolvePaths(ranker.State)
	manifest, err := ReadManifest(paths.Manifest)
	if err != nil {
		return nil, err
	}
	library, err := vectorstore.Load(ranker.EnginePath)
	if err != nil {
		return nil, err
	}
	defer library.Close()
	engine, err := library.OpenSnapshot(paths.Snapshot)
	if err != nil {
		return nil, err
	}
	defer engine.Close()
	info, err := engine.Info()
	if err != nil {
		return nil, err
	}
	if err := validateManifest(manifest, ranker.Index, info); err != nil {
		return nil, err
	}
	queryVector, err := embedding.NewClient(ranker.Endpoint).EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}
	hits, err := engine.Search(queryVector, info.Count)
	if err != nil {
		return nil, err
	}
	allowed := make(map[string]bool, len(sections))
	for _, section := range sections {
		allowed[section.ID] = true
	}
	scores := make(map[string]float64, len(sections))
	for _, hit := range hits {
		if allowed[hit.ID] {
			scores[hit.ID] = float64(hit.Score)
		}
	}
	return scores, nil
}

func Inspect(state string, index knowledge.Index, enginePath string) Info {
	paths := ResolvePaths(state)
	result := Info{State: state, Snapshot: paths.Snapshot, ExpectedIdentity: IndexIdentity(index)}
	manifest, err := ReadManifest(paths.Manifest)
	if err != nil {
		if !os.IsNotExist(err) {
			result.Error = err.Error()
		}
		return result
	}
	result.Available = true
	result.KnowledgeIdentity = manifest.KnowledgeIdentity
	result.SnapshotIdentity = manifest.SnapshotIdentity
	result.Model = manifest.Model
	result.Dimensions = manifest.Dimensions
	result.Count = manifest.Count
	if fileInfo, statErr := os.Stat(paths.Snapshot); statErr == nil {
		result.SnapshotBytes = fileInfo.Size()
	} else {
		result.Error = statErr.Error()
		return result
	}
	library, err := vectorstore.Load(enginePath)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer library.Close()
	engine, err := library.OpenSnapshot(paths.Snapshot)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer engine.Close()
	engineInfo, err := engine.Info()
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if err := validateManifest(manifest, index, engineInfo); err != nil {
		result.Error = err.Error()
		return result
	}
	result.Current = true
	return result
}

func (ranker Ranker) String() string {
	return fmt.Sprintf("knowledge vectors at %s", ResolvePaths(ranker.State).Snapshot)
}
