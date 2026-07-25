package main

import (
	"regexp"
	"strings"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

var macroTargetPattern = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)\s*(\(|$)`)
var includeGuardNamePattern = regexp.MustCompile(`^\s*#\s*ifndef\s+([A-Za-z_][A-Za-z0-9_]*)\b`)

var nonCallMacroTargets = map[string]struct{}{
	"_Alignof": {}, "_Generic": {}, "_Static_assert": {}, "alignof": {}, "defined": {},
	"do": {}, "else": {}, "for": {}, "if": {}, "return": {}, "sizeof": {}, "static_assert": {},
	"switch": {}, "typeof": {}, "typeof_unqual": {}, "while": {},
}

func isIncludeGuard(node *tree_sitter.Node, source []byte) bool {
	if node == nil || node.Kind() != "preproc_ifdef" && node.Kind() != "preproc_ifndef" {
		return false
	}
	text := nodeText(node, source)
	match := includeGuardNamePattern.FindStringSubmatch(text)
	if len(match) == 0 {
		return false
	}
	limit := len(text)
	if limit > 2048 {
		limit = 2048
	}
	definePattern := regexp.MustCompile(`(?m)^\s*#\s*define\s+` + regexp.QuoteMeta(match[1]) + `\b`)
	return definePattern.MatchString(text[:limit])
}

func macroDetails(node *tree_sitter.Node, name string, source []byte) (string, string) {
	replacement, _, _ := macroSemanticDetails(node, name, source)
	return replacement, directMacroTarget(replacement)
}

func macroSemanticDetails(node *tree_sitter.Node, name string, source []byte) (string, []string, []macroCallExpression) {
	text := nodeText(node, source)
	index := strings.Index(text, name)
	if index < 0 {
		return "", nil, nil
	}
	remainder := text[index+len(name):]
	var parameters []string
	if node.Kind() == "preproc_function_def" {
		remainder = strings.TrimLeft(remainder, " 	")
		if strings.HasPrefix(remainder, "(") {
			if end := matchingParenthesis(remainder); end >= 0 {
				parameters = macroParameterNames(remainder[1:end])
				remainder = remainder[end+1:]
			}
		}
	}
	replacement := normalizeMacroReplacement(remainder)
	return replacement, parameters, macroCallExpressions(replacement)
}

func macroParameterNames(value string) []string {
	parts := splitMacroArguments(value)
	parameters := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if part == "..." {
			part = "__VA_ARGS__"
		} else {
			part = strings.TrimSuffix(part, "...")
		}
		parameters = append(parameters, strings.TrimSpace(part))
	}
	return parameters
}

func matchingParenthesis(value string) int {
	depth := 0
	for index, character := range value {
		switch character {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return index
			}
		}
	}
	return -1
}

func normalizeMacroReplacement(value string) string {
	value = strings.ReplaceAll(value, "\\\r\n", " ")
	value = strings.ReplaceAll(value, "\\\n", " ")
	return normalizeSpace(value)
}

func directMacroTarget(replacement string) string {
	value := strings.TrimSpace(replacement)
	for strings.HasPrefix(value, "(") {
		value = strings.TrimSpace(value[1:])
	}
	match := macroTargetPattern.FindStringSubmatch(value)
	if len(match) == 0 {
		return ""
	}
	if _, excluded := nonCallMacroTargets[match[1]]; excluded {
		return ""
	}
	return match[1]
}

func macroCallExpressions(replacement string) []macroCallExpression {
	var calls []macroCallExpression
	for index := 0; index < len(replacement); {
		if replacement[index] == '"' || replacement[index] == '\'' {
			index = skipMacroQuoted(replacement, index)
			continue
		}
		if !isMacroIdentifierStart(replacement[index]) {
			index++
			continue
		}
		start := index
		callee, end := readMacroCallee(replacement, index)
		index = end
		if _, excluded := nonCallMacroTargets[callee]; excluded {
			continue
		}
		open := skipMacroSpace(replacement, end)
		if open >= len(replacement) || replacement[open] != '(' {
			continue
		}
		close := matchingDelimiter(replacement, open, '(', ')')
		if close < 0 {
			continue
		}
		if start > 0 && (replacement[start-1] == '.' || replacement[start-1] == '>') {
			index = end
			continue
		}
		expression := replacement[start : close+1]
		tokenPasting := strings.Contains(expression, "##")
		stringification := strings.Contains(strings.ReplaceAll(expression, "##", ""), "#")
		variadic := strings.Contains(expression, "__VA_ARGS__") || strings.Contains(expression, "__VA_OPT__") || strings.Contains(expression, "...")
		calls = append(calls, macroCallExpression{
			Callee: normalizeQualified(callee), Arguments: splitMacroArguments(replacement[open+1 : close]),
			TokenPasting: tokenPasting, Stringification: stringification,
			VariadicSubstitution: variadic, Unsupported: tokenPasting || stringification || variadic,
		})
		index = end
	}
	return calls
}

func readMacroCallee(value string, start int) (string, int) {
	index := start
	for {
		if index >= len(value) || !isMacroIdentifierStart(value[index]) {
			return value[start:index], index
		}
		index++
		for index < len(value) && isMacroIdentifierPart(value[index]) {
			index++
		}
		space := skipMacroSpace(value, index)
		if space+1 >= len(value) || value[space] != ':' || value[space+1] != ':' {
			return value[start:index], index
		}
		index = skipMacroSpace(value, space+2)
		if index >= len(value) || !isMacroIdentifierStart(value[index]) {
			return value[start:index], index
		}
	}
}

func splitMacroArguments(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	var arguments []string
	start, round, square, curly := 0, 0, 0, 0
	for index := 0; index < len(value); index++ {
		switch value[index] {
		case '"', '\'':
			index = skipMacroQuoted(value, index) - 1
		case '(':
			round++
		case ')':
			round--
		case '[':
			square++
		case ']':
			square--
		case '{':
			curly++
		case '}':
			curly--
		case ',':
			if round == 0 && square == 0 && curly == 0 {
				arguments = append(arguments, strings.TrimSpace(value[start:index]))
				start = index + 1
			}
		}
	}
	arguments = append(arguments, strings.TrimSpace(value[start:]))
	return arguments
}

func matchingDelimiter(value string, open int, left, right byte) int {
	depth := 0
	for index := open; index < len(value); index++ {
		switch value[index] {
		case '"', '\'':
			index = skipMacroQuoted(value, index) - 1
		case left:
			depth++
		case right:
			depth--
			if depth == 0 {
				return index
			}
		}
	}
	return -1
}

func skipMacroQuoted(value string, start int) int {
	quote := value[start]
	for index := start + 1; index < len(value); index++ {
		if value[index] == '\\' {
			index++
			continue
		}
		if value[index] == quote {
			return index + 1
		}
	}
	return len(value)
}

func skipMacroSpace(value string, index int) int {
	for index < len(value) && (value[index] == ' ' || value[index] == '	' || value[index] == '\r' || value[index] == '\n') {
		index++
	}
	return index
}

func isMacroIdentifierStart(value byte) bool {
	return value == '_' || value >= 'A' && value <= 'Z' || value >= 'a' && value <= 'z'
}

func isMacroIdentifierPart(value byte) bool {
	return isMacroIdentifierStart(value) || value >= '0' && value <= '9'
}
