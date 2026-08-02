package main

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

type token struct {
	column    int
	endColumn int
	endLine   int
	endOffset int
	line      int
	offset    int
	text      string
}

type lexError struct {
	column int
	line   int
	text   string
}

func lexJava(source string) ([]token, []lexError) {
	var tokens []token
	var errors []lexError
	for offset, line, column := 0, 1, 1; offset < len(source); {
		r, size := utf8.DecodeRuneInString(source[offset:])
		if r == '\r' {
			offset += size
			if offset < len(source) && source[offset] == '\n' {
				offset++
			}
			line++
			column = 1
			continue
		}
		if r == '\n' {
			offset += size
			line++
			column = 1
			continue
		}
		if unicode.IsSpace(r) {
			offset += size
			column++
			continue
		}
		if hasPrefix(source, offset, "//") {
			for offset < len(source) && source[offset] != '\n' && source[offset] != '\r' {
				_, width := utf8.DecodeRuneInString(source[offset:])
				offset += width
				column++
			}
			continue
		}
		if hasPrefix(source, offset, "/*") {
			startLine, startColumn := line, column
			offset += 2
			column += 2
			closed := false
			for offset < len(source) {
				if hasPrefix(source, offset, "*/") {
					offset += 2
					column += 2
					closed = true
					break
				}
				offset, line, column = advanceRune(source, offset, line, column)
			}
			if !closed {
				errors = append(errors, lexError{startColumn, startLine, "unterminated block comment"})
			}
			continue
		}

		startOffset, startLine, startColumn := offset, line, column
		if isIdentifierStart(r) {
			offset += size
			column++
			for offset < len(source) {
				next, width := utf8.DecodeRuneInString(source[offset:])
				if !isIdentifierPart(next) {
					break
				}
				offset += width
				column++
			}
			tokens = append(tokens, makeToken(source, startOffset, startLine, startColumn, offset, line, column))
			continue
		}
		if unicode.IsDigit(r) {
			offset += size
			column++
			for offset < len(source) {
				next, width := utf8.DecodeRuneInString(source[offset:])
				if !(isIdentifierPart(next) || next == '.') {
					break
				}
				offset += width
				column++
			}
			tokens = append(tokens, makeToken(source, startOffset, startLine, startColumn, offset, line, column))
			continue
		}
		if r == '"' || r == '\'' {
			quote := r
			textBlock := quote == '"' && hasPrefix(source, offset, "\"\"\"")
			width := size
			if textBlock {
				width = 3
			}
			offset += width
			column += width
			closed, escaped := false, false
			for offset < len(source) {
				if textBlock && hasPrefix(source, offset, "\"\"\"") {
					offset += 3
					column += 3
					closed = true
					break
				}
				next, nextWidth := utf8.DecodeRuneInString(source[offset:])
				if !textBlock && !escaped && next == quote {
					offset += nextWidth
					column++
					closed = true
					break
				}
				if !textBlock && (next == '\n' || next == '\r') {
					break
				}
				if !textBlock && next == '\\' && !escaped {
					escaped = true
				} else {
					escaped = false
				}
				offset, line, column = advanceRune(source, offset, line, column)
			}
			if !closed {
				errors = append(errors, lexError{startColumn, startLine, "unterminated literal"})
			}
			tokens = append(tokens, makeToken(source, startOffset, startLine, startColumn, offset, line, column))
			continue
		}
		if hasPrefix(source, offset, "...") {
			offset += 3
			column += 3
		} else {
			offset += size
			column++
		}
		tokens = append(tokens, makeToken(source, startOffset, startLine, startColumn, offset, line, column))
	}
	return tokens, errors
}

func makeToken(source string, startOffset, startLine, startColumn, endOffset, endLine, endColumn int) token {
	return token{
		column: startColumn, endColumn: endColumn, endLine: endLine, endOffset: endOffset,
		line: startLine, offset: startOffset, text: source[startOffset:endOffset],
	}
}

func advanceRune(source string, offset, line, column int) (int, int, int) {
	r, width := utf8.DecodeRuneInString(source[offset:])
	if r == '\r' {
		offset += width
		if offset < len(source) && source[offset] == '\n' {
			offset++
		}
		return offset, line + 1, 1
	}
	if r == '\n' {
		return offset + width, line + 1, 1
	}
	return offset + width, line, column + 1
}

func hasPrefix(source string, offset int, prefix string) bool {
	return offset+len(prefix) <= len(source) && source[offset:offset+len(prefix)] == prefix
}

func isIdentifierStart(r rune) bool {
	return r == '_' || r == '$' || unicode.IsLetter(r)
}

func isIdentifierPart(r rune) bool {
	return isIdentifierStart(r) || unicode.IsDigit(r) || unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Mc, r)
}

func (problem lexError) String() string {
	return fmt.Sprintf("%s at %d:%d", problem.text, problem.line, problem.column)
}
