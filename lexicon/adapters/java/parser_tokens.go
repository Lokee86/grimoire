package main

import (
	"strings"
	"unicode/utf8"
)

func qualifiedTokenName(tokens []token) (string, bool) {
	if len(tokens) == 0 || len(tokens)%2 == 0 {
		return "", false
	}
	var builder strings.Builder
	for index, item := range tokens {
		if index%2 == 0 {
			if !identifierToken(item.text) {
				return "", false
			}
			builder.WriteString(item.text)
		} else if item.text != "." {
			return "", false
		} else {
			builder.WriteByte('.')
		}
	}
	return builder.String(), true
}

func normalizedTokens(tokens []token) string {
	var builder strings.Builder
	for _, item := range tokens {
		builder.WriteString(item.text)
	}
	return builder.String()
}

func identifierToken(text string) bool {
	if text == "" {
		return false
	}
	r, _ := utf8.DecodeRuneInString(text)
	if !isIdentifierStart(r) {
		return false
	}
	for _, current := range text {
		if !isIdentifierPart(current) {
			return false
		}
	}
	return true
}

func initializerHeader(tokens []token, start, end int) bool {
	return start == end || (start+1 == end && tokens[start].text == "static")
}

func containsToken(tokens []token, start, end int, text string) bool {
	for index := start; index < end; index++ {
		if tokens[index].text == text {
			return true
		}
	}
	return false
}

func findToken(tokens []token, start, end int, text string) int {
	for index := start; index < end; index++ {
		if tokens[index].text == text {
			return index
		}
	}
	return -1
}

func decimal(value int) string {
	if value == 0 {
		return "0"
	}
	var digits [24]byte
	index := len(digits)
	for value > 0 {
		index--
		digits[index] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[index:])
}

func min(left, right int) int {
	if left < right {
		return left
	}
	return right
}
