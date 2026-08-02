package main

import (
	"sort"
	"strings"
)

type delimiterDepth struct{ braces, brackets, parentheses, angles int }

func (depth *delimiterDepth) update(text string) {
	switch text {
	case "{":
		depth.braces++
	case "}":
		if depth.braces > 0 {
			depth.braces--
		}
	case "[":
		depth.brackets++
	case "]":
		if depth.brackets > 0 {
			depth.brackets--
		}
	case "(":
		depth.parentheses++
	case ")":
		if depth.parentheses > 0 {
			depth.parentheses--
		}
	case "<":
		depth.angles++
	case ">":
		if depth.angles > 0 {
			depth.angles--
		}
	}
}

func (depth delimiterDepth) zero() bool {
	return depth.braces == 0 && depth.brackets == 0 && depth.parentheses == 0 && depth.angles == 0
}

func splitTopLevel(tokens []token, start, end int, separator string) [][2]int {
	var result [][2]int
	segmentStart := start
	depth := delimiterDepth{}
	for index := start; index < end; index++ {
		if depth.zero() && tokens[index].text == separator {
			result = append(result, [2]int{segmentStart, index})
			segmentStart = index + 1
			continue
		}
		depth.update(tokens[index].text)
	}
	result = append(result, [2]int{segmentStart, end})
	return result
}

func firstTopLevelToken(tokens []token, start, end int, wanted string) int {
	depth := delimiterDepth{}
	for index := start; index < end; index++ {
		if depth.zero() && tokens[index].text == wanted {
			return index
		}
		depth.update(tokens[index].text)
	}
	return -1
}

func lastIdentifier(tokens []token, start, end int) int {
	for index := end - 1; index >= start; index-- {
		if tokens[index].kind == tokenIdentifier {
			return index
		}
	}
	return -1
}

func lastToken(tokens []token, start, end int, wanted string) int {
	for index := end - 1; index >= start; index-- {
		if tokens[index].text == wanted {
			return index
		}
	}
	return -1
}

func compactTokenRange(tokens []token, start, end int) string {
	parts := make([]string, 0, end-start)
	for index := start; index < end && index < len(tokens); index++ {
		if tokens[index].kind != tokenNewline && tokens[index].kind != tokenEOF {
			parts = append(parts, tokens[index].text)
		}
	}
	return compactTokens(parts)
}

func compactTokens(parts []string) string {
	var output strings.Builder
	previous := ""
	for _, part := range parts {
		if part == "" {
			continue
		}
		if output.Len() != 0 && needsTokenSpace(previous, part) {
			output.WriteByte(' ')
		}
		output.WriteString(part)
		previous = part
	}
	return strings.TrimSpace(output.String())
}

func needsTokenSpace(previous, current string) bool {
	if previous == "" {
		return false
	}
	noSpaceBefore := ").,?]>:;"
	noSpaceAfter := "([<.@:"
	if strings.Contains(noSpaceBefore, current) || strings.Contains(noSpaceAfter, previous) {
		return false
	}
	if current == "<" || current == "(" || current == "[" || current == "." || current == "@" {
		return false
	}
	if previous == "." || previous == "?" || previous == "!" || previous == "~" {
		return false
	}
	return true
}

func annotationEnd(tokens []token, start, end int) int {
	index := start + 1
	for index < end && (tokens[index].kind == tokenIdentifier || tokens[index].text == "." || tokens[index].text == ":") {
		index++
	}
	if index < end && tokens[index].text == "(" {
		depth := 0
		for index < end {
			if tokens[index].text == "(" {
				depth++
			}
			if tokens[index].text == ")" {
				depth--
				if depth == 0 {
					return index + 1
				}
			}
			index++
		}
	}
	return index
}

func skipTokenKind(tokens []token, start, end int, kind tokenKind) int {
	for start < end && tokens[start].kind == kind {
		start++
	}
	return start
}

func trimTokenKind(tokens []token, start, end int, kind tokenKind) int {
	for end > start && tokens[end-1].kind == kind {
		end--
	}
	return end
}

func identifierText(value token) string {
	if strings.HasPrefix(value.text, "`") && strings.HasSuffix(value.text, "`") && len(value.text) >= 2 {
		return value.text[1 : len(value.text)-1]
	}
	return value.text
}

func uniqueSorted(values []string) []string {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value != "" {
			set[value] = struct{}{}
		}
	}
	result := make([]string, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func isVisibility(value string) bool {
	return value == "public" || value == "private" || value == "protected" || value == "internal"
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}

func nullableType(value string) bool {
	return strings.HasSuffix(strings.TrimSpace(value), "?")
}
