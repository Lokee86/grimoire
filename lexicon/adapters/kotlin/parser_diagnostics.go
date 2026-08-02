package main

import (
	"fmt"
	"strings"
)

func diagnosticExpression(content []byte, diagnostic syntaxDiagnostic) string {
	start := diagnostic.token.startOffset
	if start < 0 || start > len(content) {
		start = 0
	}
	end := diagnostic.token.endOffset
	if end < start || end > len(content) {
		end = start
	}
	for start > 0 && content[start-1] != '\n' && content[start-1] != '\r' {
		start--
	}
	for end < len(content) && content[end] != '\n' && content[end] != '\r' {
		end++
	}
	expression := strings.TrimSpace(string(content[start:end]))
	if len(expression) > 160 {
		expression = expression[:157] + "..."
	}
	if expression == "" {
		expression = diagnostic.message
	}
	return expression
}

func diagnosticAttributes(diagnostic syntaxDiagnostic) map[string]any {
	return map[string]any{"diagnostic": diagnostic.message, "parser": "lexicon-kotlin-structural"}
}

func validateParsedFile(file *parsedKotlinFile) error {
	if file.path == "" {
		return fmt.Errorf("parsed Kotlin file has no path")
	}
	return nil
}
