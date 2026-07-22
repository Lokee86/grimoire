package main

import (
	"go/token"
	"strconv"

	"golang.org/x/tools/go/ssa"
)

func (s *scanner) ensureNamedFunction(namespace, name string) NodeKey {
	namespace = s.canonicalNamespace(namespace)
	identity := "function:" + namespace + ":" + name
	key := hashIdentity(identity)
	if _, exists := s.nodes[key]; exists {
		return key
	}
	path := s.pathForNamespace(namespace)
	parent := s.repositoryKey()
	if s.isInternalNamespace(namespace) {
		if packageKey, ok := s.packageNodeForNamespace(namespace); ok {
			parent = packageKey
		}
	} else {
		parent = s.ensureNamespaceNode(namespace)
	}
	s.addNode(NodeFact{Key: key, Kind: KindFunction, Path: path, Name: name})
	s.addEdge(parent, key, RelDefines, nil)
	return key
}

func (s *scanner) ensureUnknownMethod(receiver, name string) NodeKey {
	const namespace = "go:types"
	parent := s.ensureNamespaceNode(namespace)
	identity := "dynamic-method:" + receiver + "." + name
	key := hashIdentity(identity)
	s.addNode(NodeFact{Key: key, Kind: KindMethod, Path: s.pathForNamespace(namespace), Name: name})
	s.addEdge(parent, key, RelDefines, nil)
	return key
}

func (s *scanner) ensureSyntheticSSAFunction(function *ssa.Function, set *token.FileSet) (NodeKey, bool, bool) {
	if function == nil {
		return 0, false, false
	}
	namespace := s.ssaFunctionNamespace(function)
	if namespace == "" {
		namespace = "go:ssa"
	}
	internal := s.isInternalNamespace(namespace)
	identity := "ssa-function:" + namespace + ":" + function.String()
	if position := set.PositionFor(function.Pos(), false); position.IsValid() {
		identity += ":" + strconv.Itoa(position.Line) + ":" + strconv.Itoa(position.Column)
	}
	key := hashIdentity(identity)
	if _, exists := s.nodes[key]; exists {
		return key, internal, true
	}
	path := s.pathForNamespace(namespace)
	parent := s.repositoryKey()
	if internal {
		if packageKey, ok := s.packageNodeForNamespace(namespace); ok {
			parent = packageKey
		}
	} else {
		parent = s.ensureNamespaceNode(namespace)
	}
	name := function.Name()
	if name == "" {
		name = function.String()
	}
	s.addNode(NodeFact{Key: key, Kind: KindFunction, Path: path, Name: name})
	s.addEdge(parent, key, RelDefines, nil)
	return key, internal, true
}

func (s *scanner) ssaFunctionNamespace(function *ssa.Function) string {
	for current := function; current != nil; current = current.Parent() {
		if current.Pkg != nil && current.Pkg.Pkg != nil {
			return s.canonicalNamespace(current.Pkg.Pkg.Path())
		}
		if object := current.Object(); object != nil && object.Pkg() != nil {
			return s.canonicalNamespace(object.Pkg().Path())
		}
	}
	return ""
}
