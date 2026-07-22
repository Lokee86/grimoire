package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"path"
	"path/filepath"
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

type NodeKey string
type ContentID string

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
	Repository string
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
	kind := identityKind(identity)
	payload := "lexicon:v1\x00go\x00" + kind + "\x00" + identity
	digest := sha256.Sum256([]byte(payload))
	return NodeKey("sha256:" + hex.EncodeToString(digest[:]))
}

func identityKind(identity string) string {
	prefix, _, _ := strings.Cut(identity, ":")
	switch prefix {
	case "package":
		return "module"
	case "closure", "ssa-function":
		return "function"
	case "interface-method", "dynamic-method":
		return "method"
	case "type-expression":
		return "type"
	case "capture":
		return "variable"
	default:
		return prefix
	}
}

func contentID(bytes []byte) ContentID {
	digest := sha256.Sum256(bytes)
	return ContentID("sha256:" + hex.EncodeToString(digest[:]))
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
