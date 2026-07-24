package main

import (
	"fmt"
	"os"
	"path/filepath"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_c "github.com/tree-sitter/tree-sitter-c/bindings/go"
	tree_sitter_cpp "github.com/tree-sitter/tree-sitter-cpp/bindings/go"
)

type extractor struct {
	file   *sourceFile
	source []byte
}

func analyzeRepository(root string, changedFiles, removedFiles []string, incremental bool) ([]byte, error) {
	model, err := buildRepositoryModel(root)
	if err != nil {
		return nil, err
	}
	facts := emitRepositoryFacts(model, changedFiles, removedFiles, incremental)
	return facts.render()
}

func buildRepositoryModel(root string) (*repositoryModel, error) {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	paths, err := collectSources(absoluteRoot)
	if err != nil {
		return nil, err
	}
	compileLanguages := loadCompileLanguages(absoluteRoot)
	contents := make(map[string][]byte, len(paths))
	for _, path := range paths {
		content, err := os.ReadFile(filepath.Join(absoluteRoot, filepath.FromSlash(path)))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		contents[path] = content
	}
	headerLanguages := inferHeaderLanguages(paths, contents, compileLanguages)

	cParser := tree_sitter.NewParser()
	defer cParser.Close()
	if err := cParser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_c.Language())); err != nil {
		return nil, fmt.Errorf("configure C parser: %w", err)
	}
	cppParser := tree_sitter.NewParser()
	defer cppParser.Close()
	if err := cppParser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_cpp.Language())); err != nil {
		return nil, fmt.Errorf("configure C++ parser: %w", err)
	}

	model := &repositoryModel{Repository: filepath.Base(filepath.Clean(absoluteRoot))}
	for _, path := range paths {
		content := contents[path]
		language := classifyLanguage(path, content, compileLanguages, headerLanguages)
		tree, parserLanguage, err := parseSource(content, language, isAmbiguousHeaderPath(path), cParser, cppParser)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		file := &sourceFile{
			Path: path, Language: language, ParserLanguage: parserLanguage, Content: content,
			FileID: nodeID("file", path), ModuleID: nodeID("module", path),
			ParseError: tree.RootNode().HasError(),
		}
		extractor := extractor{file: file, source: content}
		extractor.walk(tree.RootNode(), extractionContext{ContainerID: file.ModuleID})
		tree.Close()
		model.Files = append(model.Files, file)
		model.Declarations = append(model.Declarations, file.Declarations...)
	}
	return model, nil
}

func parseSource(content []byte, language string, allowFallback bool, cParser, cppParser *tree_sitter.Parser) (*tree_sitter.Tree, string, error) {
	primary := cParser
	alternate := cppParser
	alternateLanguage := "cpp"
	if language == "cpp" {
		primary = cppParser
		alternate = cParser
		alternateLanguage = "c"
	}
	tree := primary.Parse(content, nil)
	if tree == nil {
		return nil, "", fmt.Errorf("tree-sitter returned no tree")
	}
	if !allowFallback || !tree.RootNode().HasError() {
		return tree, language, nil
	}

	fallback := alternate.Parse(content, nil)
	if fallback == nil {
		return tree, language, nil
	}
	if syntaxErrorScore(fallback.RootNode()) < syntaxErrorScore(tree.RootNode()) {
		tree.Close()
		return fallback, alternateLanguage, nil
	}
	fallback.Close()
	return tree, language, nil
}

func syntaxErrorScore(node *tree_sitter.Node) int {
	if node == nil {
		return 0
	}
	score := 0
	if node.IsError() {
		score += 10
	}
	if node.IsMissing() {
		score++
	}
	for index := uint(0); index < node.ChildCount(); index++ {
		score += syntaxErrorScore(node.Child(index))
	}
	return score
}
