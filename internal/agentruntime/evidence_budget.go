package agentruntime

import (
	"strings"

	"github.com/Lokee86/grimoire/internal/agentquery"
	"github.com/Lokee86/grimoire/internal/knowledge"
)

const defaultDocumentPreviewCount = 2

func applyKnowledgePreviews(documents []knowledge.Result, explicitCount int, detail string) []knowledge.Result {
	if strings.EqualFold(strings.TrimSpace(detail), "full") {
		return documents
	}
	previewEnd := min(len(documents), explicitCount+defaultDocumentPreviewCount)
	for index := explicitCount; index < len(documents); index++ {
		if index >= previewEnd {
			documents[index].Text = ""
			documents[index].CodeLinks = nil
		}
	}
	return documents
}

func knowledgePreviewCount(documents []knowledge.Result) int {
	count := 0
	for _, document := range documents {
		if document.Text != "" {
			count++
		}
	}
	return count
}

// recordRuntimeEvidenceCoverage keeps documentation independent from code
// discovery lanes. Each provider keeps its own ranked limit; no cross-provider
// score comparison or round-robin quota is applied.
func recordRuntimeEvidenceCoverage(query *agentquery.Response, documents []knowledge.Result, available int) {
	if query == nil || (query.Mode != "search" && query.Mode != "orient") {
		return
	}
	if available < len(documents) {
		available = len(documents)
	}
	deferred := available - len(documents)
	query.Coverage = append(query.Coverage, agentquery.LaneCoverage{
		Lane:      "document_matches",
		Available: available,
		Returned:  len(documents),
		Previewed: knowledgePreviewCount(documents),
		Deferred:  deferred,
	})
	if deferred > 0 {
		query.Truncated = true
		query.TruncatedLanes = appendUniqueString(query.TruncatedLanes, "document_matches")
	}
}
