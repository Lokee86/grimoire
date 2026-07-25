package graphrank

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

const graphScorePerPromotion = 4.0

// Config controls the maximum number of positions a relationship-derived
// candidate may move. Zero values keep scoring in shadow mode.
type Config struct {
	MechanismMaxPromotion    float64
	CallChainMaxPromotion    float64
	ArchitectureMaxPromotion float64
}

// DefaultConfig retains graph diagnostics without changing candidate order.
// Promotion remains shadow-only until a corpus demonstrates a net gain.
func DefaultConfig() Config {
	return Config{}
}

// BoundedConfig exposes the conservative calibration tested by this stream.
// It is intentionally opt-in because the current corpus showed no ranking gain.
func BoundedConfig() Config {
	return Config{
		MechanismMaxPromotion:    4,
		CallChainMaxPromotion:    3,
		ArchitectureMaxPromotion: 2,
	}
}

type evaluatedCandidate struct {
	candidate  retrieve.Candidate
	position   int
	sortKey    float64
	graphScore float64
}

// Rerank records inspectable graph-derived score details in shadow mode.
func Rerank(candidates []retrieve.Candidate, intent evidence.Intent) []retrieve.Candidate {
	return RerankWithConfig(candidates, intent, DefaultConfig())
}

// RerankWithConfig can apply an explicitly calibrated bounded promotion to
// relationship-derived candidates. Direct structural seeds never reorder source
// retrieval, and exact candidates remain pinned ahead of graph-inferred candidates.
func RerankWithConfig(candidates []retrieve.Candidate, intent evidence.Intent, config Config) []retrieve.Candidate {
	if len(candidates) == 0 {
		return nil
	}
	exact := make([]retrieve.Candidate, 0)
	ranked := make([]evaluatedCandidate, 0, len(candidates))
	for position, source := range candidates {
		candidate := cloneCandidate(source)
		details := Score(candidate.Context, intent)
		graphScore := 0.0
		for _, detail := range details {
			graphScore += detail.Value
		}
		candidate.GraphScoreDetails = append(candidate.GraphScoreDetails, details...)
		if candidate.Source == "exact" {
			exact = append(exact, candidate)
			continue
		}
		promotion := graphPromotion(candidate.Context, intent, graphScore, config)
		ranked = append(ranked, evaluatedCandidate{
			candidate:  candidate,
			position:   position,
			sortKey:    float64(position) - promotion,
			graphScore: graphScore,
		})
	}
	sort.SliceStable(ranked, func(left, right int) bool {
		if ranked[left].sortKey != ranked[right].sortKey {
			return ranked[left].sortKey < ranked[right].sortKey
		}
		if ranked[left].graphScore != ranked[right].graphScore {
			return ranked[left].graphScore > ranked[right].graphScore
		}
		return ranked[left].position < ranked[right].position
	})
	result := make([]retrieve.Candidate, 0, len(candidates))
	result = append(result, exact...)
	for _, item := range ranked {
		result = append(result, item.candidate)
	}
	for index := range result {
		result[index].Rank = index + 1
	}
	return result
}

// Score returns independent graph contributions so evaluation can attribute
// each signal instead of relying on one opaque combined graph score.
func Score(descriptor *evidence.Descriptor, intent evidence.Intent) []retrieve.ScoreDetail {
	if descriptor == nil || descriptor.Graph == nil {
		return nil
	}
	graph := descriptor.Graph
	details := make([]retrieve.ScoreDetail, 0, 5)
	if value := distanceScore(graph.Distance); value > 0 {
		details = append(details, retrieve.ScoreDetail{
			Name:  fmt.Sprintf("graph distance from structural seed %d", graph.Distance),
			Value: value,
		})
	}
	if relation, value := relationScore(graph.Relations, intent); value > 0 {
		details = append(details, retrieve.ScoreDetail{
			Name:  fmt.Sprintf("graph relation %s for %s", relation, intent),
			Value: value,
		})
	}
	if value := clamp(graph.ModuleProximity, 0, 1) * 3; value > 0 {
		details = append(details, retrieve.ScoreDetail{Name: "graph module proximity", Value: value})
	}
	if value := symbolRoleScore(graph.SymbolRole, intent); value > 0 {
		details = append(details, retrieve.ScoreDetail{
			Name:  "graph symbol role " + strings.ToLower(strings.TrimSpace(graph.SymbolRole)),
			Value: value,
		})
	}
	if value := clamp(graph.Centrality, 0, 1) * 1.5; value > 0 {
		details = append(details, retrieve.ScoreDetail{Name: "graph weak centrality", Value: value})
	}
	return details
}

func graphPromotion(descriptor *evidence.Descriptor, intent evidence.Intent, score float64, config Config) float64 {
	if descriptor == nil || descriptor.Graph == nil || descriptor.Graph.Distance <= 0 {
		return 0
	}
	maximum := maxPromotion(intent, config)
	if maximum <= 0 {
		return 0
	}
	return min(score/graphScorePerPromotion, maximum)
}

func maxPromotion(intent evidence.Intent, config Config) float64 {
	switch intent {
	case evidence.IntentMechanism:
		return max(config.MechanismMaxPromotion, 0)
	case evidence.IntentCallChain:
		return max(config.CallChainMaxPromotion, 0)
	case evidence.IntentArchitecture:
		return max(config.ArchitectureMaxPromotion, 0)
	default:
		return 0
	}
}

func distanceScore(distance int) float64 {
	switch {
	case distance <= 0:
		return 6
	case distance == 1:
		return 5
	case distance == 2:
		return 3
	case distance == 3:
		return 1
	default:
		return 0
	}
}

func relationScore(relations []string, intent evidence.Intent) (string, float64) {
	bestRelation := ""
	bestScore := 0.0
	for _, relation := range relations {
		value := scoreRelation(relation, intent)
		if value > bestScore || value == bestScore && value > 0 && relation < bestRelation {
			bestRelation, bestScore = relation, value
		}
	}
	return bestRelation, bestScore
}

func scoreRelation(relation string, intent evidence.Intent) float64 {
	normalized := strings.ToLower(strings.TrimSpace(relation))
	base := normalized
	if separator := strings.IndexByte(base, ':'); separator >= 0 {
		base = base[separator+1:]
	}
	callLike := containsAny(base, "call", "invoke", "dispatch", "delegate", "handle")
	dataLike := containsAny(base, "read", "write", "use", "reference", "create", "return", "depend")
	architectureLike := containsAny(base, "own", "contain", "implement", "inherit", "override", "import", "member", "register")

	switch intent {
	case evidence.IntentCallChain:
		switch {
		case callLike && strings.HasPrefix(normalized, "outgoing:"):
			return 6
		case callLike:
			return 5
		case architectureLike:
			return 3
		case dataLike:
			return 2
		}
	case evidence.IntentArchitecture:
		switch {
		case architectureLike:
			return 6
		case dataLike:
			return 3
		case callLike:
			return 2
		}
	case evidence.IntentMechanism:
		switch {
		case callLike, dataLike:
			return 5
		case architectureLike:
			return 3
		}
	case evidence.IntentDirectLocation:
		if callLike || dataLike || architectureLike {
			return 1
		}
	case evidence.IntentMixed:
		if callLike || dataLike || architectureLike {
			return 2
		}
	}
	return 0
}

func symbolRoleScore(role string, intent evidence.Intent) float64 {
	role = strings.ToLower(strings.TrimSpace(role))
	if role == "" {
		return 0
	}
	callable := containsAny(role, "function", "method", "procedure", "constructor", "handler")
	structural := containsAny(role, "type", "class", "interface", "trait", "module", "package", "namespace")
	switch intent {
	case evidence.IntentArchitecture:
		if structural {
			return 4
		}
		if callable {
			return 2
		}
	case evidence.IntentCallChain, evidence.IntentMechanism, evidence.IntentDirectLocation:
		if callable {
			return 4
		}
		if structural {
			return 2
		}
	case evidence.IntentMixed:
		if callable || structural {
			return 2
		}
	}
	return 0
}

func cloneCandidate(candidate retrieve.Candidate) retrieve.Candidate {
	candidate.Reasons = append([]string(nil), candidate.Reasons...)
	candidate.ScoreDetails = append([]retrieve.ScoreDetail(nil), candidate.ScoreDetails...)
	candidate.GraphScoreDetails = append([]retrieve.ScoreDetail(nil), candidate.GraphScoreDetails...)
	return candidate
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func clamp(value, minimum, maximum float64) float64 {
	return math.Max(minimum, math.Min(value, maximum))
}
