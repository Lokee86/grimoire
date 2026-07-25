package interstack

import (
	"path/filepath"
	"sort"
	"strings"
)

const (
	Language       = "interstack"
	AdapterVersion = "0.1.0"
)

type Span struct {
	Path        string `json:"path"`
	StartLine   uint32 `json:"start_line"`
	StartColumn uint32 `json:"start_column"`
	EndLine     uint32 `json:"end_line"`
	EndColumn   uint32 `json:"end_column"`
}

type Node struct {
	ID            string         `json:"id"`
	Kind          string         `json:"kind"`
	Name          string         `json:"name"`
	Path          string         `json:"path"`
	QualifiedName string         `json:"qualified_name"`
	Span          *Span          `json:"span,omitempty"`
	Attributes    map[string]any `json:"attributes,omitempty"`
}

type Library struct {
	Language   string
	Repository string
	Nodes      []Node
}

type factEdge struct {
	Source     string
	Target     string
	Relation   string
	Span       *Span
	Attributes map[string]any
}

type factUnresolved struct {
	Source     string
	Relation   string
	Expression string
	Reason     string
	Span       *Span
	Attributes map[string]any
}

type Result struct {
	Repository string
	Nodes      []Node
	Edges      []factEdge
	Unresolved []factUnresolved
	Summary    Summary
}

type Summary struct {
	HTTPContracts   int
	HTTPLinks       int
	MessageChannels int
	MessageLinks    int
	ConfigKeys      int
}

type sourceIndex struct {
	callables map[string][]Node
	files     map[string]Node
	byQName   map[string][]Node
	byName    map[string][]Node
}

func newSourceIndex(libraries []Library) sourceIndex {
	index := sourceIndex{
		callables: make(map[string][]Node),
		files:     make(map[string]Node),
		byQName:   make(map[string][]Node),
		byName:    make(map[string][]Node),
	}
	for _, library := range libraries {
		for _, node := range library.Nodes {
			path := normalizeSourcePath(node.Path)
			node.Path = path
			index.byQName[node.QualifiedName] = append(index.byQName[node.QualifiedName], node)
			index.byName[node.Name] = append(index.byName[node.Name], node)
			if node.Kind == "file" && path != "" {
				index.files[path] = node
			}
			if isCallable(node.Kind) && path != "" && node.Span != nil {
				index.callables[path] = append(index.callables[path], node)
			}
		}
	}
	for path := range index.callables {
		sort.Slice(index.callables[path], func(left, right int) bool {
			leftSpan := index.callables[path][left].Span
			rightSpan := index.callables[path][right].Span
			if leftSpan.StartLine != rightSpan.StartLine {
				return leftSpan.StartLine < rightSpan.StartLine
			}
			return leftSpan.StartColumn < rightSpan.StartColumn
		})
	}
	return index
}

func (index sourceIndex) ownerAt(path string, line uint32) (Node, bool) {
	path = normalizeSourcePath(path)
	candidates := index.callables[path]
	var nearest Node
	nearestFound := false
	for _, candidate := range candidates {
		span := candidate.Span
		if span.StartLine > line {
			break
		}
		nearest = candidate
		nearestFound = true
		if span.EndLine >= line && span.StartLine <= line {
			return candidate, true
		}
	}
	if nearestFound {
		return nearest, true
	}
	file, ok := index.files[path]
	return file, ok
}

func (index sourceIndex) exactQName(name string) (Node, bool) {
	values := index.byQName[name]
	if len(values) != 1 {
		return Node{}, false
	}
	return values[0], true
}

func (index sourceIndex) callableByName(name, preferredPath string) (Node, bool) {
	values := index.byName[name]
	bestScore := -1
	var best Node
	tied := false
	preferredComponent := componentForPath(preferredPath)
	preferredDirectory := filepath.ToSlash(filepath.Dir(filepath.FromSlash(preferredPath)))
	for _, candidate := range values {
		if !isCallable(candidate.Kind) {
			continue
		}
		score := 0
		if componentForPath(candidate.Path) == preferredComponent {
			score += 4
		}
		if strings.HasPrefix(candidate.Path, preferredDirectory+"/") {
			score += 2
		}
		if score > bestScore {
			bestScore = score
			best = candidate
			tied = false
		} else if score == bestScore {
			tied = true
		}
	}
	return best, bestScore >= 0 && !tied
}

func isCallable(kind string) bool {
	switch kind {
	case "function", "method", "constructor", "test":
		return true
	default:
		return false
	}
}

func normalizeSourcePath(value string) string {
	value = filepath.ToSlash(value)
	value = strings.TrimPrefix(value, "./")
	return value
}

func componentForPath(value string) string {
	parts := strings.Split(normalizeSourcePath(value), "/")
	if len(parts) >= 2 && parts[0] == "services" {
		return strings.Join(parts[:2], "/")
	}
	if len(parts) > 0 {
		return parts[0]
	}
	return "repository"
}
