package arcanagraph

type arcanaSpan struct {
	Path        string `json:"path"`
	StartLine   int    `json:"start_line"`
	StartColumn int    `json:"start_column"`
	EndLine     int    `json:"end_line"`
	EndColumn   int    `json:"end_column"`
}

type arcanaNode struct {
	NodeID   uint32      `json:"node_id"`
	Identity string      `json:"identity"`
	Kind     string      `json:"kind"`
	Path     string      `json:"path"`
	Name     string      `json:"name"`
	Span     *arcanaSpan `json:"span"`
}

type nodeListResult struct {
	Nodes []arcanaNode `json:"nodes"`
}

type relatedNode struct {
	Relation string     `json:"relation"`
	Node     arcanaNode `json:"node"`
}

type roleResult struct {
	Node    arcanaNode    `json:"node"`
	Summary string        `json:"summary"`
	Callers []relatedNode `json:"callers"`
	Callees []relatedNode `json:"callees"`
}

type impactResult struct {
	NodeID     uint32 `json:"node_id"`
	Truncated  bool   `json:"truncated"`
	Dependents []struct {
		Depth int        `json:"depth"`
		Node  arcanaNode `json:"node"`
	} `json:"dependents"`
}

type unresolvedResult struct {
	Truncated  bool `json:"truncated"`
	Unresolved []struct {
		Relation           string      `json:"relation"`
		Expression         string      `json:"expression"`
		CandidateNamespace string      `json:"candidate_namespace"`
		CandidateName      string      `json:"candidate_name"`
		Reason             string      `json:"reason"`
		Span               *arcanaSpan `json:"span"`
	} `json:"unresolved"`
}

type chainResult struct {
	Found bool        `json:"found"`
	Chain *arcanaPath `json:"chain"`
}

type arcanaPath struct {
	Depth     int          `json:"depth"`
	Nodes     []arcanaNode `json:"nodes"`
	Relations []string     `json:"relations"`
}
