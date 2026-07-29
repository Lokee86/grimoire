package agentquery

import (
	"context"
	"strings"

	"github.com/Lokee86/grimoire/internal/structure"
)

type resolvedAnchors struct {
	lexicon []structure.Node
	arcana  []structure.Node
}

func (engine *Engine) resolveAnchors(
	ctx context.Context,
	value, query string,
	limit int,
	arcana arcanaQuery,
) (resolvedAnchors, error) {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "grimoire:v1:") {
		handle, err := parseHandle(value)
		if err != nil {
			return resolvedAnchors{}, err
		}
		if err := engine.validateSnapshot(handle); err != nil {
			return resolvedAnchors{}, err
		}
		var result resolvedAnchors
		switch handle.Provider {
		case "source":
			if engine.lexicon != nil {
				result.lexicon = engine.lexicon.ResolveSource(
					handle.Path, handle.StartLine, handle.EndLine, limit,
				)
			}
		case "lexicon":
			result.lexicon = engine.lexicon.Resolve(handle.NodeIdentity, limit)
		case "arcana":
			if arcana != nil && handle.NodeID != nil {
				node, inspectErr := arcana.Inspect(ctx, engine.arcanaSnapshot, *handle.NodeID)
				if inspectErr != nil {
					return resolvedAnchors{}, inspectErr
				}
				result.arcana = []structure.Node{node}
			}
		}
		return engine.addArcanaAnchors(ctx, result, limit, arcana), nil
	}

	anchor := value
	if anchor == "" {
		anchor = strings.TrimSpace(query)
	}
	var result resolvedAnchors
	if engine.lexicon != nil {
		result.lexicon = engine.lexicon.Resolve(anchor, limit)
		if len(result.lexicon) == 0 {
			for _, match := range engine.lexicon.Find(anchor, limit) {
				result.lexicon = append(result.lexicon, match.Node)
			}
		}
	}
	if arcana != nil && engine.arcanaSnapshot != "" && len(result.lexicon) == 0 {
		nodes, err := arcana.Resolve(ctx, engine.arcanaSnapshot, anchor, "", limit)
		if err != nil {
			return resolvedAnchors{}, err
		}
		result.arcana = nodes
	}
	return engine.addArcanaAnchors(ctx, result, limit, arcana), nil
}

func (engine *Engine) addArcanaAnchors(
	ctx context.Context,
	result resolvedAnchors,
	limit int,
	arcana arcanaQuery,
) resolvedAnchors {
	if arcana == nil || engine.arcanaSnapshot == "" || len(result.arcana) >= limit {
		return result
	}
	seen := make(map[uint32]bool)
	for _, node := range result.arcana {
		if node.NodeID != nil {
			seen[*node.NodeID] = true
		}
	}
	for _, node := range result.lexicon {
		nodes, err := arcana.ResolveTyped(
			ctx, engine.arcanaSnapshot, node.Name, node.Kind, node.Path, limit-len(result.arcana),
		)
		if err != nil {
			continue
		}
		for _, current := range nodes {
			if current.NodeID != nil && !seen[*current.NodeID] {
				seen[*current.NodeID] = true
				result.arcana = append(result.arcana, current)
			}
		}
		if len(result.arcana) >= limit {
			break
		}
	}
	return result
}

func identities(nodes []structure.Node) []string {
	result := make([]string, 0, len(nodes))
	for _, node := range nodes {
		if node.Identity != "" {
			result = append(result, node.Identity)
		}
	}
	return result
}
