package agentquery

import (
	"context"
	"strings"

	"github.com/Lokee86/grimoire/internal/arcanagraph"
	"github.com/Lokee86/grimoire/internal/lexiconfacts"
	"github.com/Lokee86/grimoire/internal/structure"
)

func (engine *Engine) trace(ctx context.Context, request Request, response *Response) error {
	traceDone := debugTiming("trace.total")
	defer traceDone()
	requestedLimit := request.Limit
	traceRequest := request
	traceRequest.Limit = traceCandidateLimit(request.Limit)

	openArcanaDone := debugTiming("trace.open_arcana")
	arcana, closeArcana := engine.openArcanaQuery(ctx, response)
	openArcanaDone()
	defer func() {
		closeArcanaDone := debugTiming("trace.close_arcana")
		closeArcana()
		closeArcanaDone()
	}()

	resolveFromDone := debugTiming("trace.resolve_from")
	from, err := engine.resolveAnchors(ctx, request.Anchor, request.Query, traceRequest.Limit, arcana)
	resolveFromDone()
	if err != nil {
		return err
	}
	var to resolvedAnchors
	if request.Target != "" {
		to, err = engine.resolveAnchors(ctx, request.Target, "", traceRequest.Limit, arcana)
		if err != nil {
			return err
		}
	}

	lexiconDone := debugTiming("trace.lexicon")
	if engine.lexicon != nil && len(from.lexicon) > 0 {
		lexiconLimit := traceRequest.Limit
		if request.Target != "" {
			lexiconLimit = min(maxTraceCandidates, max(traceRequest.Limit, requestedLimit*2))
		}
		paths := engine.lexicon.Trace(
			identities(from.lexicon), identities(to.lexicon),
			traceRequest.Direction, traceRequest.Relations, traceRequest.Depth, lexiconLimit,
		)
		for _, path := range paths {
			response.Paths = append(response.Paths, engine.lexiconPath(path, len(response.Paths)+1))
		}
		response.Truncated = response.Truncated || len(paths) == lexiconLimit
	}
	lexiconDone()

	arcanaDone := debugTiming("trace.arcana")
	needArcana := arcana != nil && engine.arcanaSnapshot != "" &&
		(request.Target != "" || distinctBehaviorEntries(response.Paths) < requestedLimit)
	if needArcana {
		arcanaResponse := Response{}
		if len(to.arcana) > 0 {
			if err := engine.arcanaPaths(ctx, from.arcana, to.arcana, traceRequest, &arcanaResponse, arcana); err != nil {
				response.Warnings = append(response.Warnings, "Arcana trace unavailable: "+err.Error())
			}
		} else if err := engine.arcanaExpansion(ctx, from.arcana, traceRequest, &arcanaResponse, arcana); err != nil {
			response.Warnings = append(response.Warnings, "Arcana trace unavailable: "+err.Error())
		}
		response.Paths = append(response.Paths, arcanaResponse.Paths...)
		response.Truncated = response.Truncated || arcanaResponse.Truncated
	}
	arcanaDone()

	finalizeDone := debugTiming("trace.finalize")
	finalizeTraceResponse(request, response, requestedLimit)
	finalizeDone()

	if len(response.Paths) == 0 || request.Detail == "full" {
		for _, node := range from.arcana {
			if node.NodeID == nil || len(response.Unresolved) >= request.Limit {
				continue
			}
			items, truncated, unresolvedErr := arcana.Unresolved(
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
	arcana arcanaQuery,
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
				paths, truncated, err := arcana.Paths(
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
	arcana arcanaQuery,
) error {
	remaining := request.Limit - len(response.Paths)
	if remaining <= 0 {
		return nil
	}
	candidateCap := min(max(remaining*2, remaining), 32)
	queryContext := request.Query + " " + request.Anchor
	type queuedPath struct {
		nodes      []structure.Node
		relations  []string
		directions []string
		visited    map[uint32]bool
		entry      string
	}
	queue := make([]queuedPath, 0, len(starts))
	globallyVisited := make(map[uint32]bool)
	for _, start := range starts {
		if start.NodeID == nil || globallyVisited[*start.NodeID] {
			continue
		}
		globallyVisited[*start.NodeID] = true
		queue = append(queue, queuedPath{
			nodes:   []structure.Node{start},
			visited: map[uint32]bool{*start.NodeID: true},
			entry:   traceStructureNodeLabel(start),
		})
	}
	entryCounts := make(map[string]int)
	var behavioral []arcanagraph.QueryPath
	var structural []arcanagraph.QueryPath
	for len(queue) > 0 && len(behavioral) < candidateCap {
		depth := len(queue[0].relations)
		frontierEnd := 0
		for frontierEnd < len(queue) && len(queue[frontierEnd].relations) == depth {
			frontierEnd++
		}
		frontier := append([]queuedPath(nil), queue[:frontierEnd]...)
		queue = queue[frontierEnd:]
		if depth >= request.Depth {
			continue
		}
		if len(frontier) > 8 {
			frontier = frontier[:8]
			response.Truncated = true
		}
		nodeIDs := make([]uint32, 0, len(frontier))
		for _, current := range frontier {
			tail := current.nodes[len(current.nodes)-1]
			if tail.NodeID != nil {
				nodeIDs = append(nodeIDs, *tail.NodeID)
			}
		}
		neighborsByNode, err := arcana.NeighborsBatch(
			ctx, engine.arcanaSnapshot, nodeIDs, request.Direction, request.Relations,
		)
		if err != nil {
			return err
		}
		nextQueue := make([]queuedPath, 0, min(candidateCap, 8))
		for _, current := range frontier {
			tail := current.nodes[len(current.nodes)-1]
			if tail.NodeID == nil {
				continue
			}
			neighbors := rankTraceNeighbors(neighborsByNode[*tail.NodeID], queryContext)
			branches := 0
			for _, neighbor := range neighbors {
				if neighbor.Node.NodeID == nil || current.visited[*neighbor.Node.NodeID] || globallyVisited[*neighbor.Node.NodeID] {
					continue
				}
				nextNodes := append(append([]structure.Node(nil), current.nodes...), neighbor.Node)
				nextRelations := append(append([]string(nil), current.relations...), neighbor.Relation)
				nextDirections := append(append([]string(nil), current.directions...), neighbor.Direction)
				entry := current.entry
				if len(current.relations) == 0 && traceContextRelation(neighbor.Relation) {
					entry = traceStructureNodeLabel(neighbor.Node)
				}
				path := arcanagraph.QueryPath{
					Nodes: nextNodes, Relations: nextRelations, Directions: nextDirections,
				}
				behavioralPath := traceRelationsHaveBehavior(nextRelations)
				if behavioralPath {
					key := strings.ToLower(entry)
					if entryCounts[key] < 4 {
						behavioral = append(behavioral, path)
						entryCounts[key]++
					}
				} else if len(structural) < candidateCap {
					structural = append(structural, path)
				}
				if len(nextRelations) >= request.Depth {
					continue
				}
				nextVisited := make(map[uint32]bool, len(current.visited)+1)
				for id := range current.visited {
					nextVisited[id] = true
				}
				nextVisited[*neighbor.Node.NodeID] = true
				globallyVisited[*neighbor.Node.NodeID] = true
				nextQueue = append(nextQueue, queuedPath{
					nodes: nextNodes, relations: nextRelations, directions: nextDirections,
					visited: nextVisited, entry: entry,
				})
				branches++
				if len(behavioral) >= candidateCap || branches >= 2 || len(nextQueue) >= 8 {
					break
				}
			}
			if len(behavioral) >= candidateCap || len(nextQueue) >= 8 {
				break
			}
		}
		queue = append(queue, nextQueue...)
	}

	paths := behavioral
	if len(paths) == 0 {
		paths = structural
	}
	for _, path := range paths {
		response.Paths = append(response.Paths, engine.arcanaPath(path, len(response.Paths)+1))
		if len(response.Paths) >= request.Limit {
			response.Truncated = len(paths) > remaining
			break
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
