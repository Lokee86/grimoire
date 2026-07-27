package arcanagraph

import (
	"context"
	"fmt"
	"strings"

	"github.com/Lokee86/grimoire/internal/structure"
)

func (client Client) seedGraphProximity(
	ctx context.Context,
	snapshot string,
	query string,
	candidates []hybridSeedCandidate,
) (map[string]float64, error) {
	result := make(map[string]float64, len(candidates))
	if strings.TrimSpace(snapshot) == "" || len(candidates) < 2 {
		return result, nil
	}

	seeds := make([]structure.Node, 0, len(candidates))
	for _, candidate := range candidates {
		seeds = append(seeds, candidate.node)
	}
	resolveRequests := make([]protocolRequest, 0, len(seeds)*2)
	for index, seed := range seeds {
		resolveRequests = append(resolveRequests,
			protocolRequest{
				ID: fmt.Sprintf("rerank-resolve-%d-exact", index), Op: "resolve_symbol",
				Name: seed.Name, Path: seed.Path, Limit: 8,
			},
			protocolRequest{
				ID: fmt.Sprintf("rerank-resolve-%d-broad", index), Op: "resolve_symbol",
				Name: seed.Name, Limit: 8,
			},
		)
	}
	responses, err := client.run(ctx, snapshot, resolveRequests)
	if err != nil {
		return result, err
	}
	resolved := resolveSeedsWithPrefix(seeds, responses, "rerank-resolve")
	if len(resolved) < 2 {
		return result, nil
	}

	requests := make([]protocolRequest, 0, len(resolved)*3)
	for index, seed := range resolved {
		nodeID := seed.node.NodeID
		requests = append(requests,
			protocolRequest{
				ID: fmt.Sprintf("rerank-neighbors-%d-out", index), Op: "neighbors",
				NodeID: &nodeID, Direction: "outgoing",
			},
			protocolRequest{
				ID: fmt.Sprintf("rerank-neighbors-%d-in", index), Op: "neighbors",
				NodeID: &nodeID, Direction: "incoming",
			},
			protocolRequest{
				ID: fmt.Sprintf("rerank-unresolved-%d", index), Op: "unresolved",
				NodeID: &nodeID, Limit: 32,
			},
		)
	}
	responses, err = client.run(ctx, snapshot, requests)
	if err != nil {
		return result, err
	}

	candidateByNodeID := make(map[uint32]int, len(resolved))
	candidateByKey := make(map[string]hybridSeedCandidate, len(candidates))
	for _, candidate := range candidates {
		candidateByKey[hybridSeedKey(candidate.node)] = candidate
	}
	for index, seed := range resolved {
		candidateByNodeID[seed.node.NodeID] = index
	}
	queryTokens := meaningfulHybridTokens(normalizedHybridText(query))
	neighborSets := make([]map[uint32]struct{}, len(resolved))
	directTokens := make([]map[string]struct{}, len(resolved))
	neighborhoodTokens := make([]map[string]struct{}, len(resolved))
	raw := make([]float64, len(resolved))
	for index, seed := range resolved {
		neighborSets[index] = make(map[uint32]struct{})
		directTokens[index] = make(map[string]struct{})
		neighborhoodTokens[index] = make(map[string]struct{})
		mergeHybridTokens(directTokens[index], meaningfulHybridTokens(normalizedHybridText(seed.seed.Name)))
		mergeHybridTokens(directTokens[index], meaningfulHybridTokens(normalizedHybridText(seed.seed.Path)))
		mergeHybridTokens(neighborhoodTokens[index], directTokens[index])
		for _, direction := range []string{"out", "in"} {
			value, ok := decodeResponse[neighborResult](responses[fmt.Sprintf("rerank-neighbors-%d-%s", index, direction)])
			if !ok {
				continue
			}
			for _, relationship := range value.Relationships {
				neighborSets[index][relationship.Node.NodeID] = struct{}{}
				relationTokens := meaningfulHybridTokens(normalizedHybridText(relationship.Relation))
				nodeNameTokens := meaningfulHybridTokens(normalizedHybridText(relationship.Node.Name))
				nodePathTokens := meaningfulHybridTokens(normalizedHybridText(relationship.Node.Path))
				mergeHybridTokens(neighborhoodTokens[index], relationTokens)
				mergeHybridTokens(neighborhoodTokens[index], nodeNameTokens)
				mergeHybridTokens(neighborhoodTokens[index], nodePathTokens)
				if direction == "out" {
					mergeHybridTokens(directTokens[index], relationTokens)
					mergeHybridTokens(directTokens[index], nodeNameTokens)
					mergeHybridTokens(directTokens[index], nodePathTokens)
				}
				other, exists := candidateByNodeID[relationship.Node.NodeID]
				if !exists || other == index {
					continue
				}
				anchor := candidateByKey[hybridSeedKey(resolved[other].seed)]
				raw[index] += hybridRelationWeight(relationship.Relation) * (0.5 + anchor.base)
			}
		}
		unresolved, ok := decodeResponse[unresolvedResult](responses[fmt.Sprintf("rerank-unresolved-%d", index)])
		if ok {
			for _, item := range unresolved.Unresolved {
				for _, value := range []string{item.Relation, item.Expression, item.CandidateNamespace, item.CandidateName} {
					tokens := meaningfulHybridTokens(normalizedHybridText(value))
					mergeHybridTokens(directTokens[index], tokens)
					mergeHybridTokens(neighborhoodTokens[index], tokens)
				}
			}
		}
	}

	baseNeighborhoodTokens := make([]map[string]struct{}, len(neighborhoodTokens))
	for index, tokens := range neighborhoodTokens {
		baseNeighborhoodTokens[index] = make(map[string]struct{}, len(tokens))
		mergeHybridTokens(baseNeighborhoodTokens[index], tokens)
	}
	for index, neighbors := range neighborSets {
		for nodeID := range neighbors {
			other, exists := candidateByNodeID[nodeID]
			if !exists || other == index {
				continue
			}
			mergeHybridTokens(neighborhoodTokens[index], baseNeighborhoodTokens[other])
		}
	}
	for left := 0; left < len(resolved); left++ {
		for right := left + 1; right < len(resolved); right++ {
			shared := sharedHybridNeighbors(neighborSets[left], neighborSets[right])
			if shared == 0 {
				continue
			}
			leftAnchor := candidateByKey[hybridSeedKey(resolved[left].seed)]
			rightAnchor := candidateByKey[hybridSeedKey(resolved[right].seed)]
			proximity := min(float64(shared), 4) * 0.15
			raw[left] += proximity * (0.5 + rightAnchor.base)
			raw[right] += proximity * (0.5 + leftAnchor.base)
		}
	}
	maximum := 0.0
	for _, score := range raw {
		maximum = max(maximum, score)
	}
	for index, seed := range resolved {
		centrality := 0.0
		if maximum > 0 {
			centrality = raw[index] / maximum
		}
		direct := hybridQueryCoverage(queryTokens, directTokens[index])
		neighborhood := hybridQueryCoverage(queryTokens, neighborhoodTokens[index])
		result[hybridSeedKey(seed.seed)] = clampHybridScore(neighborhood*0.68 + direct*0.17 + centrality*0.15)
	}
	return result, nil
}
