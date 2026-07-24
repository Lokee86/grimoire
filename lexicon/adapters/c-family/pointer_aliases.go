package main

import "unicode"

func propagateFunctionPointerAliases(declarations []*declaration) {
	aliases := map[string]struct{}{}
	for _, declaration := range declarations {
		if declaration.Kind != "type" {
			continue
		}
		functionPointer, _ := declaration.Attributes["function_pointer"].(bool)
		if functionPointer {
			aliases[declaration.Name] = struct{}{}
		}
	}

	changed := true
	for changed {
		changed = false
		for _, declaration := range declarations {
			if functionPointer, _ := declaration.Attributes["function_pointer"].(bool); functionPointer {
				continue
			}
			typeText := declarationTypeText(declaration)
			if !containsAnyIdentifier(typeText, aliases) {
				continue
			}
			declaration.Attributes["function_pointer"] = true
			changed = true
			if declaration.Kind == "type" {
				aliases[declaration.Name] = struct{}{}
			}
		}
	}
}

func declarationTypeText(declaration *declaration) string {
	if value, _ := declaration.Attributes["type"].(string); value != "" {
		return value
	}
	if value, _ := declaration.Attributes["target"].(string); value != "" {
		return value
	}
	return ""
}

func containsAnyIdentifier(value string, identifiers map[string]struct{}) bool {
	start := -1
	for index, character := range value {
		if unicode.IsLetter(character) || character == '_' || start >= 0 && unicode.IsDigit(character) {
			if start < 0 {
				start = index
			}
			continue
		}
		if start >= 0 {
			if _, ok := identifiers[value[start:index]]; ok {
				return true
			}
			start = -1
		}
	}
	if start >= 0 {
		_, ok := identifiers[value[start:]]
		return ok
	}
	return false
}
