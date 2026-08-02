package main

import "strings"

type assignmentAccess struct {
	target string
	left   string
	right  string
}

func parseAssignmentAccess(masked string) (assignmentAccess, bool) {
	trimmed := strings.TrimSpace(masked)
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "redim ") {
		rest := strings.TrimSpace(trimmed[len("redim"):])
		if strings.HasPrefix(strings.ToLower(rest), "preserve ") {
			rest = strings.TrimSpace(rest[len("preserve"):])
		}
		match := identifierChainPattern.FindStringIndex(rest)
		if match == nil || match[0] != 0 {
			return assignmentAccess{}, false
		}
		return assignmentAccess{target: rest[match[0]:match[1]], left: rest[match[1]:]}, true
	}

	working := trimmed
	for _, prefix := range []string{"set ", "let ", "for "} {
		if strings.HasPrefix(strings.ToLower(working), prefix) {
			working = strings.TrimSpace(working[len(prefix):])
			break
		}
	}
	lowerWorking := strings.ToLower(working)
	for _, prefix := range []string{
		"if ", "elseif ", "while ", "until ", "do ", "loop ", "case ", "select ",
		"on ", "call ", "print ", "error ", "return ", "exit ", "with ",
	} {
		if strings.HasPrefix(lowerWorking, prefix) {
			return assignmentAccess{}, false
		}
	}

	equals := topLevelAssignmentEquals(working)
	if equals < 0 {
		return assignmentAccess{}, false
	}
	left := strings.TrimSpace(working[:equals])
	right := strings.TrimSpace(working[equals+1:])
	match := identifierChainPattern.FindStringIndex(left)
	if match == nil || match[0] != 0 {
		return assignmentAccess{}, false
	}
	target := left[match[0]:match[1]]
	remainder := strings.TrimSpace(left[match[1]:])
	if remainder != "" && !strings.HasPrefix(remainder, "(") {
		return assignmentAccess{}, false
	}
	return assignmentAccess{target: target, left: remainder, right: right}, true
}

func topLevelAssignmentEquals(value string) int {
	depth := 0
	for index := 0; index < len(value); index++ {
		switch value[index] {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case '=':
			if depth != 0 {
				continue
			}
			if index > 0 && (value[index-1] == '<' || value[index-1] == '>') {
				continue
			}
			return index
		}
	}
	return -1
}

func splitIdentifierChain(value string) []string {
	parts := strings.Split(value, ".")
	result := parts[:0]
	for _, part := range parts {
		if part = strings.ToLower(strings.TrimSpace(part)); part != "" {
			result = append(result, part)
		}
	}
	return result
}

func maskLotusLiterals(value string) string {
	masked := []byte(value)
	inQuote := false
	inPipe := false
	inDate := false
	for index := 0; index < len(masked); index++ {
		switch masked[index] {
		case '"':
			if inPipe || inDate {
				masked[index] = ' '
				continue
			}
			if inQuote && index+1 < len(masked) && masked[index+1] == '"' {
				masked[index], masked[index+1] = ' ', ' '
				index++
				continue
			}
			inQuote = !inQuote
			masked[index] = ' '
		case '|':
			if !inQuote && !inDate {
				inPipe = !inPipe
			}
			masked[index] = ' '
		case '#':
			if !inQuote && !inPipe {
				inDate = !inDate
			}
			masked[index] = ' '
		default:
			if inQuote || inPipe || inDate {
				masked[index] = ' '
			}
		}
	}
	return string(masked)
}
