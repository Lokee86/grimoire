package main

import "strings"

func (state *parser) parseQualifiedDirective(allowWildcard bool) (string, int, bool) {
	var parts []string
	last := state.index
	expectName := true
	for state.current().kind != tokenEOF && state.current().kind != tokenNewline && !state.at(";") && !state.at("as") {
		current := state.current()
		if expectName {
			if current.kind == tokenIdentifier || (allowWildcard && current.text == "*") {
				parts = append(parts, identifierText(current))
				last = state.index
				state.index++
				expectName = false
				continue
			}
			return "", last, false
		}
		if current.text != "." {
			break
		}
		parts = append(parts, ".")
		last = state.index
		state.index++
		expectName = true
	}
	if len(parts) == 0 || expectName {
		return "", last, false
	}
	return strings.Join(parts, ""), last, true
}

func (state *parser) skipBalanced(open, close string) bool {
	if !state.at(open) {
		return false
	}
	matching := state.matchingDelimiter(state.index, open, close)
	if matching < 0 {
		state.addDiagnostic(state.index, "unclosed "+open+" delimiter")
		state.index = state.findRecoveryBoundary(state.index + 1)
		return false
	}
	state.index = matching + 1
	return true
}

func (state *parser) matchingDelimiter(openIndex int, open, close string) int {
	depth := 0
	for index := openIndex; index < len(state.tokens); index++ {
		switch state.tokens[index].text {
		case open:
			depth++
		case close:
			depth--
			if depth == 0 {
				return index
			}
		}
	}
	return -1
}

func (state *parser) findHeaderToken(wanted string) int {
	angle := 0
	for index := state.index; index < len(state.tokens); index++ {
		current := state.tokens[index]
		if current.kind == tokenEOF || current.text == "{" || current.text == "=" || current.text == ";" {
			return -1
		}
		if current.kind == tokenNewline && angle == 0 {
			return -1
		}
		if current.text == "<" {
			angle++
		} else if current.text == ">" && angle > 0 {
			angle--
		} else if current.text == wanted && angle == 0 {
			return index
		}
	}
	return -1
}

func (state *parser) findTypeEnd() int {
	depth := delimiterDepth{}
	for index := state.index; index < len(state.tokens); index++ {
		current := state.tokens[index]
		if depth.zero() && (current.kind == tokenEOF || current.kind == tokenNewline || current.text == "{" || current.text == "=" || current.text == ";" || current.text == "where") {
			return index
		}
		depth.update(current.text)
	}
	return len(state.tokens) - 1
}

func (state *parser) findPropertyHeaderEnd() int {
	depth := delimiterDepth{}
	for index := state.index; index < len(state.tokens); index++ {
		current := state.tokens[index]
		if depth.zero() && (current.kind == tokenEOF || current.kind == tokenNewline || current.text == ";" || current.text == "{") {
			return index
		}
		depth.update(current.text)
	}
	return len(state.tokens) - 1
}

func (state *parser) skipExpression() {
	depth := delimiterDepth{}
	for state.current().kind != tokenEOF {
		current := state.current()
		if depth.zero() && (current.kind == tokenNewline || current.text == ";" || current.text == "}") {
			return
		}
		depth.update(current.text)
		state.index++
	}
}

func (state *parser) skipStatement() {
	depth := delimiterDepth{}
	for state.current().kind != tokenEOF {
		current := state.current()
		if depth.zero() && (current.kind == tokenNewline || current.text == ";" || current.text == "}") {
			if current.kind == tokenNewline || current.text == ";" {
				state.index++
			}
			return
		}
		depth.update(current.text)
		state.index++
	}
}

func (state *parser) skipStatementHeader() { state.skipStatement() }

func (state *parser) skipToLineEnd() {
	for state.current().kind != tokenEOF && state.current().kind != tokenNewline && !state.at(";") {
		state.index++
	}
}

func (state *parser) findRecoveryBoundary(start int) int {
	for index := start; index < len(state.tokens); index++ {
		if state.tokens[index].kind == tokenNewline || state.tokens[index].text == ";" || state.tokens[index].text == "}" || state.tokens[index].kind == tokenEOF {
			return index
		}
	}
	return len(state.tokens) - 1
}

func (state *parser) skipSeparators() {
	for state.current().kind == tokenNewline || state.at(";") {
		state.index++
	}
}

func (state *parser) skipNewlines() {
	for state.current().kind == tokenNewline {
		state.index++
	}
}

func (state *parser) current() token {
	if state.index >= len(state.tokens) {
		return state.tokens[len(state.tokens)-1]
	}
	return state.tokens[state.index]
}

func (state *parser) at(text string) bool { return state.current().text == text }

func (state *parser) nextNonNewline(index int) int {
	for index < len(state.tokens) && state.tokens[index].kind == tokenNewline {
		index++
	}
	return index
}

func (state *parser) span(start, end int) sourceSpan {
	if start < 0 {
		start = 0
	}
	if end < start {
		end = start
	}
	if start >= len(state.tokens) {
		start = len(state.tokens) - 1
	}
	if end >= len(state.tokens) {
		end = len(state.tokens) - 1
	}
	first, last := state.tokens[start], state.tokens[end]
	return sourceSpan{
		EndColumn: last.endColumn, EndLine: last.endLine, Path: state.path,
		StartColumn: first.startColumn, StartLine: first.startLine,
	}
}

func (state *parser) tokenText(start, end int) string {
	return compactTokenRange(state.tokens, start, end)
}

func (state *parser) addDiagnostic(index int, message string) {
	if len(state.tokens) == 0 {
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= len(state.tokens) {
		index = len(state.tokens) - 1
	}
	state.diagnostics = append(state.diagnostics, syntaxDiagnostic{message: message, token: state.tokens[index]})
}
