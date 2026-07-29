package arcanagraph

import (
	"context"
	"fmt"
	"strings"

	"github.com/Lokee86/grimoire/internal/structure"
)

type QueryPath struct {
	Nodes      []structure.Node
	Relations  []string
	Directions []string
}

type QueryNeighbor struct {
	Direction string
	Relation  string
	Certainty string
	Node      structure.Node
}

type QueryImpact struct {
	Depth int
	QueryNeighbor
}

type neighborResult struct {
	Node          arcanaNode    `json:"node"`
	Direction     string        `json:"direction"`
	Relationships []relatedNode `json:"relationships"`
}

type pathsResult struct {
	Truncated bool         `json:"truncated"`
	Paths     []arcanaPath `json:"paths"`
}

const maxNeighborResults = 64

type protocolBatchRun func([]protocolRequest) (map[string]protocolResponse, error)

type neighborBatchQuery interface {
	NeighborsBatch(context.Context, string, []uint32, string, []string) (map[uint32][]QueryNeighbor, error)
}

func (client Client) Resolve(
	ctx context.Context,
	snapshot, name, path string,
	limit int,
) ([]structure.Node, error) {
	return client.ResolveTyped(ctx, snapshot, name, "", path, limit)
}

func (client Client) ResolveTyped(
	ctx context.Context,
	snapshot, name, kind, path string,
	limit int,
) ([]structure.Node, error) {
	return resolveWithRun(name, kind, path, limit, func(requests []protocolRequest) (map[string]protocolResponse, error) {
		return client.run(ctx, snapshot, requests)
	})
}

func resolveWithRun(name, kind, path string, limit int, run protocolBatchRun) ([]structure.Node, error) {
	if strings.TrimSpace(name) == "" || limit <= 0 {
		return nil, nil
	}
	response, err := run([]protocolRequest{{
		ID: "resolve", Op: "resolve_symbol", Name: name, Kind: kind, Path: path, Limit: limit,
	}})
	if err != nil {
		return nil, err
	}
	nodes := decodeNodeList(response["resolve"]).Nodes
	result := make([]structure.Node, len(nodes))
	for index, node := range nodes {
		result[index] = node.toStructure()
	}
	return result, nil
}

func (client Client) Inspect(
	ctx context.Context,
	snapshot string,
	nodeID uint32,
) (structure.Node, error) {
	return inspectWithRun(nodeID, func(requests []protocolRequest) (map[string]protocolResponse, error) {
		return client.run(ctx, snapshot, requests)
	})
}

func inspectWithRun(nodeID uint32, run protocolBatchRun) (structure.Node, error) {
	response, err := run([]protocolRequest{{
		ID: "inspect", Op: "neighbors", NodeID: &nodeID, Direction: "outgoing",
	}})
	if err != nil {
		return structure.Node{}, err
	}
	value, ok := decodeResponse[neighborResult](response["inspect"])
	if !ok {
		return structure.Node{}, fmt.Errorf("Arcana did not return node %d", nodeID)
	}
	return value.Node.toStructure(), nil
}

func (client Client) Neighbors(
	ctx context.Context,
	snapshot string,
	nodeID uint32,
	direction string,
	relations []string,
) ([]QueryNeighbor, error) {
	results, err := client.NeighborsBatch(ctx, snapshot, []uint32{nodeID}, direction, relations)
	return results[nodeID], err
}

func (client Client) NeighborsBatch(
	ctx context.Context,
	snapshot string,
	nodeIDs []uint32,
	direction string,
	relations []string,
) (map[uint32][]QueryNeighbor, error) {
	if snapshot == "" {
		return map[uint32][]QueryNeighbor{}, nil
	}
	return neighborsBatchWithRun(nodeIDs, direction, relations, func(requests []protocolRequest) (map[string]protocolResponse, error) {
		return client.run(ctx, snapshot, requests)
	})
}

func neighborsBatchWithRun(
	nodeIDs []uint32,
	direction string,
	relations []string,
	run protocolBatchRun,
) (map[uint32][]QueryNeighbor, error) {
	result := make(map[uint32][]QueryNeighbor, len(nodeIDs))
	if len(nodeIDs) == 0 {
		return result, nil
	}
	directions := []string{direction}
	if direction == "" || direction == "both" {
		directions = []string{"incoming", "outgoing"}
	}
	relationValues := relations
	protocolRelations := []string(nil)
	allowedRelations := make(map[string]bool)
	if len(relationValues) == 0 {
		relationValues = []string{""}
	} else if len(relationValues) > 1 {
		protocolRelations = append([]string(nil), relationValues...)
		for _, relation := range relationValues {
			allowedRelations[relation] = true
		}
		relationValues = []string{""}
	}
	type requestOwner struct {
		nodeID uint32
	}
	owners := make(map[string]requestOwner)
	requests := make([]protocolRequest, 0, len(nodeIDs)*len(directions)*len(relationValues))
	seenNodes := make(map[uint32]bool, len(nodeIDs))
	for _, nodeID := range nodeIDs {
		if seenNodes[nodeID] {
			continue
		}
		seenNodes[nodeID] = true
		for _, currentDirection := range directions {
			for _, relation := range relationValues {
				id := fmt.Sprintf("neighbors-%d", len(requests))
				currentNodeID := nodeID
				requests = append(requests, protocolRequest{
					ID: id, Op: "neighbors", NodeID: &currentNodeID,
					Direction: currentDirection, Relation: relation,
					Relations: protocolRelations, Limit: maxNeighborResults,
				})
				owners[id] = requestOwner{nodeID: nodeID}
			}
		}
	}
	responses, err := run(requests)
	if err != nil {
		return nil, err
	}
	seen := make(map[uint32]map[string]bool, len(nodeIDs))
	for _, request := range requests {
		owner := owners[request.ID]
		value, ok := decodeResponse[neighborResult](responses[request.ID])
		if !ok {
			continue
		}
		if seen[owner.nodeID] == nil {
			seen[owner.nodeID] = make(map[string]bool)
		}
		for _, related := range value.Relationships {
			if len(allowedRelations) > 0 && !allowedRelations[related.Relation] {
				continue
			}
			key := fmt.Sprintf("%s\x00%s\x00%d", value.Direction, related.Relation, related.Node.NodeID)
			if seen[owner.nodeID][key] {
				continue
			}
			seen[owner.nodeID][key] = true
			result[owner.nodeID] = append(result[owner.nodeID], QueryNeighbor{
				Direction: value.Direction, Relation: related.Relation,
				Certainty: relationCertainty(related.Relation), Node: related.Node.toStructure(),
			})
		}
	}
	return result, nil
}

func (client Client) Paths(
	ctx context.Context,
	snapshot string,
	fromNodeID, toNodeID uint32,
	relations []string,
	maxDepth, limit int,
) ([]QueryPath, bool, error) {
	return pathsWithRun(fromNodeID, toNodeID, relations, maxDepth, limit, func(requests []protocolRequest) (map[string]protocolResponse, error) {
		return client.run(ctx, snapshot, requests)
	})
}

func pathsWithRun(
	fromNodeID, toNodeID uint32,
	relations []string,
	maxDepth, limit int,
	run protocolBatchRun,
) ([]QueryPath, bool, error) {
	response, err := run([]protocolRequest{{
		ID: "paths", Op: "paths", FromNodeID: &fromNodeID, ToNodeID: &toNodeID,
		Relations: relations, MaxDepth: maxDepth, Limit: limit,
	}})
	if err != nil {
		return nil, false, err
	}
	value, ok := decodeResponse[pathsResult](response["paths"])
	if !ok {
		return nil, false, nil
	}
	result := make([]QueryPath, len(value.Paths))
	for pathIndex, path := range value.Paths {
		nodes := make([]structure.Node, len(path.Nodes))
		for nodeIndex, node := range path.Nodes {
			nodes[nodeIndex] = node.toStructure()
		}
		directions := make([]string, len(path.Relations))
		for index := range directions {
			directions[index] = "outgoing"
		}
		result[pathIndex] = QueryPath{
			Nodes: nodes, Relations: path.Relations, Directions: directions,
		}
	}
	return result, value.Truncated, nil
}

func (client Client) ImpactQuery(
	ctx context.Context,
	snapshot string,
	startNodeID uint32,
	direction string,
	relations []string,
	maxDepth, limit int,
) ([]QueryImpact, bool, error) {
	return impactWithQuery(ctx, snapshot, startNodeID, direction, relations, maxDepth, limit, client)
}

func impactWithQuery(
	ctx context.Context,
	snapshot string,
	startNodeID uint32,
	direction string,
	relations []string,
	maxDepth, limit int,
	query neighborBatchQuery,
) ([]QueryImpact, bool, error) {
	type queued struct {
		id    uint32
		depth int
	}
	seen := map[uint32]bool{startNodeID: true}
	queue := []queued{{id: startNodeID}}
	var result []QueryImpact
	truncated := false
	for len(queue) > 0 && len(result) < limit {
		depth := queue[0].depth
		frontierEnd := 0
		for frontierEnd < len(queue) && queue[frontierEnd].depth == depth {
			frontierEnd++
		}
		frontier := append([]queued(nil), queue[:frontierEnd]...)
		queue = queue[frontierEnd:]
		if depth >= maxDepth {
			continue
		}
		nodeIDs := make([]uint32, 0, len(frontier))
		for _, current := range frontier {
			nodeIDs = append(nodeIDs, current.id)
		}
		neighborsByNode, err := query.NeighborsBatch(ctx, snapshot, nodeIDs, direction, relations)
		if err != nil {
			return nil, false, err
		}
		for _, current := range frontier {
			for _, neighbor := range neighborsByNode[current.id] {
				if neighbor.Node.NodeID == nil || seen[*neighbor.Node.NodeID] {
					continue
				}
				id := *neighbor.Node.NodeID
				seen[id] = true
				result = append(result, QueryImpact{
					Depth: current.depth + 1, QueryNeighbor: neighbor,
				})
				queue = append(queue, queued{id: id, depth: current.depth + 1})
				if len(result) == limit {
					truncated = len(queue) > 0 || len(frontier) > 1
					break
				}
			}
			if len(result) == limit {
				break
			}
		}
	}
	return result, truncated, nil
}

func (client Client) Unresolved(
	ctx context.Context,
	snapshot string,
	nodeID uint32,
	limit int,
) ([]structure.Unresolved, bool, error) {
	return unresolvedWithRun(nodeID, limit, func(requests []protocolRequest) (map[string]protocolResponse, error) {
		return client.run(ctx, snapshot, requests)
	})
}

func unresolvedWithRun(
	nodeID uint32,
	limit int,
	run protocolBatchRun,
) ([]structure.Unresolved, bool, error) {
	response, err := run([]protocolRequest{{
		ID: "unresolved", Op: "unresolved", NodeID: &nodeID, Limit: limit,
	}})
	if err != nil {
		return nil, false, err
	}
	value, ok := decodeResponse[unresolvedResult](response["unresolved"])
	if !ok {
		return nil, false, nil
	}
	items := make([]structure.Unresolved, len(value.Unresolved))
	for index, item := range value.Unresolved {
		items[index] = structure.Unresolved{
			Relation: item.Relation, Expression: item.Expression,
			CandidateNamespace: item.CandidateNamespace, CandidateName: item.CandidateName,
			Reason: item.Reason, Span: item.Span.toStructure(),
		}
	}
	return items, value.Truncated, nil
}

func (client Client) run(
	ctx context.Context,
	snapshot string,
	requests []protocolRequest,
) (map[string]protocolResponse, error) {
	run := client.Run
	if run == nil {
		run = runProtocol
	}
	command := strings.TrimSpace(client.Command)
	if command == "" {
		command = "arcana"
	}
	return run(ctx, command, snapshot, requests)
}
