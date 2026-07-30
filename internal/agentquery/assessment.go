package agentquery

import (
	"path/filepath"
	"strings"
)

var narrowEvidenceDimensions = []string{
	"owner",
	"control_flow",
	"public_boundary",
	"tests",
}

func assessEvidence(request Request, response *Response) {
	if response == nil {
		return
	}
	assessment := &EvidenceAssessment{
		Stage: request.Mode,
		Scope: request.Breadth,
	}
	observed := make(map[string]bool)
	paths := make(map[string]bool)

	observeResult := func(result Result) {
		assessment.EvidenceItems++
		observeNodeDimensions(result.Node, result.Provider, result.Excerpt, observed, paths)
	}
	for _, lane := range [][]Result{response.ExactMatches, response.SourceMatches, response.SymbolMatches} {
		for _, result := range lane {
			observeResult(result)
		}
	}
	for _, inspection := range response.Inspections {
		assessment.EvidenceItems++
		if inspection.Node != nil {
			observeNodeDimensions(*inspection.Node, inspection.Handle.Provider, inspection.Source, observed, paths)
		}
		observePathDimensions(inspection.ContainingSpan.Path, observed, paths)
		if strings.TrimSpace(inspection.Source) != "" {
			observed["control_flow"] = true
		}
	}
	assessment.EvidenceItems += len(response.Paths) + len(response.Dependents) + len(response.RelationshipMatches)
	assessment.DistinctPaths = len(paths)
	assessment.ObservedDimensions, assessment.MissingDimensions = dimensionLists(observed)

	switch request.Mode {
	case "orient":
		if assessment.EvidenceItems == 0 {
			assessment.Status = "insufficient"
			assessment.NextAction = "use-direct-search-or-specific-query"
			assessment.Reason = "orientation returned no concrete repository anchors"
		} else {
			assessment.Status = "discovery-ready"
			assessment.NextAction = "search-specific-behavior"
			assessment.Reason = "repository anchors are available; continue with a concrete query"
		}
	case "search":
		assessSearch(request, assessment, observed)
	case "inspect":
		assessInspection(assessment, observed)
	case "trace", "impact":
		if assessment.EvidenceItems == 0 {
			assessment.Status = "insufficient"
			assessment.NextAction = "inspect-anchor-or-refine-query"
			assessment.Reason = "structural expansion returned no evidence"
		} else if len(response.Unresolved) > 0 {
			assessment.Status = "partial"
			assessment.NextAction = "resolve-named-structural-gap"
			assessment.Reason = "bounded structural evidence is available, but unresolved relationships remain"
		} else {
			assessment.Status = "ready-to-synthesize"
			assessment.NextAction = "synthesize"
			assessment.Reason = "bounded structural evidence is available without unresolved relationships"
		}
	default:
		return
	}
	response.Assessment = assessment
}

func assessSearch(request Request, assessment *EvidenceAssessment, observed map[string]bool) {
	if assessment.EvidenceItems == 0 {
		assessment.Status = "insufficient"
		assessment.NextAction = "use-direct-search-or-balanced-search"
		assessment.Reason = "search returned no concrete evidence"
		return
	}
	if request.Breadth == "narrow" {
		if observed["owner"] && observed["control_flow"] {
			assessment.Status = "ready-for-targeted-inspection"
			assessment.NextAction = "inspect-selected-handles"
			assessment.Reason = "narrow discovery found owner and control-flow anchors; inspect only missing task dimensions"
			return
		}
		assessment.Status = "partial"
		assessment.NextAction = "inspect-best-anchor-or-switch-balanced"
		assessment.Reason = "narrow discovery found evidence but did not establish both owner and control-flow anchors"
		return
	}
	assessment.Status = "discovery-ready"
	assessment.NextAction = "inspect-or-expand-selected-handle"
	assessment.Reason = "balanced discovery returned ranked evidence; select a handle before expanding"
}

func assessInspection(assessment *EvidenceAssessment, observed map[string]bool) {
	if assessment.EvidenceItems == 0 {
		assessment.Status = "insufficient"
		assessment.NextAction = "inspect-valid-handle"
		assessment.Reason = "inspection returned no exact evidence"
		return
	}
	if len(assessment.MissingDimensions) == 0 {
		assessment.Status = "ready-to-synthesize"
		assessment.NextAction = "synthesize"
		assessment.Reason = "inspection covers owner, control flow, public boundary, and tests"
		return
	}
	if observed["owner"] && observed["control_flow"] {
		assessment.Status = "partial"
		assessment.NextAction = "inspect-missing-dimensions"
		assessment.Reason = "core implementation is grounded; inspect only the listed missing dimensions if the task requires them"
		return
	}
	assessment.Status = "partial"
	assessment.NextAction = "inspect-owning-implementation"
	assessment.Reason = "exact evidence is available but the owning implementation is not yet grounded"
}

func observeNodeDimensions(node Node, provider, excerpt string, observed map[string]bool, paths map[string]bool) {
	kind := strings.ToLower(strings.TrimSpace(node.Kind))
	label := strings.TrimSpace(node.QualifiedName)
	if label == "" {
		label = strings.TrimSpace(node.Name)
	}
	if label != "" && kind != "file" && kind != "document" {
		observed["owner"] = true
	}
	switch kind {
	case "function", "method", "constructor", "handler", "type", "struct", "class", "interface", "enum":
		observed["owner"] = true
	}
	if provider == "exact" || provider == "source" || provider == "lexical" {
		if strings.TrimSpace(node.Path) != "" {
			observed["owner"] = true
		}
		observed["control_flow"] = true
	} else if strings.TrimSpace(excerpt) != "" {
		observed["control_flow"] = true
	}
	switch kind {
	case "function", "method", "constructor", "handler":
		observed["control_flow"] = true
	case "interface", "http-endpoint", "route", "command", "config", "schema", "contract":
		observed["public_boundary"] = true
	}
	observePathDimensions(node.Path, observed, paths)
}

func observePathDimensions(path string, observed map[string]bool, paths map[string]bool) {
	path = normalizePath(path)
	if path == "" {
		return
	}
	paths[strings.ToLower(path)] = true
	kind, _ := classifyPath(path)
	if kind == "test" {
		observed["tests"] = true
	}
	lower := strings.ToLower(path)
	ext := strings.ToLower(filepath.Ext(lower))
	if kind == "contract" || strings.Contains(lower, "/include/") || strings.HasPrefix(lower, "include/") ||
		strings.Contains(lower, "/public/") || strings.Contains(lower, "api") ||
		ext == ".h" || ext == ".hpp" || ext == ".proto" {
		observed["public_boundary"] = true
	}
}

func dimensionLists(observed map[string]bool) ([]string, []string) {
	found := make([]string, 0, len(narrowEvidenceDimensions))
	missing := make([]string, 0, len(narrowEvidenceDimensions))
	for _, dimension := range narrowEvidenceDimensions {
		if observed[dimension] {
			found = append(found, dimension)
		} else {
			missing = append(missing, dimension)
		}
	}
	return found, missing
}
