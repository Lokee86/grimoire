package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"
)

type semanticEdge struct {
	target   NodeKey
	relation RelationKind
}

type callClass uint8

const (
	callClassInternal callClass = iota
	callClassExternal
	callClassBuiltin
	callClassConversion
	callClassDynamic
	callClassInterface
)

func (s *scanner) repositoryKey() NodeKey {
	return hashIdentity("repository:" + s.repository)
}

func (s *scanner) canonicalNamespace(namespace string) string {
	if namespace == "" {
		return namespace
	}
	if strings.HasSuffix(namespace, "_test") {
		base := strings.TrimSuffix(namespace, "_test")
		if _, ok := s.moduleForNamespace(base); ok {
			return base
		}
	}
	return namespace
}

func (s *scanner) isInternalNamespace(namespace string) bool {
	namespace = s.canonicalNamespace(namespace)
	_, ok := s.moduleForNamespace(namespace)
	return ok
}

func (s *scanner) registerSemanticID(id string, key NodeKey) {
	s.semanticIDs[id] = appendUniqueKey(s.semanticIDs[id], key)
}

func (s *scanner) semanticIDCandidates(id string) []NodeKey {
	candidates := append([]NodeKey(nil), s.baseSemanticIDs[id]...)
	for _, key := range s.semanticIDs[id] {
		candidates = appendUniqueKey(candidates, key)
	}
	return candidates
}

func astSemanticFunctionID(importPath string, declaration *ast.FuncDecl) string {
	if declaration.Recv == nil {
		return "function:" + importPath + ":" + declaration.Name.Name
	}
	return "method:" + importPath + ":" + receiverName(declaration.Recv) + "." + declaration.Name.Name
}

func interfaceMethodIdentity(importPath, interfaceName, methodName string) string {
	return "interface-method:" + importPath + ":" + interfaceName + "." + methodName
}

func closureIdentity(importPath, path string, position token.Position) string {
	return fmt.Sprintf("closure:%s:%s:%d:%d", importPath, filepath.ToSlash(path), position.Line, position.Column)
}

func closurePositionKey(path string, position token.Position) string {
	return fmt.Sprintf("%s/%d/%d", filepath.ToSlash(path), position.Line, position.Column)
}

func callsiteStartKey(source NodeKey, path string, position token.Position) string {
	return fmt.Sprintf("%s/%s/%d/%d", source, filepath.ToSlash(path), position.Line, position.Column)
}

func semanticFunctionID(function *types.Func, namespace string) string {
	if origin := function.Origin(); origin != nil {
		function = origin
	}
	signature, _ := function.Type().(*types.Signature)
	if signature != nil && signature.Recv() != nil {
		return "method:" + namespace + ":" + receiverTypeName(signature.Recv().Type()) + "." + function.Name()
	}
	return "function:" + namespace + ":" + function.Name()
}

func (s *scanner) internalFunctionCandidates(function *types.Func, targets semanticTargets) []NodeKey {
	if target, exists := targets.byObject[function]; exists {
		return []NodeKey{target}
	}
	namespace := s.canonicalNamespace(objectNamespace(function))
	id := semanticFunctionID(function, namespace)
	candidates := append([]NodeKey(nil), targets.byID[id]...)
	for _, key := range s.semanticIDCandidates(id) {
		candidates = appendUniqueKey(candidates, key)
	}
	return candidates
}

func (s *scanner) ensureFunctionNode(
	function *types.Func,
	targets semanticTargets,
	set *token.FileSet,
) (NodeKey, bool, bool) {
	namespace := s.canonicalNamespace(objectNamespace(function))
	if s.isInternalNamespace(namespace) {
		id := semanticFunctionID(function, namespace)
		canonical := hashIdentity(id)
		if s.hasNode(canonical) {
			return canonical, true, true
		}
		candidates := s.internalFunctionCandidates(function, targets)
		if len(candidates) == 1 {
			return candidates[0], true, true
		}
		if len(candidates) > 1 {
			return "", true, false
		}
		key := s.ensureTypedInternalFunction(function, namespace, set)
		return key, true, true
	}
	return s.ensureExternalFunction(function, namespace), false, true
}
