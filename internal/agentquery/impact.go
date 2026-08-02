package agentquery

import "context"

func (engine *Engine) impact(ctx context.Context, request Request, response *Response) error {
	arcana, closeArcana := engine.openArcanaQuery(ctx, response)
	defer closeArcana()

	starts, err := engine.resolveAnchors(ctx, request.Anchor, request.Query, request.Limit, arcana)
	if err != nil {
		return err
	}

	candidateLimit := impactCandidateLimit(request.Limit)
	candidates := make([]Dependent, 0, candidateLimit*2)
	providerTruncated := false
	if engine.lexicon != nil {
		items := engine.lexicon.Impact(
			identities(starts.lexicon), request.Direction, request.Relations,
			request.Depth, candidateLimit,
		)
		providerTruncated = providerTruncated || len(items) == candidateLimit
		for _, item := range items {
			spans, evidence := siteRanges(item.Sites, engine.source.Identity())
			candidates = append(candidates, Dependent{
				Depth: item.Depth, Direction: item.Direction, Relation: item.Relation,
				Certainty: certainty(item.Relation),
				Node:      engine.node("lexicon", engine.lexiconSnapshot, item.Node),
				Evidence:  evidence, Spans: spans,
			})
		}
	}

	if arcana != nil && engine.arcanaSnapshot != "" {
		arcanaCandidates := 0
		for _, start := range starts.arcana {
			if start.NodeID == nil || arcanaCandidates >= candidateLimit {
				continue
			}
			items, truncated, impactErr := arcana.ImpactQuery(
				ctx, engine.arcanaSnapshot, *start.NodeID, request.Direction,
				request.Relations, request.Depth, candidateLimit-arcanaCandidates,
			)
			if impactErr != nil {
				response.Warnings = append(response.Warnings, "Arcana impact unavailable: "+impactErr.Error())
				break
			}
			providerTruncated = providerTruncated || truncated
			for _, item := range items {
				dependent := Dependent{
					Depth: item.Depth, Direction: item.Direction,
					Relation: item.Relation, Certainty: item.Certainty,
					Node:     engine.node("arcana", engine.arcanaSnapshotID, item.Node),
					Evidence: []string{"Arcana bounded graph traversal"},
				}
				if dependent.Node.Span != nil {
					dependent.Spans = []Range{*dependent.Node.Span}
				}
				candidates = append(candidates, dependent)
				arcanaCandidates++
			}
		}
	}

	response.Dependents = rankImpactDependents(request, candidates, request.Limit)
	response.Truncated = response.Truncated || providerTruncated || impactDistinctCandidateCount(candidates) > len(response.Dependents)
	if len(response.Dependents) == 0 {
		response.Warnings = append(response.Warnings, "no dependents matched the supplied anchor and filters")
	}
	return nil
}
