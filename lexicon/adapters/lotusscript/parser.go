package main

import (
	"regexp"
	"strings"
)

var (
	callablePattern   = regexp.MustCompile(`(?i)^((?:(?:Public|Private|Protected|Static)\s+)*)(Sub|Function|Property\s+(?:Get|Set))\s+([A-Za-z_][A-Za-z0-9_]*)\s*(\([^)]*\))?`)
	classPattern      = regexp.MustCompile(`(?i)^(?:(Public|Private|Protected)\s+)?Class\s+([A-Za-z_][A-Za-z0-9_]*)(?:\s+As\s+([A-Za-z_][A-Za-z0-9_.]*))?`)
	constPattern      = regexp.MustCompile(`(?i)^((?:(?:Public|Private|Protected)\s+)*)Const\s+(.+)$`)
	externalPattern   = regexp.MustCompile(`(?i)^Declare\s+((?:(?:Public|Private)\s+)*)(Sub|Function)\s+([A-Za-z_][A-Za-z0-9_]*)\s*(\([^)]*\))?(.*)$`)
	typePattern       = regexp.MustCompile(`(?i)^(?:(Public|Private)\s+)?Type\s+([A-Za-z_][A-Za-z0-9_]*)`)
	usePattern        = regexp.MustCompile(`(?i)^(UseLSX|Use)\s+(.+)$`)
	withMemberPattern = regexp.MustCompile(`(^|[\s(=,+\-*/&<>])\.([A-Za-z_][A-Za-z0-9_]*)`)
)

func parseLotusScript(state *analysisState, file *parsedFile) {
	var currentClass *declaration
	var currentCallable *declaration
	var withStack []string
	for _, line := range logicalLines(file.path, string(file.content)) {
		lower := strings.ToLower(strings.TrimSpace(line.text))
		if lower == "option public" {
			state.modulePublic[file.path] = true
			continue
		}
		switch lower {
		case "end sub", "end function", "end property":
			currentCallable = nil
			withStack = nil
			continue
		case "end class", "end type":
			currentCallable = nil
			currentClass = nil
			withStack = nil
			continue
		case "end with":
			if len(withStack) > 0 {
				withStack = withStack[:len(withStack)-1]
			}
			continue
		}

		if match := usePattern.FindStringSubmatch(line.text); match != nil {
			state.addUse(file, line, match[1], match[2])
			continue
		}
		if match := classPattern.FindStringSubmatch(line.text); match != nil {
			decl := state.addType(file, line, "type", match[2], match[1], match[3], false)
			currentClass = &decl
			currentCallable = nil
			continue
		}
		if match := typePattern.FindStringSubmatch(line.text); match != nil {
			decl := state.addType(file, line, "type", match[2], match[1], "", true)
			currentClass = &decl
			currentCallable = nil
			continue
		}
		if match := externalPattern.FindStringSubmatch(line.text); match != nil && containsWord(match[5], "lib") {
			state.addCallable(file, line, currentClass, match[2], match[3], match[1], match[4], true, match[5])
			continue
		}
		if match := callablePattern.FindStringSubmatch(line.text); match != nil {
			decl := state.addCallable(file, line, currentClass, match[2], match[3], match[1], match[4], false, "")
			currentCallable = &decl
			withStack = nil
			continue
		}

		if currentCallable != nil && strings.HasPrefix(lower, "with ") {
			expression := strings.TrimSpace(line.text[len("With"):])
			if len(withStack) > 0 {
				expression = expandWithMembers(expression, withStack[len(withStack)-1])
			}
			withLine := line
			withLine.text = expression
			state.collectCalls(withLine, *currentCallable)
			state.collectAccess(withLine, *currentCallable)
			withStack = append(withStack, expression)
			continue
		}

		callLine := line
		if currentCallable != nil && len(withStack) > 0 {
			callLine.text = expandWithMembers(line.text, withStack[len(withStack)-1])
		}
		if state.addVariables(file, line, currentClass, currentCallable) {
			if currentCallable != nil {
				state.collectCalls(callLine, *currentCallable)
				if strings.HasPrefix(lower, "redim ") {
					state.collectAccess(callLine, *currentCallable)
				}
			}
			continue
		}
		if currentCallable != nil {
			state.collectCalls(callLine, *currentCallable)
			state.collectAccess(callLine, *currentCallable)
		}
	}
}

func expandWithMembers(text, receiver string) string {
	if receiver == "" || !strings.Contains(text, ".") {
		return text
	}
	return withMemberPattern.ReplaceAllStringFunc(text, func(match string) string {
		index := strings.LastIndex(match, ".")
		if index < 0 {
			return match
		}
		return match[:index] + receiver + match[index:]
	})
}
