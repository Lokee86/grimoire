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
		content, err := os.ReadFile(filepath.Join(absoluteRoot, filepath.FromSlash(path)))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		language := classifyLanguage(path, content, compileLanguages)
		parser := cParser
		if language == "cpp" {
			parser = cppParser
		}
		tree := parser.Parse(content, nil)
		if tree == nil {
			return nil, fmt.Errorf("parse %s: tree-sitter returned no tree", path)
		}
		file := &sourceFile{
			Path: path, Language: language, Content: content,
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
