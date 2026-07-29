package agentquery

import "context"

func (engine *Engine) impact(ctx context.Context, request Request, response *Response) error {
	arcana, closeArcana := engine.openArcanaQuery(ctx, response)
	defer closeArcana()

	starts, err := engine.resolveAnchors(ctx, request.Anchor, request.Query, request.Limit, arcana)
	if err != nil {
		return err
	}
	if engine.lexicon != nil {
		for _, item := range engine.lexicon.Impact(
			identities(starts.lexicon), request.Direction, request.Relations,
			request.Depth, request.Limit,
		) {
			spans, evidence := siteRanges(item.Sites, engine.source.Identity())
			response.Dependents = append(response.Dependents, Dependent{
				Depth: item.Depth, Direction: item.Direction, Relation: item.Relation,
				Certainty: certainty(item.Relation),
				Node:      engine.node("lexicon", engine.lexiconSnapshot, item.Node),
				Evidence:  evidence, Spans: spans,
			})
			if len(response.Dependents) == request.Limit {
				response.Truncated = true
				return nil
			}
		}
	}
	if arcana != nil && engine.arcanaSnapshot != "" {
		for _, start := range starts.arcana {
			if start.NodeID == nil {
				continue
			}
			items, truncated, impactErr := arcana.ImpactQuery(
				ctx, engine.arcanaSnapshot, *start.NodeID, request.Direction,
				request.Relations, request.Depth, request.Limit-len(response.Dependents),
			)
			if impactErr != nil {
				response.Warnings = append(response.Warnings, "Arcana impact unavailable: "+impactErr.Error())
				break
			}
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
				response.Dependents = append(response.Dependents, dependent)
			}
			response.Truncated = response.Truncated || truncated
			if len(response.Dependents) >= request.Limit {
				response.Truncated = true
				break
			}
		}
	}
	if len(response.Dependents) == 0 {
		response.Warnings = append(response.Warnings, "no dependents matched the supplied anchor and filters")
	}
	return nil
}
