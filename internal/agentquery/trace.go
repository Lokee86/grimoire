package agentquery

import (
	"context"

	"github.com/Lokee86/grimoire/internal/arcanagraph"
	"github.com/Lokee86/grimoire/internal/lexiconfacts"
	"github.com/Lokee86/grimoire/internal/structure"
)

func (engine *Engine) trace(ctx context.Context, request Request, response *Response) error {
	from, err := engine.resolveAnchors(ctx, request.Anchor, request.Query, request.Limit)
	if err != nil {
		return err
	}
	var to resolvedAnchors
	if request.Target != "" {
		to, err = engine.resolveAnchors(ctx, request.Target, "", request.Limit)
		if err != nil {
			return err
		}
	}

	if engine.lexicon != nil && len(from.lexicon) > 0 {
		paths := engine.lexicon.Trace(
			identities(from.lexicon), identities(to.lexicon),
			request.Direction, request.Relations, request.Depth, request.Limit,
		)
		for _, path := range paths {
			response.Paths = append(response.Paths, engine.lexiconPath(path, len(response.Paths)+1))
			if len(response.Paths) == request.Limit {
				response.Truncated = true
				break
			}
		}
	}

	if engine.arcanaSnapshot != "" && len(response.Paths) < request.Limit {
		if len(to.arcana) > 0 {
			if err := engine.arcanaPaths(ctx, from.arcana, to.arcana, request, response); err != nil {
				response.Warnings = append(response.Warnings, "Arcana trace unavailable: "+err.Error())
			}
		} else if err := engine.arcanaExpansion(ctx, from.arcana, request, response); err != nil {
			response.Warnings = append(response.Warnings, "Arcana trace unavailable: "+err.Error())
		}
	}
	for _, node := range from.arcana {
		if node.NodeID == nil || len(response.Unresolved) >= request.Limit {
			continue
		}
		items, truncated, unresolvedErr := engine.arcana.Unresolved(
			ctx, engine.arcanaSnapshot, *node.NodeID, request.Limit-len(response.Unresolved),
		)
		if unresolvedErr != nil {
			response.Warnings = append(response.Warnings, "Arcana unresolved alternatives unavailable: "+unresolvedErr.Error())
			break
		}
		for _, item := range items {
			response.Unresolved = append(response.Unresolved, unresolvedFromStructure(item, engine.source.Identity()))
		}
		response.Truncated = response.Truncated || truncated
	}
	if len(response.Paths) == 0 {
		response.Warnings = append(response.Warnings, "no structural path matched the supplied anchor and filters")
	}
	return nil
}

func (engine *Engine) lexiconPath(value lexiconfacts.QueryPath, rank int) Path {
	result := Path{Rank: rank, Nodes: make([]Node, len(value.Nodes))}
	for index, node := range value.Nodes {
		result.Nodes[index] = engine.node("lexicon", engine.lexiconSnapshot, node)
	}
	for index, edge := range value.Edges {
		spans, evidence := siteRanges(edge.Sites, engine.source.Identity())
		result.Steps = append(result.Steps, PathStep{
			From: result.Nodes[index].Handle, To: result.Nodes[index+1].Handle,
			Direction: edge.Direction, Relation: edge.Relation,
			Certainty: edge.Certainty, Evidence: evidence, Spans: spans,
		})
	}
	return result
}

func (engine *Engine) arcanaPaths(
	ctx context.Context,
	from, to []structure.Node,
	request Request,
	response *Response,
) error {
	for _, start := range from {
		if start.NodeID == nil {
			continue
		}
		for _, target := range to {
			if target.NodeID == nil {
				continue
			}
			type pathQuery struct {
				from, to uint32
				reverse  bool
			}
			queries := []pathQuery{{from: *start.NodeID, to: *target.NodeID}}
			if request.Direction == "incoming" {
				queries = []pathQuery{{from: *target.NodeID, to: *start.NodeID, reverse: true}}
			} else if request.Direction == "both" {
				queries = append(queries, pathQuery{
					from: *target.NodeID, to: *start.NodeID, reverse: true,
				})
			}
			for _, query := range queries {
				paths, truncated, err := engine.arcana.Paths(
					ctx, engine.arcanaSnapshot, query.from, query.to,
					request.Relations, request.Depth, request.Limit-len(response.Paths),
				)
				if err != nil {
					return err
				}
				for _, path := range paths {
					if query.reverse {
						path = reverseArcanaPath(path)
					}
					response.Paths = append(response.Paths, engine.arcanaPath(path, len(response.Paths)+1))
				}
				response.Truncated = response.Truncated || truncated
				if len(response.Paths) >= request.Limit {
					return nil
				}
			}
		}
	}
	return nil
}

func (engine *Engine) arcanaExpansion(
	ctx context.Context,
	starts []structure.Node,
	request Request,
	response *Response,
) error {
	for _, start := range starts {
		if start.NodeID == nil {
			continue
		}
		visited := map[uint32]bool{*start.NodeID: true}
		var walk func([]structure.Node, []string, []string) error
		walk = func(nodes []structure.Node, relations, directions []string) error {
			if len(relations) >= request.Depth || len(response.Paths) >= request.Limit {
				return nil
			}
			current := nodes[len(nodes)-1]
			neighbors, err := engine.arcana.Neighbors(
				ctx, engine.arcanaSnapshot, *current.NodeID, request.Direction, request.Relations,
			)
			if err != nil {
				return err
			}
			for _, neighbor := range neighbors {
				if neighbor.Node.NodeID == nil || visited[*neighbor.Node.NodeID] {
					continue
				}
				nextNodes := append(append([]structure.Node(nil), nodes...), neighbor.Node)
				nextRelations := append(append([]string(nil), relations...), neighbor.Relation)
				nextDirections := append(append([]string(nil), directions...), neighbor.Direction)
				response.Paths = append(response.Paths, engine.arcanaPath(
					arcanagraph.QueryPath{
						Nodes: nextNodes, Relations: nextRelations,
						Directions: nextDirections,
					},
					len(response.Paths)+1,
				))
				if len(response.Paths) >= request.Limit {
					response.Truncated = true
					return nil
				}
				visited[*neighbor.Node.NodeID] = true
				if err := walk(nextNodes, nextRelations, nextDirections); err != nil {
					return err
				}
				delete(visited, *neighbor.Node.NodeID)
			}
			return nil
		}
		if err := walk([]structure.Node{start}, nil, nil); err != nil {
			return err
		}
	}
	return nil
}

func (engine *Engine) arcanaPath(value arcanagraph.QueryPath, rank int) Path {
	result := Path{Rank: rank, Nodes: make([]Node, len(value.Nodes))}
	for index, node := range value.Nodes {
		result.Nodes[index] = engine.node("arcana", engine.arcanaSnapshotID, node)
	}
	for index, relation := range value.Relations {
		direction := "outgoing"
		if index < len(value.Directions) && value.Directions[index] != "" {
			direction = value.Directions[index]
		}
		step := PathStep{
			From: result.Nodes[index].Handle, To: result.Nodes[index+1].Handle,
			Direction: direction, Relation: relation, Certainty: certainty(relation),
			Evidence: []string{"Arcana immutable graph edge"},
		}
		if result.Nodes[index].Span != nil {
			step.Spans = []Range{*result.Nodes[index].Span}
		}
		result.Steps = append(result.Steps, step)
	}
	return result
}

func reverseArcanaPath(value arcanagraph.QueryPath) arcanagraph.QueryPath {
	for left, right := 0, len(value.Nodes)-1; left < right; left, right = left+1, right-1 {
		value.Nodes[left], value.Nodes[right] = value.Nodes[right], value.Nodes[left]
	}
	for left, right := 0, len(value.Relations)-1; left < right; left, right = left+1, right-1 {
		value.Relations[left], value.Relations[right] = value.Relations[right], value.Relations[left]
	}
	value.Directions = make([]string, len(value.Relations))
	for index := range value.Directions {
		value.Directions[index] = "incoming"
	}
	return value
}
