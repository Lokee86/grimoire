package main

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type tokenKind int

const (
	tokenIdentifier tokenKind = iota
	tokenNumber
	tokenString
	tokenSymbol
	tokenNewline
	tokenUnknown
	tokenEOF
)

type token struct {
	kind        tokenKind
	text        string
	startOffset int
	endOffset   int
	startLine   int
	startColumn int
	endLine     int
	endColumn   int
}

type syntaxDiagnostic struct {
	message string
	token   token
}

type lexer struct {
	content     []byte
	offset      int
	line        int
	column      int
	tokens      []token
	diagnostics []syntaxDiagnostic
}

func lex(content []byte) ([]token, []syntaxDiagnostic) {
	state := &lexer{content: content, line: 1, column: 1}
	for state.offset < len(content) {
		state.scanToken()
	}
	state.tokens = append(state.tokens, token{
		kind: tokenEOF, startOffset: len(content), endOffset: len(content),
		startLine: state.line, endLine: state.line, startColumn: state.column, endColumn: state.column,
	})
	return state.tokens, state.diagnostics
}

func (state *lexer) scanToken() {
	current := state.peekRune()
	if current == '\r' || current == '\n' {
		state.scanNewline()
		return
	}
	if unicode.IsSpace(current) {
		state.advanceRune()
		return
	}
	if state.hasPrefix("//") {
		for state.offset < len(state.content) && state.peekRune() != '\r' && state.peekRune() != '\n' {
			state.advanceRune()
		}
		return
	}
	if state.hasPrefix("/*") {
		state.scanBlockComment()
		return
	}
	if state.hasPrefix("\"\"\"") {
		state.scanQuotedString(true)
		return
	}
	if current == '"' || current == '\'' {
		state.scanQuotedString(false)
		return
	}
	if current == '`' {
		state.scanBacktickIdentifier()
		return
	}
	if isIdentifierStart(current) {
		state.scanIdentifier()
		return
	}
	if unicode.IsDigit(current) {
		state.scanNumber()
		return
	}
	if strings.ContainsRune("{}()[]<>,.:;=?@*+-/!&|%^~", current) {
		start := state.mark()
		state.advanceRune()
		state.emit(tokenSymbol, start)
		return
	}
	start := state.mark()
	state.advanceRune()
	state.emit(tokenUnknown, start)
	state.diagnostics = append(state.diagnostics, syntaxDiagnostic{
		message: fmt.Sprintf("unsupported character %q", current),
		token:   state.tokens[len(state.tokens)-1],
	})
}

func (state *lexer) scanNewline() {
	start := state.mark()
	if state.hasPrefix("\r\n") {
		state.offset += 2
	} else {
		state.offset++
	}
	state.line++
	state.column = 1
	state.emit(tokenNewline, start)
}

func (state *lexer) scanIdentifier() {
	start := state.mark()
	state.advanceRune()
	for state.offset < len(state.content) && isIdentifierPart(state.peekRune()) {
		state.advanceRune()
	}
	state.emit(tokenIdentifier, start)
}

func (state *lexer) scanNumber() {
	start := state.mark()
	state.advanceRune()
	for state.offset < len(state.content) {
		current := state.peekRune()
		if !unicode.IsDigit(current) && !unicode.IsLetter(current) && current != '_' && current != '.' {
			break
		}
		state.advanceRune()
	}
	state.emit(tokenNumber, start)
}

func (state *lexer) scanBacktickIdentifier() {
	start := state.mark()
	state.advanceRune()
	for state.offset < len(state.content) && state.peekRune() != '`' && state.peekRune() != '\r' && state.peekRune() != '\n' {
		state.advanceRune()
	}
	if state.offset >= len(state.content) || state.peekRune() != '`' {
		state.emit(tokenUnknown, start)
		state.diagnostics = append(state.diagnostics, syntaxDiagnostic{message: "unterminated backtick identifier", token: state.tokens[len(state.tokens)-1]})
		return
	}
	state.advanceRune()
	state.emit(tokenIdentifier, start)
}

func (state *lexer) scanQuotedString(triple bool) {
	start := state.mark()
	quote := state.peekRune()
	if triple {
		state.advanceRune()
		state.advanceRune()
		state.advanceRune()
		for state.offset < len(state.content) && !state.hasPrefix("\"\"\"") {
			state.advanceRune()
		}
		if !state.hasPrefix("\"\"\"") {
			state.emit(tokenUnknown, start)
			state.diagnostics = append(state.diagnostics, syntaxDiagnostic{message: "unterminated triple-quoted string", token: state.tokens[len(state.tokens)-1]})
			return
		}
		state.advanceRune()
		state.advanceRune()
		state.advanceRune()
		state.emit(tokenString, start)
		return
	}

	state.advanceRune()
	escaped := false
	for state.offset < len(state.content) {
		current := state.peekRune()
		if current == '\r' || current == '\n' {
			break
		}
		state.advanceRune()
		if current == quote && !escaped {
			state.emit(tokenString, start)
			return
		}
		if current == '\\' && !escaped {
			escaped = true
		} else {
			escaped = false
		}
	}
	state.emit(tokenUnknown, start)
	state.diagnostics = append(state.diagnostics, syntaxDiagnostic{message: "unterminated quoted literal", token: state.tokens[len(state.tokens)-1]})
}

func (state *lexer) scanBlockComment() {
	start := state.mark()
	state.advanceRune()
	state.advanceRune()
	depth := 1
	for state.offset < len(state.content) && depth > 0 {
		if state.hasPrefix("/*") {
			state.advanceRune()
			state.advanceRune()
			depth++
			continue
		}
		if state.hasPrefix("*/") {
			state.advanceRune()
			state.advanceRune()
			depth--
			continue
		}
		state.advanceRune()
	}
	if depth != 0 {
		state.tokens = append(state.tokens, state.tokenFrom(tokenUnknown, start))
		state.diagnostics = append(state.diagnostics, syntaxDiagnostic{message: "unterminated block comment", token: state.tokens[len(state.tokens)-1]})
	}
}

type lexerMark struct {
	offset int
	line   int
	column int
}

func (state *lexer) mark() lexerMark {
	return lexerMark{offset: state.offset, line: state.line, column: state.column}
}

func (state *lexer) emit(kind tokenKind, start lexerMark) {
	state.tokens = append(state.tokens, state.tokenFrom(kind, start))
}

func (state *lexer) tokenFrom(kind tokenKind, start lexerMark) token {
	return token{
		kind: kind, text: string(state.content[start.offset:state.offset]),
		startOffset: start.offset, endOffset: state.offset,
		startLine: start.line, startColumn: start.column, endLine: state.line, endColumn: state.column,
	}
}

func (state *lexer) advanceRune() rune {
	if state.offset >= len(state.content) {
		return 0
	}
	if state.hasPrefix("\r\n") {
		state.offset += 2
		state.line++
		state.column = 1
		return '\n'
	}
	value, size := utf8.DecodeRune(state.content[state.offset:])
	if value == utf8.RuneError && size == 0 {
		return 0
	}
	state.offset += size
	if value == '\n' || value == '\r' {
		state.line++
		state.column = 1
	} else {
		state.column++
	}
	return value
}

func (state *lexer) peekRune() rune {
	value, _ := utf8.DecodeRune(state.content[state.offset:])
	return value
}

func (state *lexer) hasPrefix(value string) bool {
	return len(state.content)-state.offset >= len(value) && string(state.content[state.offset:state.offset+len(value)]) == value
}

func isIdentifierStart(value rune) bool {
	return value == '_' || unicode.IsLetter(value)
}

func isIdentifierPart(value rune) bool {
	return isIdentifierStart(value) || unicode.IsDigit(value)
}
