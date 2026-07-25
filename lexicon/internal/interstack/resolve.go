package interstack

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type sourceFile struct {
	Path  string
	Ext   string
	Lines []string
}

type resolver struct {
	index         sourceIndex
	result        Result
	nodes         map[string]Node
	edges         map[string]factEdge
	unresolved    map[string]factUnresolved
	constants     map[string]map[string]struct{}
	http          []httpContract
	httpSources   []httpProducer
	httpProviders map[string][]httpProducer
}

func Resolve(sourceRoot string, libraries []Library) (Result, error) {
	allowedLanguages := make(map[string]struct{}, len(libraries))
	for _, library := range libraries {
		allowedLanguages[library.Language] = struct{}{}
	}
	files, err := collectSourceFiles(sourceRoot, allowedLanguages)
	if err != nil {
		return Result{}, err
	}
	repository := ""
	for _, library := range libraries {
		if library.Repository != "" {
			repository = library.Repository
			break
		}
	}
	if repository == "" {
		repository = filepath.Base(filepath.Clean(sourceRoot))
	}
	value := resolver{
		index:         newSourceIndex(libraries),
		result:        Result{Repository: repository},
		nodes:         make(map[string]Node),
		edges:         make(map[string]factEdge),
		unresolved:    make(map[string]factUnresolved),
		constants:     make(map[string]map[string]struct{}),
		httpProviders: make(map[string][]httpProducer),
	}
	for _, file := range files {
		value.collectConstants(file)
	}
	for _, file := range files {
		value.detectHTTPConsumers(file)
		value.detectMessageConsumers(file)
	}
	for _, file := range files {
		value.collectHTTPPathProviders(file)
	}
	for _, file := range files {
		value.detectHTTPProducers(file)
		value.detectMessageProducers(file)
		value.detectConfigReads(file)
	}
	value.resolveHTTPProducers()
	value.finish()
	return value.result, nil
}

func (r *resolver) finish() {
	r.result.Nodes = make([]Node, 0, len(r.nodes))
	for _, node := range r.nodes {
		r.result.Nodes = append(r.result.Nodes, node)
	}
	r.result.Edges = make([]factEdge, 0, len(r.edges))
	for _, edge := range r.edges {
		r.result.Edges = append(r.result.Edges, edge)
	}
	r.result.Unresolved = make([]factUnresolved, 0, len(r.unresolved))
	for _, item := range r.unresolved {
		r.result.Unresolved = append(r.result.Unresolved, item)
	}
}

func (r *resolver) addNode(node Node) {
	if node.ID == "" {
		return
	}
	r.nodes[node.ID] = node
}

func (r *resolver) addEdge(edge factEdge) {
	if edge.Source == "" || edge.Target == "" || edge.Relation == "" {
		return
	}
	key := edgeSortKey(edge)
	r.edges[key] = edge
}

func (r *resolver) addUnresolved(item factUnresolved) {
	if item.Source == "" || item.Relation == "" || item.Expression == "" || item.Reason == "" {
		return
	}
	key := unresolvedSortKey(item)
	r.unresolved[key] = item
}

func lineSpan(path string, line int, text string) *Span {
	endColumn := len([]rune(text)) + 1
	if endColumn < 2 {
		endColumn = 2
	}
	return &Span{
		Path: path, StartLine: uint32(line), StartColumn: 1,
		EndLine: uint32(line), EndColumn: uint32(endColumn),
	}
}

func collectSourceFiles(root string, allowedLanguages map[string]struct{}) ([]sourceFile, error) {
	result := make([]sourceFile, 0)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if os.IsNotExist(walkErr) {
			return nil
		}
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != root && ignoredDirectory(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if ignoredSourcePath(relative) {
			return nil
		}
		extension := strings.ToLower(filepath.Ext(path))
		language := sourceLanguage(extension)
		if language == "" {
			return nil
		}
		if _, enabled := allowedLanguages[language]; !enabled {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Size() > 2*1024*1024 {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.IndexByte(string(data), 0) >= 0 {
			return nil
		}
		text := strings.ReplaceAll(string(data), "\r\n", "\n")
		result = append(result, sourceFile{Path: relative, Ext: extension, Lines: strings.Split(text, "\n")})
		return nil
	})
	sort.Slice(result, func(left, right int) bool { return result[left].Path < result[right].Path })
	return result, err
}

func ignoredDirectory(name string) bool {
	switch name {
	case ".git", ".lexicon", ".worktrees", ".workingtrees", ".ddocs", ".obsidian",
		"node_modules", "vendor", "target", "build", "dist", "coverage", ".bundle":
		return true
	default:
		return false
	}
}

func ignoredSourcePath(path string) bool {
	for _, part := range strings.Split(filepath.ToSlash(path), "/") {
		switch strings.ToLower(part) {
		case "test", "tests", "spec", "fixtures":
			return true
		}
	}
	return false
}

func sourceLanguage(extension string) string {
	switch extension {
	case ".go":
		return "go"
	case ".rb":
		return "ruby"
	case ".gd":
		return "gdscript"
	case ".ts", ".tsx", ".js", ".jsx":
		return "typescript"
	case ".py":
		return "python"
	default:
		return ""
	}
}

func uniqueString(values map[string]struct{}) (string, bool) {
	if len(values) != 1 {
		return "", false
	}
	for value := range values {
		return value, true
	}
	panic(fmt.Sprintf("unreachable map size %d", len(values)))
}
