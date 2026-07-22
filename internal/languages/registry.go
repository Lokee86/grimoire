package languages

import (
	"path/filepath"
	"sort"
	"strings"
)

type Definition struct {
	Language    string
	Directory   string
	Extensions  []string
	ConfigFiles []string
}

var definitions = []Definition{
	{Language: "gdscript", Directory: "gdscript", Extensions: []string{".gd"}, ConfigFiles: []string{"project.godot"}},
	{Language: "go", Directory: "go", Extensions: []string{".go"}, ConfigFiles: []string{"go.mod", "go.sum"}},
	{Language: "python", Directory: "python", Extensions: []string{".py"}, ConfigFiles: []string{"pyproject.toml", "setup.cfg", "requirements.txt"}},
	{Language: "ruby", Directory: "ruby", Extensions: []string{".rb", ".gemspec"}, ConfigFiles: []string{"Gemfile", "Gemfile.lock"}},
	{Language: "rust", Directory: "rust", Extensions: []string{".rs"}, ConfigFiles: []string{"Cargo.toml", "Cargo.lock"}},
	{Language: "typescript", Directory: "typescript", Extensions: []string{".ts", ".tsx", ".mts", ".cts", ".js", ".jsx", ".mjs", ".cjs"}, ConfigFiles: []string{"package.json", "package-lock.json", "tsconfig.json", "jsconfig.json"}},
}

func Definitions() []Definition {
	result := make([]Definition, len(definitions))
	for index, definition := range definitions {
		result[index] = clone(definition)
	}
	return result
}

func Lookup(language string) (Definition, bool) {
	for _, definition := range definitions {
		if definition.Language == language {
			return clone(definition), true
		}
	}
	return Definition{}, false
}

func Supported() []string {
	result := make([]string, 0, len(definitions))
	for _, definition := range definitions {
		result = append(result, definition.Language)
	}
	sort.Strings(result)
	return result
}

func ForPath(path string) []string {
	name := filepath.Base(path)
	extension := strings.ToLower(filepath.Ext(name))
	result := make([]string, 0, 1)
	for _, definition := range definitions {
		if contains(definition.ConfigFiles, name) || contains(definition.Extensions, extension) {
			result = append(result, definition.Language)
		}
	}
	if len(result) == 0 {
		return nil
	}
	sort.Strings(result)
	return result
}

func OwnsSource(language, path string) bool {
	return SourceExtension(language, filepath.Ext(path))
}

func SourceExtension(language, extension string) bool {
	definition, ok := Lookup(language)
	return ok && contains(definition.Extensions, strings.ToLower(extension))
}

func clone(definition Definition) Definition {
	definition.Extensions = append([]string(nil), definition.Extensions...)
	definition.ConfigFiles = append([]string(nil), definition.ConfigFiles...)
	return definition
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
