package app

import (
	"path/filepath"
	"strings"

	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/queryshape"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

func candidateIntentBoost(candidate retrieve.Candidate, planned queryshape.RetrievalIntent) (string, float64) {
	normalizedPath := strings.ToLower(filepath.ToSlash(candidate.Chunk.Path))
	base := strings.ToLower(filepath.Base(normalizedPath))
	text := strings.ToLower(candidate.Chunk.Text)
	firstLine := text
	if newline := strings.IndexByte(firstLine, '\n'); newline >= 0 {
		firstLine = firstLine[:newline]
	}
	artifactPenalty := generatedEvaluationPenalty(normalizedPath, planned.Query)

	switch planned.Intent {
	case evidence.IntentDirectLocation:
		value := artifactPenalty
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
		if implementationSourcePath(normalizedPath) {
			value += 16
			if containsDeclaration(text) {
				value += 8
			}
		}
		return "direct-location implementation and declaration priority", value
	case evidence.IntentCallChain:
		value := artifactPenalty
		if implementationSourcePath(normalizedPath) && containsDeclaration(text) {
			value += 24
		} else if implementationSourcePath(normalizedPath) {
			value += 15
		}
		return "call-chain implementation priority", value
	case evidence.IntentArchitecture:
		value := artifactPenalty
		ownershipQuery := containsAny(strings.ToLower(planned.Query), " owns ", " own ", "ownership")
		switch {
		case implementationSourcePath(normalizedPath) && containsDeclaration(text):
			if ownershipQuery {
				value += 22
			} else {
				value += 16
			}
		case implementationSourcePath(normalizedPath):
			if ownershipQuery {
				value += 14
			} else {
				value += 10
			}
		case base == "readme.md", base == "go.mod", base == "cargo.toml", base == "package.json":
			value += 4
		case strings.Contains(normalizedPath, "/architecture/"), strings.Contains(normalizedPath, "/docs/architecture"):
			value += 2
		case containsAny(base, "interface", "registry", "factory", "module", "container", "router"):
			value += 5
		case containsAny(firstLine, "package ", "module ", "namespace "):
			value += 3
		}
		return "architecture ownership implementation priority", value
	case evidence.IntentMechanism:
		value := artifactPenalty
		switch {
		case implementationSourcePath(normalizedPath) && containsDeclaration(text):
			value += 16
		case implementationSourcePath(normalizedPath):
			value += 10
		case supportingSourcePath(normalizedPath):
			value++
		}
		return "mechanism implementation priority", value
	}
	if artifactPenalty != 0 {
		return "generated evaluation artifact penalty", artifactPenalty
	}
	return "", 0
}

func generatedEvaluationPenalty(path, query string) float64 {
	query = strings.ToLower(filepath.ToSlash(query))
	base := strings.ToLower(filepath.Base(path))
	if strings.Contains(query, path) || base != "." && strings.Contains(query, base) {
		return 0
	}
	switch {
	case strings.Contains(path, "evaluation/results/"):
		return -80
	case strings.Contains(path, "evaluation/retrieval/"):
		return -60
	default:
		return 0
	}
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

func containsDeclaration(text string) bool {
	return containsAny(text, "func ", "function ", "def ", "class ", "type ", "interface ", "fn ")
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}
