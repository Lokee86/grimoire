package main

import "path/filepath"

func emitRepositoryFacts(model *repositoryModel, changedFiles, removedFiles []string, incremental bool) *factSet {
	facts := newFactSet(model.Repository, changedFiles, removedFiles, incremental)
	index := buildDeclarationIndex(model.Declarations)
	files := buildFileIndex(model.Files)

	for _, file := range model.Files {
		fileAttributes := map[string]any{"language": file.Language, "parser": "tree-sitter", "parser_language": file.ParserLanguage}
		if file.ParseError {
			fileAttributes["parse_error"] = true
		}
		facts.addNode(file.Path, map[string]any{
			"attributes": fileAttributes, "content_id": contentID(file.Content), "id": file.FileID,
			"kind": "file", "name": filepath.Base(filepath.FromSlash(file.Path)), "owner": file.Path,
			"path": file.Path, "qualified_name": file.Path, "record": "node",
		})
		facts.addNode(file.Path, map[string]any{
			"attributes": map[string]any{"language": file.Language}, "id": file.ModuleID,
			"kind": "module", "name": moduleName(file.Path), "owner": file.Path,
			"path": file.Path, "qualified_name": file.Path, "record": "node",
		})
		facts.addEdge(file.Path, map[string]any{
			"owner": file.Path, "record": "edge", "relation": "contains", "source": file.FileID, "target": file.ModuleID,
		})
		for _, declaration := range file.Declarations {
			record := map[string]any{
				"attributes": declaration.Attributes, "id": declaration.ID, "kind": declaration.Kind,
				"name": declaration.Name, "owner": declaration.Path, "path": declaration.Path,
				"qualified_name": declaration.QualifiedName, "record": "node", "span": declaration.Span.record(),
			}
			facts.addNode(file.Path, record)
			facts.addEdge(file.Path, map[string]any{
				"owner": file.Path, "record": "edge", "relation": "defines", "source": declaration.ContainerID,
				"span": declaration.Span.record(), "target": declaration.ID,
			})
		}
		for _, include := range file.Includes {
			facts.addNode(file.Path, map[string]any{
				"attributes": map[string]any{"expression": include.Expression, "language": file.Language, "system": include.System, "target": include.Target},
				"id":         include.ID, "kind": "import", "name": include.Target, "owner": file.Path, "path": file.Path,
				"qualified_name": file.Path + "::include::" + include.Target, "record": "node", "span": include.Span.record(),
			})
			facts.addEdge(file.Path, map[string]any{
				"owner": file.Path, "record": "edge", "relation": "defines", "source": file.ModuleID,
				"span": include.Span.record(), "target": include.ID,
			})
			resolveInclude(facts, files, include)
		}
		if file.ParseError {
			facts.addUnresolved(file.Path, map[string]any{
				"attributes": map[string]any{"parser": "tree-sitter"}, "expression": file.Path,
				"owner": file.Path, "reason": "unsupported-form", "record": "unresolved",
				"relation": "references", "source": file.ModuleID,
			})
		}
	}

	for _, file := range model.Files {
		for _, observation := range file.Inheritance {
			resolveInheritance(facts, index, observation)
		}
		for _, observation := range file.Calls {
			resolveCall(facts, index, observation)
		}
		for _, observation := range file.Accesses {
			resolveAccess(facts, index, observation)
		}
	}
	return facts
}

func moduleName(path string) string {
	base := filepath.Base(filepath.FromSlash(path))
	return base[:len(base)-len(filepath.Ext(base))]
}
