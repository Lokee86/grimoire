package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Summary struct {
	Nodes, Edges, DirectCalls, PossibleCallTargets, CallExpressions, UnresolvedCalls int
	Directories, Files, Packages, Imports                                            int
	BuiltinCalls, ConversionCalls, ExternalCalls, DynamicCalls, InterfaceCalls       int
	Closures, Captures, SemanticErrors                                               int
}

type packageInfo struct {
	key       NodeKey
	importKey string
	name      string
}

type callable struct {
	packageKey string
	namespace  string
	source     NodeKey
	body       *ast.BlockStmt
	path       string
}

type scanner struct {
	root, module  string
	set           *token.FileSet
	facts         RepositoryFacts
	nodes         map[NodeKey]NodeFact
	edges         map[string]EdgeFact
	packages      map[string]packageInfo
	callables     []callable
	targets       map[string]map[string][]NodeKey
	semanticCalls map[string]semanticCall
	semanticIDs   map[string][]NodeKey
	closureKeys   map[string]NodeKey
	callsiteKeys  map[string]string
	fileImports   map[string]map[string]string
	summary       Summary
}

func scanRepository(root string) (RepositoryFacts, Summary, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return RepositoryFacts{}, Summary{}, err
	}
	module, err := readModule(root)
	if err != nil {
		return RepositoryFacts{}, Summary{}, err
	}
	s := &scanner{
		root: root, module: module, set: token.NewFileSet(), facts: RepositoryFacts{Repository: module},
		nodes: make(map[NodeKey]NodeFact), edges: make(map[string]EdgeFact),
		packages: make(map[string]packageInfo), targets: make(map[string]map[string][]NodeKey),
		semanticCalls: make(map[string]semanticCall),
		semanticIDs:   make(map[string][]NodeKey), closureKeys: make(map[string]NodeKey),
		callsiteKeys: make(map[string]string), fileImports: make(map[string]map[string]string),
	}
	files, dirs, err := discover(root)
	if err != nil {
		return RepositoryFacts{}, Summary{}, err
	}
	if err := s.addRootAndDirectories(dirs); err != nil {
		return RepositoryFacts{}, Summary{}, err
	}
	for _, file := range files {
		if err := s.addFile(file); err != nil {
			return RepositoryFacts{}, Summary{}, err
		}
	}
	for _, file := range files {
		if filepath.Ext(file) == ".go" {
			if err := s.parseGoFile(file); err != nil {
				return RepositoryFacts{}, Summary{}, err
			}
		}
	}
	if err := s.loadSemanticCalls(); err != nil {
		return RepositoryFacts{}, Summary{}, err
	}
	s.addCallEdges()
	for _, node := range s.nodes {
		s.facts.Nodes = append(s.facts.Nodes, node)
	}
	for _, edge := range s.edges {
		s.facts.Edges = append(s.facts.Edges, edge)
	}
	s.summary.Nodes = len(s.facts.Nodes)
	s.summary.Edges = len(s.facts.Edges)
	return s.facts, s.summary, nil
}

func discover(root string) ([]string, []string, error) {
	var files, dirs []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != root && ignoredDir(entry.Name()) {
				return filepath.SkipDir
			}
			if path != root {
				dirs = append(dirs, path)
			}
			return nil
		}
		name := entry.Name()
		if name == "go.mod" || filepath.Ext(name) == ".go" {
			files = append(files, path)
		}
		return nil
	})
	sort.Strings(files)
	sort.Strings(dirs)
	return files, dirs, err
}

func ignoredDir(name string) bool {
	switch name {
	case ".git", ".worktrees", ".workingtrees", ".ddocs", ".lexicon", ".arcana", ".grimoire", ".pitlord", ".cantrip", ".homunculus", ".incubus", ".ritual", ".warlock", "vendor":
		return true
	default:
		return false
	}
}

func readModule(root string) (string, error) {
	file, err := os.Open(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("open go.mod: %w", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(strings.TrimSpace(scanner.Text()))
		if len(fields) >= 2 && fields[0] == "module" {
			return fields[1], nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read go.mod: %w", err)
	}
	return "", fmt.Errorf("go.mod has no module directive")
}

func (s *scanner) addRootAndDirectories(dirs []string) error {
	rootPath := ".lexicon-repository"
	rootKey := hashIdentity("repository:" + s.module)
	s.addNode(NodeFact{Key: rootKey, Kind: KindRepository, Path: rootPath, Name: s.module})
	for _, absolute := range dirs {
		rel, err := s.relative(absolute)
		if err != nil {
			return err
		}
		key := hashIdentity("directory:" + rel)
		s.addNode(NodeFact{Key: key, Kind: KindDirectory, Path: rel, Name: filepath.Base(rel)})
		s.addEdge(s.parentKey(rel), key, RelContains, nil)
	}
	s.summary.Directories = len(dirs)
	return nil
}

func (s *scanner) addFile(absolute string) error {
	rel, err := s.relative(absolute)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(absolute)
	if err != nil {
		return fmt.Errorf("read %s: %w", rel, err)
	}
	key := hashIdentity("file:" + rel)
	id := contentID(data)
	s.addNode(NodeFact{Key: key, Kind: KindFile, Path: rel, Name: filepath.Base(rel), ContentID: &id})
	s.addEdge(s.parentKey(rel), key, RelContains, nil)
	s.summary.Files++
	return nil
}

func (s *scanner) relative(absolute string) (string, error) {
	rel, err := filepath.Rel(s.root, absolute)
	if err != nil {
		return "", err
	}
	return normalizePath(rel)
}

func (s *scanner) parentKey(rel string) NodeKey {
	dir := filepath.ToSlash(filepath.Dir(rel))
	if dir == "." {
		return hashIdentity("repository:" + s.module)
	}
	return hashIdentity("directory:" + dir)
}

func (s *scanner) addNode(node NodeFact) {
	if _, exists := s.nodes[node.Key]; !exists {
		s.nodes[node.Key] = node
	}
}

func (s *scanner) addEdge(source, target NodeKey, relation RelationKind, span *SourceSpan) {
	key := fmt.Sprintf("%s/%s/%s", source, target, relation)
	if _, exists := s.edges[key]; !exists {
		s.edges[key] = EdgeFact{Source: source, Target: target, Relation: relation, Span: span}
	}
}

func (s *scanner) parseGoFile(absolute string) error {
	rel, err := s.relative(absolute)
	if err != nil {
		return err
	}
	file, err := parser.ParseFile(s.set, absolute, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse %s: %w", rel, err)
	}
	dir := filepath.ToSlash(filepath.Dir(rel))
	if dir == "." {
		dir = ""
	}
	importPath := s.module
	if dir != "" {
		importPath += "/" + dir
	}
	packageKey := dir + "\x00" + file.Name.Name
	pkg, exists := s.packages[packageKey]
	if !exists {
		pkgKey := hashIdentity("package:" + importPath + ":" + file.Name.Name)
		pkg = packageInfo{key: pkgKey, importKey: importPath, name: file.Name.Name}
		s.packages[packageKey] = pkg
		pkgPath := dir
		if pkgPath == "" {
			pkgPath = ".lexicon-repository"
		}
		s.addNode(NodeFact{Key: pkgKey, Kind: KindPackage, Path: pkgPath, Name: file.Name.Name, Span: s.span(file.Name.Pos(), file.Name.End(), rel)})
		s.addEdge(s.parentKey(rel), pkgKey, RelContains, nil)
		s.summary.Packages++
	}
	fileKey := hashIdentity("file:" + rel)
	s.addEdge(pkg.key, fileKey, RelContains, s.span(file.Name.Pos(), file.Name.End(), rel))
	for _, spec := range file.Imports {
		importName, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return fmt.Errorf("parse import in %s: %w", rel, err)
		}
		alias := filepath.Base(importName)
		if spec.Name != nil {
			alias = spec.Name.Name
		}
		if alias != "_" && alias != "." {
			if s.fileImports[rel] == nil {
				s.fileImports[rel] = make(map[string]string)
			}
			s.fileImports[rel][alias] = importName
		}
		internal := importName == s.module || strings.HasPrefix(importName, s.module+"/")
		class := "external"
		if internal {
			class = "internal"
		}
		logical := "@" + class + "/" + importName
		importKey := hashIdentity("import:" + class + ":" + importName)
		s.addNode(NodeFact{Key: importKey, Kind: KindImport, Path: logical, Name: importName, Span: s.span(spec.Pos(), spec.End(), rel)})
		s.addEdge(pkg.key, importKey, RelImports, s.span(spec.Pos(), spec.End(), rel))
		s.summary.Imports++
	}
	for _, declaration := range file.Decls {
		s.parseDeclaration(declaration, pkg, rel)
	}
	return nil
}
