package scope

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	lexfiles "github.com/Lokee86/lexicon/internal/files"
)

func Build(sourceRoot, temporaryRoot, language string, contextFiles []string) (string, error) {
	repository := filepath.Join(temporaryRoot, language, "source")
	if err := os.RemoveAll(filepath.Dir(repository)); err != nil {
		return "", err
	}
	selected := make(map[string]struct{}, len(contextFiles))
	for _, path := range contextFiles {
		selected[filepath.ToSlash(path)] = struct{}{}
	}
	if err := expandSemanticUnits(sourceRoot, language, selected); err != nil {
		return "", err
	}
	err := filepath.WalkDir(sourceRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(sourceRoot, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		_, chosen := selected[relative]
		if !chosen && !languageConfig(language, relative) {
			return nil
		}
		return copyFile(path, filepath.Join(repository, filepath.FromSlash(relative)))
	})
	if err != nil {
		return "", fmt.Errorf("build %s analysis scope: %w", language, err)
	}
	return repository, nil
}

func expandSemanticUnits(root, language string, selected map[string]struct{}) error {
	switch language {
	case "go":
		for path := range cloneSet(selected) {
			if strings.EqualFold(filepath.Ext(path), ".go") {
				if err := includeDirectorySources(root, filepath.Dir(filepath.FromSlash(path)), ".go", selected); err != nil {
					return err
				}
			}
		}
	case "rust":
		for path := range cloneSet(selected) {
			if strings.EqualFold(filepath.Ext(path), ".rs") {
				crate := nearestConfig(root, filepath.Dir(filepath.FromSlash(path)), "Cargo.toml")
				if err := includeTreeSources(root, crate, ".rs", selected); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func includeDirectorySources(root, relativeDir, extension string, selected map[string]struct{}) error {
	entries, err := os.ReadDir(filepath.Join(root, relativeDir))
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), extension) {
			selected[filepath.ToSlash(filepath.Join(relativeDir, entry.Name()))] = struct{}{}
		}
	}
	return nil
}

func includeTreeSources(root, relativeRoot, extension string, selected map[string]struct{}) error {
	return filepath.WalkDir(filepath.Join(root, relativeRoot), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(entry.Name()), extension) {
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			selected[filepath.ToSlash(relative)] = struct{}{}
		}
		return nil
	})
}

func nearestConfig(root, relativeDir, name string) string {
	for current := filepath.Clean(relativeDir); ; current = filepath.Dir(current) {
		if _, err := os.Stat(filepath.Join(root, current, name)); err == nil {
			return current
		}
		if current == "." || current == "" {
			return "."
		}
	}
}

func languageConfig(language, path string) bool {
	if sourceExtension(language, filepath.Ext(path)) {
		return false
	}
	for _, candidate := range lexfiles.Languages(path) {
		if candidate == language {
			return true
		}
	}
	return false
}

func sourceExtension(language, extension string) bool {
	extension = strings.ToLower(extension)
	switch language {
	case "go":
		return extension == ".go"
	case "python":
		return extension == ".py"
	case "ruby":
		return extension == ".rb" || extension == ".gemspec"
	case "gdscript":
		return extension == ".gd"
	case "rust":
		return extension == ".rs"
	case "typescript":
		return extension == ".ts" || extension == ".tsx" || extension == ".mts" || extension == ".cts"
	}
	return false
}

func copyFile(source, destination string) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.Create(destination)
	if err != nil {
		return err
	}
	if _, err := io.Copy(output, input); err != nil {
		_ = output.Close()
		return err
	}
	return output.Close()
}

func cloneSet(source map[string]struct{}) map[string]struct{} {
	result := make(map[string]struct{}, len(source))
	for key := range source {
		result[key] = struct{}{}
	}
	return result
}
