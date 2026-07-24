package main

import tree_sitter "github.com/tree-sitter/go-tree-sitter"

func (extractor *extractor) collectPointerBinding(left, right *tree_sitter.Node, context extractionContext) {
	if left == nil || right == nil || context.CallableID == "" {
		return
	}
	target := callableReferenceCandidate(right, extractor.source)
	if target == "" {
		return
	}
	candidate := ""
	member := false
	switch left.Kind() {
	case "identifier":
		candidate = nodeText(left, extractor.source)
	case "field_expression":
		candidate = nodeText(left.ChildByFieldName("field"), extractor.source)
		member = true
	}
	candidate = lastQualifiedPart(candidate)
	if candidate == "" {
		return
	}
	extractor.file.PointerBindings = append(extractor.file.PointerBindings, pointerBindingObservation{
		SourceID: context.CallableID, SourceScope: context.CallableScope, Path: extractor.file.Path,
		Candidate: candidate, Target: target, Member: member, Span: spanForNode(extractor.file.Path, left),
	})
}

func (extractor *extractor) collectDesignatedPointerBindings(node *tree_sitter.Node, context extractionContext) {
	if node == nil {
		return
	}
	if node.Kind() == "initializer_pair" {
		designator := node.ChildByFieldName("designator")
		value := node.ChildByFieldName("value")
		field := firstDescendant(designator, "field_identifier")
		target := callableReferenceCandidate(value, extractor.source)
		if field != nil && target != "" {
			extractor.file.PointerBindings = append(extractor.file.PointerBindings, pointerBindingObservation{
				SourceID: context.CallableID, SourceScope: context.CallableScope, Path: extractor.file.Path,
				Candidate: nodeText(field, extractor.source), Target: target, Member: true,
				Span: spanForNode(extractor.file.Path, node),
			})
		}
	}
	for _, child := range namedChildren(node) {
		extractor.collectDesignatedPointerBindings(child, context)
	}
}
