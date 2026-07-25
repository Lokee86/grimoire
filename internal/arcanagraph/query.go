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

func (client Client) Resolve(
	ctx context.Context,
	snapshot, name, path string,
	limit int,
) ([]structure.Node, error) {
	if snapshot == "" || strings.TrimSpace(name) == "" || limit <= 0 {
		return nil, nil
	}
	response, err := client.run(ctx, snapshot, []protocolRequest{{
		ID: "resolve", Op: "resolve_symbol", Name: name, Path: path, Limit: limit,
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
	response, err := client.run(ctx, snapshot, []protocolRequest{{
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
	directions := []string{direction}
	if direction == "" || direction == "both" {
		directions = []string{"incoming", "outgoing"}
	}
	relationValues := relations
	if len(relationValues) == 0 {
		relationValues = []string{""}
	}
	requests := make([]protocolRequest, 0, len(directions)*len(relationValues))
	for _, currentDirection := range directions {
		for _, relation := range relationValues {
			requests = append(requests, protocolRequest{
				ID: fmt.Sprintf("neighbors-%d", len(requests)), Op: "neighbors",
				NodeID: &nodeID, Direction: currentDirection, Relation: relation,
			})
		}
	}
	responses, err := client.run(ctx, snapshot, requests)
	if err != nil {
		return nil, err
	}
	var result []QueryNeighbor
	seen := make(map[string]bool)
	for _, request := range requests {
		value, ok := decodeResponse[neighborResult](responses[request.ID])
		if !ok {
			continue
		}
		for _, related := range value.Relationships {
			key := fmt.Sprintf("%s\x00%s\x00%d", value.Direction, related.Relation, related.Node.NodeID)
			if seen[key] {
				continue
			}
			seen[key] = true
			result = append(result, QueryNeighbor{
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
	response, err := client.run(ctx, snapshot, []protocolRequest{{
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
	type queued struct {
		id    uint32
		depth int
	}
	seen := map[uint32]bool{startNodeID: true}
	queue := []queued{{id: startNodeID}}
	var result []QueryImpact
	truncated := false
	for len(queue) > 0 && len(result) < limit {
		current := queue[0]
		queue = queue[1:]
		if current.depth >= maxDepth {
			continue
		}
		neighbors, err := client.Neighbors(ctx, snapshot, current.id, direction, relations)
		if err != nil {
			return nil, false, err
		}
		for _, neighbor := range neighbors {
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
				truncated = len(queue) > 0
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
	response, err := client.run(ctx, snapshot, []protocolRequest{{
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
