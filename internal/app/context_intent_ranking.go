package app

import (
	"path/filepath"
	"strings"

	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

func candidateIntentBoost(candidate retrieve.Candidate, intent evidence.Intent) (string, float64) {
	normalizedPath := strings.ToLower(filepath.ToSlash(candidate.Chunk.Path))
	base := strings.ToLower(filepath.Base(normalizedPath))
	text := strings.ToLower(candidate.Chunk.Text)
	firstLine := text
	if newline := strings.IndexByte(firstLine, '\n'); newline >= 0 {
		firstLine = firstLine[:newline]
	}

	switch intent {
	case evidence.IntentDirectLocation:
		value := 0.0
		for _, detail := range candidate.ScoreDetails {
			switch {
			case strings.HasPrefix(detail.Name, "filename matches"):
				value += 10
			case strings.HasPrefix(detail.Name, "path matches"):
				value += 6
			case strings.HasPrefix(detail.Name, "leading line matches"):
				value += 3
			}
		}
		return "direct-location path and declaration priority", value
	case evidence.IntentCallChain:
		if implementationSourcePath(normalizedPath) && containsDeclaration(firstLine) {
			return "call-chain implementation declaration priority", 6
		}
		if implementationSourcePath(normalizedPath) {
			return "call-chain implementation priority", 3
		}
	case evidence.IntentArchitecture:
		switch {
		case base == "readme.md", base == "go.mod", base == "cargo.toml", base == "package.json":
			return "architecture boundary-file priority", 8
		case strings.Contains(normalizedPath, "/architecture/"), strings.Contains(normalizedPath, "/docs/architecture"):
			return "architecture documentation priority", 7
		case containsAny(base, "interface", "registry", "factory", "module", "container", "router"):
			return "architecture ownership declaration priority", 5
		case implementationSourcePath(normalizedPath) && containsAny(firstLine, "package ", "module ", "namespace "):
			return "architecture package declaration priority", 3
		}
	case evidence.IntentMechanism:
		switch {
		case implementationSourcePath(normalizedPath) && containsDeclaration(firstLine):
			return "mechanism implementation declaration priority", 5
		case implementationSourcePath(normalizedPath):
			return "mechanism implementation priority", 3
		case supportingSourcePath(normalizedPath):
			return "mechanism supporting evidence priority", 2
		}
	}
	return "", 0
}

func supportingSourcePath(path string) bool {
	normalized := strings.ToLower(filepath.ToSlash(path))
	base := filepath.Base(normalized)
	if strings.Contains(base, "_test.") || strings.Contains(base, ".test.") ||
		containsAny(normalized, "/test/", "/tests/", "/spec/", "/specs/") {
		return true
	}
	if strings.Contains(normalized, "/docs/") || strings.HasSuffix(base, ".md") || strings.HasSuffix(base, ".rst") {
		return true
	}
	switch filepath.Ext(base) {
	case ".json", ".toml", ".yaml", ".yml", ".ini", ".conf":
		return true
	default:
		return false
	}
}

func implementationSourcePath(path string) bool {
	return !supportingSourcePath(path)
}

func containsDeclaration(line string) bool {
	return containsAny(line, "func ", "function ", "def ", "class ", "type ", "interface ", "fn ")
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}
