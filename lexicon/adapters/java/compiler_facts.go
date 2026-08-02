package main

type compilerEvidenceKind uint8

const (
	compilerEvidenceEdge compilerEvidenceKind = iota
	compilerEvidenceUnresolved
)

type compilerEvidenceRef struct {
	kind compilerEvidenceKind
	key  string
}

type compilerEvidenceSite struct {
	source   string
	relation string
	owner    string
	line     int
	column   int
}

type compilerEvidenceLine struct {
	source   string
	relation string
	owner    string
	line     int
}

type compilerEvidenceIndex struct {
	exact map[compilerEvidenceSite][]compilerEvidenceRef
	lines map[compilerEvidenceLine][]compilerEvidenceRef
}

func buildCompilerEvidenceIndex(facts *factSet) compilerEvidenceIndex {
	index := compilerEvidenceIndex{
		exact: make(map[compilerEvidenceSite][]compilerEvidenceRef),
		lines: make(map[compilerEvidenceLine][]compilerEvidenceRef),
	}
	add := func(kind compilerEvidenceKind, key string, record map[string]any) {
		source, _ := record["source"].(string)
		relation, _ := record["relation"].(string)
		owner, _ := record["owner"].(string)
		value, ok := record["span"].(*span)
		if source == "" || relation == "" || !ok || value == nil {
			return
		}
		reference := compilerEvidenceRef{kind: kind, key: key}
		exact := compilerEvidenceSite{
			source: source, relation: relation, owner: owner,
			line: value.StartLine, column: value.StartColumn,
		}
		line := compilerEvidenceLine{
			source: source, relation: relation, owner: owner, line: value.StartLine,
		}
		index.exact[exact] = append(index.exact[exact], reference)
		index.lines[line] = append(index.lines[line], reference)
	}
	for key, record := range facts.edges {
		add(compilerEvidenceEdge, key, record)
	}
	for key, record := range facts.unresolved {
		add(compilerEvidenceUnresolved, key, record)
	}
	return index
}

func (state *analysisState) applyCompilerFact(fact compilerFact, evidence *compilerEvidenceIndex) bool {
	if fact.Record == "failure" {
		state.facts.addUnresolved(
			state.repositoryID,
			"defines",
			fact.Path,
			"compiler-analysis-failed",
			fact.Path,
			nil,
			map[string]any{"engine": "javac", "failure_type": fact.Reason},
		)
		return true
	}
	if fact.SourceKind == "" || fact.Relation == "" {
		return false
	}
	source := nodeID(fact.SourceKind, fact.SourceIdentity)
	if state.facts.nodes[source] == nil {
		return false
	}
	state.removeCompetingRuntimeEvidence(source, fact, evidence)
	if fact.Record == "suppression" {
		return true
	}
	if fact.Record != "edge" || fact.TargetKind == "" {
		return false
	}
	target := nodeID(fact.TargetKind, fact.TargetIdentity)
	if state.facts.nodes[target] == nil {
		return false
	}
	state.facts.addEdge(source, target, fact.Relation, fact.Path, &span{
		Path: fact.Path, StartLine: fact.StartLine, StartColumn: fact.StartColumn,
		EndLine: fact.EndLine, EndColumn: fact.EndColumn,
	}, map[string]any{"resolution": fact.Engine})
	return true
}

func (state *analysisState) removeCompetingRuntimeEvidence(
	source string,
	fact compilerFact,
	evidence *compilerEvidenceIndex,
) {
	if evidence == nil {
		return
	}
	relations := []string{fact.Relation}
	if fact.Relation == "calls" {
		relations = append(relations, "possible-calls")
	}
	for _, relation := range relations {
		if fact.Record == "suppression" {
			for line := fact.StartLine - 1; line <= fact.StartLine+1; line++ {
				state.deleteCompilerEvidence(evidence.lines[compilerEvidenceLine{
					source: source, relation: relation, owner: fact.Path, line: line,
				}])
			}
			continue
		}
		state.deleteCompilerEvidence(evidence.exact[compilerEvidenceSite{
			source: source, relation: relation, owner: fact.Path,
			line: fact.StartLine, column: fact.StartColumn,
		}])
	}
}

func (state *analysisState) deleteCompilerEvidence(references []compilerEvidenceRef) {
	for _, reference := range references {
		switch reference.kind {
		case compilerEvidenceEdge:
			delete(state.facts.edges, reference.key)
		case compilerEvidenceUnresolved:
			delete(state.facts.unresolved, reference.key)
		}
	}
}
