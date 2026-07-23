package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
)

// collectSemanticDataflow emits only edges whose target is a types.Object
// declared by this repository. It deliberately omits imported, builtin, and
// unresolved names instead of manufacturing unsound symbol nodes.
func (s *scanner) collectSemanticDataflow(pkg *packages.Package, rel string, source NodeKey, body *ast.BlockStmt) {
	if pkg.TypesInfo == nil || pkg.Fset == nil || body == nil {
		return
	}
	visitor := &dataflowVisitor{scanner: s, pkg: pkg, rel: rel, source: source}
	visitor.visitBlock(body)
}

type dataflowVisitor struct {
	scanner *scanner
	pkg     *packages.Package
	rel     string
	source  NodeKey
}

func (v *dataflowVisitor) visitBlock(block *ast.BlockStmt) {
	for _, statement := range block.List {
		v.visitStmt(statement)
	}
}

func (v *dataflowVisitor) visitStmt(statement ast.Stmt) {
	switch statement := statement.(type) {
	case *ast.AssignStmt:
		for _, value := range statement.Rhs {
			v.visitExpr(value)
		}
		for _, target := range statement.Lhs {
			v.visitTarget(target, statement.Tok != token.ASSIGN)
		}
	case *ast.IncDecStmt:
		v.visitTarget(statement.X, true)
	case *ast.DeclStmt:
		if declaration, ok := statement.Decl.(*ast.GenDecl); ok {
			v.visitGenDecl(declaration)
		}
	default:
		ast.Inspect(statement, func(node ast.Node) bool {
			if node == statement {
				return true
			}
			if nested, ok := node.(ast.Stmt); ok {
				v.visitStmt(nested)
				return false
			}
			if _, nested := node.(*ast.FuncLit); nested {
				return false
			}
			if expression, ok := node.(ast.Expr); ok {
				v.visitExpr(expression)
				return false
			}
			return true
		})
	}
}

func (v *dataflowVisitor) visitGenDecl(declaration *ast.GenDecl) {
	for _, specification := range declaration.Specs {
		value, ok := specification.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for _, initializer := range value.Values {
			v.visitExpr(initializer)
		}
		for _, name := range value.Names {
			v.visitTarget(name, false)
		}
	}
}

func (v *dataflowVisitor) visitTarget(expression ast.Expr, compound bool) {
	switch expression := expression.(type) {
	case *ast.Ident:
		if compound {
			v.addObject(expression, true)
		}
		v.addObject(expression, false, true)
	case *ast.SelectorExpr:
		v.visitExpr(expression.X)
		if compound {
			v.addObject(expression.Sel, true)
		}
		v.addObject(expression.Sel, false, true)
	case *ast.ParenExpr:
		v.visitTarget(expression.X, compound)
	case *ast.StarExpr:
		v.visitExpr(expression.X)
	case *ast.IndexExpr:
		v.visitExpr(expression.X)
		v.visitExpr(expression.Index)
	case *ast.IndexListExpr:
		v.visitExpr(expression.X)
		for _, index := range expression.Indices {
			v.visitExpr(index)
		}
	default:
		v.visitExpr(expression)
	}
}

func (v *dataflowVisitor) visitExpr(expression ast.Expr) {
	switch expression := expression.(type) {
	case *ast.Ident:
		v.addObject(expression, false)
	case *ast.SelectorExpr:
		v.visitExpr(expression.X)
		v.addObject(expression.Sel, false)
	case *ast.FuncLit:
		return
	default:
		ast.Inspect(expression, func(node ast.Node) bool {
			if node == expression {
				return true
			}
			if _, nested := node.(*ast.FuncLit); nested {
				return false
			}
			if identifier, ok := node.(*ast.Ident); ok {
				v.addObject(identifier, false)
				return false
			}
			if selector, ok := node.(*ast.SelectorExpr); ok {
				v.visitExpr(selector)
				return false
			}
			return true
		})
	}
}

func (v *dataflowVisitor) addObject(identifier *ast.Ident, write bool, forceWrite ...bool) {
	object := v.pkg.TypesInfo.ObjectOf(identifier)
	if object == nil || !v.scanner.isInternalObject(object) {
		return
	}
	target := v.scanner.ensureDataSymbol(object, v.pkg.Fset)
	if target == "" {
		return
	}
	relation := RelReads
	if write || (len(forceWrite) > 0 && forceWrite[0]) {
		relation = RelWrites
	}
	v.scanner.addEdge(v.source, target, relation, spanFromSet(v.pkg.Fset, identifier.Pos(), identifier.End(), v.rel))
}

func (s *scanner) isInternalObject(object types.Object) bool {
	return object != nil && object.Pkg() != nil && s.isInternalNamespace(object.Pkg().Path())
}

func (s *scanner) ensureDataSymbol(object types.Object, set *token.FileSet) NodeKey {
	if !s.isInternalObject(object) {
		return ""
	}
	position := set.PositionFor(object.Pos(), false)
	rel, err := s.relative(position.Filename)
	if err != nil {
		return ""
	}
	kind := KindVariable
	switch value := object.(type) {
	case *types.Const:
		kind = KindConstant
	case *types.Var:
		if value.IsField() {
			kind = KindField
		}
	}
	identity := string(kind) + ":" + object.Pkg().Path() + ":" + rel + ":" + fmt.Sprintf("%d:%d", position.Line, position.Column) + ":" + object.Name()
	key := hashIdentity(identity)
	if _, exists := s.nodes[key]; !exists {
		span := &SourceSpan{
			Path: rel, StartLine: uint32(position.Line), StartColumn: uint32(position.Column),
			EndLine: uint32(position.Line), EndColumn: uint32(position.Column + len(object.Name())),
		}
		s.addNode(NodeFact{Key: key, Kind: kind, Path: rel, Name: object.Name(), Span: span})
	}
	return key
}
