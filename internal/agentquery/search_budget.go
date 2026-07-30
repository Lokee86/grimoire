package agentquery

import "strings"

const defaultLanePreviewCount = 2

func applyNarrowSearchBudget(response *Response, limit int) map[string]int {
	suppressed := make(map[string]int)
	if response == nil || limit <= 0 {
		return suppressed
	}

	type lane struct {
		name     string
		results  []Result
		selected []Result
		index    int
	}
	lanes := []lane{
		{name: "exact_matches", results: response.ExactMatches},
		{name: "symbol_matches", results: response.SymbolMatches},
		{name: "source_matches", results: response.SourceMatches},
	}
	selected := make([]Result, 0, limit)
	for len(selected) < limit {
		progressed := false
		for laneIndex := range lanes {
			current := &lanes[laneIndex]
			for current.index < len(current.results) {
				candidate := current.results[current.index]
				current.index++
				if narrowDuplicate(candidate, selected) {
					suppressed[current.name]++
					continue
				}
				current.selected = append(current.selected, candidate)
				selected = append(selected, candidate)
				progressed = true
				break
			}
			if len(selected) == limit {
				break
			}
		}
		if !progressed {
			break
		}
	}

	response.ExactMatches = rankResults(lanes[0].selected)
	response.SymbolMatches = rankResults(lanes[1].selected)
	response.SourceMatches = rankResults(lanes[2].selected)
	return suppressed
}

func narrowDuplicate(candidate Result, selected []Result) bool {
	candidatePath := strings.ToLower(strings.TrimSpace(candidate.Node.Path))
	candidateLabel := narrowResultLabel(candidate)
	for _, existing := range selected {
		if handleKey(candidate.Node.Handle) == handleKey(existing.Node.Handle) {
			return true
		}
		if candidatePath == "" || candidatePath != strings.ToLower(strings.TrimSpace(existing.Node.Path)) {
			continue
		}
		existingLabel := narrowResultLabel(existing)
		if candidate.Node.Span == nil || existing.Node.Span == nil {
			if candidateLabel != "" && candidateLabel == existingLabel {
				return true
			}
			continue
		}
		candidateSpan := candidate.Node.Span
		existingSpan := existing.Node.Span
		overlaps := candidateSpan.StartLine <= existingSpan.EndLine && existingSpan.StartLine <= candidateSpan.EndLine
		if !overlaps {
			continue
		}
		if candidateSpan.StartLine == existingSpan.StartLine && candidateSpan.EndLine == existingSpan.EndLine {
			return true
		}
		if candidateLabel != "" && candidateLabel == existingLabel {
			return true
		}
	}
	return false
}

func narrowResultLabel(result Result) string {
	label := result.Node.QualifiedName
	if label == "" {
		label = result.Node.Name
	}
	return strings.ToLower(strings.TrimSpace(label))
}

func applyResultPreviews(results []Result, detail string) int {
	detail = strings.ToLower(strings.TrimSpace(detail))
	if detail == "full" {
		return len(results)
	}
	if detail == "handles" {
		for index := range results {
			results[index].Excerpt = ""
			results[index].Node.Span = nil
		}
		return 0
	}
	previewed := min(defaultLanePreviewCount, len(results))
	for index := previewed; index < len(results); index++ {
		results[index].Excerpt = ""
	}
	return previewed
}

// recordLaneCoverage reports provider-local selection without comparing scores
// across independent retrieval lanes.
func recordLaneCoverage(response *Response, lane string, available, returned, previewed, suppressedDuplicates int) {
	if response == nil {
		return
	}
	if available < returned {
		available = returned
	}
	if previewed > returned {
		previewed = returned
	}
	deferred := available - returned
	response.Coverage = append(response.Coverage, LaneCoverage{
		Lane:                 lane,
		Available:            available,
		Returned:             returned,
		Previewed:            previewed,
		Deferred:             deferred,
		SuppressedDuplicates: suppressedDuplicates,
	})
	markLaneTruncated(response, lane, deferred > 0)
}

func deferRelationshipExpansion(response *Response, candidateCount int) {
	if response == nil || candidateCount <= 0 {
		return
	}
	response.DeferredExpansions = append(response.DeferredExpansions, DeferredExpansion{
		Kind:           "relationships",
		CandidateCount: candidateCount,
		FollowUpModes:  []string{"trace", "impact"},
		Reason:         "broad search returns ranked discovery evidence; expand a selected handle explicitly",
	})
}
