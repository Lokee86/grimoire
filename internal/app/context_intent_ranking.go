package app

import (
	"path/filepath"
	"strings"

	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/queryshape"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

type facetRankingSignals struct {
	weightedCoverage float64
}

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
		if planned.Weight < 0.999 || !explicitLocationQuery(planned.Query) {
			return "direct-location implementation and declaration priority", legacyDirectLocationBoost(candidate, normalizedPath, text, artifactPenalty)
		}
		signals := candidateFacetRankingSignals(candidate)
		value := artifactPenalty + boundedCoverageBoost(signals.weightedCoverage, 4, 20)
		if implementationSourcePath(normalizedPath) {
			if containsDeclaration(text) {
				value += 18
			} else {
				value += 8
			}
		}
		return "direct-location facet specificity", value
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

func explicitLocationQuery(query string) bool {
	normalized := strings.ToLower(strings.TrimSpace(query))
	for _, prefix := range []string{
		"where ", "find ", "locate ", "which file ", "which function ",
		"which method ", "which type ", "which constant ",
	} {
		if strings.HasPrefix(normalized, prefix) {
			return true
		}
	}
	return false
}

func legacyDirectLocationBoost(candidate retrieve.Candidate, path, text string, artifactPenalty float64) float64 {
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
	if implementationSourcePath(path) {
		value += 16
		if containsDeclaration(text) {
			value += 8
		}
	}
	return value
}

func candidateFacetRankingSignals(candidate retrieve.Candidate) facetRankingSignals {
	termWeights := make(map[string]float64)
	signals := facetRankingSignals{}
	for _, detail := range candidate.ScoreDetails {
		name := detail.Name
		var term string
		weight := 0.0
		switch {
		case strings.HasPrefix(name, "BM25 content matches "):
			term = strings.TrimPrefix(name, "BM25 content matches ")
			weight = 1
		case strings.HasPrefix(name, "declaration alias "):
			term = strings.TrimPrefix(name, "declaration alias ")
			if separator := strings.Index(term, " -> "); separator >= 0 {
				term = term[:separator]
			}
			weight = 1.25
		case strings.HasPrefix(name, "leading line matches "):
			term = strings.TrimPrefix(name, "leading line matches ")
			weight = 0.75
		case strings.HasPrefix(name, "filename matches "):
			term = strings.TrimPrefix(name, "filename matches ")
			weight = 0.5
		case strings.HasPrefix(name, "path matches "):
			term = strings.TrimPrefix(name, "path matches ")
			weight = 0.25
		}
		term = strings.TrimSpace(term)
		if term != "" && weight > termWeights[term] {
			termWeights[term] = weight
		}
	}
	for _, weight := range termWeights {
		signals.weightedCoverage += weight
	}
	return signals
}

func boundedCoverageBoost(coverage, multiplier, maximum float64) float64 {
	value := coverage * multiplier
	if value > maximum {
		return maximum
	}
	return value
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
