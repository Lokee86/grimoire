package queryshape

import (
	"sort"
	"strings"

	"github.com/Lokee86/grimoire/internal/evidence"
)

type markdownSection struct {
	Heading string
	Lines   []string
	Order   int
}

type scoredSectionLine struct {
	Text  string
	Score int
	Order int
}

func markdownRetrievalClauses(query string) ([]retrievalClause, bool) {
	cleaned := stripFencedBlocks(query)
	sections := parseMarkdownSections(cleaned)
	if len(sections) < 2 {
		return nil, false
	}
	clauses := make([]retrievalClause, 0, len(sections))
	for _, section := range sections {
		clause, ok := sectionRetrievalClause(section)
		if ok {
			clauses = append(clauses, clause)
		}
	}
	return clauses, true
}

func parseMarkdownSections(query string) []markdownSection {
	var sections []markdownSection
	var current *markdownSection
	for _, raw := range strings.Split(query, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if level, heading, ok := markdownHeading(line); ok && level <= 2 {
			sections = append(sections, markdownSection{Heading: cleanSectionHeading(heading), Order: len(sections)})
			current = &sections[len(sections)-1]
			continue
		}
		plain := cleanMarkdownLine(line)
		if plain == "" {
			continue
		}
		if current == nil {
			sections = append(sections, markdownSection{Order: len(sections)})
			current = &sections[len(sections)-1]
		}
		current.Lines = append(current.Lines, plain)
	}
	return sections
}

func sectionRetrievalClause(section markdownSection) (retrievalClause, bool) {
	heading := cleanClauseText(section.Heading)
	if heading == "" && len(section.Lines) == 0 {
		return retrievalClause{}, false
	}
	if excludedSectionHeading(heading) {
		return retrievalClause{}, false
	}
	lines := make([]scoredSectionLine, 0, len(section.Lines))
	for index, line := range section.Lines {
		line = compactWords(line, 32)
		score := scoreRetrievalClause(line)
		if score <= 0 || isSectionMetadata(line) {
			continue
		}
		lines = append(lines, scoredSectionLine{Text: line, Score: score, Order: index})
	}
	sort.SliceStable(lines, func(left, right int) bool {
		if lines[left].Score != lines[right].Score {
			return lines[left].Score > lines[right].Score
		}
		return lines[left].Order < lines[right].Order
	})
	if len(lines) > 3 {
		lines = lines[:3]
	}
	sort.SliceStable(lines, func(left, right int) bool { return lines[left].Order < lines[right].Order })

	parts := make([]string, 0, 1+len(lines))
	if heading != "" {
		parts = append(parts, heading)
	}
	for _, line := range lines {
		parts = append(parts, line.Text)
	}
	query := compactWords(strings.Join(parts, ": "), 56)
	if len(strings.Fields(query)) < 3 {
		return retrievalClause{}, false
	}
	score := scoreRetrievalClause(query) + headingPriority(heading)
	return retrievalClause{
		Query: query, Intent: classifySectionIntent(heading, query), Topic: clauseTopic(query),
		Score: score, Order: section.Order,
	}, true
}

func stripFencedBlocks(query string) string {
	var result []string
	inFence := false
	fence := ""
	for _, line := range strings.Split(query, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			marker := trimmed[:3]
			if !inFence {
				inFence = true
				fence = marker
			} else if marker == fence {
				inFence = false
				fence = ""
			}
			continue
		}
		if !inFence {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

func looksStructuredQuery(query string) bool {
	if strings.Contains(query, "```") || strings.Contains(query, "~~~") {
		return true
	}
	headings := 0
	for _, line := range strings.Split(query, "\n") {
		if level, _, ok := markdownHeading(strings.TrimSpace(line)); ok && level <= 2 {
			headings++
		}
	}
	return headings >= 2
}

func markdownHeading(line string) (int, string, bool) {
	if !strings.HasPrefix(line, "#") {
		return 0, "", false
	}
	level := 0
	for level < len(line) && line[level] == '#' {
		level++
	}
	if level == len(line) || line[level] != ' ' {
		return 0, "", false
	}
	return level, strings.TrimSpace(line[level:]), true
}

func cleanSectionHeading(heading string) string {
	heading = cleanMarkdownLine(heading)
	lower := strings.ToLower(heading)
	if strings.HasPrefix(lower, "phase ") {
		for _, separator := range []string{" — ", " – ", " - ", ": "} {
			if index := strings.Index(heading, separator); index >= 0 {
				return strings.TrimSpace(heading[index+len(separator):])
			}
		}
	}
	return heading
}

func cleanMarkdownLine(line string) string {
	line = strings.TrimSpace(line)
	line = strings.TrimLeft(line, ">")
	line = strings.TrimSpace(line)
	for _, prefix := range []string{"- ", "* ", "+ "} {
		if strings.HasPrefix(line, prefix) {
			line = strings.TrimSpace(strings.TrimPrefix(line, prefix))
			break
		}
	}
	if dot := strings.Index(line, ". "); dot > 0 && dot <= 3 {
		numeric := true
		for _, r := range line[:dot] {
			if r < '0' || r > '9' {
				numeric = false
				break
			}
		}
		if numeric {
			line = strings.TrimSpace(line[dot+2:])
		}
	}
	line = strings.NewReplacer("**", "", "__", "", "`", "").Replace(line)
	return strings.TrimSpace(line)
}

func excludedSectionHeading(heading string) bool {
	lower := strings.ToLower(strings.TrimSpace(heading))
	if lower == "goal" || containsAnyText(lower, "deliverables", "completion criteria", "initial categories") {
		return true
	}
	return strings.HasPrefix(lower, "build the ") &&
		(strings.Contains(lower, " corpus") || strings.Contains(lower, " snapshot")) &&
		!strings.Contains(lower, "evaluation")
}

func headingPriority(heading string) int {
	lower := strings.ToLower(heading)
	score := 0
	if actionCueIndex(strings.Fields(heading)) >= 0 {
		score += 4
	}
	if containsAnyText(lower, "evaluation", "runner", "scoring", "report", "baseline", "contract", "format", "corpus") {
		score += 4
	}
	if containsAnyText(lower, "evaluation contract", "evaluation format", "corpus model") {
		score += 6
	}
	if strings.HasPrefix(lower, "build the ") && strings.Contains(lower, " corpus") && !strings.Contains(lower, "evaluation") {
		score -= 4
	}
	if containsAnyText(lower, "goal", "initial categories", "deliverables", "completion criteria") {
		score -= 10
	}
	return score
}

func classifySectionIntent(heading, query string) evidence.Intent {
	lower := strings.ToLower(heading)
	switch {
	case containsAnyText(lower, "architecture", "ownership", "boundary"):
		return evidence.IntentArchitecture
	case containsAnyText(lower, "trace", "call chain", "execution flow"):
		return evidence.IntentCallChain
	case containsAnyText(lower, "where", "locate", "find"):
		return evidence.IntentDirectLocation
	case actionCueIndex(strings.Fields(heading)) >= 0:
		return evidence.IntentMechanism
	default:
		return classifyClauseIntent(query)
	}
}

func isSectionMetadata(line string) bool {
	lower := strings.ToLower(strings.TrimSpace(line))
	return containsAnyText(lower,
		"suggested location", "suggested output", "initial categories",
		"deliverables", "completion criteria", "the evaluation must answer")
}
