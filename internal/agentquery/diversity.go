package agentquery

import "strings"

const maxResultsPerPath = 2

type resultGroup struct {
	path       string
	candidates []Result
	selected   map[int]bool
}

func selectDiverseResults(candidates []Result, limit int) []Result {
	if limit <= 0 || len(candidates) == 0 {
		return nil
	}
	groups := make([]resultGroup, 0)
	groupIndex := make(map[string]int)
	for _, candidate := range candidates {
		path := resultPathKey(candidate)
		index, exists := groupIndex[path]
		if !exists {
			index = len(groups)
			groupIndex[path] = index
			groups = append(groups, resultGroup{path: path, selected: make(map[int]bool)})
		}
		groups[index].candidates = append(groups[index].candidates, candidate)
	}

	selected := make([]Result, 0, min(limit, len(candidates)))
	for groupIndex := range groups {
		group := &groups[groupIndex]
		for _, candidateIndex := range representativeIndexes(group.candidates) {
			group.selected[candidateIndex] = true
			selected = append(selected, group.candidates[candidateIndex])
			if len(selected) == limit {
				return rankResults(selected)
			}
		}
	}
	for groupIndex := range groups {
		group := &groups[groupIndex]
		for candidateIndex, candidate := range group.candidates {
			if group.selected[candidateIndex] {
				continue
			}
			selected = append(selected, candidate)
			if len(selected) == limit {
				return rankResults(selected)
			}
		}
	}
	return rankResults(selected)
}

func representativeIndexes(candidates []Result) []int {
	if len(candidates) == 0 {
		return nil
	}
	indexes := []int{0}
	if len(candidates) == 1 || maxResultsPerPath == 1 {
		return indexes
	}
	bestIndex := 1
	bestPriority := resultBehaviorPriority(candidates[1])
	for index := 2; index < len(candidates); index++ {
		priority := resultBehaviorPriority(candidates[index])
		if priority < bestPriority {
			bestIndex = index
			bestPriority = priority
		}
	}
	return append(indexes, bestIndex)
}

func resultBehaviorPriority(result Result) int {
	kind := strings.ToLower(strings.TrimSpace(result.Kind))
	if kind == "" {
		kind = strings.ToLower(strings.TrimSpace(result.Node.Kind))
	}
	switch kind {
	case "function", "test", "handler", "constructor":
		return 0
	case "method":
		return 1
	case "type", "struct", "class", "interface", "enum", "constant":
		return 2
	default:
		return 3
	}
}

func resultPathKey(result Result) string {
	path := strings.ToLower(strings.TrimSpace(result.Node.Path))
	if path == "" {
		path = handleKey(result.Node.Handle)
	}
	return path
}

func rankResults(results []Result) []Result {
	for index := range results {
		results[index].Rank = index + 1
	}
	return results
}

func resultSemanticKey(result Result) string {
	parts := []string{
		result.Kind,
		result.Node.Kind,
		result.Node.QualifiedName,
		result.Node.Name,
		result.Node.Path,
	}
	return strings.ToLower(strings.Join(parts, "\x00"))
}
