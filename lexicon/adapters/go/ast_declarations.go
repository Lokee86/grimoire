package main

import (
	"fmt"
	"go/ast"
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
			if interfaceType, ok := typeSpec.Type.(*ast.InterfaceType); ok {
				s.parseInterfaceMethods(pkg, path, typeSpec.Name.Name, key, interfaceType)
			}
		}
	case *ast.FuncDecl:
		kind, identity := declarationIdentity(pkg.importKey, path, declaration)
		key := hashIdentity(identity)
		node := NodeFact{Key: key, Kind: kind, Path: path, Name: declaration.Name.Name, Span: s.span(declaration.Pos(), declaration.End(), path)}
		s.addNode(node)
		s.addEdge(pkg.key, key, RelDefines, node.Span)
		s.registerSemanticID(astSemanticFunctionID(pkg.importKey, declaration), key)

		scope := pkg.importKey + "\x00" + pkg.name
		if declaration.Recv == nil {
			s.targets[scope] = appendTarget(s.targets[scope], declaration.Name.Name, key)
		}
		if declaration.Body != nil {
			s.callables = append(s.callables, callable{
				packageKey: scope,
				namespace:  pkg.importKey,
				source:     key,
				body:       declaration.Body,
				path:       path,
			})
			s.collectClosures(pkg, path, key, declaration.Body)
		}
	}
}

func (s *scanner) parseInterfaceMethods(
	pkg packageInfo,
	path string,
	interfaceName string,
	interfaceKey NodeKey,
	interfaceType *ast.InterfaceType,
) {
	if interfaceType.Methods == nil {
		return
	}
	for _, field := range interfaceType.Methods.List {
		if len(field.Names) == 0 {
			continue
		}
		if _, ok := field.Type.(*ast.FuncType); !ok {
			continue
		}
		for _, name := range field.Names {
			key := hashIdentity(interfaceMethodIdentity(pkg.importKey, interfaceName, name.Name))
			span := s.span(field.Pos(), field.End(), path)
			s.addNode(NodeFact{Key: key, Kind: KindMethod, Path: path, Name: name.Name, Span: span})
			s.addEdge(interfaceKey, key, RelDefines, span)
		}
	}
}

func (s *scanner) collectClosures(pkg packageInfo, path string, parent NodeKey, body *ast.BlockStmt) {
	ast.Inspect(body, func(node ast.Node) bool {
		literal, ok := node.(*ast.FuncLit)
		if !ok {
			return true
		}
		position := s.set.PositionFor(literal.Pos(), false)
		identity := closureIdentity(pkg.importKey, path, position)
		key := hashIdentity(identity)
		span := s.span(literal.Pos(), literal.End(), path)
		name := fmt.Sprintf("closure@%d:%d", position.Line, position.Column)
		s.addNode(NodeFact{Key: key, Kind: KindFunction, Path: path, Name: name, Span: span})
		s.addEdge(parent, key, RelDefines, span)
		s.closureKeys[closurePositionKey(path, position)] = key
		s.callables = append(s.callables, callable{
			packageKey: pkg.importKey + "\x00" + pkg.name,
			namespace:  pkg.importKey,
			source:     key,
			body:       literal.Body,
			path:       path,
		})
		s.summary.Closures++
		s.collectClosures(pkg, path, key, literal.Body)
		return false
	})
}

func declarationIdentity(importPath, path string, declaration *ast.FuncDecl) (NodeKind, string) {
	name := declaration.Name.Name
	if declaration.Recv != nil {
		receiver := receiverName(declaration.Recv)
		return KindMethod, "method:" + importPath + ":" + receiver + "." + name
	}
	if strings.HasSuffix(path, "_test.go") && strings.HasPrefix(name, "Test") {
		return KindTest, "test:" + importPath + ":" + name
	}
	return KindFunction, "function:" + importPath + ":" + name
}

func declarationKey(importPath, path string, declaration *ast.FuncDecl) NodeKey {
	_, identity := declarationIdentity(importPath, path, declaration)
	return hashIdentity(identity)
}

func appendTarget(targets map[string][]NodeKey, name string, key NodeKey) map[string][]NodeKey {
	if targets == nil {
		targets = make(map[string][]NodeKey)
	}
	targets[name] = appendUniqueKey(targets[name], key)
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
	case *ast.ParenExpr:
		return expressionName(expression.X)
	default:
		return "anonymous"
	}
}
