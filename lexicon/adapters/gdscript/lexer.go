package main

import (
	"fmt"
	"unicode"
)

type tokenKind uint8

const (
	tokenIdentifier tokenKind = iota
	tokenString
	tokenNumber
	tokenSymbol
)

type token struct {
	kind      tokenKind
	text      string
	line      int
	column    int
	endLine   int
	endColumn int
}

type statement struct {
	tokens []token
	indent int
	start  token
	end    token
}

func makeStatements(tokens []token) []statement {
	var statements []statement
	var current []token
	depth := 0
	flush := func() {
		if len(current) == 0 {
			return
		}
		statements = append(statements, statement{tokens: append([]token(nil), current...), indent: current[0].column - 1, start: current[0], end: current[len(current)-1]})
		current = nil
	}
	for _, tok := range tokens {
		if tok.text == "\n" {
			if depth == 0 {
				flush()
			}
			continue
		}
		current = append(current, tok)
		switch tok.text {
		case "(", "[", "{":
			depth++
		case ")", "]", "}":
			if depth > 0 {
				depth--
			}
		}
	}
	flush()
	return statements
}

func lex(source string) ([]token, error) {
	var tokens []token
	line, column := 1, 1
	for i := 0; i < len(source); {
		ch := source[i]
		if ch == '\r' {
			i++
			continue
		}
		if ch == '\n' {
			tokens = append(tokens, token{text: "\n", line: line, column: column, endLine: line, endColumn: column + 1})
			i++
			line++
			column = 1
			continue
		}
		if ch == ' ' || ch == '\t' {
			i++
			if ch == '\t' {
				column += 4
			} else {
				column++
			}
			continue
		}
		if ch == '#' {
			for i < len(source) && source[i] != '\n' {
				i++
				column++
			}
			continue
		}
		startLine, startColumn, start := line, column, i
		if ch == '\'' || ch == '"' {
			quote := ch
			triple := i+2 < len(source) && source[i+1] == quote && source[i+2] == quote
			if triple {
				i += 3
				column += 3
			} else {
				i++
				column++
			}
			closed := false
			for i < len(source) {
				if triple && i+2 < len(source) && source[i] == quote && source[i+1] == quote && source[i+2] == quote {
					i += 3
					column += 3
					closed = true
					break
				}
				if !triple && source[i] == quote {
					i++
					column++
					closed = true
					break
				}
				if source[i] == '\\' && i+1 < len(source) {
					i += 2
					column += 2
					continue
				}
				if source[i] == '\n' {
					i++
					line++
					column = 1
					continue
				}
				i++
				column++
			}
			if !closed {
				return nil, fmt.Errorf("unterminated string at %d:%d", startLine, startColumn)
			}
			tokens = append(tokens, token{kind: tokenString, text: source[start:i], line: startLine, column: startColumn, endLine: line, endColumn: column})
			continue
		}
		if isIdentifierStart(ch) {
			i++
			column++
			for i < len(source) && isIdentifierPart(source[i]) {
				i++
				column++
			}
			tokens = append(tokens, token{kind: tokenIdentifier, text: source[start:i], line: startLine, column: startColumn, endLine: line, endColumn: column})
			continue
		}
		if unicode.IsDigit(rune(ch)) {
			i++
			column++
			for i < len(source) && (isIdentifierPart(source[i]) || source[i] == '.') {
				i++
				column++
			}
			tokens = append(tokens, token{kind: tokenNumber, text: source[start:i], line: startLine, column: startColumn, endLine: line, endColumn: column})
			continue
		}
		symbol := string(ch)
		if i+1 < len(source) {
			candidate := source[i : i+2]
			switch candidate {
			case "->", ":=", "==", "!=", "<=", ">=", "&&", "||", "+=", "-=", "*=", "/=", "++", "--":
				symbol = candidate
			}
		}
		i += len(symbol)
		column += len(symbol)
		tokens = append(tokens, token{kind: tokenSymbol, text: symbol, line: startLine, column: startColumn, endLine: line, endColumn: column})
	}
	return tokens, nil
}

func spanFromTokens(path string, start, end token) sourceSpan {
	span := sourceSpan{"end_column": end.endColumn, "end_line": end.endLine, "start_column": start.column, "start_line": start.line}
	if path != "" {
		span["path"] = path
	}
	return span
}

func isIdentifierStart(ch byte) bool {
	return ch == '_' || ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z'
}

func isIdentifierPart(ch byte) bool { return isIdentifierStart(ch) || ch >= '0' && ch <= '9' }
