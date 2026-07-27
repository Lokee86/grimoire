package arcanagraph

import (
	"fmt"
	"strings"

	"github.com/Lokee86/grimoire/internal/structure"
)

func resolveSeedsWithPrefix(
	seeds []structure.Node,
	responses map[string]protocolResponse,
	prefix string,
) []resolvedSeed {
	seenNodes := make(map[uint32]struct{})
	result := make([]resolvedSeed, 0, len(seeds))
	for index, seed := range seeds {
		exact := decodeNodeList(responses[fmt.Sprintf("%s-%d-exact", prefix, index)])
		broad := decodeNodeList(responses[fmt.Sprintf("%s-%d-broad", prefix, index)])
		node, found := chooseResolvedNode(seed, exact.Nodes)
		if !found {
			node, found = chooseResolvedNode(seed, broad.Nodes)
		}
		if !found {
			continue
		}
		if _, exists := seenNodes[node.NodeID]; exists {
			continue
		}
		seenNodes[node.NodeID] = struct{}{}
		result = append(result, resolvedSeed{seed: seed, node: node})
	}
	return result
}

func hybridRelationWeight(relation string) float64 {
	switch strings.ToLower(strings.TrimSpace(relation)) {
	case "calls", "routes", "handles", "publishes", "subscribes", "implements", "overrides":
		return 1
	case "reads", "writes", "passes_to", "inherits", "extends", "uses":
		return 0.8
	case "contains", "declares", "owns", "belongs_to":
		return 0.4
	default:
		return 0.6
	}
}

func sharedHybridNeighbors(left, right map[uint32]struct{}) int {
	if len(left) > len(right) {
		left, right = right, left
	}
	count := 0
	for nodeID := range left {
		if _, exists := right[nodeID]; exists {
			count++
		}
	}
	return count
}
