package interstack

import "strings"

var boundaryConfigKeys = []string{
	"LEXICON_STATE_DIR",
	"GRIMOIRE_LEXICON_COMMAND",
	"GRIMOIRE_ARCANA_COMMAND",
	"GRIMOIRE_HOME",
}

func (r *resolver) detectBoundaryConfig(file sourceFile) {
	for index, line := range file.Lines {
		owner, ok := r.index.ownerAt(file.Path, uint32(index+1))
		if !ok {
			continue
		}
		for _, key := range boundaryConfigKeys {
			if !strings.Contains(line, `"`+key+`"`) && !strings.Contains(line, `'`+key+`'`) {
				continue
			}
			node := r.addConfigKey(key)
			r.addEdge(factEdge{
				Source: owner.ID, Target: node.ID, Relation: "reads-config",
				Span: lineSpan(file.Path, index+1, line),
				Attributes: map[string]any{
					"confidence": 1.0,
					"evidence":   []string{strings.TrimSpace(line)},
					"transport":  "boundary-configuration",
				},
			})
		}
	}
}
