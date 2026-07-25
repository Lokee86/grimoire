package knowledge

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/Lokee86/grimoire/internal/lexical"
)

func extractSections(path, content string) []Section {
	lines := lineRanges(content)
	type heading struct {
		line, start, level int
		title              string
	}
	headings := make([]heading, 0)
	for _, line := range lines {
		if level, title, ok := markdownHeading(line.text); ok {
			headings = append(headings, heading{line: line.number, start: line.start, level: level, title: title})
		}
	}
	if len(headings) == 0 {
		return []Section{makeSection(path, content, 0, len(content), 1, len(lines), "", nil, stableID(path, "document"))}
	}

	sections := make([]Section, 0, len(headings))
	if headings[0].start > 0 {
		sections = append(sections, makeSection(path, content, 0, headings[0].start, 1, headings[0].line-1, "", nil, stableID(path, "preamble")))
	}
	pathStack := make([]string, 0, 6)
	occurrences := make(map[string]int)
	for index, current := range headings {
		end := len(content)
		endLine := len(lines)
		if index+1 < len(headings) {
			end = headings[index+1].start
			endLine = headings[index+1].line - 1
		}
		for len(pathStack) >= current.level {
			pathStack = pathStack[:len(pathStack)-1]
		}
		pathStack = append(pathStack, current.title)
		headingPath := append([]string(nil), pathStack...)
		key := strings.Join(headingPath, "\x1f")
		occurrences[key]++
		sectionID := stableID(path, key, fmt.Sprint(occurrences[key]))
		sections = append(sections, makeSection(path, content, current.start, end, current.line, endLine, current.title, headingPath, sectionID))
	}
	return sections
}

// ExtractSections exposes deterministic Markdown section extraction for callers
// that need to inspect spans without building repository state.
func ExtractSections(path, content string) []Section {
	return extractSections(path, content)
}

type lineRange struct {
	number int
	start  int
	text   string
}

func lineRanges(content string) []lineRange {
	if content == "" {
		return nil
	}
	lines := make([]lineRange, 0, strings.Count(content, "\n")+1)
	start, number := 0, 1
	for start < len(content) {
		end := strings.IndexByte(content[start:], '\n')
		if end < 0 {
			end = len(content)
		} else {
			end += start + 1
		}
		textEnd := end
		if textEnd > start && content[textEnd-1] == '\n' {
			textEnd--
		}
		if textEnd > start && content[textEnd-1] == '\r' {
			textEnd--
		}
		lines = append(lines, lineRange{number: number, start: start, text: content[start:textEnd]})
		start, number = end, number+1
	}
	return lines
}

func markdownHeading(line string) (int, string, bool) {
	trimmed := strings.TrimLeft(line, " \t")
	level := 0
	for level < len(trimmed) && trimmed[level] == '#' {
		level++
	}
	if level == 0 || level > 6 || len(trimmed) <= level || trimmed[level] != ' ' {
		return 0, "", false
	}
	title := strings.TrimSpace(trimmed[level:])
	title = strings.TrimRight(title, "#")
	title = strings.TrimSpace(title)
	if title == "" {
		return 0, "", false
	}
	return level, title, true
}

func makeSection(path, content string, start, end, startLine, endLine int, heading string, headingPath []string, id string) Section {
	text := content[start:end]
	if endLine < startLine {
		endLine = startLine
	}
	return Section{
		ID: id, Heading: heading, HeadingPath: headingPath,
		StartByte: start, EndByte: end, StartLine: startLine, EndLine: endLine,
		Hash: hashBytes([]byte(text)), Text: text, Terms: termFrequencies(text),
	}
}

func termFrequencies(text string) []TermFrequency {
	counts := make(map[string]int)
	for _, term := range lexical.Tokens(text) {
		counts[term]++
	}
	terms := make([]TermFrequency, 0, len(counts))
	for term, frequency := range counts {
		terms = append(terms, TermFrequency{Term: term, Frequency: frequency})
	}
	for left := 0; left < len(terms); left++ {
		for right := left + 1; right < len(terms); right++ {
			if terms[right].Term < terms[left].Term {
				terms[left], terms[right] = terms[right], terms[left]
			}
		}
	}
	return terms
}

func stableID(parts ...string) string {
	return hashBytes([]byte(strings.Join(parts, "\x00")))[:16]
}

func hashBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}
