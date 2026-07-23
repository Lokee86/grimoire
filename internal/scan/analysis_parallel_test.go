package scan

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/Lokee86/lexicon/internal/adapters"
	"github.com/Lokee86/lexicon/internal/config"
	"github.com/Lokee86/lexicon/internal/objectstore"
)

type blockingAnalyzer struct {
	mu         sync.Mutex
	active     int
	maximum    int
	twoStarted chan struct{}
	release    chan struct{}
	once       sync.Once
}

func (analyzer *blockingAnalyzer) Run(_ context.Context, request adapters.Request) error {
	analyzer.mu.Lock()
	analyzer.active++
	if analyzer.active > analyzer.maximum {
		analyzer.maximum = analyzer.active
	}
	if analyzer.active >= 2 {
		analyzer.once.Do(func() { close(analyzer.twoStarted) })
	}
	analyzer.mu.Unlock()

	<-analyzer.release
	header, err := json.Marshal(map[string]any{
		"adapter_version": "test", "language": request.Language, "record": "lexicon",
		"repository": "test", "schema_version": 1,
	})
	if err != nil {
		return err
	}
	if err := os.WriteFile(request.Output, append(header, '\n'), 0o644); err != nil {
		return err
	}

	analyzer.mu.Lock()
	analyzer.active--
	analyzer.mu.Unlock()
	return nil
}

func TestAnalyzePlansRunsIndependentLanguagesConcurrently(t *testing.T) {
	if runtime.GOMAXPROCS(0) < 2 {
		t.Skip("requires at least two scheduler slots")
	}
	repository := t.TempDir()
	stateRoot := t.TempDir()
	sourceRoot := filepath.Join(stateRoot, "source")
	if err := os.MkdirAll(sourceRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	for relative, contents := range map[string]string{
		"main.py": "value = 1\n",
		"main.rb": "value = 1\n",
	} {
		if err := os.WriteFile(filepath.Join(sourceRoot, relative), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	analyzer := &blockingAnalyzer{twoStarted: make(chan struct{}), release: make(chan struct{})}
	output := &bytes.Buffer{}
	scanner := &Scanner{
		Repository: repository, StateRoot: stateRoot, Analyzer: analyzer,
		Store: objectstore.Store{Root: config.StateRoot(repository)}, Output: output,
	}

	result := make(chan error, 1)
	go func() {
		_, err := scanner.analyzePlans(context.Background(), objectstore.Manifest{Version: objectstore.SnapshotVersion}, []analysisPlan{
			{Language: "python", Full: true},
			{Language: "ruby", Full: true},
		})
		result <- err
	}()

	select {
	case <-analyzer.twoStarted:
		close(analyzer.release)
	case <-time.After(5 * time.Second):
		close(analyzer.release)
		t.Fatal("language adapters did not overlap")
	}
	if err := <-result; err != nil {
		t.Fatal(err)
	}
	analyzer.mu.Lock()
	maximum := analyzer.maximum
	analyzer.mu.Unlock()
	if maximum < 2 {
		t.Fatalf("maximum concurrent adapters = %d, want at least 2", maximum)
	}
}
