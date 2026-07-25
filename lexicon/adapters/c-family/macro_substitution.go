package main

import "strings"

func macroInvocationBindings(macro *declaration, arguments []string) (map[string]string, bool) {
	bindings := make(map[string]string, len(macro.MacroParameters))
	if !macro.MacroFunction {
		return bindings, true
	}
	if len(macro.MacroParameters) != len(arguments) {
		return nil, false
	}
	for index, parameter := range macro.MacroParameters {
		if parameter == "__VA_ARGS__" {
			return nil, false
		}
		bindings[parameter] = strings.TrimSpace(arguments[index])
	}
	return bindings, true
}

func substituteMacroCall(call macroCallExpression, bindings map[string]string) macroCallExpression {
	expanded := call
	expanded.Callee = substituteMacroCallee(call.Callee, bindings)
	expanded.Arguments = make([]string, len(call.Arguments))
	for index, argument := range call.Arguments {
		expanded.Arguments[index] = substituteMacroTokens(argument, bindings)
	}
	return expanded
}

func substituteMacroCallee(callee string, bindings map[string]string) string {
	if replacement, exists := bindings[callee]; exists {
		return normalizeQualified(stripWrappingParentheses(replacement))
	}
	return normalizeQualified(substituteMacroTokens(callee, bindings))
}

func substituteMacroTokens(expression string, bindings map[string]string) string {
	if len(bindings) == 0 || expression == "" {
		return expression
	}
	var output strings.Builder
	for index := 0; index < len(expression); {
		if expression[index] == '"' || expression[index] == '\'' {
			end := skipMacroQuoted(expression, index)
			output.WriteString(expression[index:end])
			index = end
			continue
		}
		if !isMacroIdentifierStart(expression[index]) {
			output.WriteByte(expression[index])
			index++
			continue
		}
		start := index
		index++
		for index < len(expression) && isMacroIdentifierPart(expression[index]) {
			index++
		}
		identifier := expression[start:index]
		if replacement, exists := bindings[identifier]; exists {
			output.WriteByte('(')
			output.WriteString(strings.TrimSpace(replacement))
			output.WriteByte(')')
			continue
		}
		output.WriteString(identifier)
	}
	return output.String()
}

func stripWrappingParentheses(expression string) string {
	expression = strings.TrimSpace(expression)
	for len(expression) >= 2 && expression[0] == '(' {
		close := matchingDelimiter(expression, 0, '(', ')')
		if close != len(expression)-1 {
			break
		}
		expression = strings.TrimSpace(expression[1:close])
	}
	return expression
}

func macroArgumentCandidates(arguments []string) []string {
	candidates := make([]string, len(arguments))
	for index, argument := range arguments {
		if identifier, ok := simpleIdentifierName(argument); ok {
			candidates[index] = identifier
		}
	}
	return candidates
}

func macroSubstitutionAttributes(bindings map[string]string) map[string]string {
	if len(bindings) == 0 {
		return nil
	}
	result := make(map[string]string, len(bindings))
	for parameter, expression := range bindings {
		result[parameter] = expression
	}
	return result
}

func macroHasUnsupportedSubstitution(macro *declaration) bool {
	replacement, _ := macro.Attributes["replacement"].(string)
	for index := 0; index < len(replacement); {
		if replacement[index] == '"' || replacement[index] == '\'' {
			index = skipMacroQuoted(replacement, index)
			continue
		}
		if replacement[index] == '#' {
			return true
		}
		if isMacroIdentifierStart(replacement[index]) {
			start := index
			index++
			for index < len(replacement) && isMacroIdentifierPart(replacement[index]) {
				index++
			}
			identifier := replacement[start:index]
			if identifier == "__VA_ARGS__" || identifier == "__VA_OPT__" {
				return true
			}
			continue
		}
		index++
	}
	return false
}
