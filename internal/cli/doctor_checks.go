package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Lokee86/lexicon/internal/objectstore"
	"github.com/Lokee86/lexicon/internal/state"
)

var doctorRuntimeExecutables = map[string][][]string{
	"go":         {{"go"}},
	"gdscript":   {{"go"}},
	"python":     {{"python", "python3"}},
	"ruby":       {{"ruby"}},
	"rust":       {{"cargo"}},
	"typescript": {{"node"}, {"npm", "npm.cmd"}},
}

func verifyStateRepository(path string) error {
	repository, err := state.Open(path)
	if err != nil {
		return err
	}
	_, err = repository.Head()
	return err
}

func verifySnapshot(store objectstore.Store) (objectstore.Manifest, error) {
	_, manifest, err := doctorCurrentSnapshot(store)
	if err != nil {
		return manifest, err
	}
	ids := referencedObjectIDs(manifest)
	failures := make([]error, 0)
	for _, id := range ids {
		if _, err := doctorLoadObject(store, id); err != nil {
			failures = append(failures, fmt.Errorf("snapshot object %s: %w", id, err))
		}
	}
	return manifest, errors.Join(failures...)
}

func referencedObjectIDs(manifest objectstore.Manifest) []string {
	ids := make(map[string]struct{})
	for _, language := range manifest.Languages {
		if language.SharedObjectID != "" {
			ids[language.SharedObjectID] = struct{}{}
		}
		for _, file := range language.Files {
			ids[file.ObjectID] = struct{}{}
		}
	}
	result := make([]string, 0, len(ids))
	for id := range ids {
		result = append(result, id)
	}
	sort.Strings(result)
	return result
}

func directoryExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func checkAdapterDirectory(root, language string) error {
	path := filepath.Join(root, language)
	if !directoryExists(path) {
		return fmt.Errorf("adapter directory is missing: %s", path)
	}
	return nil
}

func manifestLanguages(manifest objectstore.Manifest) []string {
	set := make(map[string]struct{}, len(manifest.Languages))
	for _, language := range manifest.Languages {
		if language.Language != "" {
			set[language.Language] = struct{}{}
		}
	}
	languages := make([]string, 0, len(set))
	for language := range set {
		languages = append(languages, language)
	}
	sort.Strings(languages)
	return languages
}

func checkRuntime(language string) error {
	requirements, ok := doctorRuntimeExecutables[language]
	if !ok {
		return fmt.Errorf("no runtime definition for detected language %q", language)
	}
	for _, candidates := range requirements {
		found := false
		for _, candidate := range candidates {
			if _, err := doctorLookPath(candidate); err == nil {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("required executable not found: %s", strings.Join(candidates, " or "))
		}
	}
	return nil
}

func checkConsumerCommand(command string) error {
	if _, err := doctorLookPath(command); err != nil {
		return fmt.Errorf("required executable not found: %s", command)
	}
	return nil
}

func consumerDefinitionPaths(repository string) ([]string, error) {
	directory := filepath.Join(repository, ".lexicon", "consumers")
	entries, err := os.ReadDir(directory)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read Lexicon consumers: %w", err)
	}
	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		paths = append(paths, filepath.Join(directory, entry.Name()))
	}
	sort.Strings(paths)
	return paths, nil
}
