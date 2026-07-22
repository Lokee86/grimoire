package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"path/filepath"
)

func classifyNonIdentifier(expression ast.Expr) (UnresolvedReason, string, string) {
	if selector, ok := expression.(*ast.SelectorExpr); ok {
		return ReasonUnsupportedForm, expressionName(selector.X), selector.Sel.Name
	}
	return ReasonDynamicTarget, "", expressionName(expression)
}

func expressionText(set *token.FileSet, expression ast.Expr) string {
	var output bytes.Buffer
	if err := printer.Fprint(&output, set, expression); err != nil {
		return expressionName(expression)
	}
	return output.String()
}

func isBuiltin(name string) bool {
	switch name {
	case "append", "cap", "clear", "close", "complex", "copy", "delete", "imag", "len",
		"make", "max", "min", "new", "panic", "print", "println", "real", "recover":
		return true
	default:
		return false
	}
}

func (s *scanner) span(start, end token.Pos, path string) *SourceSpan {
	return spanFromSet(s.set, start, end, path)
}

func spanFromSet(set *token.FileSet, start, end token.Pos, path string) *SourceSpan {
	if !start.IsValid() || !end.IsValid() {
		return nil
	}
	begin := set.PositionFor(start, false)
	finish := set.PositionFor(end, false)
	return &SourceSpan{
		Path: path, StartLine: uint32(begin.Line), StartColumn: uint32(begin.Column),
		EndLine: uint32(finish.Line), EndColumn: uint32(finish.Column),
	}
}

func callsiteKey(source NodeKey, span *SourceSpan) string {
	if span == nil {
		return fmt.Sprintf("%016x/-", source)
	}
	return fmt.Sprintf("%016x/%s/%d/%d/%d/%d", source, span.Path, span.StartLine,
		span.StartColumn, span.EndLine, span.EndColumn)
}

func isGoFile(path string) bool { return filepath.Ext(path) == ".go" }
