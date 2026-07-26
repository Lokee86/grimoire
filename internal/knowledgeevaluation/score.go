package knowledgeevaluation

import (
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/Lokee86/grimoire/internal/knowledge"
)

func ScoreCase(entry Case, response knowledge.SearchResponse, latency time.Duration, recallAtK []int) CaseResult {
	recallAtK = normalizeKs(recallAtK)
	result := CaseResult{
		CaseID:      entry.ID,
		Query:       entry.Query,
		Category:    entry.Category,
		RecallAtK:   make([]RecallMetric, 0, len(recallAtK)),
		VectorUsed:  response.VectorUsed,
		VectorError: response.VectorError,
		LatencyMS:   float64(latency) / float64(time.Millisecond),
		Results:     make([]Result, 0, len(response.Results)),
	}
	if result.LatencyMS < 0 {
		result.LatencyMS = 0
	}

	requiredMatched := make([]bool, len(entry.Required))
	for rank, hit := range response.Results {
		required := matchedExpectation(hit, entry.Required)
		supporting := matchedExpectation(hit, entry.Supporting)
		forbidden := matchedExpectation(hit, entry.Forbidden)
		relevant := required || supporting
		if required {
			markMatches(hit, entry.Required, requiredMatched)
		}
		if relevant {
			result.RelevantSelections++
		}
		if !relevant {
			result.IrrelevantSelections++
		}
		if forbidden {
			result.ForbiddenSelections++
		}
		result.Results = append(result.Results, Result{
			Handle: hit.Handle, Path: hit.Path, Heading: hit.Heading, Score: hit.Score, Rank: rank + 1,
			Required: required, Supporting: supporting, Forbidden: forbidden, Relevant: relevant,
		})
	}
	result.ResultCount = len(result.Results)
	if result.ResultCount > 0 {
		result.IrrelevantSelectionRate = float64(result.IrrelevantSelections) / float64(result.ResultCount)
	}
	for index, expectation := range entry.Required {
		if requiredMatched[index] {
			result.RequiredMatched = append(result.RequiredMatched, expectation)
		} else {
			result.RequiredMissing = append(result.RequiredMissing, expectation)
		}
	}
	if len(entry.Required) > 0 {
		result.RequiredSectionRecall = float64(len(result.RequiredMatched)) / float64(len(entry.Required))
	}
	for _, k := range recallAtK {
		matched := countRequiredAtK(entry.Required, response.Results, k)
		value := 0.0
		if len(entry.Required) > 0 {
			value = float64(matched) / float64(len(entry.Required))
		}
		result.RecallAtK = append(result.RecallAtK, RecallMetric{K: k, Value: value})
	}
	for rank, hit := range response.Results {
		if matchedExpectation(hit, append(append([]SectionExpectation(nil), entry.Required...), entry.Supporting...)) {
			result.FirstRelevantRank = rank + 1
			result.MRR = 1 / float64(rank+1)
			break
		}
	}
	result.Pass = len(result.RequiredMissing) == 0 && result.ForbiddenSelections == 0
	return result
}

func matchedExpectation(hit knowledge.Result, expectations []SectionExpectation) bool {
	for _, expectation := range expectations {
		if matches(hit, expectation) {
			return true
		}
	}
	return false
}

func markMatches(hit knowledge.Result, expectations []SectionExpectation, matched []bool) {
	for index, expectation := range expectations {
		if matches(hit, expectation) {
			matched[index] = true
		}
	}
}

func countRequiredAtK(expectations []SectionExpectation, hits []knowledge.Result, k int) int {
	matched := make([]bool, len(expectations))
	if k > len(hits) {
		k = len(hits)
	}
	for _, hit := range hits[:k] {
		markMatches(hit, expectations, matched)
	}
	count := 0
	for _, value := range matched {
		if value {
			count++
		}
	}
	return count
}

func matches(hit knowledge.Result, expectation SectionExpectation) bool {
	if filepath.ToSlash(hit.Path) != filepath.ToSlash(expectation.Path) {
		return false
	}
	if expectation.SectionID != "" && hit.SectionID != expectation.SectionID {
		return false
	}
	if expectation.Heading != "" && hit.Heading != expectation.Heading {
		return false
	}
	if len(expectation.HeadingPath) > 0 && !sameStrings(hit.HeadingPath, expectation.HeadingPath) {
		return false
	}
	return true
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if strings.TrimSpace(left[index]) != strings.TrimSpace(right[index]) {
			return false
		}
	}
	return true
}

func BuildAggregate(cases []CaseResult, recallAtK []int) Aggregate {
	aggregate := Aggregate{Cases: len(cases), RecallAtK: make([]RecallMetric, 0)}
	if len(cases) == 0 {
		return aggregate
	}
	latencies := make([]float64, 0, len(cases))
	for _, result := range cases {
		if result.Pass {
			aggregate.Passes++
		}
		aggregate.RequiredSectionRecall += result.RequiredSectionRecall
		aggregate.MRR += result.MRR
		aggregate.IrrelevantSelections += result.IrrelevantSelections
		aggregate.VectorUsedCases += boolInt(result.VectorUsed)
		if result.VectorError != "" {
			aggregate.VectorErrorCases++
		}
		latencies = append(latencies, result.LatencyMS)
	}
	aggregate.PassRate = float64(aggregate.Passes) / float64(aggregate.Cases)
	aggregate.RequiredSectionRecall /= float64(aggregate.Cases)
	aggregate.MRR /= float64(aggregate.Cases)
	if total := totalResults(cases); total > 0 {
		aggregate.IrrelevantSelectionRate = float64(aggregate.IrrelevantSelections) / float64(total)
	}
	aggregate.VectorUsageRate = float64(aggregate.VectorUsedCases) / float64(aggregate.Cases)
	for _, k := range normalizeKs(recallAtK) {
		total := 0.0
		for _, result := range cases {
			for _, metric := range result.RecallAtK {
				if metric.K == k {
					total += metric.Value
					break
				}
			}
		}
		aggregate.RecallAtK = append(aggregate.RecallAtK, RecallMetric{K: k, Value: total / float64(len(cases))})
	}
	sortFloat64s(latencies)
	aggregate.MedianLatencyMS = percentile(latencies, 0.50)
	aggregate.P95LatencyMS = percentile(latencies, 0.95)
	return aggregate
}

func totalResults(cases []CaseResult) int {
	total := 0
	for _, result := range cases {
		total += result.ResultCount
	}
	return total
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func percentile(values []float64, fraction float64) float64 {
	if len(values) == 0 {
		return 0
	}
	index := int(math.Ceil(fraction*float64(len(values)))) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(values) {
		index = len(values) - 1
	}
	return values[index]
}

func sortFloat64s(values []float64) {
	for left := 0; left < len(values); left++ {
		for right := left + 1; right < len(values); right++ {
			if values[right] < values[left] {
				values[left], values[right] = values[right], values[left]
			}
		}
	}
}
