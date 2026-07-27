package agentruntime

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
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

func investigationResponse(query agentquery.Response, documents []knowledge.Result) investigation.Response {
	response := investigation.Response{Snapshot: investigation.Snapshot{
		Repository: query.Snapshot.Source,
		Providers:  query.Snapshot.Providers,
	}}
	nodes := make(map[string]investigation.Node)
	ranges := make(map[string]investigation.SourceRange)
	addNode := func(node agentquery.Node) {
		if node.Handle.Value == "" {
			return
		}
		label := node.QualifiedName
		if label == "" {
			label = node.Name
		}
		nodes[node.Handle.Value] = investigation.Node{
			ID: node.Handle.Value, Kind: node.Kind, Label: label, Path: node.Path,
			Metadata: map[string]string{"provider": node.Handle.Provider},
		}
		if node.Span != nil {
			addRange(ranges, *node.Span, "")
		}
	}
	for _, lane := range [][]agentquery.Result{
		query.ExactMatches,
		query.SourceMatches,
		query.SymbolMatches,
	} {
		for _, result := range lane {
			addNode(result.Node)
		}
	}
	for _, relationship := range query.RelationshipMatches {
		addNode(relationship.Subject)
		addNode(relationship.Object)
		nodes := []string{relationship.Subject.Handle.Value, relationship.Object.Handle.Value}
		edges := []string{relationship.Relation}
		response.GraphPaths = append(response.GraphPaths, investigation.GraphPath{
			ID: graphPathID(nodes, edges), Nodes: nodes, Edges: edges, Label: relationship.Relation,
		})
		for _, span := range relationship.Spans {
			addRange(ranges, span, "")
		}
	}
	for _, path := range query.Paths {
		graph := investigation.GraphPath{}
		for _, node := range path.Nodes {
			addNode(node)
			graph.Nodes = append(graph.Nodes, node.Handle.Value)
		}
		for _, step := range path.Steps {
			graph.Edges = append(graph.Edges, step.Relation)
			for _, span := range step.Spans {
				addRange(ranges, span, "")
			}
		}
		if len(graph.Nodes) == 0 {
			graph.Nodes = append(graph.Nodes, path.ContinuationHandles...)
			graph.Edges = append(graph.Edges, path.Relations...)
			for _, evidence := range path.Evidence {
				addTraceEvidence(ranges, evidence)
			}
		}
		graph.ID = graphPathID(graph.Nodes, graph.Edges)
		graph.Label = path.Summary
		if graph.Label == "" {
			graph.Label = query.Mode
		}
		response.GraphPaths = append(response.GraphPaths, graph)
	}
	for _, dependent := range query.Dependents {
		addNode(dependent.Node)
		for _, span := range dependent.Spans {
			addRange(ranges, span, "")
		}
	}
	for _, inspection := range query.Inspections {
		if inspection.Node != nil {
			addNode(*inspection.Node)
		}
		if inspection.Declaration != nil {
			addRange(ranges, *inspection.Declaration, "")
		}
		addRange(ranges, inspection.ContainingSpan, inspection.Source)
	}
	for _, unresolved := range query.Unresolved {
		question := strings.TrimSpace(unresolved.Relation + " " + unresolved.Expression)
		if question == "" {
			question = strings.TrimSpace(unresolved.Reason)
		}
		if question != "" {
			response.UnresolvedQuestions = append(response.UnresolvedQuestions, investigation.UnresolvedQuestion{
				Question: question,
				Context:  unresolved.Reason,
			})
		}
		if unresolved.Span != nil {
			addRange(ranges, *unresolved.Span, "")
		}
	}
	nodeKeys := make([]string, 0, len(nodes))
	for key := range nodes {
		nodeKeys = append(nodeKeys, key)
	}
	sort.Strings(nodeKeys)
	for _, key := range nodeKeys {
		response.Nodes = append(response.Nodes, nodes[key])
	}
	rangeKeys := make([]string, 0, len(ranges))
	for key := range ranges {
		rangeKeys = append(rangeKeys, key)
	}
	sort.Strings(rangeKeys)
	for _, key := range rangeKeys {
		response.SourceRanges = append(response.SourceRanges, ranges[key])
	}
	for _, document := range documents {
		title := document.Heading
		if title == "" {
			title = document.Path
		}
		response.Documents = append(response.Documents, investigation.Document{
			ID: document.Handle, URI: document.Handle, Title: title, Content: document.Text,
			Metadata: map[string]string{
				"path": document.Path,
				"kind": string(document.Kind),
			},
		})
	}
	return response
}

func addTraceEvidence(target map[string]investigation.SourceRange, value agentquery.TraceEvidence) {
	if value.Path == "" || value.StartLine <= 0 || value.EndLine < value.StartLine {
		return
	}
	key := value.Handle
	if key == "" {
		key = value.Path + ":" + strconv.Itoa(value.StartLine) + ":" + strconv.Itoa(value.EndLine)
	}
	target[key] = investigation.SourceRange{
		Path: value.Path, StartLine: value.StartLine, EndLine: value.EndLine,
	}
}

func addRange(target map[string]investigation.SourceRange, value agentquery.Range, text string) {
	if value.Path == "" || value.StartLine <= 0 || value.EndLine < value.StartLine {
		return
	}
	key := value.Handle.Value
	if key == "" {
		key = value.Path + ":" + strconv.Itoa(value.StartLine) + ":" + strconv.Itoa(value.EndLine)
	}
	target[key] = investigation.SourceRange{
		Path: value.Path, StartLine: value.StartLine, StartColumn: value.StartColumn,
		EndLine: value.EndLine, EndColumn: value.EndColumn, Text: text,
	}
}

func graphPathID(nodes, edges []string) string {
	identity := strings.Join(nodes, "\x00") + "\x01" + strings.Join(edges, "\x00")
	digest := sha256.Sum256([]byte(identity))
	return "path:" + hex.EncodeToString(digest[:])
}
