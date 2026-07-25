package repostate

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func fixtureRunner(t *testing.T, root string, mu *sync.Mutex, calls *[]string, useNewID bool) CommandRunner {
	t.Helper()
	id := testID('a')
	if useNewID {
		id = testID('c')
	}
	return fixtureRunnerWithLexicon(t, root, mu, calls, id)
}

func fixtureRunnerWithLexicon(t *testing.T, root string, mu *sync.Mutex, calls *[]string, lexiconID string) CommandRunner {
	t.Helper()
	return func(_ context.Context, command string, arguments ...string) error {
		mu.Lock()
		*calls = append(*calls, command+":"+arguments[0])
		mu.Unlock()
		switch command + ":" + arguments[0] {
		case "lexicon:init", "lexicon:scan", "lexicon:rebuild":
			writeLexicon(t, root, lexiconID)
		case "arcana:sync":
			writeArcana(t, root, currentLexicon(t, root))
		case "grimoire:index":
			if err := os.RemoveAll(filepath.Join(root, ".grimoire")); err != nil {
				t.Fatal(err)
			}
			writeGrimoire(t, root, currentLexicon(t, root))
		default:
			t.Fatalf("unexpected command %s %v", command, arguments)
		}
		return nil
	}
}

func currentLexicon(t *testing.T, root string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".lexicon", "CURRENT"))
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(data))
}
