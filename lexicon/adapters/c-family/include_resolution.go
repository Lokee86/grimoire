package main

import (
	"path/filepath"
	"strings"
)

func resolveInclude(facts *factSet, files fileIndex, observation includeObservation) {
	if target := localIncludeTarget(files, observation); target != nil {
		facts.addEdge(observation.Path, map[string]any{
			"owner": observation.Path, "record": "edge", "relation": "includes", "source": observation.ID,
			"span": observation.Span.record(), "target": target.FileID,
		})
		return
	}
	reason := "missing-target"
	if observation.System {
		reason = "external-target"
	}
	facts.addUnresolved(observation.Path, map[string]any{
		"expression": observation.Expression, "owner": observation.Path, "reason": reason,
		"record": "unresolved", "relation": "imports", "source": observation.ID, "span": observation.Span.record(),
	})
}

func localIncludeTarget(files fileIndex, observation includeObservation) *sourceFile {
	target := filepath.ToSlash(filepath.Clean(filepath.FromSlash(observation.Target)))
	if direct := files.byPath[target]; direct != nil {
		return direct
	}
	relative := filepath.ToSlash(filepath.Clean(filepath.Join(filepath.Dir(filepath.FromSlash(observation.Path)), filepath.FromSlash(target))))
	if !strings.HasPrefix(relative, "../") {
		if local := files.byPath[relative]; local != nil {
			return local
		}
	}
	matches := files.byBaseName[strings.ToLower(filepath.Base(filepath.FromSlash(target)))]
	if len(matches) == 1 {
		return matches[0]
	}
	return nil
}
