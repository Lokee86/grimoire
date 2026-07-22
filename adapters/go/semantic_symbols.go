package main

import (
	"go/token"
	"go/types"
	"strings"
)

func (s *scanner) ensureTypedInternalFunction(function *types.Func, namespace string, set *token.FileSet) NodeKey {
	id := semanticFunctionID(function, namespace)
	key := hashIdentity(id)
	if _, exists := s.nodes[key]; exists {
		s.registerSemanticID(id, key)
		return key
	}
	kind := KindFunction
	if signature, ok := function.Type().(*types.Signature); ok && signature.Recv() != nil {
		kind = KindMethod
	}
	path := s.pathForNamespace(namespace)
	var span *SourceSpan
	if set != nil && function.Pos().IsValid() {
		position := set.PositionFor(function.Pos(), false)
		if rel, err := s.relative(position.Filename); err == nil {
			path = rel
			span = &SourceSpan{
				Path: rel, StartLine: uint32(position.Line), StartColumn: uint32(position.Column),
				EndLine: uint32(position.Line), EndColumn: uint32(position.Column),
			}
		}
	}
	s.addNode(NodeFact{Key: key, Kind: kind, Path: path, Name: function.Name(), Span: span})
	if parent, ok := s.packageNodeForNamespace(namespace); ok {
		s.addEdge(parent, key, RelDefines, span)
	}
	s.registerSemanticID(id, key)
	return key
}

func (s *scanner) ensureExternalFunction(function *types.Func, namespace string) NodeKey {
	if namespace == "" {
		namespace = "go:unknown"
	}
	id := semanticFunctionID(function, namespace)
	key := hashIdentity(id)
	if _, exists := s.nodes[key]; exists {
		return key
	}
	kind := KindFunction
	if signature, ok := function.Type().(*types.Signature); ok && signature.Recv() != nil {
		kind = KindMethod
	}
	parent := s.ensureNamespaceNode(namespace)
	s.addNode(NodeFact{Key: key, Kind: kind, Path: s.pathForNamespace(namespace), Name: function.Name()})
	s.addEdge(parent, key, RelDefines, nil)
	return key
}

func (s *scanner) ensureBuiltinNode(name string) NodeKey {
	const namespace = "go:builtins"
	parent := s.ensureNamespaceNode(namespace)
	key := hashIdentity("function:" + namespace + ":" + name)
	s.addNode(NodeFact{Key: key, Kind: KindFunction, Path: s.pathForNamespace(namespace), Name: name})
	s.addEdge(parent, key, RelDefines, nil)
	return key
}

func (s *scanner) ensureTypeNode(value types.Type) NodeKey {
	value = types.Unalias(value)
	for {
		pointer, ok := value.(*types.Pointer)
		if !ok {
			break
		}
		value = pointer.Elem()
	}
	if named, ok := value.(*types.Named); ok {
		object := named.Obj()
		namespace := s.canonicalNamespace(objectNamespace(object))
		key := hashIdentity("type:" + namespace + ":" + object.Name())
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
		s.addNode(NodeFact{Key: key, Kind: KindType, Path: path, Name: object.Name()})
		s.addEdge(parent, key, RelDefines, nil)
		return key
	}
	name := types.TypeString(value, func(pkg *types.Package) string { return pkg.Path() })
	const namespace = "go:types"
	parent := s.ensureNamespaceNode(namespace)
	key := hashIdentity("type-expression:" + name)
	s.addNode(NodeFact{Key: key, Kind: KindType, Path: s.pathForNamespace(namespace), Name: name})
	s.addEdge(parent, key, RelDefines, nil)
	return key
}

func (s *scanner) ensureNamespaceNode(namespace string) NodeKey {
	key := hashIdentity("namespace:" + namespace)
	if _, exists := s.nodes[key]; exists {
		return key
	}
	s.addNode(NodeFact{Key: key, Kind: KindNamespace, Path: s.pathForNamespace(namespace), Name: namespace})
	s.addEdge(s.repositoryKey(), key, RelContains, nil)
	return key
}

func (s *scanner) packageNodeForNamespace(namespace string) (NodeKey, bool) {
	namespace = s.canonicalNamespace(namespace)
	for _, pkg := range s.packages {
		if pkg.importKey == namespace {
			return pkg.key, true
		}
	}
	return "", false
}

func (s *scanner) pathForNamespace(namespace string) string {
	namespace = s.canonicalNamespace(namespace)
	switch namespace {
	case "go:builtins":
		return "@builtin/go"
	case "go:types":
		return "@types/go"
	case "go:unknown":
		return "@external/go-unknown"
	}
	if s.isInternalNamespace(namespace) {
		relative := strings.TrimPrefix(namespace, s.module)
		relative = strings.TrimPrefix(relative, "/")
		if relative == "" {
			return ".lexicon-repository"
		}
		return relative
	}
	if isStandardLibraryNamespace(namespace) {
		return "@stdlib/" + namespace
	}
	return "@external/" + namespace
}
