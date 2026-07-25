package lexical

import (
	"strings"
	"unicode"
)

func Tokens(value string) []string {
	var tokens []string
	scanTokens(value, func(token []rune) {
		normalized := strings.ToLower(string(token))
		if normalized != "" {
			tokens = append(tokens, normalized)
		}
	})
	return tokens
}

func TokenSet(value string) []string {
	seen := make(map[string]struct{})
	for _, token := range Tokens(value) {
		seen[token] = struct{}{}
	}
	return sortedKeys(seen)
}

func scanTokens(value string, yield func([]rune)) {
	segment := make([]rune, 0, 32)
	flush := func() {
		if len(segment) == 0 {
			return
		}
		yieldIdentifierTokens(segment, yield)
		segment = segment[:0]
	}
	for _, current := range value {
		if unicode.IsLetter(current) || unicode.IsDigit(current) {
			segment = append(segment, current)
			continue
		}
		flush()
	}
	flush()
}

func yieldIdentifierTokens(identifier []rune, yield func([]rune)) {
	yield(identifier)
	start := 0
	for index := 1; index < len(identifier); index++ {
		previous := identifier[index-1]
		current := identifier[index]
		nextIsLower := index+1 < len(identifier) && unicode.IsLower(identifier[index+1])
		boundary := unicode.IsLower(previous) && unicode.IsUpper(current)
		boundary = boundary || unicode.IsLetter(previous) && unicode.IsDigit(current)
		boundary = boundary || unicode.IsDigit(previous) && unicode.IsLetter(current)
		boundary = boundary || unicode.IsUpper(previous) && unicode.IsUpper(current) && nextIsLower
		if !boundary {
			continue
		}
		yield(identifier[start:index])
		start = index
	}
	if start > 0 {
		yield(identifier[start:])
	}
}
