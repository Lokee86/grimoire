package main

import tree_sitter "github.com/tree-sitter/go-tree-sitter"

func callArguments(node *tree_sitter.Node, source []byte) []string {
	if node == nil {
		return nil
	}
	arguments := make([]string, 0, node.NamedChildCount())
	for _, child := range namedChildren(node) {
		arguments = append(arguments, callableReferenceCandidate(child, source))
	}
	return arguments
}

func callArgumentExpressions(node *tree_sitter.Node, source []byte) []string {
	if node == nil {
		return nil
	}
	expressions := make([]string, 0, node.NamedChildCount())
	for _, child := range namedChildren(node) {
		expressions = append(expressions, nodeText(child, source))
	}
	return expressions
}

func callableReferenceCandidate(node *tree_sitter.Node, source []byte) string {
	if node == nil {
		return ""
	}
	switch node.Kind() {
	case "identifier", "type_identifier", "qualified_identifier", "scoped_identifier", "operator_name":
		return normalizeQualified(nodeText(node, source))
	case "pointer_expression", "unary_expression", "parenthesized_expression", "cast_expression":
		for _, child := range namedChildren(node) {
			if candidate := callableReferenceCandidate(child, source); candidate != "" {
				return candidate
			}
		}
	}
	return ""
}
