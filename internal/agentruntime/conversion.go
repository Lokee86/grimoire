package agentruntime

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Lokee86/grimoire/internal/agentquery"
	"github.com/Lokee86/grimoire/internal/investigation"
	"github.com/Lokee86/grimoire/internal/knowledge"
)

func copyDiscoveryResponse(target *Response, source agentquery.Response) {
	target.ExactMatches = source.ExactMatches
	target.SourceMatches = source.SourceMatches
	target.SymbolMatches = source.SymbolMatches
	target.RelationshipMatches = source.RelationshipMatches
	target.Paths = source.Paths
	target.Dependents = source.Dependents
	target.Inspections = source.Inspections
	target.Unresolved = source.Unresolved
}

type investigationBuilder struct {
	response      *investigation.Response
	nodeIndex     map[string]int
	rangeIndex    map[string]int
	pathIndex     map[string]int
	documentIndex map[string]int
}

func newInvestigationBuilder(response *investigation.Response) *investigationBuilder {
	return &investigationBuilder{
		response:      response,
		nodeIndex:     make(map[string]int),
		rangeIndex:    make(map[string]int),
		pathIndex:     make(map[string]int),
		documentIndex: make(map[string]int),
	}
}

func (builder *investigationBuilder) addNode(node agentquery.Node) (investigation.EvidenceRef, bool) {
	if node.Handle.Value == "" {
		return investigation.EvidenceRef{}, false
	}
	if index, exists := builder.nodeIndex[node.Handle.Value]; exists {
		return investigation.EvidenceRef{Kind: "node", Index: index}, true
	}
	label := node.QualifiedName
	if label == "" {
		label = node.Name
	}
	index := len(builder.response.Nodes)
	builder.nodeIndex[node.Handle.Value] = index
	builder.response.Nodes = append(builder.response.Nodes, investigation.Node{
		ID: node.Handle.Value, Kind: node.Kind, Label: label, Path: node.Path,
	})
	return investigation.EvidenceRef{Kind: "node", Index: index}, true
}

func (builder *investigationBuilder) addNodeEvidence(node agentquery.Node, text string) (investigation.EvidenceRef, investigation.EvidenceRef, bool) {
	nodeRef, ok := builder.addNode(node)
	if !ok {
		return investigation.EvidenceRef{}, investigation.EvidenceRef{}, false
	}
	primary := nodeRef
	if node.Span != nil {
		if sourceRef, sourceOK := builder.addRange(*node.Span, text); sourceOK {
			primary = sourceRef
		}
	}
	return nodeRef, primary, true
}

func (builder *investigationBuilder) addRange(value agentquery.Range, text string) (investigation.EvidenceRef, bool) {
	if value.Path == "" || value.StartLine <= 0 || value.EndLine < value.StartLine {
		return investigation.EvidenceRef{}, false
	}
	key := sourceRangeKey(value)
	if index, exists := builder.rangeIndex[key]; exists {
		current := &builder.response.SourceRanges[index]
		if current.Text == "" {
			current.Text = text
		}
		return investigation.EvidenceRef{Kind: "source", Index: index}, true
	}
	index := len(builder.response.SourceRanges)
	builder.rangeIndex[key] = index
	builder.response.SourceRanges = append(builder.response.SourceRanges, investigation.SourceRange{
		Path: value.Path, StartLine: value.StartLine, StartColumn: value.StartColumn,
		EndLine: value.EndLine, EndColumn: value.EndColumn,
		Text: compactKnowledgeText(text, maxInvestigationEvidenceTextBytes),
	})
	return investigation.EvidenceRef{Kind: "source", Index: index}, true
}

func (builder *investigationBuilder) addGraphPath(graph investigation.GraphPath) (investigation.EvidenceRef, bool) {
	if len(graph.Nodes) == 0 {
		return investigation.EvidenceRef{}, false
	}
	if graph.ID == "" {
		graph.ID = graphPathID(graph.Nodes, graph.Edges)
	}
	if index, exists := builder.pathIndex[graph.ID]; exists {
		return investigation.EvidenceRef{Kind: "path", Index: index}, true
	}
	index := len(builder.response.GraphPaths)
	builder.pathIndex[graph.ID] = index
	builder.response.GraphPaths = append(builder.response.GraphPaths, graph)
	return investigation.EvidenceRef{Kind: "path", Index: index}, true
}

func (builder *investigationBuilder) addDocument(document investigation.Document) investigation.EvidenceRef {
	if index, exists := builder.documentIndex[document.ID]; exists {
		return investigation.EvidenceRef{Kind: "document", Index: index}
	}
	index := len(builder.response.Documents)
	builder.documentIndex[document.ID] = index
	builder.response.Documents = append(builder.response.Documents, document)
	return investigation.EvidenceRef{Kind: "document", Index: index}
}

func (builder *investigationBuilder) addHit(hit investigation.RetrievalHit) {
	builder.response.RetrievalHits = append(builder.response.RetrievalHits, hit)
}

func investigationResponse(query agentquery.Response, documents []knowledge.Result) investigation.Response {
	response := investigation.Response{Snapshot: investigation.Snapshot{
		Repository: query.Snapshot.Source, Providers: query.Snapshot.Providers,
	}}
	builder := newInvestigationBuilder(&response)

	addResults := func(lane string, results []agentquery.Result) {
		for _, result := range results {
			nodeRef, evidence, ok := builder.addNodeEvidence(result.Node, result.Excerpt)
			if !ok {
				continue
			}
			builder.addHit(investigation.RetrievalHit{
				Evidence: evidence, RelatedEvidence: relatedEvidence(evidence, nodeRef),
				Lane: lane, Provider: result.Provider,
				Rank: result.Rank, Score: result.Score, Reasons: append([]string(nil), result.Reasons...),
				DuplicateOf: result.DuplicateOf,
			})
		}
	}
	addResults("exact_matches", query.ExactMatches)
	addResults("source_matches", query.SourceMatches)
	addResults("symbol_matches", query.SymbolMatches)

	for _, match := range query.RelationshipMatches {
		subjectNode, subjectEvidence, subjectOK := builder.addNodeEvidence(match.Subject, "")
		objectNode, objectEvidence, objectOK := builder.addNodeEvidence(match.Object, "")
		if !subjectOK || !objectOK {
			continue
		}
		related := []investigation.EvidenceRef{subjectNode, subjectEvidence, objectNode, objectEvidence}
		var seed *investigation.RetrievalSeed
		if match.Seed != nil {
			seedRef, ok := builder.addNode(*match.Seed)
			if ok {
				seed = &investigation.RetrievalSeed{
					Evidence: seedRef, Lane: match.SeedLane, Provider: match.Seed.Handle.Provider,
					Rank: match.SeedRank, Score: match.SeedScore, Reasons: append([]string(nil), match.SeedReasons...),
				}
			}
		}
		for _, span := range match.Spans {
			if spanRef, ok := builder.addRange(span, ""); ok {
				related = append(related, spanRef)
			}
		}
		nodes := []string{match.Subject.Handle.Value, match.Object.Handle.Value}
		edges := []string{match.Relation}
		pathRef, ok := builder.addGraphPath(investigation.GraphPath{
			ID: graphPathID(nodes, edges), Nodes: nodes, Edges: edges, Label: match.Relation,
		})
		if !ok {
			continue
		}
		builder.addHit(investigation.RetrievalHit{
			Evidence: pathRef, RelatedEvidence: relatedEvidence(pathRef, related...),
			Lane: "relationship_matches", Provider: match.Provider,
			Rank: match.Rank, Reasons: append([]string(nil), match.Reasons...),
			Direction: match.Direction, Relation: match.Relation, Certainty: match.Certainty,
			Support: append([]string(nil), match.Evidence...), Seed: seed,
		})
	}

	for _, path := range query.Paths {
		graph := investigation.GraphPath{Label: path.Summary}
		related := make([]investigation.EvidenceRef, 0)
		for _, node := range path.Nodes {
			nodeRef, evidence, ok := builder.addNodeEvidence(node, "")
			if ok {
				related = append(related, nodeRef, evidence)
			}
			graph.Nodes = append(graph.Nodes, node.Handle.Value)
		}
		for _, step := range path.Steps {
			graph.Edges = append(graph.Edges, step.Relation)
			for _, span := range step.Spans {
				if spanRef, ok := builder.addRange(span, ""); ok {
					related = append(related, spanRef)
				}
			}
		}
		if len(graph.Nodes) == 0 {
			graph.Nodes = append(graph.Nodes, path.ContinuationHandles...)
			graph.Edges = append(graph.Edges, path.Relations...)
			for _, evidence := range path.Evidence {
				if evidenceRef, ok := builder.addTraceEvidence(evidence); ok {
					related = append(related, evidenceRef)
				}
			}
		}
		if graph.Label == "" {
			graph.Label = query.Mode
		}
		pathRef, ok := builder.addGraphPath(graph)
		if !ok {
			continue
		}
		builder.addHit(investigation.RetrievalHit{
			Evidence: pathRef, RelatedEvidence: relatedEvidence(pathRef, related...),
			Lane: "paths", Rank: path.Rank, Score: path.Score,
		})
	}

	for _, dependent := range query.Dependents {
		nodeRef, evidence, ok := builder.addNodeEvidence(dependent.Node, "")
		if !ok {
			continue
		}
		related := []investigation.EvidenceRef{nodeRef}
		for _, span := range dependent.Spans {
			if spanRef, ok := builder.addRange(span, ""); ok {
				related = append(related, spanRef)
			}
		}
		builder.addHit(investigation.RetrievalHit{
			Evidence: evidence, RelatedEvidence: relatedEvidence(evidence, related...),
			Lane: "dependents", Provider: dependent.Node.Handle.Provider,
			Depth: dependent.Depth, Direction: dependent.Direction, Relation: dependent.Relation,
			Certainty: dependent.Certainty, Support: append([]string(nil), dependent.Evidence...),
		})
	}
	for _, inspection := range query.Inspections {
		var primary investigation.EvidenceRef
		hasPrimary := false
		related := make([]investigation.EvidenceRef, 0, 3)
		if inspection.Node != nil {
			nodeRef, evidence, ok := builder.addNodeEvidence(*inspection.Node, "")
			if ok {
				primary, hasPrimary = evidence, true
				related = append(related, nodeRef, evidence)
			}
		}
		if inspection.Declaration != nil {
			if declaration, ok := builder.addRange(*inspection.Declaration, ""); ok {
				related = append(related, declaration)
				if !hasPrimary {
					primary, hasPrimary = declaration, true
				}
			}
		}
		if containing, ok := builder.addRange(inspection.ContainingSpan, inspection.Source); ok {
			primary, hasPrimary = containing, true
			related = append(related, containing)
		}
		if hasPrimary {
			builder.addHit(investigation.RetrievalHit{
				Evidence: primary, RelatedEvidence: relatedEvidence(primary, related...),
				Lane: "inspections", Provider: inspection.Handle.Provider,
			})
		}
	}
	for _, unresolved := range query.Unresolved {
		question := strings.TrimSpace(unresolved.Relation + " " + unresolved.Expression)
		if question == "" {
			question = strings.TrimSpace(unresolved.Reason)
		}
		if question != "" {
			response.UnresolvedQuestions = append(response.UnresolvedQuestions, investigation.UnresolvedQuestion{
				Question: question, Context: unresolved.Reason,
			})
		}
		if unresolved.Span != nil {
			if evidence, ok := builder.addRange(*unresolved.Span, ""); ok {
				builder.addHit(investigation.RetrievalHit{
					Evidence: evidence, Lane: "unresolved", Relation: unresolved.Relation,
					Reasons: nonEmptyStrings(unresolved.Reason),
				})
			}
		}
	}
	for index, document := range documents {
		title := document.Heading
		if title == "" {
			title = document.Path
		}
		documentRef := builder.addDocument(investigation.Document{
			ID: document.Handle, URI: document.Path, Title: title,
			Content: compactKnowledgeText(document.Text, maxInvestigationEvidenceTextBytes),
		})
		builder.addHit(investigation.RetrievalHit{
			Evidence: documentRef, Lane: "document_matches", Provider: "knowledge",
			Rank: index + 1, Score: document.Score, Reasons: append([]string(nil), document.Reasons...),
		})
	}
	return response
}

func (builder *investigationBuilder) addTraceEvidence(value agentquery.TraceEvidence) (investigation.EvidenceRef, bool) {
	if value.Path == "" || value.StartLine <= 0 || value.EndLine < value.StartLine {
		return investigation.EvidenceRef{}, false
	}
	return builder.addRange(agentquery.Range{
		Path: value.Path, StartLine: value.StartLine, EndLine: value.EndLine,
		Handle: agentquery.Handle{Value: value.Handle},
	}, "")
}

func relatedEvidence(primary investigation.EvidenceRef, values ...investigation.EvidenceRef) []investigation.EvidenceRef {
	seen := map[investigation.EvidenceRef]bool{primary: true}
	result := make([]investigation.EvidenceRef, 0, len(values))
	for _, value := range values {
		if value.Kind == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func sourceRangeKey(value agentquery.Range) string {
	return filepath.ToSlash(value.Path) + ":" +
		strconv.Itoa(value.StartLine) + ":" + strconv.Itoa(value.StartColumn) + ":" +
		strconv.Itoa(value.EndLine) + ":" + strconv.Itoa(value.EndColumn)
}

func nonEmptyStrings(values ...string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			result = append(result, value)
		}
	}
	return result
}

func graphPathID(nodes, edges []string) string {
	identity := strings.Join(nodes, "\x00") + "\x01" + strings.Join(edges, "\x00")
	digest := sha256.Sum256([]byte(identity))
	return "path:" + hex.EncodeToString(digest[:])
}
