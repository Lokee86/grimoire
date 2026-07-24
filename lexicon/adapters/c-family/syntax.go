package main

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"regexp"
	"strings"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

var whitespacePattern = regexp.MustCompile(`\s+`)

func nodeText(node *tree_sitter.Node, source []byte) string {
	if node == nil {
		return ""
	}
	return node.Utf8Text(source)
}

func namedChildren(node *tree_sitter.Node) []*tree_sitter.Node {
	if node == nil {
		return nil
	}
	children := make([]*tree_sitter.Node, 0, node.NamedChildCount())
	for index := uint(0); index < node.NamedChildCount(); index++ {
		if child := node.NamedChild(index); child != nil {
			children = append(children, child)
		}
	}
	return children
}

func firstDescendant(node *tree_sitter.Node, kinds ...string) *tree_sitter.Node {
	if node == nil {
		return nil
	}
	for _, kind := range kinds {
		if node.Kind() == kind {
			return node
		}
	}
	for _, child := range namedChildren(node) {
		if match := firstDescendant(child, kinds...); match != nil {
			return match
		}
	}
	return nil
}

func descendants(node *tree_sitter.Node, kind string) []*tree_sitter.Node {
	var result []*tree_sitter.Node
	var visit func(*tree_sitter.Node)
	visit = func(current *tree_sitter.Node) {
		if current == nil {
			return
		}
		if current.Kind() == kind {
			result = append(result, current)
		}
		for _, child := range namedChildren(current) {
			visit(child)
		}
	}
	visit(node)
	return result
}

func spanForNode(path string, node *tree_sitter.Node) sourceSpan {
	start := node.StartPosition()
	end := node.EndPosition()
	return sourceSpan{
		Path:        filepath.ToSlash(path),
		StartLine:   int(start.Row) + 1,
		StartColumn: int(start.Column) + 1,
		EndLine:     int(end.Row) + 1,
		EndColumn:   int(end.Column) + 1,
	}
}

func normalizeSpace(value string) string {
	return strings.TrimSpace(whitespacePattern.ReplaceAllString(value, " "))
}

func anonymousName(kind string, node *tree_sitter.Node, source []byte) string {
	hash := sha256.Sum256([]byte(normalizeSpace(nodeText(node, source))))
	return "(anonymous " + kind + " " + hex.EncodeToString(hash[:6]) + ")"
}

func lastQualifiedPart(value string) string {
	value = normalizeQualified(value)
	if index := strings.LastIndex(value, "::"); index >= 0 {
		return value[index+2:]
	}
	return value
}

func stripIncludeTarget(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "<")
	value = strings.TrimSuffix(value, ">")
	value = strings.TrimPrefix(value, `"`)
	value = strings.TrimSuffix(value, `"`)
	return filepath.ToSlash(value)
}

func isExpressionKind(kind string) bool {
	return strings.HasSuffix(kind, "_expression") || kind == "initializer_list" || kind == "argument_list"
}
