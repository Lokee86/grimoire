package spandiscovery

import (
	"regexp"
	"strings"
)

type braceDeclaration struct {
	line           int
	depth          int
	kind           Kind
	name           string
	receiverMethod bool
}

var (
	goFunction     = regexp.MustCompile(`^func\s+(\([^)]*\)\s*)?([A-Za-z_][A-Za-z0-9_]*)`)
	goType         = regexp.MustCompile(`^type\s+([A-Za-z_][A-Za-z0-9_]*)(?:\[[^]]+\])?\s+(struct|interface)\b`)
	rustFunction   = regexp.MustCompile(`^(?:(?:pub(?:\([^)]*\))?)\s+)?(?:async\s+)?fn\s+([A-Za-z_][A-Za-z0-9_]*)`)
	rustType       = regexp.MustCompile(`^(?:(?:pub(?:\([^)]*\))?)\s+)?(struct|enum|trait|union|mod)\s+([^\s<{(]+)`)
	rustImplFor    = regexp.MustCompile(`^(?:unsafe\s+)?impl(?:<[^>]+>)?\s+[^\s{]+\s+for\s+([^\s<{]+)`)
	rustImpl       = regexp.MustCompile(`^(?:unsafe\s+)?impl(?:<[^>]+>)?\s+([^\s<{]+)`)
	scriptFunction = regexp.MustCompile(`^(?:export\s+)?(?:default\s+)?(?:async\s+)?function\s+([A-Za-z_$][A-Za-z0-9_$]*)`)
	scriptType     = regexp.MustCompile(`^(?:export\s+)?(?:default\s+)?(?:abstract\s+)?(class|interface|enum|namespace)\s+([A-Za-z_$][A-Za-z0-9_$]*)`)
	scriptArrow    = regexp.MustCompile(`^(?:export\s+)?(?:const|let|var)\s+([A-Za-z_$][A-Za-z0-9_$]*)\b.*=>`)
	cType          = regexp.MustCompile(`^(?:(?:public|private|protected|internal|abstract|sealed|static|partial)\s+)*(class|struct|interface|enum|record|union|namespace)\s+([A-Za-z_][A-Za-z0-9_]*)`)
)

func discoverBraceDeclarations(lines []string, language string) []Span {
	content := strings.Join(lines, "\n")
	pairs, lineDepths := scanBracePairs(content)
	lineOffsets := sourceLineOffsets(content)
	declarations := make([]braceDeclaration, 0)
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || braceCommentLine(trimmed) {
			continue
		}
		depth := 0
		if index < len(lineDepths) {
			depth = lineDepths[index]
		}
		kind, name, receiverMethod, ok := matchBraceDeclaration(trimmed, language, depth)
		if !ok {
			continue
		}
		declarations = append(declarations, braceDeclaration{
			line: index + 1, depth: depth, kind: kind, name: name,
			receiverMethod: receiverMethod,
		})
	}

	spans := make([]Span, 0, len(declarations))
	for index, declaration := range declarations {
		nextLine := len(lines) + 1
		for next := index + 1; next < len(declarations); next++ {
			if declarations[next].depth <= declaration.depth {
				nextLine = declarations[next].line
				break
			}
		}
		pair, found := declarationBracePair(
			pairs, lineOffsets, declaration.line, nextLine, declaration.depth,
		)
		endLine := declaration.line
		if found {
			endLine = pair.closeLine
		} else if declaration.kind == KindFunction || declaration.receiverMethod {
			continue
		}
		kind := declaration.kind
		if declaration.receiverMethod {
			kind = KindMethod
		}
		spans = append(spans, Span{
			StartLine: declaration.line, EndLine: endLine,
			Kind: kind, Name: declaration.name, Language: language,
		})
	}

	for index := range spans {
		if spans[index].Kind != KindFunction {
			continue
		}
		for _, container := range spans {
			if container.Kind != KindType || container.StartLine >= spans[index].StartLine ||
				container.EndLine < spans[index].EndLine {
				continue
			}
			spans[index].Kind = KindMethod
			break
		}
	}
	return spans
}

func matchBraceDeclaration(line, language string, depth int) (Kind, string, bool, bool) {
	switch language {
	case "go":
		if match := goFunction.FindStringSubmatch(line); len(match) == 3 {
			return KindFunction, match[2], match[1] != "", true
		}
		if match := goType.FindStringSubmatch(line); len(match) == 3 {
			return KindType, match[1], false, true
		}
		return "", "", false, false
	case "rust":
		if match := rustFunction.FindStringSubmatch(line); len(match) == 2 {
			return KindFunction, match[1], false, true
		}
		if match := rustType.FindStringSubmatch(line); len(match) == 3 {
			kind := KindType
			if match[1] == "mod" {
				kind = KindBlock
			}
			return kind, match[2], false, true
		}
		if match := rustImplFor.FindStringSubmatch(line); len(match) == 2 {
			return KindType, match[1], false, true
		}
		if match := rustImpl.FindStringSubmatch(line); len(match) == 2 {
			return KindType, match[1], false, true
		}
		return "", "", false, false
	case "javascript", "typescript":
		if match := scriptFunction.FindStringSubmatch(line); len(match) == 2 {
			return KindFunction, match[1], false, true
		}
		if match := scriptArrow.FindStringSubmatch(line); len(match) == 2 {
			return KindFunction, match[1], false, true
		}
		if match := scriptType.FindStringSubmatch(line); len(match) == 3 {
			kind := KindType
			if match[1] == "namespace" {
				kind = KindBlock
			}
			return kind, match[2], false, true
		}
	}

	if match := cType.FindStringSubmatch(line); len(match) == 3 {
		kind := KindType
		if match[1] == "namespace" {
			kind = KindBlock
		}
		return kind, match[2], false, true
	}
	if name, ok := genericFunctionName(line, depth); ok {
		return KindFunction, name, false, true
	}
	return "", "", false, false
}

func genericFunctionName(line string, depth int) (string, bool) {
	open := strings.IndexByte(line, '(')
	if open <= 0 || strings.HasSuffix(strings.TrimSpace(line), ";") {
		return "", false
	}
	prefix := strings.TrimSpace(line[:open])
	if strings.ContainsAny(prefix, ".=") {
		return "", false
	}
	fields := strings.Fields(prefix)
	if len(fields) == 0 ||
		(len(fields) == 1 && depth == 0 && !strings.Contains(fields[0], "::")) {
		return "", false
	}
	first := strings.Trim(fields[0], "*&")
	for _, blocked := range []string{"if", "for", "while", "switch", "catch", "return", "throw", "new", "sizeof"} {
		if first == blocked {
			return "", false
		}
	}
	name := strings.Trim(fields[len(fields)-1], "*&")
	if qualifier := strings.LastIndex(name, "::"); qualifier >= 0 {
		name = name[qualifier+2:]
	}
	name = strings.TrimPrefix(name, "~")
	if name == "" || strings.ContainsAny(name, "[]{}<>,") {
		return "", false
	}
	return name, true
}

func declarationBracePair(
	pairs []bracePair,
	lineOffsets []int,
	startLine, nextLine, depth int,
) (bracePair, bool) {
	if startLine <= 0 || startLine > len(lineOffsets) {
		return bracePair{}, false
	}
	startOffset := lineOffsets[startLine-1]
	var best bracePair
	found := false
	for _, pair := range pairs {
		if pair.openOffset < startOffset || pair.openLine < startLine ||
			pair.openLine >= nextLine || pair.openLine > startLine+12 || pair.depth != depth {
			continue
		}
		if !found || pair.closeOffset > best.closeOffset {
			best = pair
			found = true
		}
	}
	return best, found
}

func sourceLineOffsets(content string) []int {
	offsets := []int{0}
	for index := 0; index < len(content); index++ {
		if content[index] == '\n' {
			offsets = append(offsets, index+1)
		}
	}
	return offsets
}

func braceCommentLine(line string) bool {
	return strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") ||
		strings.HasPrefix(line, "*")
}
