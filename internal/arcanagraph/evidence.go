package arcanagraph

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Lokee86/grimoire/internal/structure"
)

func evidenceFromResponses(
	resolved []resolvedSeed,
	responses map[string]protocolResponse,
) []structure.Evidence {
	result := make([]structure.Evidence, 0, len(resolved)*3)
	for index, seed := range resolved {
		if role, ok := decodeResponse[roleResult](responses[fmt.Sprintf("role-%d", index)]); ok {
			result = append(result, roleEvidence(seed, role))
		}
		if impact, ok := decodeResponse[impactResult](responses[fmt.Sprintf("impact-%d", index)]); ok && len(impact.Dependents) > 0 {
			result = append(result, impactEvidence(seed, impact))
		}
		if unresolved, ok := decodeResponse[unresolvedResult](responses[fmt.Sprintf("unresolved-%d", index)]); ok && len(unresolved.Unresolved) > 0 {
			result = append(result, unresolvedEvidence(seed, unresolved))
		}
	}
	chainSeeds := min(len(resolved), maxChainSeeds)
	for from := 0; from < chainSeeds; from++ {
		for to := 0; to < chainSeeds; to++ {
			if from == to {
				continue
			}
			chain, ok := decodeResponse[chainResult](responses[fmt.Sprintf("chain-%d-%d", from, to)])
			if !ok || !chain.Found || chain.Chain == nil {
				continue
			}
			result = append(result, callChainEvidence(resolved[from], resolved[to], *chain.Chain))
		}
	}
	return result
}

func roleEvidence(seed resolvedSeed, role roleResult) structure.Evidence {
	relationships := make([]structure.Relationship, 0, len(role.Callers)+len(role.Callees))
	for _, caller := range role.Callers {
		relationships = append(relationships, relationship("incoming", caller))
	}
	for _, callee := range role.Callees {
		relationships = append(relationships, relationship("outgoing", callee))
	}
	truncated := len(relationships) > 16
	if truncated {
		relationships = relationships[:16]
	}
	node := role.Node.toStructure()
	return structure.Evidence{
		Provider: "arcana", Kind: "operational_role",
		Reasons: []string{"Arcana graph role for Lexicon-matched symbol " + seed.seed.Name},
		Node:    &node, Summary: role.Summary, Relationships: relationships, Truncated: truncated,
	}
}

func impactEvidence(seed resolvedSeed, impact impactResult) structure.Evidence {
	dependents := make([]structure.DepthNode, len(impact.Dependents))
	for index, dependent := range impact.Dependents {
		dependents[index] = structure.DepthNode{Depth: dependent.Depth, Node: dependent.Node.toStructure()}
	}
	node := seed.node.toStructure()
	return structure.Evidence{
		Provider: "arcana", Kind: "impact",
		Reasons: []string{"Arcana transitive dependents for Lexicon-matched symbol " + seed.seed.Name},
		Node:    &node, Dependents: dependents, Truncated: impact.Truncated,
	}
}

func unresolvedEvidence(seed resolvedSeed, unresolved unresolvedResult) structure.Evidence {
	items := make([]structure.Unresolved, len(unresolved.Unresolved))
	for index, item := range unresolved.Unresolved {
		items[index] = structure.Unresolved{
			Relation: item.Relation, Expression: item.Expression,
			CandidateNamespace: item.CandidateNamespace, CandidateName: item.CandidateName,
			Reason: item.Reason, Span: item.Span.toStructure(),
		}
	}
	node := seed.node.toStructure()
	return structure.Evidence{
		Provider: "arcana", Kind: "unresolved",
		Reasons: []string{"Arcana unresolved references owned by Lexicon-matched symbol " + seed.seed.Name},
		Node:    &node, Unresolved: items, Truncated: unresolved.Truncated,
	}
}

func callChainEvidence(from, to resolvedSeed, chain arcanaPath) structure.Evidence {
	nodes := make([]structure.Node, len(chain.Nodes))
	for index, node := range chain.Nodes {
		nodes[index] = node.toStructure()
	}
	return structure.Evidence{
		Provider: "arcana", Kind: "call_chain",
		Reasons: []string{fmt.Sprintf("Arcana shortest call chain from %s to %s", from.seed.Name, to.seed.Name)},
		Chain:   &structure.Path{Depth: chain.Depth, Nodes: nodes, Relations: chain.Relations},
	}
}

func relationship(direction string, related relatedNode) structure.Relationship {
	return structure.Relationship{
		Direction: direction, Relation: related.Relation,
		Certainty: relationCertainty(related.Relation), Node: related.Node.toStructure(),
	}
}

func relationCertainty(relation string) string {
	if strings.Contains(relation, "possible") {
		return "possible"
	}
	return "definite"
}

func decodeNodeList(response protocolResponse) nodeListResult {
	value, _ := decodeResponse[nodeListResult](response)
	return value
}

func decodeResponse[T any](response protocolResponse) (T, bool) {
	var result T
	if !response.OK || len(response.Result) == 0 {
		return result, false
	}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		return result, false
	}
	return result, true
}

func (node arcanaNode) toStructure() structure.Node {
	nodeID := node.NodeID
	return structure.Node{
		Identity: node.Identity, NodeID: &nodeID, Kind: node.Kind,
		Name: node.Name, Path: node.Path, Span: node.Span.toStructure(),
	}
}

func (span *arcanaSpan) toStructure() *structure.Span {
	if span == nil {
		return nil
	}
	return &structure.Span{
		Path: span.Path, StartLine: span.StartLine, StartColumn: span.StartColumn,
		EndLine: span.EndLine, EndColumn: span.EndColumn,
	}
}
