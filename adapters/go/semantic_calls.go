package main

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
)

func (s *scanner) collectSemanticCalls(pkg *packages.Package, targets semanticTargets) {
	if pkg.TypesInfo == nil || pkg.Fset == nil {
		return
	}
	for _, file := range pkg.Syntax {
		rel, ok := s.semanticFilePath(pkg.Fset, file)
		if !ok {
			continue
		}
		importPath := s.importPathFor(rel)
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Body == nil {
				continue
			}
			source := declarationKey(importPath, rel, function)
			if _, exists := s.nodes[source]; !exists {
				continue
			}
			s.collectCallableSemanticCalls(pkg, rel, source, function.Body, targets)
		}
	}
}

func (s *scanner) collectCallableSemanticCalls(
	pkg *packages.Package,
	rel string,
	source NodeKey,
	body *ast.BlockStmt,
	targets semanticTargets,
) {
	ast.Inspect(body, func(node ast.Node) bool {
		if node != body {
			if literal, nested := node.(*ast.FuncLit); nested {
				position := pkg.Fset.PositionFor(literal.Pos(), false)
				closure, exists := s.closureKeys[closurePositionKey(rel, position)]
				if exists {
					s.collectCallableSemanticCalls(pkg, rel, closure, literal.Body, targets)
				}
				return false
			}
		}
		switch statement := node.(type) {
		case *ast.GoStmt:
			s.registerCallsitePosition(pkg, rel, source, statement.Call, statement.Pos())
		case *ast.DeferStmt:
			s.registerCallsitePosition(pkg, rel, source, statement.Call, statement.Pos())
		}
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		span := spanFromSet(pkg.Fset, call.Pos(), call.End(), rel)
		key := callsiteKey(source, span)
		s.registerCallsitePosition(pkg, rel, source, call, call.Pos())
		if call.Lparen.IsValid() {
			s.registerCallsitePosition(pkg, rel, source, call, call.Lparen)
		}
		resolution := s.resolveTypedCall(pkg, call, targets)
		s.mergeSemanticCall(key, resolution)
		return true
	})
}

func (s *scanner) registerCallsitePosition(
	pkg *packages.Package,
	rel string,
	source NodeKey,
	call *ast.CallExpr,
	position token.Pos,
) {
	span := spanFromSet(pkg.Fset, call.Pos(), call.End(), rel)
	key := callsiteKey(source, span)
	start := pkg.Fset.PositionFor(position, false)
	s.callsiteKeys[callsiteStartKey(source, rel, start)] = key
}

func (s *scanner) resolveTypedCall(
	pkg *packages.Package,
	call *ast.CallExpr,
	targets semanticTargets,
) semanticCall {
	if literal := calledFuncLiteral(call.Fun); literal != nil {
		position := pkg.Fset.PositionFor(literal.Pos(), false)
		if rel, err := s.relative(position.Filename); err == nil {
			if closure, exists := s.closureKeys[closurePositionKey(rel, position)]; exists {
				return semanticCall{
					edges:    []semanticEdge{{target: closure, relation: RelCalls}},
					resolved: true,
					class:    callClassDynamic,
				}
			}
		}
	}
	if typed, exists := pkg.TypesInfo.Types[call.Fun]; exists && typed.IsType() {
		return semanticCall{
			edges:    []semanticEdge{{target: s.ensureTypeNode(typed.Type), relation: RelConvertsTo}},
			resolved: true,
			class:    callClassConversion,
		}
	}
	object := calledObject(pkg.TypesInfo, call.Fun)
	switch object := object.(type) {
	case *types.Builtin:
		return semanticCall{
			edges:    []semanticEdge{{target: s.ensureBuiltinNode(object.Name()), relation: RelCalls}},
			resolved: true,
			class:    callClassBuiltin,
		}
	case *types.TypeName:
		return semanticCall{
			edges:    []semanticEdge{{target: s.ensureTypeNode(object.Type()), relation: RelConvertsTo}},
			resolved: true,
			class:    callClassConversion,
		}
	case *types.Func:
		target, internal, unambiguous := s.ensureFunctionNode(object, targets, pkg.Fset)
		namespace := s.canonicalNamespace(objectNamespace(object))
		if !unambiguous {
			return semanticCall{reason: ReasonAmbiguousTarget, namespace: namespace, name: object.Name()}
		}
		class := callClassExternal
		if internal {
			class = callClassInternal
		}
		if isInterfaceCall(pkg.TypesInfo, call.Fun) {
			class = callClassInterface
		}
		return semanticCall{
			edges:    []semanticEdge{{target: target, relation: RelCalls}},
			resolved: true,
			class:    class,
		}
	case nil:
		reason, namespace, name := classifyNonIdentifier(call.Fun)
		return semanticCall{reason: reason, namespace: namespace, name: name, class: callClassDynamic}
	default:
		return semanticCall{reason: ReasonDynamicTarget, name: expressionName(call.Fun), class: callClassDynamic}
	}
}

func calledFuncLiteral(expression ast.Expr) *ast.FuncLit {
	for {
		switch typed := expression.(type) {
		case *ast.FuncLit:
			return typed
		case *ast.ParenExpr:
			expression = typed.X
		default:
			return nil
		}
	}
}

func calledObject(info *types.Info, expression ast.Expr) types.Object {
	switch expression := expression.(type) {
	case *ast.Ident:
		return info.Uses[expression]
	case *ast.SelectorExpr:
		if selection := info.Selections[expression]; selection != nil {
			return selection.Obj()
		}
		return info.Uses[expression.Sel]
	case *ast.IndexExpr:
		return calledObject(info, expression.X)
	case *ast.IndexListExpr:
		return calledObject(info, expression.X)
	case *ast.ParenExpr:
		return calledObject(info, expression.X)
	default:
		return nil
	}
}

func isInterfaceCall(info *types.Info, expression ast.Expr) bool {
	for {
		switch typed := expression.(type) {
		case *ast.IndexExpr:
			expression = typed.X
		case *ast.IndexListExpr:
			expression = typed.X
		case *ast.ParenExpr:
			expression = typed.X
		default:
			selector, ok := expression.(*ast.SelectorExpr)
			if !ok {
				return false
			}
			selection := info.Selections[selector]
			return selection != nil && isInterfaceType(selection.Recv())
		}
	}
}
