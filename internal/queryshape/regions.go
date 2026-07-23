package queryshape

import (
	"math"
	"path/filepath"
	"strings"

	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/structure"
)

const candidateSampleLimit = 20

func candidateRegions(candidates []retrieve.Candidate) []string {
	limit := min(len(candidates), candidateSampleLimit)
	seen := make(map[string]struct{}, limit)
	var regions []string
	for _, candidate := range candidates[:limit] {
		region := PathRegion(candidate.Chunk.Path)
		if region == "" {
			continue
		}
		if _, exists := seen[region]; exists {
			continue
		}
		seen[region] = struct{}{}
		regions = append(regions, region)
	}
	return regions
}

func structuralRegions(evidence []structure.Evidence) []string {
	seen := make(map[string]struct{})
	var regions []string
	add := func(path string) {
		region := PathRegion(path)
		if region == "" {
			return
		}
		if _, exists := seen[region]; exists {
			return
		}
		seen[region] = struct{}{}
		regions = append(regions, region)
	}
	for _, item := range evidence {
		if item.Node != nil {
			add(item.Node.Path)
		}
		for _, relation := range item.Relationships {
			add(relation.Node.Path)
		}
		for _, dependent := range item.Dependents {
			add(dependent.Node.Path)
		}
		if item.Chain != nil {
			for _, node := range item.Chain.Nodes {
				add(node.Path)
			}
		}
	}
	return regions
}

// PathRegion returns the stable repository region used by query and assembly analysis.
func PathRegion(path string) string {
	path = filepath.ToSlash(strings.TrimSpace(path))
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return ""
	}
	if len(parts) > 1 {
		switch parts[0] {
		case "cmd", "docs", "internal", "native", "pkg", "src", "test", "tests":
			return parts[0] + "/" + parts[1]
		}
	}
	return parts[0]
}

func topScoreGap(candidates []retrieve.Candidate) float64 {
	if len(candidates) < 2 {
		if len(candidates) == 1 {
			return 1
		}
		return 0
	}
	denominator := math.Max(math.Abs(candidates[0].Score), 1)
	gap := (candidates[0].Score - candidates[1].Score) / denominator
	return math.Max(0, math.Min(1, gap))
}

func candidateDispersion(candidates []retrieve.Candidate) float64 {
	limit := min(len(candidates), candidateSampleLimit)
	if limit < 2 {
		return 0
	}
	regions := make(map[string]struct{}, limit)
	for _, candidate := range candidates[:limit] {
		if region := PathRegion(candidate.Chunk.Path); region != "" {
			regions[region] = struct{}{}
		}
	}
	if len(regions) < 2 {
		return 0
	}
	return float64(len(regions)-1) / float64(limit-1)
}
