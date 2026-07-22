package main

import (
	"fmt"
	"hash/fnv"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

type NodeKind string

const (
	KindRepository NodeKind = "repository"
	KindDirectory  NodeKind = "directory"
	KindFile       NodeKind = "file"
	KindPackage    NodeKind = "module"
	KindNamespace  NodeKind = "namespace"
	KindImport     NodeKind = "import"
	KindType       NodeKind = "type"
	KindFunction   NodeKind = "function"
	KindMethod     NodeKind = "method"
	KindVariable   NodeKind = "variable"
	KindTest       NodeKind = "test"
)

type RelationKind string

const (
	RelContains      RelationKind = "contains"
	RelDefines       RelationKind = "defines"
	RelImports       RelationKind = "imports"
	RelCalls         RelationKind = "calls"
	RelPossibleCalls RelationKind = "possible-calls"
	RelConvertsTo    RelationKind = "converts-to"
	RelImplements    RelationKind = "implements"
	RelExtends       RelationKind = "extends"
	RelReferences    RelationKind = "references"
)

type NodeKey uint64
type ContentID uint64

type SourceSpan struct {
	Path                   string
	StartLine, StartColumn uint32
	EndLine, EndColumn     uint32
}

type NodeFact struct {
	Key       NodeKey
	Kind      NodeKind
	Path      string
	Name      string
	ContentID *ContentID
	Span      *SourceSpan
}

type EdgeFact struct {
	Source, Target NodeKey
	Relation       RelationKind
	Span           *SourceSpan
}

type UnresolvedReason string

const (
	ReasonMissingTarget   UnresolvedReason = "missing-target"
	ReasonAmbiguousTarget UnresolvedReason = "ambiguous-target"
	ReasonUnsupportedForm UnresolvedReason = "unsupported-form"
	ReasonDynamicTarget   UnresolvedReason = "dynamic-target"
	ReasonExternalTarget  UnresolvedReason = "external-target"
	ReasonBuiltinTarget   UnresolvedReason = "builtin-target"
	ReasonTypeConversion  UnresolvedReason = "type-conversion"
	ReasonSelfTarget      UnresolvedReason = "self-target"
)

type UnresolvedReferenceFact struct {
	Source             NodeKey
	Relation           RelationKind
	Expression         string
	CandidateNamespace string
	CandidateName      string
	Reason             UnresolvedReason
	Span               *SourceSpan
}

type RepositoryFacts struct {
	Nodes      []NodeFact
	Edges      []EdgeFact
	Unresolved []UnresolvedReferenceFact
}

func hashBytes(bytes []byte) uint64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write(bytes)
	return hasher.Sum64()
}

func hashIdentity(identity string) NodeKey {
	return NodeKey(hashBytes([]byte(identity)))
}

func contentID(bytes []byte) ContentID {
	return ContentID(hashBytes(bytes))
}

func formatID(value uint64) string {
	return fmt.Sprintf("%016x", value)
}

func encodeFacts(facts RepositoryFacts) string {
	nodes := append([]NodeFact(nil), facts.Nodes...)
	edges := append([]EdgeFact(nil), facts.Edges...)
	unresolved := append([]UnresolvedReferenceFact(nil), facts.Unresolved...)
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Key != nodes[j].Key {
			return nodes[i].Key < nodes[j].Key
		}
		if nodes[i].Kind != nodes[j].Kind {
			return nodes[i].Kind < nodes[j].Kind
		}
		if nodes[i].Path != nodes[j].Path {
			return nodes[i].Path < nodes[j].Path
		}
		return nodes[i].Name < nodes[j].Name
	})
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].Source != edges[j].Source {
			return edges[i].Source < edges[j].Source
		}
		if edges[i].Target != edges[j].Target {
			return edges[i].Target < edges[j].Target
		}
		if edges[i].Relation != edges[j].Relation {
			return edges[i].Relation < edges[j].Relation
		}
		return spanText(edges[i].Span) < spanText(edges[j].Span)
	})
	sort.Slice(unresolved, func(i, j int) bool {
		left, right := unresolved[i], unresolved[j]
		if left.Source != right.Source {
			return left.Source < right.Source
		}
		if left.Relation != right.Relation {
			return left.Relation < right.Relation
		}
		if left.Expression != right.Expression {
			return left.Expression < right.Expression
		}
		if left.CandidateNamespace != right.CandidateNamespace {
			return left.CandidateNamespace < right.CandidateNamespace
		}
		if left.CandidateName != right.CandidateName {
			return left.CandidateName < right.CandidateName
		}
		if left.Reason != right.Reason {
			return left.Reason < right.Reason
		}
		return spanText(left.Span) < spanText(right.Span)
	})

	var output strings.Builder
	output.WriteString("version\t2\n")
	for _, node := range nodes {
		fields := []string{"N", formatID(uint64(node.Key)), string(node.Kind), node.Path, node.Name, "-"}
		if node.ContentID != nil {
			fields[5] = formatID(uint64(*node.ContentID))
		}
		fields = append(fields, spanFields(node.Span)...)
		writeRecord(&output, fields)
	}
	for _, edge := range edges {
		fields := []string{"E", formatID(uint64(edge.Source)), formatID(uint64(edge.Target)), string(edge.Relation)}
		fields = append(fields, spanFields(edge.Span)...)
		writeRecord(&output, fields)
	}
	for _, reference := range unresolved {
		fields := []string{
			"U", formatID(uint64(reference.Source)), string(reference.Relation),
			string(reference.Reason), reference.Expression, optionalField(reference.CandidateNamespace),
			optionalField(reference.CandidateName),
		}
		fields = append(fields, spanFields(reference.Span)...)
		writeRecord(&output, fields)
	}
	return output.String()
}

func optionalField(value string) string {
	if value == "" {
		return "-"
	}
	return value
}

func writeRecord(output *strings.Builder, fields []string) {
	for index, field := range fields {
		if index > 0 {
			output.WriteByte('\t')
		}
		escapeField(output, field)
	}
	output.WriteByte('\n')
}

func spanFields(span *SourceSpan) []string {
	if span == nil {
		return []string{"-", "-", "-", "-", "-"}
	}
	return []string{span.Path, fmt.Sprint(span.StartLine), fmt.Sprint(span.StartColumn), fmt.Sprint(span.EndLine), fmt.Sprint(span.EndColumn)}
}

func spanText(span *SourceSpan) string {
	return strings.Join(spanFields(span), "\x00")
}

func escapeField(output *strings.Builder, value string) {
	for _, character := range value {
		switch character {
		case '\\':
			output.WriteString("\\\\")
		case '\t':
			output.WriteString("\\t")
		case '\n':
			output.WriteString("\\n")
		case '\r':
			output.WriteString("\\r")
		case '\x00':
			output.WriteString("\\0")
		default:
			output.WriteRune(character)
		}
	}
}

func normalizePath(value string) (string, error) {
	value = filepath.ToSlash(value)
	if value == "" || strings.HasPrefix(value, "/") || filepath.VolumeName(value) != "" {
		return "", fmt.Errorf("invalid repository path %q", value)
	}
	cleaned := path.Clean(value)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("invalid repository path %q", value)
	}
	return cleaned, nil
}
