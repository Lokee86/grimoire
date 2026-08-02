package main

import (
	"sort"
	"strings"
)

func visibilityAttributes(value string) map[string]any {
	attributes := make(map[string]any)
	for _, word := range strings.Fields(strings.ToLower(value)) {
		switch word {
		case "public", "private", "protected":
			attributes["visibility"] = word
		case "static":
			attributes["static"] = true
		}
	}
	return attributes
}

func declarationPublic(modifiers string, defaultPublic bool) bool {
	for _, word := range strings.Fields(strings.ToLower(modifiers)) {
		switch word {
		case "public":
			return true
		case "private", "protected":
			return false
		}
	}
	return defaultPublic
}

func parameterName(value string) string {
	fields := strings.Fields(value)
	for _, field := range fields {
		switch strings.ToLower(field) {
		case "byval", "byref", "optional", "paramarray":
			continue
		}
		return identifierPrefix(field)
	}
	return ""
}

func declaredType(value string) string {
	lower := strings.ToLower(value)
	index := strings.Index(lower, " as ")
	if index < 0 {
		return ""
	}
	tail := strings.TrimSpace(value[index+4:])
	if equal := strings.Index(tail, "="); equal >= 0 {
		tail = strings.TrimSpace(tail[:equal])
	}
	if len(tail) >= 4 && strings.EqualFold(tail[:4], "new ") {
		tail = strings.TrimSpace(tail[4:])
	}
	return strings.TrimSpace(tail)
}

func isVariableModifier(value string) bool {
	switch strings.ToLower(value) {
	case "dim", "global", "public", "private", "protected", "static":
		return true
	default:
		return false
	}
}

func looksLikeMemberDeclaration(value string) bool {
	return identifierPrefix(strings.TrimSpace(value)) != ""
}

func redimVariableBody(value string) (string, bool) {
	fields := strings.Fields(value)
	if len(fields) < 2 || !strings.EqualFold(fields[0], "ReDim") {
		return "", false
	}
	index := 1
	if index < len(fields) && strings.EqualFold(fields[index], "Preserve") {
		index++
	}
	if index >= len(fields) {
		return "", false
	}
	return strings.TrimSpace(strings.Join(fields[index:], " ")), true
}

func containsWord(value, word string) bool {
	for _, field := range strings.FieldsFunc(value, func(r rune) bool {
		return !(r == '_' || r == '-' || r == '.' || r >= '0' && r <= '9' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z')
	}) {
		if strings.EqualFold(field, word) {
			return true
		}
	}
	return false
}

func libraryName(tail string) string {
	lower := strings.ToLower(tail)
	index := strings.Index(lower, "lib")
	if index < 0 {
		return ""
	}
	value := strings.TrimSpace(tail[index+3:])
	if fields := splitTopLevel(value, ' '); len(fields) > 0 {
		if literal, ok := literalValue(fields[0]); ok {
			return literal
		}
	}
	return ""
}

func sortDeclarations(items []declaration) {
	sort.Slice(items, func(left, right int) bool { return items[left].id < items[right].id })
}
