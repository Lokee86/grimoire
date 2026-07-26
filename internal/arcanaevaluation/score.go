package arcanaevaluation

import (
	"encoding/json"
	"math"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Lokee86/grimoire/internal/evaluation"
	"github.com/Lokee86/grimoire/internal/structure"
)

func ScoreCase(entry Case, measurement Measurement, recallAtK []int) CaseResult {
	result := CaseResult{
		CaseID: entry.ID, Query: entry.Query, Category: entry.Category, Mode: measurement.Mode,
		Error: measurement.Error, VectorUsed: measurement.VectorUsed, ProviderCalls: measurement.ProviderCalls,
		Timings: measurement.Timings, Seeds: make([]SeedResult, 0, len(measurement.Seeds)),
		RecallAtK:          make([]RecallMetric, 0, len(recallAtK)),
		StructuralEvidence: append([]structure.Evidence(nil), measurement.Structural...),
	}
	if result.Timings.TotalMS > 0 && result.Timings.TotalMS < 0.001 {
		result.Timings.TotalMS = 0.001
	}
	seedPayload, _ := json.Marshal(measurement.Seeds)
	structuralPayload, _ := json.Marshal(measurement.Structural)
	result.SeedPayloadBytes = len(seedPayload)
	result.StructuralPayloadBytes = len(structuralPayload)
	result.PayloadBytes = result.SeedPayloadBytes + result.StructuralPayloadBytes

	requiredMatched := make([]bool, len(entry.RequiredSeeds))
	for rank, seed := range measurement.Seeds {
		required := seedMatchesAny(seed.Node, entry.RequiredSeeds)
		supporting := seedMatchesAny(seed.Node, entry.SupportingSeeds)
		if required {
			markSeedMatches(seed.Node, entry.RequiredSeeds, requiredMatched)
		}
		if result.FirstRequiredSeedRank == 0 && required {
			result.FirstRequiredSeedRank = rank + 1
			result.MRR = 1 / float64(rank+1)
		}
		result.Seeds = append(result.Seeds, SeedResult{
			Rank: rank + 1, Source: seed.Source, Node: seed.Node,
			Required: required, Supporting: supporting, Relevant: required || supporting,
		})
	}
	for index, expectation := range entry.RequiredSeeds {
		if requiredMatched[index] {
			result.RequiredSeedsMatched = append(result.RequiredSeedsMatched, expectation)
		} else {
			result.RequiredSeedsMissing = append(result.RequiredSeedsMissing, expectation)
		}
	}
	result.RequiredSeedRecall = ratio(len(result.RequiredSeedsMatched), len(entry.RequiredSeeds))
	for _, k := range normalizeKs(recallAtK) {
		result.RecallAtK = append(result.RecallAtK, RecallMetric{
			K: k, Value: ratio(countRequiredSeedsAtK(entry.RequiredSeeds, measurement.Seeds, k), len(entry.RequiredSeeds)),
		})
	}

	for _, expected := range entry.RequiredStructural {
		matched := structuralMatchesAny(expected, measurement.Structural)
		result.StructuralJudgments = append(result.StructuralJudgments, StructuralJudgment{Expectation: expected, Matched: matched})
		if matched {
			result.RequiredStructuralMatched++
		} else {
			result.RequiredStructuralMissing++
		}
	}
	result.RequiredStructuralRecall = ratio(result.RequiredStructuralMatched, len(entry.RequiredStructural))
	result.Pass = result.Error == "" && len(result.RequiredSeedsMissing) == 0 && result.RequiredStructuralMissing == 0
	return result
}

func seedMatchesAny(node structure.Node, expectations []SeedExpectation) bool {
	for _, expectation := range expectations {
		if seedMatches(node, expectation) {
			return true
		}
	}
	return false
}

func seedMatches(node structure.Node, expectation SeedExpectation) bool {
	if expectation.Kind != "" && !strings.EqualFold(strings.TrimSpace(node.Kind), strings.TrimSpace(expectation.Kind)) {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(node.Name), strings.TrimSpace(expectation.Name)) &&
		!strings.EqualFold(strings.TrimSpace(node.QualifiedName), strings.TrimSpace(expectation.Name)) {
		return false
	}
	path := node.Path
	if path == "" && node.Span != nil {
		path = node.Span.Path
	}
	return filepath.ToSlash(path) == filepath.ToSlash(expectation.Path)
}

func markSeedMatches(node structure.Node, expectations []SeedExpectation, matched []bool) {
	for index, expectation := range expectations {
		if seedMatches(node, expectation) {
			matched[index] = true
		}
	}
}

func countRequiredSeedsAtK(expectations []SeedExpectation, seeds []RankedSeed, k int) int {
	matched := make([]bool, len(expectations))
	if k > len(seeds) {
		k = len(seeds)
	}
	for _, seed := range seeds[:k] {
		markSeedMatches(seed.Node, expectations, matched)
	}
	count := 0
	for _, value := range matched {
		if value {
			count++
		}
	}
	return count
}

func structuralMatchesAny(expected evaluation.StructuralExpectation, evidence []structure.Evidence) bool {
	for _, item := range evidence {
		if evaluation.StructuralExpectationMatches(expected, item) {
			return true
		}
	}
	return false
}

func BuildAggregates(report *Report) {
	groups := map[string][]CaseResult{
		ModeLexiconSeeds:       nil,
		ModeLexiconVectorSeeds: nil,
	}
	for _, result := range report.Cases {
		groups[result.Mode] = append(groups[result.Mode], result)
	}
	report.Aggregates = make([]Aggregate, 0, 2)
	for _, mode := range []string{ModeLexiconSeeds, ModeLexiconVectorSeeds} {
		report.Aggregates = append(report.Aggregates, BuildAggregate(mode, groups[mode], report.RecallAtK))
	}
	report.Comparison = Compare(report.Aggregates[0], report.Aggregates[1])
}

func BuildAggregate(mode string, cases []CaseResult, recallAtK []int) Aggregate {
	aggregate := Aggregate{Mode: mode, Cases: len(cases)}
	if len(cases) == 0 {
		return aggregate
	}
	latencies := make([]float64, 0, len(cases))
	payloads := make([]float64, 0, len(cases))
	for _, result := range cases {
		if result.Pass {
			aggregate.Passes++
		}
		if result.Error != "" {
			aggregate.ErrorCases++
		}
		if result.VectorUsed {
			aggregate.VectorUsedCases++
		}
		aggregate.RequiredSeedRecall += result.RequiredSeedRecall
		aggregate.MRR += result.MRR
		aggregate.RequiredStructuralRecall += result.RequiredStructuralRecall
		aggregate.MeanProviderCalls += float64(result.ProviderCalls)
		latencies = append(latencies, result.Timings.TotalMS)
		payloads = append(payloads, float64(result.PayloadBytes))
	}
	count := float64(len(cases))
	aggregate.PassRate = float64(aggregate.Passes) / count
	aggregate.RequiredSeedRecall /= count
	aggregate.MRR /= count
	aggregate.RequiredStructuralRecall /= count
	aggregate.MeanProviderCalls /= count
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
		aggregate.RecallAtK = append(aggregate.RecallAtK, RecallMetric{K: k, Value: total / count})
	}
	sort.Float64s(latencies)
	sort.Float64s(payloads)
	aggregate.MedianLatencyMS = percentile(latencies, 0.50)
	aggregate.P95LatencyMS = percentile(latencies, 0.95)
	aggregate.MedianPayloadBytes = percentile(payloads, 0.50)
	aggregate.P95PayloadBytes = percentile(payloads, 0.95)
	return aggregate
}

func Compare(baseline, vector Aggregate) Comparison {
	comparison := Comparison{
		BaselineMode: baseline.Mode, VectorMode: vector.Mode,
		PassRateDelta:                 vector.PassRate - baseline.PassRate,
		RequiredSeedRecallDelta:       vector.RequiredSeedRecall - baseline.RequiredSeedRecall,
		MRRDelta:                      vector.MRR - baseline.MRR,
		RequiredStructuralRecallDelta: vector.RequiredStructuralRecall - baseline.RequiredStructuralRecall,
		MedianLatencyMSDelta:          vector.MedianLatencyMS - baseline.MedianLatencyMS,
		MedianPayloadBytesDelta:       vector.MedianPayloadBytes - baseline.MedianPayloadBytes,
		MeanProviderCallsDelta:        vector.MeanProviderCalls - baseline.MeanProviderCalls,
	}
	for _, vectorMetric := range vector.RecallAtK {
		baselineValue := 0.0
		for _, baselineMetric := range baseline.RecallAtK {
			if baselineMetric.K == vectorMetric.K {
				baselineValue = baselineMetric.Value
				break
			}
		}
		comparison.RecallAtKDelta = append(comparison.RecallAtKDelta, RecallMetric{K: vectorMetric.K, Value: vectorMetric.Value - baselineValue})
	}
	return comparison
}

func ratio(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
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
