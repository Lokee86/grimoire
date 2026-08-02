package main

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

var gradleDependencyConfigurations = map[string]struct{}{
	"api": {}, "compileOnly": {}, "implementation": {}, "kapt": {},
	"ksp": {}, "runtimeOnly": {}, "testImplementation": {},
}

func parseGradleDependencies(path string, content []byte) []dependencyEvidence {
	var result []dependencyEvidence
	var blocks []bool
	dependenciesDepth := 0
	for offset := 0; offset < len(content); {
		if next := skipGradleTrivia(content, offset); next != offset {
			offset = next
			continue
		}
		if content[offset] == '"' || content[offset] == '\'' {
			offset = skipGradleString(content, offset)
			continue
		}
		switch content[offset] {
		case '{':
			blocks = append(blocks, false)
			offset++
			continue
		case '}':
			if len(blocks) != 0 {
				if blocks[len(blocks)-1] {
					dependenciesDepth--
				}
				blocks = blocks[:len(blocks)-1]
			}
			offset++
			continue
		}
		if !isGradleIdentifierStart(content[offset]) {
			offset++
			continue
		}
		start := offset
		offset = scanGradleIdentifier(content, offset)
		name := string(content[start:offset])
		if name == "dependencies" {
			next := skipGradleSpacesAndComments(content, offset)
			if next < len(content) && content[next] == '{' {
				blocks = append(blocks, true)
				dependenciesDepth++
				offset = next + 1
			}
			continue
		}
		if dependenciesDepth == 0 || len(blocks) == 0 || !blocks[len(blocks)-1] {
			continue
		}
		if _, supported := gradleDependencyConfigurations[name]; !supported {
			continue
		}
		dependency, end, ok := parseGradleDeclaration(path, content, start, offset, name)
		if ok {
			result = append(result, dependency)
			offset = end
		}
	}
	return result
}

func parseGradleDeclaration(path string, content []byte, start, afterName int, configuration string) (dependencyEvidence, int, bool) {
	expressionStart := afterName
	for expressionStart < len(content) && (content[expressionStart] == ' ' || content[expressionStart] == '	') {
		expressionStart++
	}
	if expressionStart >= len(content) || content[expressionStart] == '\n' || content[expressionStart] == '\r' {
		return dependencyEvidence{}, afterName, false
	}
	expressionEnd, declarationEnd := expressionStart, expressionStart
	if content[expressionStart] == '(' {
		close, ok := scanGradleBalanced(content, expressionStart)
		if !ok {
			close = len(content)
		}
		expressionStart++
		expressionEnd = close
		declarationEnd = close
		if close < len(content) {
			declarationEnd++
		}
	} else {
		expressionEnd = scanGradleLineExpression(content, expressionStart)
		declarationEnd = expressionEnd
	}
	expressionStart, expressionEnd = trimByteRange(content, expressionStart, expressionEnd)
	if expressionStart >= expressionEnd {
		return dependencyEvidence{}, declarationEnd, false
	}
	expression := string(content[expressionStart:expressionEnd])
	coordinate, group, artifact, version, resolved := literalGradleCoordinate(expression)
	return dependencyEvidence{
		artifact: artifact, configuration: configuration, coordinate: coordinate,
		expression: expression, group: group, resolved: resolved,
		span: offsetSpan(path, content, start, declarationEnd), version: version,
	}, declarationEnd, true
}

func literalGradleCoordinate(expression string) (string, string, string, string, bool) {
	value := strings.TrimSpace(expression)
	if len(value) < 2 || (value[0] != '"' && value[0] != '\'') || value[len(value)-1] != value[0] {
		return "", "", "", "", false
	}
	coordinate := value[1 : len(value)-1]
	if coordinate == "" || strings.ContainsAny(coordinate, "$\\\"'(){}[] \t\r\n/") {
		return "", "", "", "", false
	}
	parts := strings.Split(coordinate, ":")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", "", "", false
	}
	version := ""
	if len(parts) > 2 {
		version = parts[2]
	}
	return coordinate, parts[0], parts[1], version, true
}

func scanGradleBalanced(content []byte, open int) (int, bool) {
	depth := 0
	for offset := open; offset < len(content); {
		if next := skipGradleTrivia(content, offset); next != offset {
			offset = next
			continue
		}
		if content[offset] == '"' || content[offset] == '\'' {
			offset = skipGradleString(content, offset)
			continue
		}
		if content[offset] == '(' {
			depth++
		} else if content[offset] == ')' {
			depth--
			if depth == 0 {
				return offset, true
			}
		}
		offset++
	}
	return len(content), false
}

func scanGradleLineExpression(content []byte, offset int) int {
	depth := 0
	for offset < len(content) {
		if content[offset] == '"' || content[offset] == '\'' {
			offset = skipGradleString(content, offset)
			continue
		}
		if content[offset] == '(' {
			depth++
		} else if content[offset] == ')' && depth > 0 {
			depth--
		} else if depth == 0 && (content[offset] == '\n' || content[offset] == '\r' || content[offset] == ';' || content[offset] == '}') {
			break
		} else if depth == 0 && offset+1 < len(content) && content[offset] == '/' && content[offset+1] == '/' {
			break
		}
		offset++
	}
	return offset
}

func skipGradleTrivia(content []byte, offset int) int {
	if unicode.IsSpace(rune(content[offset])) {
		return offset + 1
	}
	if offset+1 >= len(content) || content[offset] != '/' {
		return offset
	}
	if content[offset+1] == '/' {
		for offset < len(content) && content[offset] != '\n' {
			offset++
		}
		return offset
	}
	if content[offset+1] == '*' {
		if end := strings.Index(string(content[offset+2:]), "*/"); end >= 0 {
			return offset + end + 4
		}
		return len(content)
	}
	return offset
}

func skipGradleSpacesAndComments(content []byte, offset int) int {
	for offset < len(content) {
		next := skipGradleTrivia(content, offset)
		if next == offset {
			break
		}
		offset = next
	}
	return offset
}

func skipGradleString(content []byte, offset int) int {
	quote := content[offset]
	triple := quote == '"' && offset+2 < len(content) && content[offset+1] == quote && content[offset+2] == quote
	if triple {
		offset += 3
		for offset+2 < len(content) {
			if content[offset] == quote && content[offset+1] == quote && content[offset+2] == quote {
				return offset + 3
			}
			offset++
		}
		return len(content)
	}
	for offset++; offset < len(content); offset++ {
		if content[offset] == '\\' {
			offset++
		} else if offset < len(content) && content[offset] == quote {
			return offset + 1
		}
	}
	return len(content)
}

func scanGradleIdentifier(content []byte, offset int) int {
	for offset < len(content) && (isGradleIdentifierStart(content[offset]) || content[offset] >= '0' && content[offset] <= '9') {
		offset++
	}
	return offset
}

func isGradleIdentifierStart(value byte) bool {
	return value == '_' || value >= 'A' && value <= 'Z' || value >= 'a' && value <= 'z'
}

func trimByteRange(content []byte, start, end int) (int, int) {
	for start < end && unicode.IsSpace(rune(content[start])) {
		start++
	}
	for end > start && unicode.IsSpace(rune(content[end-1])) {
		end--
	}
	return start, end
}

func offsetSpan(path string, content []byte, start, end int) sourceSpan {
	startLine, startColumn := offsetPosition(content, start)
	endLine, endColumn := offsetPosition(content, end)
	return sourceSpan{Path: path, StartLine: startLine, StartColumn: startColumn, EndLine: endLine, EndColumn: endColumn}
}

func offsetPosition(content []byte, target int) (int, int) {
	line, column := 1, 1
	for offset := 0; offset < target && offset < len(content); {
		if content[offset] == '\r' {
			offset++
			if offset < target && offset < len(content) && content[offset] == '\n' {
				offset++
			}
			line, column = line+1, 1
			continue
		}
		if content[offset] == '\n' {
			offset++
			line, column = line+1, 1
			continue
		}
		_, width := utf8.DecodeRune(content[offset:])
		offset += width
		column++
	}
	return line, column
}
