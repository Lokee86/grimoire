package main

import "strings"

type callReference struct {
	callee   string
	name     string
	receiver []token
	args     [][]token
	expr     string
	span     sourceSpan
}

func findCalls(stmt statement, path string) []callReference {
	return findCallsInTokens(stmt.tokens, path)
}

func findCallsInTokens(tokens []token, path string) []callReference {
	var calls []callReference
	for i := 0; i+1 < len(tokens); i++ {
		current := tokens[i]
		if current.kind != tokenIdentifier || tokens[i+1].text != "(" || isCallKeyword(current.text) {
			continue
		}
		if current.text == "preload" || current.text == "load" || isCallDeclarationName(tokens, i) {
			continue
		}
		close := matchingParen(tokens, i+1)
		if close < 0 {
			close = i + 1
		}
		start := i
		var receiver []token
		if i > 0 && tokens[i-1].text == "." {
			start = receiverStart(tokens, i-1)
			receiver = append([]token(nil), tokens[start:i-1]...)
		}
		calls = append(calls, callReference{
			callee:   joinTokens(tokens[start : i+1]),
			name:     current.text,
			receiver: receiver,
			args:     splitArguments(tokens[i+2 : close]),
			expr:     joinTokens(tokens[start : close+1]),
			span:     spanFromTokens(path, tokens[start], tokens[close]),
		})
	}
	return calls
}

func terminalCall(tokens []token) *callReference {
	tokens = trimExpression(tokens)
	if len(tokens) < 3 || tokens[len(tokens)-1].text != ")" {
		return nil
	}
	calls := findCallsInTokens(tokens, "")
	for i := len(calls) - 1; i >= 0; i-- {
		call := calls[i]
		if call.span["end_line"] == tokens[len(tokens)-1].endLine && call.span["end_column"] == tokens[len(tokens)-1].endColumn {
			return &call
		}
	}
	return nil
}

func splitArguments(tokens []token) [][]token {
	if len(tokens) == 0 {
		return nil
	}
	var result [][]token
	start, depth := 0, 0
	for i := 0; i <= len(tokens); i++ {
		if i == len(tokens) || (tokens[i].text == "," && depth == 0) {
			part := trimExpression(tokens[start:i])
			if len(part) > 0 {
				result = append(result, append([]token(nil), part...))
			}
			start = i + 1
			continue
		}
		switch tokens[i].text {
		case "(", "[", "{":
			depth++
		case ")", "]", "}":
			if depth > 0 {
				depth--
			}
		}
	}
	return result
}

func receiverStart(tokens []token, dot int) int {
	depth := 0
	for i := dot - 1; i >= 0; i-- {
		switch tokens[i].text {
		case ")", "]", "}":
			depth++
		case "(", "[", "{":
			if depth > 0 {
				depth--
				continue
			}
			return i + 1
		}
		if depth == 0 && isExpressionBoundary(tokens[i].text) {
			return i + 1
		}
	}
	return 0
}

func isExpressionBoundary(text string) bool {
	switch text {
	case "=", ":=", "+=", "-=", "*=", "/=", ",", ":", "return", "if", "elif", "while", "for", "in", "and", "or", "not", "await":
		return true
	default:
		return false
	}
}

func isCallDeclarationName(tokens []token, index int) bool {
	for i := index - 1; i >= 0; i-- {
		if tokens[i].text == "func" || tokens[i].text == "signal" {
			return true
		}
		if tokens[i].text == ":" || tokens[i].text == "=" || tokens[i].text == ";" {
			break
		}
	}
	return false
}

func matchingOpen(tokens []token, close int) int {
	if close < 0 || close >= len(tokens) {
		return -1
	}
	opening := map[string]string{")": "(", "]": "[", "}": "{"}[tokens[close].text]
	if opening == "" {
		return -1
	}
	depth := 0
	for i := close; i >= 0; i-- {
		if tokens[i].text == tokens[close].text {
			depth++
		} else if tokens[i].text == opening {
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func trimExpression(tokens []token) []token {
	for len(tokens) > 0 && (tokens[0].text == "await" || tokens[0].text == "return") {
		tokens = tokens[1:]
	}
	for len(tokens) > 1 && tokens[0].text == "(" && tokens[len(tokens)-1].text == ")" && matchingParen(tokens, 0) == len(tokens)-1 {
		tokens = tokens[1 : len(tokens)-1]
	}
	return tokens
}

func simpleIdentifier(tokens []token) string {
	tokens = trimExpression(tokens)
	if len(tokens) == 1 && tokens[0].kind == tokenIdentifier {
		return tokens[0].text
	}
	return ""
}

func stringLiteral(tokens []token) string {
	tokens = trimExpression(tokens)
	if len(tokens) == 2 && tokens[0].text == "&" {
		tokens = tokens[1:]
	}
	if len(tokens) != 1 || tokens[0].kind != tokenString || len(tokens[0].text) < 2 {
		return ""
	}
	value := tokens[0].text
	if strings.HasPrefix(value, `"""`) || strings.HasPrefix(value, `'''`) {
		if len(value) >= 6 {
			return value[3 : len(value)-3]
		}
	}
	return value[1 : len(value)-1]
}
