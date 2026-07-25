package diffcontext

import "strings"

// Change identifies one changed source span in the current repository tree.
// Lines are one-based and inclusive. A zero-length new-side hunk is represented
// by a one-line anchor at its new-side start line.
type Change struct {
	Path      string
	OldPath   string
	StartLine int
	EndLine   int
	Deleted   bool
	Summary   string
}

func normalizePath(path string) string {
	path = strings.TrimSpace(path)
	path = strings.ReplaceAll(path, "\\", "/")
	path = strings.TrimPrefix(path, "./")
	return path
}

func normalizeChange(change Change) Change {
	change.Path = normalizePath(change.Path)
	change.OldPath = normalizePath(change.OldPath)
	if change.StartLine <= 0 {
		change.StartLine = 1
	}
	if change.EndLine < change.StartLine {
		change.EndLine = change.StartLine
	}
	return change
}
