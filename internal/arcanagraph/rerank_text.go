package arcanagraph

import (
	"math"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/Lokee86/grimoire/internal/structure"
)

func hybridIdentifierScore(query string, queryTokens map[string]struct{}, node structure.Node) float64 {
	name := normalizedHybridText(node.Name)
	qualified := normalizedHybridText(node.QualifiedName)
	if name == "" {
		return 0
	}
	nameTokens := meaningfulHybridTokens(name)
	if len(nameTokens) == 0 {
		return 0
	}
	matchedWeight, totalWeight := hybridTokenOverlap(queryTokens, nameTokens)
	if matchedWeight == 0 || totalWeight == 0 {
		return 0
	}
	if len(nameTokens) == 1 {
		matchedToken := ""
		for token := range nameTokens {
			matchedToken = token
		}
		score := 0.35 + math.Min(float64(len(matchedToken)), 10)/40
		if hybridGenericTokens[matchedToken] {
			score *= 0.45
		}
		return clampHybridScore(score)
	}
	score := matchedWeight / totalWeight
	matchedCount := 0
	for token := range nameTokens {
		if _, exists := queryTokens[token]; exists {
			matchedCount++
		}
	}
	if matchedCount >= 2 {
		score = math.Min(1, score+0.15)
	}
	if len(name) >= 6 && strings.Contains(query, name) {
		score = max(score, 0.9)
	}
	if qualified != "" && len(qualified) >= 6 && strings.Contains(query, qualified) {
		score = max(score, 0.95)
	}
	return clampHybridScore(score)
}

func hybridPathScore(queryTokens map[string]struct{}, path string) float64 {
	pathTokens := meaningfulHybridTokens(normalizedHybridText(filepath.ToSlash(path)))
	if len(pathTokens) == 0 {
		return 0
	}
	matchedWeight, totalWeight := hybridTokenOverlap(queryTokens, pathTokens)
	if totalWeight == 0 {
		return 0
	}
	return clampHybridScore(matchedWeight / totalWeight)
}

func hybridTokenOverlap(query, candidate map[string]struct{}) (float64, float64) {
	matched, total := 0.0, 0.0
	for token := range candidate {
		weight := math.Min(float64(len(token)), 10) / 10
		total += weight
		if _, exists := query[token]; exists {
			matched += weight
		}
	}
	return matched, total
}

func meaningfulHybridTokens(value string) map[string]struct{} {
	result := make(map[string]struct{})
	for _, token := range strings.Fields(value) {
		token = hybridTokenRoot(token)
		if len(token) < 3 || hybridStopWords[token] {
			continue
		}
		result[token] = struct{}{}
	}
	return result
}

func hybridTokenRoot(token string) string {
	token = strings.ToLower(strings.TrimSpace(token))
	if alias := hybridTokenAliases[token]; alias != "" {
		return alias
	}
	switch {
	case len(token) > 6 && strings.HasSuffix(token, "ies"):
		token = strings.TrimSuffix(token, "ies") + "y"
	case len(token) > 6 && strings.HasSuffix(token, "ing"):
		token = strings.TrimSuffix(token, "ing")
	case len(token) > 5 && strings.HasSuffix(token, "ed"):
		token = strings.TrimSuffix(token, "ed")
	case len(token) > 5 && strings.HasSuffix(token, "es") && !strings.HasSuffix(token, "ses"):
		token = strings.TrimSuffix(token, "es")
	case len(token) > 4 && strings.HasSuffix(token, "s") && !strings.HasSuffix(token, "ss"):
		token = strings.TrimSuffix(token, "s")
	}
	if alias := hybridTokenAliases[token]; alias != "" {
		return alias
	}
	return token
}

func mergeHybridTokens(target, source map[string]struct{}) {
	for token := range source {
		target[token] = struct{}{}
	}
}

func hybridQueryCoverage(query, evidence map[string]struct{}) float64 {
	if len(query) == 0 || len(evidence) == 0 {
		return 0
	}
	matched, total := 0.0, 0.0
	for token := range query {
		weight := math.Min(float64(len(token)), 10) / 10
		total += weight
		if _, exists := evidence[token]; exists {
			matched += weight
		}
	}
	if total == 0 {
		return 0
	}
	return clampHybridScore(matched / total)
}

func normalizedHybridText(value string) string {
	var output strings.Builder
	var previous rune
	for _, current := range value {
		if unicode.IsUpper(current) && (unicode.IsLower(previous) || unicode.IsDigit(previous)) {
			output.WriteByte(' ')
		}
		if unicode.IsLetter(current) || unicode.IsDigit(current) {
			output.WriteRune(unicode.ToLower(current))
		} else {
			output.WriteByte(' ')
		}
		previous = current
	}
	return strings.Join(strings.Fields(output.String()), " ")
}

func hybridDeclarationQuality(node structure.Node) float64 {
	path := strings.ToLower(filepath.ToSlash(node.Path))
	if strings.Contains(path, "/test/") || strings.Contains(path, "/tests/") ||
		strings.Contains(path, "_test.") || strings.HasSuffix(path, "tests.rs") {
		return 0.15
	}
	switch strings.ToLower(strings.TrimSpace(node.Kind)) {
	case "function", "method", "constructor":
		return 1
	case "type", "class", "struct", "interface", "trait", "enum":
		return 0.82
	case "module", "package", "namespace":
		return 0.15
	case "file", "directory":
		return 0.05
	case "variable", "parameter", "field", "constant", "import", "export", "closure", "lambda":
		return 0.05
	default:
		return 0.35
	}
}
