package main

import (
	"strings"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

func (extractor *extractor) extractExpression(node *tree_sitter.Node, context extractionContext) {
	if node == nil || context.CallableID == "" {
		return
	}
	switch node.Kind() {
	case "call_expression":
		function := node.ChildByFieldName("function")
		candidate, member := callCandidate(function, extractor.source)
		if candidate != "" {
			extractor.file.Calls = append(extractor.file.Calls, callObservation{
				SourceID: context.CallableID, SourceScope: context.CallableScope, Path: extractor.file.Path,
				Expression: nodeText(function, extractor.source), Candidate: candidate, Member: member,
				Span: spanForNode(extractor.file.Path, node),
			})
		}
		if function != nil && function.Kind() == "field_expression" {
			extractor.extractExpression(function.ChildByFieldName("argument"), context)
		}
		if arguments := node.ChildByFieldName("arguments"); arguments != nil {
			for _, child := range namedChildren(arguments) {
				extractor.extractExpression(child, context)
			}
		}
		return
	case "assignment_expression":
		left := node.ChildByFieldName("left")
		right := node.ChildByFieldName("right")
		extractor.collectTarget(left, context, "writes")
		if assignmentOperator(node, left, right, extractor.source) != "=" {
			extractor.collectTarget(left, context, "reads")
		}
		extractor.extractExpression(right, context)
		return
	case "update_expression":
		target := firstExpressionChild(node)
		extractor.collectTarget(target, context, "reads")
		extractor.collectTarget(target, context, "writes")
		return
	case "field_expression":
		field := node.ChildByFieldName("field")
		if field != nil {
			extractor.addAccess(field, context, "reads", true)
		}
		extractor.extractExpression(node.ChildByFieldName("argument"), context)
		return
	case "subscript_expression":
		extractor.extractExpression(node.ChildByFieldName("argument"), context)
		extractor.extractExpression(node.ChildByFieldName("index"), context)
		return
	case "identifier":
		extractor.addAccess(node, context, "reads", false)
		return
	case "field_identifier":
		extractor.addAccess(node, context, "reads", true)
		return
	case "type_identifier", "primitive_type", "number_literal", "string_literal", "char_literal", "true", "false", "null", "nullptr":
		return
	}
	for _, child := range namedChildren(node) {
		extractor.extractExpression(child, context)
	}
}

func (extractor *extractor) collectTarget(node *tree_sitter.Node, context extractionContext, relation string) {
	if node == nil {
		return
	}
	switch node.Kind() {
	case "identifier":
		extractor.addAccess(node, context, relation, false)
	case "field_identifier":
		extractor.addAccess(node, context, relation, true)
	case "field_expression":
		if field := node.ChildByFieldName("field"); field != nil {
			extractor.addAccess(field, context, relation, true)
		}
		extractor.extractExpression(node.ChildByFieldName("argument"), context)
	case "subscript_expression":
		extractor.collectTarget(node.ChildByFieldName("argument"), context, relation)
		extractor.extractExpression(node.ChildByFieldName("index"), context)
	default:
		if declarator := node.ChildByFieldName("argument"); declarator != nil {
			extractor.collectTarget(declarator, context, relation)
			return
		}
		for _, child := range namedChildren(node) {
			extractor.collectTarget(child, context, relation)
		}
	}
}

func (extractor *extractor) addAccess(node *tree_sitter.Node, context extractionContext, relation string, member bool) {
	candidate := lastQualifiedPart(nodeText(node, extractor.source))
	if candidate == "" || candidate == "this" || candidate == "self" {
		return
	}
	extractor.file.Accesses = append(extractor.file.Accesses, accessObservation{
		SourceID: context.CallableID, SourceScope: context.CallableScope, ParentType: context.TypeID,
		Path: extractor.file.Path, Expression: nodeText(node, extractor.source), Candidate: candidate,
		Relation: relation, Member: member, Span: spanForNode(extractor.file.Path, node),
	})
}

func callCandidate(node *tree_sitter.Node, source []byte) (string, bool) {
	if node == nil {
		return "", false
	}
	switch node.Kind() {
	case "identifier", "type_identifier", "qualified_identifier", "scoped_identifier", "operator_name":
		return normalizeQualified(nodeText(node, source)), false
	case "field_expression":
		field := node.ChildByFieldName("field")
		return lastQualifiedPart(nodeText(field, source)), true
	case "template_function":
		if name := node.ChildByFieldName("name"); name != nil {
			return normalizeQualified(nodeText(name, source)), false
		}
	}
	return "", false
}

func assignmentOperator(node, left, right *tree_sitter.Node, source []byte) string {
	if node == nil || left == nil || right == nil || left.EndByte() > right.StartByte() {
		return "="
	}
	return strings.TrimSpace(string(source[left.EndByte():right.StartByte()]))
}

func firstExpressionChild(node *tree_sitter.Node) *tree_sitter.Node {
	for _, child := range namedChildren(node) {
		if child.Kind() != "comment" {
			return child
		}
	}
	return nil
}
