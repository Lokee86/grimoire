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
	text := nodeText(node, source)
	index := strings.Index(text, name)
	if index < 0 {
		return "", ""
	}
	remainder := text[index+len(name):]
	if node.Kind() == "preproc_function_def" {
		remainder = strings.TrimLeft(remainder, " \t")
		if strings.HasPrefix(remainder, "(") {
			if end := matchingParenthesis(remainder); end >= 0 {
				remainder = remainder[end+1:]
			}
		}
	}
	replacement := normalizeMacroReplacement(remainder)
	return replacement, directMacroTarget(replacement)
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
