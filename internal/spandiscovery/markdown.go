package spandiscovery

import "strings"

type markdownHeading struct {
	line  int
	level int
	name  string
}

func discoverMarkdown(lines []string, language string) []Span {
	headings := make([]markdownHeading, 0)
	fence := ""
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if marker := fenceMarker(trimmed); marker != "" {
			if fence == "" {
				fence = marker
			} else if marker == fence {
				fence = ""
			}
			continue
		}
		if fence != "" {
			continue
		}
		level, name, ok := markdownHeadingLine(line)
		if ok {
			headings = append(headings, markdownHeading{line: index + 1, level: level, name: name})
		}
	}

	spans := make([]Span, 0, len(headings))
	for index, heading := range headings {
		end := len(lines)
		for next := index + 1; next < len(headings); next++ {
			if headings[next].level <= heading.level {
				end = headings[next].line - 1
				break
			}
		}
		spans = append(spans, Span{
			StartLine: heading.line,
			EndLine:   end,
			Kind:      KindSection,
			Name:      heading.name,
			Language:  language,
		})
	}
	return spans
}

func markdownHeadingLine(line string) (int, string, bool) {
	if len(line)-len(strings.TrimLeft(line, " ")) > 3 {
		return 0, "", false
	}
	trimmed := strings.TrimLeft(line, " ")
	level := 0
	for level < len(trimmed) && trimmed[level] == '#' {
		level++
	}
	if level == 0 || level > 6 || level >= len(trimmed) || trimmed[level] != ' ' {
		return 0, "", false
	}
	name := strings.TrimSpace(trimmed[level:])
	name = strings.TrimSpace(strings.TrimRight(name, "#"))
	return level, name, name != ""
}

func fenceMarker(line string) string {
	if strings.HasPrefix(line, "```") {
		return "```"
	}
	if strings.HasPrefix(line, "~~~") {
		return "~~~"
	}
	return ""
}

func discoverTOMLSections(lines []string, language string) []Span {
	type section struct {
		line int
		name string
	}
	sections := make([]section, 0)
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) < 3 || !strings.HasPrefix(trimmed, "[") || !strings.HasSuffix(trimmed, "]") {
			continue
		}
		name := strings.TrimSpace(strings.Trim(trimmed, "[]"))
		if name == "" {
			continue
		}
		sections = append(sections, section{line: index + 1, name: name})
	}

	spans := make([]Span, 0, len(sections))
	for index, current := range sections {
		end := len(lines)
		if index+1 < len(sections) {
			end = sections[index+1].line - 1
		}
		spans = append(spans, Span{
			StartLine: current.line,
			EndLine:   end,
			Kind:      KindSection,
			Name:      current.name,
			Language:  language,
		})
	}
	return spans
}
