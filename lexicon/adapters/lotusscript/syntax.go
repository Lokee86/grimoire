package main

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func logicalLines(path, source string) []logicalLine {
	physical := strings.Split(strings.ReplaceAll(source, "\r\n", "\n"), "\n")
	result := make([]logicalLine, 0, len(physical))
	var parts []string
	startLine := 0
	inBlockComment := false
	for index, raw := range physical {
		lineNumber := index + 1
		trimmed := strings.TrimSpace(strings.TrimSuffix(raw, "\r"))
		if inBlockComment {
			if isBlockCommentEnd(trimmed) {
				inBlockComment = false
			}
			continue
		}
		if isBlockCommentStart(trimmed) {
			inBlockComment = true
			continue
		}
		clean := strings.TrimSpace(stripLotusComment(trimmed))
		if startLine == 0 && clean == "" {
			continue
		}
		if startLine == 0 {
			startLine = lineNumber
		}
		continued := strings.HasSuffix(clean, "_")
		if continued {
			clean = strings.TrimSpace(strings.TrimSuffix(clean, "_"))
		}
		if clean != "" {
			parts = append(parts, clean)
		}
		if continued {
			continue
		}
		text := strings.TrimSpace(strings.Join(parts, " "))
		if text != "" {
			lineSpan := span{
				EndColumn: utf8.RuneCountInString(raw) + 1,
				EndLine:   lineNumber, Path: path,
				StartColumn: 1, StartLine: startLine,
			}
			result = append(result, splitStatements(lineSpan, text)...)
		}
		parts = nil
		startLine = 0
	}
	return result
}

func isBlockCommentStart(line string) bool {
	fields := strings.Fields(strings.ToLower(line))
	return len(fields) == 1 && fields[0] == "%rem"
}

func isBlockCommentEnd(line string) bool {
	fields := strings.Fields(strings.ToLower(line))
	return len(fields) == 2 && fields[0] == "%end" && fields[1] == "rem"
}

func stripLotusComment(line string) string {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) >= 3 && strings.EqualFold(trimmed[:3], "rem") {
		if len(trimmed) == 3 || unicode.IsSpace(rune(trimmed[3])) {
			return ""
		}
	}
	inQuote := false
	inPipe := false
	for index := 0; index < len(line); index++ {
		switch line[index] {
		case '"':
			if inPipe {
				continue
			}
			if inQuote && index+1 < len(line) && line[index+1] == '"' {
				index++
				continue
			}
			inQuote = !inQuote
		case '|':
			if !inQuote {
				inPipe = !inPipe
			}
		case '\'':
			if !inQuote && !inPipe {
				return line[:index]
			}
		}
	}
	return line
}

func splitStatements(lineSpan span, text string) []logicalLine {
	parts := splitTopLevel(text, ':')
	if len(parts) == 1 {
		return []logicalLine{{span: lineSpan, text: text}}
	}
	if first := strings.TrimSpace(parts[0]); first != "" && identifierPrefix(first) == first {
		parts = parts[1:]
	}
	result := make([]logicalLine, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			result = append(result, logicalLine{span: lineSpan, text: part})
		}
	}
	return result
}

func splitTopLevel(value string, separator byte) []string {
	var result []string
	start := 0
	depth := 0
	inQuote := false
	inPipe := false
	inDate := false
	for index := 0; index < len(value); index++ {
		switch value[index] {
		case '"':
			if inPipe {
				continue
			}
			if inQuote && index+1 < len(value) && value[index+1] == '"' {
				index++
				continue
			}
			inQuote = !inQuote
		case '|':
			if !inQuote && !inDate {
				inPipe = !inPipe
			}
		case '#':
			if !inQuote && !inPipe {
				inDate = !inDate
			}
		case '(':
			if !inQuote && !inPipe && !inDate {
				depth++
			}
		case ')':
			if !inQuote && !inPipe && !inDate && depth > 0 {
				depth--
			}
		default:
			if value[index] == separator && depth == 0 && !inQuote && !inPipe && !inDate {
				result = append(result, strings.TrimSpace(value[start:index]))
				start = index + 1
			}
		}
	}
	result = append(result, strings.TrimSpace(value[start:]))
	return result
}

func literalValue(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if len(value) < 2 {
		return "", false
	}
	if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '|' && value[len(value)-1] == '|') {
		return value[1 : len(value)-1], true
	}
	return "", false
}

func identifierPrefix(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || !(value[0] == '_' || unicode.IsLetter(rune(value[0]))) {
		return ""
	}
	end := 1
	for end < len(value) {
		r := rune(value[end])
		if value[end] != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			break
		}
		end++
	}
	return value[:end]
}
