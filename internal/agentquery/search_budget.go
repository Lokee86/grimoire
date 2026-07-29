package agentquery

import "strings"

const defaultLanePreviewCount = 2

func applyResultPreviews(results []Result, detail string) int {
	if strings.EqualFold(strings.TrimSpace(detail), "full") {
		return len(results)
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
