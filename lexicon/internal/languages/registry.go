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

const genericLanguagePrefix = "generic-"

var definitions = []Definition{
	{Language: "c-family", Directory: "c-family", Extensions: []string{".c", ".cc", ".cp", ".cpp", ".cxx", ".c++", ".h", ".hh", ".hpp", ".hxx", ".h++", ".inc", ".inl", ".ipp", ".tpp"}, ConfigFiles: []string{"compile_commands.json", "CMakeLists.txt"}},
	{Language: "gdscript", Directory: "gdscript", Extensions: []string{".gd"}, ConfigFiles: []string{"project.godot"}},
	{Language: "go", Directory: "go", Extensions: []string{".go"}, ConfigFiles: []string{"go.mod", "go.sum"}},
	{Language: "python", Directory: "python", Extensions: []string{".py"}, ConfigFiles: []string{"pyproject.toml", "setup.cfg", "requirements.txt"}},
	{Language: "ruby", Directory: "ruby", Extensions: []string{".rb", ".gemspec"}, ConfigFiles: []string{"Gemfile", "Gemfile.lock"}},
	{Language: "rust", Directory: "rust", Extensions: []string{".rs"}, ConfigFiles: []string{"Cargo.toml", "Cargo.lock"}},
	{Language: "typescript", Directory: "typescript", Extensions: []string{".ts", ".tsx", ".mts", ".cts", ".js", ".jsx", ".mjs", ".cjs", ".svelte"}, ConfigFiles: []string{"package.json", "package-lock.json", "tsconfig.json", "jsconfig.json"}},
	{Language: "generic", Directory: "generic"},
}

var genericSourceExtensions = map[string]struct{}{
	".asm": {}, ".bash": {}, ".bat": {}, ".clj": {}, ".cljs": {},
	".cmd": {}, ".cr": {}, ".cs": {}, ".dart": {}, ".elm": {}, ".erl": {},
	".ex": {}, ".exs": {}, ".f03": {}, ".f90": {}, ".f95": {}, ".fish": {}, ".fs": {},
	".fsx": {}, ".groovy": {}, ".hs": {}, ".java": {},
	".jl": {}, ".kt": {}, ".kts": {}, ".lhs": {}, ".lua": {}, ".m": {}, ".ml": {},
	".mli": {}, ".mm": {}, ".nim": {}, ".nims": {}, ".pas": {}, ".php": {}, ".pl": {},
	".pm": {}, ".proto": {}, ".ps1": {}, ".r": {}, ".scala": {}, ".sc": {}, ".s": {},
	".sh": {}, ".sol": {}, ".sql": {}, ".swift": {}, ".sv": {}, ".v": {}, ".vb": {},
	".vbs": {}, ".vim": {}, ".zig": {},
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
	if IsGeneric(language) {
		return Definition{Language: language, Directory: "generic", Extensions: []string{GenericExtension(language)}}, true
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
		if definition.Language == "generic" {
			continue
		}
		if contains(definition.ConfigFiles, name) || contains(definition.Extensions, extension) {
			result = append(result, definition.Language)
		}
	}
	if len(result) > 0 {
		sort.Strings(result)
		return result
	}
	if _, ok := genericSourceExtensions[extension]; ok {
		return []string{genericLanguagePrefix + strings.TrimPrefix(extension, ".")}
	}
	return nil
}

func OwnsSource(language, path string) bool {
	return SourceExtension(language, filepath.Ext(path))
}

func SourceExtension(language, extension string) bool {
	extension = strings.ToLower(extension)
	if IsGeneric(language) {
		return GenericExtension(language) == extension
	}
	definition, ok := Lookup(language)
	return ok && contains(definition.Extensions, extension)
}

func IsGeneric(language string) bool {
	if !strings.HasPrefix(language, genericLanguagePrefix) {
		return false
	}
	extension := "." + strings.TrimPrefix(strings.ToLower(language), genericLanguagePrefix)
	_, ok := genericSourceExtensions[extension]
	return ok
}

func GenericExtension(language string) string {
	if !IsGeneric(language) {
		return ""
	}
	return "." + strings.TrimPrefix(strings.ToLower(language), genericLanguagePrefix)
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
