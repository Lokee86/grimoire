package lexical

import (
	"path/filepath"
	"sort"
	"strings"
)

type Input struct {
	Key  string
	Path string
	Text string
}

type TermFrequency struct {
	Term      string `json:"term"`
	Frequency int    `json:"frequency"`
}

type Document struct {
	Key               string          `json:"key"`
	Length            int             `json:"length"`
	Terms             []TermFrequency `json:"terms"`
	BaseTokens        []string        `json:"base_tokens"`
	PathTokens        []string        `json:"path_tokens"`
	LeadingTokens     []string        `json:"leading_tokens"`
	DeclarationTokens []string        `json:"declaration_tokens"`
}

func Analyze(input Input) Document {
	frequencies := make(map[string]int)
	length := 0
	scanTokens(input.Text, func(token []rune) {
		length++
		frequencies[strings.ToLower(string(token))]++
	})
	terms := make([]TermFrequency, 0, len(frequencies))
	for term, frequency := range frequencies {
		terms = append(terms, TermFrequency{Term: term, Frequency: frequency})
	}
	sort.Slice(terms, func(left, right int) bool { return terms[left].Term < terms[right].Term })

	firstLine := input.Text
	if newline := strings.IndexByte(firstLine, '\n'); newline >= 0 {
		firstLine = firstLine[:newline]
	}
	return Document{
		Key:               input.Key,
		Length:            length,
		Terms:             terms,
		BaseTokens:        TokenSet(filepath.Base(input.Path)),
		PathTokens:        TokenSet(input.Path),
		LeadingTokens:     TokenSet(firstLine),
		DeclarationTokens: TokenSet(filepath.Base(input.Path) + "\n" + input.Path + "\n" + declarationHeader(input.Text)),
	}
}

func declarationHeader(text string) string {
	const maxLines = 6
	const maxBytes = 768
	lines := strings.Split(text, "\n")
	var header strings.Builder
	linesAdded := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || commentOnlyLine(line) {
			continue
		}
		if header.Len() > 0 {
			header.WriteByte('\n')
		}
		header.WriteString(line)
		linesAdded++
		if linesAdded >= maxLines || header.Len() >= maxBytes {
			break
		}
	}
	value := header.String()
	if len(value) > maxBytes {
		value = value[:maxBytes]
	}
	return value
}

func commentOnlyLine(line string) bool {
	for _, prefix := range []string{"//", "#", "--", "/*", "*", "<!--"} {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}

func sortedKeys(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
