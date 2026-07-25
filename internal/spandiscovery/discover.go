package spandiscovery

import (
	"path/filepath"
	"strings"
)

type languageFamily int

const (
	familyUnknown languageFamily = iota
	familyMarkdown
	familyIndent
	familyBrace
	familyTOML
)

type languageSpec struct {
	name   string
	family languageFamily
}

// Discover returns deterministic language-aware source boundaries for a file.
// Unsupported files return no spans so callers can retain fallback chunking.
func Discover(path, content string) []Span {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.TrimSuffix(content, "\n")
	if strings.TrimSpace(content) == "" {
		return nil
	}

	lines := strings.Split(content, "\n")
	spec := languageForPath(path)
	var spans []Span
	switch spec.family {
	case familyMarkdown:
		spans = discoverMarkdown(lines, spec.name)
	case familyIndent:
		spans = discoverIndentDeclarations(lines, spec.name)
	case familyBrace:
		spans = discoverBraceDeclarations(lines, spec.name)
	case familyTOML:
		spans = discoverTOMLSections(lines, spec.name)
	default:
		return nil
	}
	return normalizeSpans(spans, len(lines))
}

func languageForPath(path string) languageSpec {
	base := strings.ToLower(filepath.Base(path))
	extension := strings.ToLower(filepath.Ext(base))
	switch extension {
	case ".md", ".markdown":
		return languageSpec{name: "markdown", family: familyMarkdown}
	case ".py":
		return languageSpec{name: "python", family: familyIndent}
	case ".gd":
		return languageSpec{name: "gdscript", family: familyIndent}
	case ".rb":
		return languageSpec{name: "ruby", family: familyIndent}
	case ".go":
		return languageSpec{name: "go", family: familyBrace}
	case ".rs":
		return languageSpec{name: "rust", family: familyBrace}
	case ".js", ".jsx":
		return languageSpec{name: "javascript", family: familyBrace}
	case ".ts", ".tsx":
		return languageSpec{name: "typescript", family: familyBrace}
	case ".java":
		return languageSpec{name: "java", family: familyBrace}
	case ".cs":
		return languageSpec{name: "csharp", family: familyBrace}
	case ".c", ".h":
		return languageSpec{name: "c", family: familyBrace}
	case ".cc", ".cpp", ".hpp", ".hh", ".hxx":
		return languageSpec{name: "cpp", family: familyBrace}
	case ".toml":
		return languageSpec{name: "toml", family: familyTOML}
	}
	if base == "gemfile" || base == "rakefile" {
		return languageSpec{name: "ruby", family: familyIndent}
	}
	return languageSpec{}
}
