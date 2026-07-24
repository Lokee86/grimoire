package main

import (
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var quotedIncludePattern = regexp.MustCompile(`(?m)^\s*#\s*include\s*"([^"]+)"`)

func inferHeaderLanguages(paths []string, contents map[string][]byte, compileLanguages map[string]string) map[string]string {
	pathSet := make(map[string]struct{}, len(paths))
	byBase := make(map[string][]string)
	for _, path := range paths {
		pathSet[path] = struct{}{}
		base := strings.ToLower(filepath.Base(filepath.FromSlash(path)))
		byBase[base] = append(byBase[base], path)
	}
	for base := range byBase {
		sort.Strings(byBase[base])
	}

	includes := make(map[string][]string, len(paths))
	for _, path := range paths {
		for _, match := range quotedIncludePattern.FindAllSubmatch(contents[path], -1) {
			if target := resolveSourceInclude(path, string(match[1]), pathSet, byBase); target != "" {
				includes[path] = append(includes[path], target)
			}
		}
		sort.Strings(includes[path])
	}

	type evidence struct {
		c   bool
		cpp bool
	}
	evidenceByPath := make(map[string]evidence)
	queue := make([]string, 0, len(paths))
	for _, path := range paths {
		if isHeaderPath(path) {
			continue
		}
		language := sourceFileLanguage(path, compileLanguages)
		evidenceByPath[path] = evidence{c: language == "c", cpp: language == "cpp"}
		queue = append(queue, path)
	}

	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]
		current := evidenceByPath[path]
		for _, target := range includes[path] {
			if !isHeaderPath(target) {
				continue
			}
			previous := evidenceByPath[target]
			next := evidence{c: previous.c || current.c, cpp: previous.cpp || current.cpp}
			if next == previous {
				continue
			}
			evidenceByPath[target] = next
			queue = append(queue, target)
		}
	}

	result := make(map[string]string)
	for path, evidence := range evidenceByPath {
		if !isHeaderPath(path) {
			continue
		}
		if evidence.c {
			result[path] = "c"
		} else if evidence.cpp {
			result[path] = "cpp"
		}
	}
	return result
}

func sourceFileLanguage(path string, compileLanguages map[string]string) string {
	path = filepath.ToSlash(path)
	if language := compileLanguages[path]; language != "" {
		return language
	}
	if filepath.Ext(path) == ".C" || strings.ToLower(filepath.Ext(path)) != ".c" {
		return "cpp"
	}
	return "c"
}

func resolveSourceInclude(sourcePath, target string, paths map[string]struct{}, byBase map[string][]string) string {
	target = filepath.ToSlash(filepath.Clean(filepath.FromSlash(target)))
	if _, ok := paths[target]; ok {
		return target
	}
	relative := filepath.ToSlash(filepath.Clean(filepath.Join(filepath.Dir(filepath.FromSlash(sourcePath)), filepath.FromSlash(target))))
	if !strings.HasPrefix(relative, "../") {
		if _, ok := paths[relative]; ok {
			return relative
		}
	}
	matches := byBase[strings.ToLower(filepath.Base(filepath.FromSlash(target)))]
	if len(matches) == 1 {
		return matches[0]
	}
	return ""
}
