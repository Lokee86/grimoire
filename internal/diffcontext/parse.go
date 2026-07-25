package diffcontext

import (
	"bufio"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var hunkHeader = regexp.MustCompile(`^@@ -([0-9]+)(,([0-9]+))? \+([0-9]+)(,([0-9]+))? @@ ?(.*)$`)

type diffFile struct {
	oldPath string
	newPath string
	deleted bool
	newFile bool
	touched bool
	changes []Change
}

// Parse converts git diff --unified=0 output into current-tree source changes.
// It deliberately accepts the text as data and performs no repository or
// command access. Malformed and non-hunk lines are ignored.
func Parse(diff string) []Change {
	var current *diffFile
	var parsed []Change
	flush := func() {
		if current == nil {
			return
		}
		parsed = append(parsed, current.changes...)
		if len(current.changes) == 0 && current.touched {
			path := current.newPath
			if path == "" || path == "/dev/null" {
				path = current.oldPath
			}
			if path != "" && path != "/dev/null" {
				oldPath := current.oldPath
				if oldPath == path || oldPath == "/dev/null" {
					oldPath = ""
				}
				summary := "changed file"
				if current.deleted {
					summary = "deleted file"
				} else if current.newFile {
					summary = "new file"
				} else if oldPath != "" {
					summary = "renamed file"
				}
				parsed = append(parsed, Change{
					Path: path, OldPath: oldPath, StartLine: 1, EndLine: 1,
					Deleted: current.deleted, Summary: summary,
				})
			}
		}
		current = nil
	}

	scanner := bufio.NewScanner(strings.NewReader(diff))
	for scanner.Scan() {
		line := strings.TrimSuffix(scanner.Text(), "\r")
		switch {
		case strings.HasPrefix(line, "diff --git "):
			flush()
			oldPath, newPath := parseDiffHeaderPaths(strings.TrimPrefix(line, "diff --git "))
			current = &diffFile{oldPath: oldPath, newPath: newPath, touched: true}
		case strings.HasPrefix(line, "new file mode "):
			ensureDiffFile(&current)
			current.newFile = true
		case strings.HasPrefix(line, "deleted file mode "):
			ensureDiffFile(&current)
			current.deleted = true
		case strings.HasPrefix(line, "rename from "):
			ensureDiffFile(&current)
			current.oldPath = parsePath(line[len("rename from "):], "")
		case strings.HasPrefix(line, "rename to "):
			ensureDiffFile(&current)
			current.newPath = parsePath(line[len("rename to "):], "")
		case strings.HasPrefix(line, "--- "):
			ensureDiffFile(&current)
			current.oldPath = parsePath(line[len("--- "):], "a/")
		case strings.HasPrefix(line, "+++ "):
			ensureDiffFile(&current)
			current.newPath = parsePath(line[len("+++ "):], "b/")
			current.deleted = current.newPath == "/dev/null"
			current.newFile = current.oldPath == "/dev/null"
		case strings.HasPrefix(line, "@@ "):
			ensureDiffFile(&current)
			change, ok := parseHunk(line, *current)
			if ok {
				current.changes = append(current.changes, change)
			}
		}
	}
	flush()

	for index := range parsed {
		parsed[index] = normalizeChange(parsed[index])
	}
	return deduplicateChanges(parsed)
}

// ParseWithError is an error-shaped companion for callers that prefer a parser
// API with an error return. The unified-diff parser is intentionally tolerant,
// so it currently returns no parse errors.
func ParseWithError(diff string) ([]Change, error) {
	return Parse(diff), nil
}

func ensureDiffFile(file **diffFile) {
	if *file == nil {
		*file = &diffFile{}
	}
}

func parseHunk(line string, file diffFile) (Change, bool) {
	matches := hunkHeader.FindStringSubmatch(line)
	if matches == nil {
		return Change{}, false
	}
	newStart := parseNumber(matches[4], 1)
	newCount := parseNumber(matches[6], 1)
	start := newStart
	end := newStart + newCount - 1
	if newCount == 0 {
		// Deletions have no current-file lines. Keep an anchor so the
		// deletion remains addressable and can still produce evidence.
		if start <= 0 {
			start = 1
		}
		end = start
	}
	path := file.newPath
	if path == "" || path == "/dev/null" {
		path = file.oldPath
	}
	if path == "" || path == "/dev/null" {
		return Change{}, false
	}
	oldPath := file.oldPath
	if oldPath == path || oldPath == "/dev/null" {
		oldPath = ""
	}
	return Change{
		Path: path, OldPath: oldPath,
		StartLine: start, EndLine: end,
		Deleted: file.newPath == "/dev/null",
		Summary: strings.TrimSpace(matches[7]),
	}, true
}

func parseNumber(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	number, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return number
}

func parseDiffHeaderPaths(value string) (string, string) {
	left, remainder := nextDiffPath(value)
	right, _ := nextDiffPath(remainder)
	return parsePath(left, "a/"), parsePath(right, "b/")
}

func nextDiffPath(value string) (string, string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ""
	}
	if value[0] == '"' {
		for index := 1; index < len(value); index++ {
			if value[index] != '"' || value[index-1] == '\\' {
				continue
			}
			return value[:index+1], value[index+1:]
		}
		return value, ""
	}
	if space := strings.IndexByte(value, ' '); space >= 0 {
		return value[:space], value[space+1:]
	}
	return value, ""
}

func parsePath(value, stripPrefix string) string {
	value = strings.TrimSpace(value)
	if tab := strings.IndexByte(value, '\t'); tab >= 0 {
		value = value[:tab]
	}
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		if unquoted, err := strconv.Unquote(value); err == nil {
			value = unquoted
		}
	}
	if value == "/dev/null" {
		return value
	}
	if stripPrefix != "" && strings.HasPrefix(value, stripPrefix) {
		value = strings.TrimPrefix(value, stripPrefix)
	}
	return normalizePath(value)
}

func deduplicateChanges(changes []Change) []Change {
	seen := make(map[string]struct{}, len(changes))
	result := make([]Change, 0, len(changes))
	for _, change := range changes {
		key := change.Path + "\x00" + change.OldPath + "\x00" +
			strconv.Itoa(change.StartLine) + "\x00" + strconv.Itoa(change.EndLine) + "\x00" +
			strconv.FormatBool(change.Deleted) + "\x00" + change.Summary
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, change)
	}
	sort.Slice(result, func(left, right int) bool {
		a, b := result[left], result[right]
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		if a.StartLine != b.StartLine {
			return a.StartLine < b.StartLine
		}
		if a.EndLine != b.EndLine {
			return a.EndLine < b.EndLine
		}
		if a.OldPath != b.OldPath {
			return a.OldPath < b.OldPath
		}
		if a.Deleted != b.Deleted {
			return !a.Deleted
		}
		return a.Summary < b.Summary
	})
	return result
}
