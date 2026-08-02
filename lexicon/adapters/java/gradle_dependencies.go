package main

import (
	"regexp"
	"strings"
)

var gradleConfiguration = regexp.MustCompile(`\b(implementation|api|compileOnly|runtimeOnly|testImplementation|kapt)\b`)

func parseGradleDependencies(path, kind string, content []byte) []dependencyEvidence {
	masked := maskGradleComments(content)
	var dependencies []dependencyEvidence
	for cursor := 0; cursor < len(masked); {
		location := gradleConfiguration.FindSubmatchIndex(masked[cursor:])
		if location == nil {
			break
		}
		start := cursor + location[0]
		nameEnd := cursor + location[1]
		configuration := string(masked[cursor+location[2] : cursor+location[3]])
		cursor = nameEnd
		if insideGradleString(masked, start) || (start > 0 && masked[start-1] == '.') {
			continue
		}
		expressionStart := skipGradleSpace(masked, nameEnd)
		if expressionStart < len(masked) && masked[expressionStart] == '.' {
			continue
		}
		end := gradleStatementEnd(masked, expressionStart)
		if end <= start {
			continue
		}
		cursor = end
		statement := strings.TrimSpace(string(content[start:end]))
		argument := gradleArgument(string(content[expressionStart:end]), kind)
		coordinate, version, resolved := gradleCoordinate(argument)
		dependencies = append(dependencies, dependencyEvidence{
			category: gradleCategory(configuration), configuration: configuration, constraint: version,
			coordinate: coordinate, expression: statement, resolved: resolved,
			source: "gradle:dependencies", span: dependencySpan(path, content, start, end),
		})
	}
	return dependencies
}

func gradleStatementEnd(content []byte, start int) int {
	if start >= len(content) {
		return start
	}
	if content[start] == '(' {
		if end := matchingGradleParenthesis(content, start); end > start {
			return end
		}
	}
	quote := byte(0)
	escaped := false
	for index := start; index < len(content); index++ {
		value := content[index]
		if quote != 0 {
			if escaped {
				escaped = false
			} else if value == '\\' {
				escaped = true
			} else if value == quote {
				quote = 0
			}
			continue
		}
		if value == '\'' || value == '"' {
			quote = value
			continue
		}
		if value == '\n' || value == ';' {
			return trimGradleEnd(content, start, index)
		}
	}
	return trimGradleEnd(content, start, len(content))
}

func matchingGradleParenthesis(content []byte, start int) int {
	depth := 0
	quote := byte(0)
	escaped := false
	for index := start; index < len(content); index++ {
		value := content[index]
		if quote != 0 {
			if escaped {
				escaped = false
			} else if value == '\\' {
				escaped = true
			} else if value == quote {
				quote = 0
			}
			continue
		}
		if value == '\'' || value == '"' {
			quote = value
			continue
		}
		switch value {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return index + 1
			}
		}
	}
	return -1
}

func gradleArgument(expression, kind string) string {
	expression = strings.TrimSpace(expression)
	if strings.HasPrefix(expression, "(") && strings.HasSuffix(expression, ")") {
		expression = strings.TrimSpace(expression[1 : len(expression)-1])
	}
	if len(expression) < 2 {
		return ""
	}
	quote := expression[0]
	if quote != '\'' && quote != '"' || expression[len(expression)-1] != quote {
		return ""
	}
	if kind == "gradle-kotlin" && quote == '\'' {
		return ""
	}
	value := expression[1 : len(expression)-1]
	if strings.Contains(value, "\\") || quote == '"' && strings.Contains(value, "$") {
		return ""
	}
	return value
}

func gradleCoordinate(argument string) (string, string, bool) {
	parts := strings.Split(argument, ":")
	if len(parts) != 3 {
		return "", "", false
	}
	for _, part := range parts {
		if part == "" || strings.TrimSpace(part) != part {
			return "", "", false
		}
	}
	return parts[0] + ":" + parts[1], parts[2], true
}

func gradleCategory(configuration string) string {
	switch configuration {
	case "testImplementation":
		return "test"
	case "compileOnly", "kapt":
		return "build"
	default:
		return "runtime"
	}
}

func maskGradleComments(content []byte) []byte {
	masked := append([]byte(nil), content...)
	quote := byte(0)
	lineComment, blockComment, escaped := false, false, false
	for index := 0; index < len(masked); index++ {
		value := masked[index]
		if lineComment {
			if value == '\n' {
				lineComment = false
			} else {
				masked[index] = ' '
			}
			continue
		}
		if blockComment {
			if value == '*' && index+1 < len(masked) && masked[index+1] == '/' {
				masked[index], masked[index+1] = ' ', ' '
				index++
				blockComment = false
			} else if value != '\n' {
				masked[index] = ' '
			}
			continue
		}
		if quote != 0 {
			if escaped {
				escaped = false
			} else if value == '\\' {
				escaped = true
			} else if value == quote {
				quote = 0
			}
			continue
		}
		if value == '\'' || value == '"' {
			quote = value
		} else if value == '/' && index+1 < len(masked) && masked[index+1] == '/' {
			masked[index], masked[index+1] = ' ', ' '
			index++
			lineComment = true
		} else if value == '/' && index+1 < len(masked) && masked[index+1] == '*' {
			masked[index], masked[index+1] = ' ', ' '
			index++
			blockComment = true
		}
	}
	return masked
}

func insideGradleString(content []byte, offset int) bool {
	quote := byte(0)
	escaped := false
	for index := 0; index < offset; index++ {
		value := content[index]
		if quote != 0 {
			if escaped {
				escaped = false
			} else if value == '\\' {
				escaped = true
			} else if value == quote {
				quote = 0
			}
		} else if value == '\'' || value == '"' {
			quote = value
		}
	}
	return quote != 0
}

func skipGradleSpace(content []byte, index int) int {
	for index < len(content) && (content[index] == ' ' || content[index] == '\t' || content[index] == '\r' || content[index] == '\n') {
		index++
	}
	return index
}

func trimGradleEnd(content []byte, start, end int) int {
	for end > start && (content[end-1] == ' ' || content[end-1] == '\t' || content[end-1] == '\r') {
		end--
	}
	return end
}
