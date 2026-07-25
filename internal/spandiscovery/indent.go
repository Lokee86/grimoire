package spandiscovery

import (
	"regexp"
	"strings"
)

type indentDeclaration struct {
	line   int
	indent int
	kind   Kind
	name   string
}

var (
	pythonDeclaration   = regexp.MustCompile(`^(async\s+def|def|class)\s+([A-Za-z_][A-Za-z0-9_]*)`)
	gdscriptDeclaration = regexp.MustCompile(`^(static\s+func|func|class)\s+([A-Za-z_][A-Za-z0-9_]*)`)
	rubyDeclaration     = regexp.MustCompile(`^(class|module|def)\s+([A-Za-z_][A-Za-z0-9_:!?=.]*|<<\s*self)`)
)

func discoverIndentDeclarations(lines []string, language string) []Span {
	declarations := make([]indentDeclaration, 0)
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || indentCommentLine(trimmed, language) {
			continue
		}
		kind, name, ok := matchIndentDeclaration(trimmed, language)
		if !ok {
			continue
		}
		declarations = append(declarations, indentDeclaration{
			line: index + 1, indent: sourceIndent(line), kind: kind, name: name,
		})
	}

	spans := make([]Span, 0, len(declarations))
	for index, declaration := range declarations {
		kind := declaration.kind
		if kind == KindFunction && declarationInsideType(declarations, index) {
			kind = KindMethod
		}
		spans = append(spans, Span{
			StartLine: declaration.line,
			EndLine:   declarationEnd(lines, declaration.line, declaration.indent, language),
			Kind:      kind,
			Name:      declaration.name,
			Language:  language,
		})
	}
	return spans
}

func matchIndentDeclaration(line, language string) (Kind, string, bool) {
	var match []string
	switch language {
	case "python":
		match = pythonDeclaration.FindStringSubmatch(line)
	case "gdscript":
		match = gdscriptDeclaration.FindStringSubmatch(line)
	case "ruby":
		match = rubyDeclaration.FindStringSubmatch(line)
	default:
		return "", "", false
	}
	if len(match) != 3 {
		return "", "", false
	}
	keyword := match[1]
	kind := KindFunction
	if keyword == "class" || keyword == "module" {
		kind = KindType
	}
	return kind, strings.TrimSpace(match[2]), true
}

func declarationInsideType(declarations []indentDeclaration, index int) bool {
	current := declarations[index]
	for previous := index - 1; previous >= 0; previous-- {
		candidate := declarations[previous]
		if candidate.indent >= current.indent {
			continue
		}
		return candidate.kind == KindType
	}
	return false
}

func declarationEnd(lines []string, startLine, indent int, language string) int {
	for index := startLine; index < len(lines); index++ {
		trimmed := strings.TrimSpace(lines[index])
		if trimmed == "" || indentCommentLine(trimmed, language) {
			continue
		}
		currentIndent := sourceIndent(lines[index])
		if language == "ruby" && trimmed == "end" && currentIndent <= indent {
			return index + 1
		}
		if currentIndent <= indent {
			return index
		}
	}
	return len(lines)
}

func sourceIndent(line string) int {
	indent := 0
	for _, current := range line {
		switch current {
		case ' ':
			indent++
		case '\t':
			indent += 4
		default:
			return indent
		}
	}
	return indent
}

func indentCommentLine(line, language string) bool {
	if strings.HasPrefix(line, "#") {
		return true
	}
	return language == "ruby" && strings.HasPrefix(line, "=begin")
}
