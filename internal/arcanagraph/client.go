package arcanagraph

import (
	"context"
	"fmt"

	"github.com/Lokee86/grimoire/internal/structure"
)

const (
	maxSeeds        = 6
	maxChainSeeds   = 3
	impactLimit     = 12
	unresolvedLimit = 8
	graphDepth      = 8
)

// Client queries one immutable Arcana snapshot through its process JSONL
// protocol. Run is replaceable for deterministic tests.
type Client struct {
	Command string
	Run     protocolRun
}

func (client Client) Search(
	ctx context.Context,
	snapshot string,
	seeds []structure.Node,
) ([]structure.Evidence, error) {
	if snapshot == "" || len(seeds) == 0 {
		return nil, nil
	}
	seeds = uniqueSeeds(seeds, maxSeeds)
	run := client.Run
	if run == nil {
		run = runProtocol
	}
	command := client.Command
	if command == "" {
		command = "arcana"
	}

	resolveRequests := make([]protocolRequest, 0, len(seeds)*2)
	for index, seed := range seeds {
		resolveRequests = append(resolveRequests, protocolRequest{
			ID: fmt.Sprintf("resolve-%d-exact", index), Op: "resolve_symbol",
			Name: seed.Name, Path: seed.Path, Limit: 8,
		})
		resolveRequests = append(resolveRequests, protocolRequest{
			ID: fmt.Sprintf("resolve-%d-broad", index), Op: "resolve_symbol",
			Name: seed.Name, Limit: 8,
		})
	}
	resolvedResponses, err := run(ctx, command, snapshot, resolveRequests)
	if err != nil {
		return nil, err
	}
	resolved := resolveSeeds(seeds, resolvedResponses)
	if len(resolved) == 0 {
		return nil, nil
	}

	detailRequests := make([]protocolRequest, 0, len(resolved)*3+maxChainSeeds*maxChainSeeds)
	includePossible := true
	for index, seed := range resolved {
		nodeID := seed.node.NodeID
		detailRequests = append(detailRequests,
			protocolRequest{
				ID: fmt.Sprintf("role-%d", index), Op: "operational_role",
				NodeID: &nodeID, IncludePossible: &includePossible, MaxDepth: graphDepth,
			},
			protocolRequest{
				ID: fmt.Sprintf("impact-%d", index), Op: "impact",
				NodeID: &nodeID, MaxDepth: 4, Limit: impactLimit,
			},
			protocolRequest{
				ID: fmt.Sprintf("unresolved-%d", index), Op: "unresolved",
				NodeID: &nodeID, Limit: unresolvedLimit,
			},
		)
	}
	chainSeeds := min(len(resolved), maxChainSeeds)
	for from := 0; from < chainSeeds; from++ {
		for to := 0; to < chainSeeds; to++ {
			if from == to {
				continue
			}
			fromID, toID := resolved[from].node.NodeID, resolved[to].node.NodeID
			detailRequests = append(detailRequests, protocolRequest{
				ID: fmt.Sprintf("chain-%d-%d", from, to), Op: "shortest_call_chain",
				FromNodeID: &fromID, ToNodeID: &toID,
				IncludePossible: &includePossible, MaxDepth: graphDepth,
			})
		}
	}
	detailResponses, err := run(ctx, command, snapshot, detailRequests)
	if err != nil {
		return nil, err
	}
	evidence := evidenceFromResponses(resolved, detailResponses)
	for index := range evidence {
		evidence[index].Rank = index + 1
	}
	return evidence, nil
}

type resolvedSeed struct {
	seed structure.Node
	node arcanaNode
}

func uniqueSeeds(seeds []structure.Node, limit int) []structure.Node {
	seen := make(map[string]struct{}, len(seeds))
	result := make([]structure.Node, 0, min(len(seeds), limit))
	for _, seed := range seeds {
		if seed.Name == "" {
			continue
		}
		key := seed.Name + "\x00" + seed.Path
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, seed)
		if len(result) == limit {
			break
		}
	}
	return result
}

func resolveSeeds(
	seeds []structure.Node,
	responses map[string]protocolResponse,
) []resolvedSeed {
	seenNodes := make(map[uint32]struct{})
	result := make([]resolvedSeed, 0, len(seeds))
	for index, seed := range seeds {
		exact := decodeNodeList(responses[fmt.Sprintf("resolve-%d-exact", index)])
		broad := decodeNodeList(responses[fmt.Sprintf("resolve-%d-broad", index)])
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

func chooseResolvedNode(seed structure.Node, nodes []arcanaNode) (arcanaNode, bool) {
	for _, node := range nodes {
		if seed.Path != "" && node.Path == seed.Path {
			return node, true
		}
	}
	if len(nodes) > 0 {
		return nodes[0], true
	}
	return arcanaNode{}, false
}
