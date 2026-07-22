package main

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"
)

func (s *scanner) parseDeclaration(declaration ast.Decl, pkg packageInfo, path string) {
	switch declaration := declaration.(type) {
	case *ast.GenDecl:
		for _, specification := range declaration.Specs {
			typeSpec, ok := specification.(*ast.TypeSpec)
			if !ok {
				continue
			}
			key := hashIdentity("type:" + pkg.importKey + ":" + typeSpec.Name.Name)
			node := NodeFact{Key: key, Kind: KindType, Path: path, Name: typeSpec.Name.Name, Span: s.span(typeSpec.Pos(), typeSpec.End(), path)}
			s.addNode(node)
			s.addEdge(pkg.key, key, RelDefines, node.Span)
		}
	case *ast.FuncDecl:
		name := declaration.Name.Name
		kind := KindFunction
		identity := "function:" + pkg.importKey + ":" + name
		if declaration.Recv != nil {
			kind = KindMethod
			receiver := receiverName(declaration.Recv)
			identity = "method:" + pkg.importKey + ":" + receiver + "." + name
		} else if strings.HasSuffix(path, "_test.go") && strings.HasPrefix(name, "Test") {
			kind = KindTest
			identity = "test:" + pkg.importKey + ":" + name
		}
		key := hashIdentity(identity)
		node := NodeFact{Key: key, Kind: kind, Path: path, Name: name, Span: s.span(declaration.Pos(), declaration.End(), path)}
		s.addNode(node)
		s.addEdge(pkg.key, key, RelDefines, node.Span)
		if declaration.Recv == nil && declaration.Body != nil {
			scope := pkg.importKey + "\x00" + pkg.name
			s.targets[scope] = appendTarget(s.targets[scope], name, key)
			s.callables = append(s.callables, callable{packageKey: scope, source: key, body: declaration.Body, path: path})
		}
	}
}

func appendTarget(targets map[string][]NodeKey, name string, key NodeKey) map[string][]NodeKey {
	if targets == nil {
		targets = make(map[string][]NodeKey)
	}
	targets[name] = append(targets[name], key)
	return targets
}

func receiverName(fields *ast.FieldList) string {
	if fields == nil || len(fields.List) == 0 {
		return ""
	}
	return expressionName(fields.List[0].Type)
}

func expressionName(expression ast.Expr) string {
	switch expression := expression.(type) {
	case *ast.Ident:
		return expression.Name
	case *ast.StarExpr:
		return "*" + expressionName(expression.X)
	case *ast.SelectorExpr:
		return expressionName(expression.X) + "." + expression.Sel.Name
	case *ast.IndexExpr:
		return expressionName(expression.X)
	case *ast.IndexListExpr:
		return expressionName(expression.X)
	default:
		return "anonymous"
	}
}

func (s *scanner) addCallEdges() {
	for _, function := range s.callables {
		targets := s.targets[function.packageKey]
		ast.Inspect(function.body, func(node ast.Node) bool {
			if node != function.body {
				if _, nested := node.(*ast.FuncLit); nested {
					return false
				}
			}
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			s.summary.CallExpressions++
			identifier, ok := call.Fun.(*ast.Ident)
			if !ok || len(targets[identifier.Name]) != 1 {
				s.summary.UnresolvedCalls++
				return true
			}
			target := targets[identifier.Name][0]
			if target == function.source {
				s.summary.UnresolvedCalls++
				return true
			}
			s.addEdge(function.source, target, RelCalls, s.span(call.Pos(), call.End(), function.path))
			s.summary.DirectCalls++
			return true
		})
	}
}

func (s *scanner) span(start, end token.Pos, path string) *SourceSpan {
	if !start.IsValid() || !end.IsValid() {
		return nil
	}
	begin := s.set.PositionFor(start, false)
	finish := s.set.PositionFor(end, false)
	return &SourceSpan{Path: path, StartLine: uint32(begin.Line), StartColumn: uint32(begin.Column), EndLine: uint32(finish.Line), EndColumn: uint32(finish.Column)}
}

func isGoFile(path string) bool { return filepath.Ext(path) == ".go" }
