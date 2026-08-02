package agentruntime

import (
	"encoding/json"
	"errors"
	"sort"

	"github.com/Lokee86/grimoire/internal/investigation"
)

const (
	maxInvestigationDeltaBytes        = 22 * 1024
	maxInvestigationEvidenceTextBytes = 1200
	maxInvestigationReasonTextBytes   = 220
	maxInvestigationReasons           = 3
	maxInvestigationSupport           = 3
	investigationEvidencePerResult    = 3
)

func boundInvestigationResponse(ledger *investigation.Ledger, response investigation.Response, limit int) (investigation.Response, bool, error) {
	if ledger == nil {
		return investigation.Response{}, false, errors.New("investigation ledger is required")
	}
	candidate, hitBudgeted := budgetInvestigationHits(response, limit)
	candidate = pruneInvestigationResponse(candidate)
	candidate, _ = compactInvestigationResponse(candidate)
	maxEvidence := investigationEvidenceLimit(limit)
	within, err := investigationResponseWithinBudget(ledger, candidate, maxEvidence)
	if err != nil {
		return investigation.Response{}, false, err
	}
	if within {
		return candidate, hitBudgeted, nil
	}

	truncated := hitBudgeted
	for {
		within, err := investigationResponseWithinBudget(ledger, candidate, maxEvidence)
		if err != nil {
			return investigation.Response{}, false, err
		}
		if within {
			return candidate, truncated, nil
		}
		if dropLastInvestigationAnnotation(&candidate) {
			truncated = true
			continue
		}
		if len(candidate.RetrievalHits) > 0 {
			candidate.RetrievalHits = append([]investigation.RetrievalHit(nil), candidate.RetrievalHits[:len(candidate.RetrievalHits)-1]...)
			candidate = pruneInvestigationResponse(candidate)
			truncated = true
			continue
		}
		return investigation.Response{}, false, errors.New("investigation snapshot exceeds serialized delta budget")
	}
}

func investigationResponseWithinBudget(ledger *investigation.Ledger, response investigation.Response, maxEvidence int) (bool, error) {
	delta, err := ledger.DeltaFor(response)
	if err != nil {
		return false, err
	}
	payload, err := json.Marshal(delta)
	if err != nil {
		return false, err
	}
	return investigationDeltaEvidenceCount(delta) <= maxEvidence && len(payload) <= maxInvestigationDeltaBytes, nil
}

func compactInvestigationResponse(response investigation.Response) (investigation.Response, bool) {
	compacted := false

	response.Nodes = append([]investigation.Node(nil), response.Nodes...)
	for index := range response.Nodes {
		response.Nodes[index].Label, compacted = compactInvestigationString(response.Nodes[index].Label, maxInvestigationReasonTextBytes, compacted)
		response.Nodes[index].Metadata, compacted = compactInvestigationMetadata(response.Nodes[index].Metadata, compacted)
	}
	response.SourceRanges = append([]investigation.SourceRange(nil), response.SourceRanges...)
	for index := range response.SourceRanges {
		response.SourceRanges[index].Text, compacted = compactInvestigationString(response.SourceRanges[index].Text, maxInvestigationEvidenceTextBytes, compacted)
		response.SourceRanges[index].Metadata, compacted = compactInvestigationMetadata(response.SourceRanges[index].Metadata, compacted)
	}
	response.GraphPaths = append([]investigation.GraphPath(nil), response.GraphPaths...)
	for index := range response.GraphPaths {
		response.GraphPaths[index].Label, compacted = compactInvestigationString(response.GraphPaths[index].Label, maxInvestigationReasonTextBytes, compacted)
		response.GraphPaths[index].Metadata, compacted = compactInvestigationMetadata(response.GraphPaths[index].Metadata, compacted)
	}
	response.Documents = append([]investigation.Document(nil), response.Documents...)
	for index := range response.Documents {
		response.Documents[index].Title, compacted = compactInvestigationString(response.Documents[index].Title, maxInvestigationReasonTextBytes, compacted)
		response.Documents[index].Content, compacted = compactInvestigationString(response.Documents[index].Content, maxInvestigationEvidenceTextBytes, compacted)
		response.Documents[index].Metadata, compacted = compactInvestigationMetadata(response.Documents[index].Metadata, compacted)
	}
	response.RetrievalHits = append([]investigation.RetrievalHit(nil), response.RetrievalHits...)
	for index := range response.RetrievalHits {
		hit := &response.RetrievalHits[index]
		hit.RelatedEvidence = append([]investigation.EvidenceRef(nil), hit.RelatedEvidence...)
		hit.Reasons, compacted = compactInvestigationStrings(hit.Reasons, maxInvestigationReasons, compacted)
		hit.Support, compacted = compactInvestigationStrings(hit.Support, maxInvestigationSupport, compacted)
		if hit.Seed != nil {
			seed := *hit.Seed
			seed.Reasons, compacted = compactInvestigationStrings(seed.Reasons, maxInvestigationReasons, compacted)
			hit.Seed = &seed
		}
	}
	response.UnresolvedQuestions = append([]investigation.UnresolvedQuestion(nil), response.UnresolvedQuestions...)
	for index := range response.UnresolvedQuestions {
		response.UnresolvedQuestions[index].Question, compacted = compactInvestigationString(response.UnresolvedQuestions[index].Question, maxInvestigationReasonTextBytes, compacted)
		response.UnresolvedQuestions[index].Context, compacted = compactInvestigationString(response.UnresolvedQuestions[index].Context, maxInvestigationEvidenceTextBytes, compacted)
	}
	response.RejectedBranches = append([]investigation.Branch(nil), response.RejectedBranches...)
	for index := range response.RejectedBranches {
		compactInvestigationBranch(&response.RejectedBranches[index], &compacted)
	}
	response.AcceptedBranches = append([]investigation.Branch(nil), response.AcceptedBranches...)
	for index := range response.AcceptedBranches {
		compactInvestigationBranch(&response.AcceptedBranches[index], &compacted)
	}
	return response, compacted
}

func compactInvestigationBranch(branch *investigation.Branch, compacted *bool) {
	branch.Description, *compacted = compactInvestigationString(branch.Description, maxInvestigationEvidenceTextBytes, *compacted)
	branch.Reason, *compacted = compactInvestigationString(branch.Reason, maxInvestigationEvidenceTextBytes, *compacted)
	branch.References, *compacted = compactInvestigationStrings(branch.References, maxInvestigationSupport, *compacted)
}

func compactInvestigationMetadata(metadata map[string]string, compacted bool) (map[string]string, bool) {
	if len(metadata) == 0 {
		return nil, compacted
	}
	copy := make(map[string]string, len(metadata))
	for key, value := range metadata {
		copy[key], compacted = compactInvestigationString(value, maxInvestigationReasonTextBytes, compacted)
	}
	return copy, compacted
}

func compactInvestigationStrings(values []string, maximum int, compacted bool) ([]string, bool) {
	if len(values) == 0 {
		return nil, compacted
	}
	if len(values) > maximum {
		values = values[:maximum]
		compacted = true
	}
	copy := make([]string, len(values))
	for index, value := range values {
		copy[index], compacted = compactInvestigationString(value, maxInvestigationReasonTextBytes, compacted)
	}
	return copy, compacted
}

func compactInvestigationString(value string, maximum int, compacted bool) (string, bool) {
	result := compactKnowledgeText(value, maximum)
	return result, compacted || result != value
}

func investigationHitLimit(limit int) int {
	if limit <= 0 {
		return 8
	}
	return min(16, max(8, limit*2))
}

func investigationEvidenceLimit(limit int) int {
	return max(24, investigationHitLimit(limit)*investigationEvidencePerResult)
}

func budgetInvestigationHits(response investigation.Response, limit int) (investigation.Response, bool) {
	maximum := investigationHitLimit(limit)
	if len(response.RetrievalHits) <= maximum {
		return response, false
	}

	laneOrder := make([]string, 0)
	laneIndices := make(map[string][]int)
	for index, hit := range response.RetrievalHits {
		lane := hit.Lane
		if lane == "" {
			lane = "other"
		}
		if _, exists := laneIndices[lane]; !exists {
			laneOrder = append(laneOrder, lane)
		}
		laneIndices[lane] = append(laneIndices[lane], index)
	}

	positions := make(map[string]int, len(laneOrder))
	selected := make([]int, 0, maximum)
	for len(selected) < maximum {
		progressed := false
		for _, lane := range laneOrder {
			position := positions[lane]
			indices := laneIndices[lane]
			if position >= len(indices) {
				continue
			}
			selected = append(selected, indices[position])
			positions[lane] = position + 1
			progressed = true
			if len(selected) == maximum {
				break
			}
		}
		if !progressed {
			break
		}
	}
	sort.Ints(selected)
	hits := make([]investigation.RetrievalHit, 0, len(selected))
	for _, index := range selected {
		hits = append(hits, response.RetrievalHits[index])
	}
	response.RetrievalHits = hits
	return response, true
}

func investigationDeltaEvidenceCount(delta investigation.Delta) int {
	return len(delta.NewNodes) + len(delta.PriorNodeHandles) +
		len(delta.NewSourceRanges) + len(delta.PriorSourceRanges) +
		len(delta.NewGraphPaths) + len(delta.PriorGraphPaths) +
		len(delta.NewDocuments) + len(delta.PriorDocuments)
}

func pruneInvestigationResponse(response investigation.Response) investigation.Response {
	keep := make(map[investigation.EvidenceRef]bool)
	for _, hit := range response.RetrievalHits {
		keep[hit.Evidence] = true
		for _, related := range hit.RelatedEvidence {
			keep[related] = true
		}
		if hit.Seed != nil {
			keep[hit.Seed.Evidence] = true
		}
	}

	nodeMap := make(map[int]int)
	rangeMap := make(map[int]int)
	pathMap := make(map[int]int)
	documentMap := make(map[int]int)

	nodes := make([]investigation.Node, 0, len(response.Nodes))
	for index, value := range response.Nodes {
		if !keep[investigation.EvidenceRef{Kind: "node", Index: index}] {
			continue
		}
		nodeMap[index] = len(nodes)
		nodes = append(nodes, value)
	}
	ranges := make([]investigation.SourceRange, 0, len(response.SourceRanges))
	for index, value := range response.SourceRanges {
		if !keep[investigation.EvidenceRef{Kind: "source", Index: index}] {
			continue
		}
		rangeMap[index] = len(ranges)
		ranges = append(ranges, value)
	}
	paths := make([]investigation.GraphPath, 0, len(response.GraphPaths))
	for index, value := range response.GraphPaths {
		if !keep[investigation.EvidenceRef{Kind: "path", Index: index}] {
			continue
		}
		pathMap[index] = len(paths)
		paths = append(paths, value)
	}
	documents := make([]investigation.Document, 0, len(response.Documents))
	for index, value := range response.Documents {
		if !keep[investigation.EvidenceRef{Kind: "document", Index: index}] {
			continue
		}
		documentMap[index] = len(documents)
		documents = append(documents, value)
	}

	response.Nodes = nodes
	response.SourceRanges = ranges
	response.GraphPaths = paths
	response.Documents = documents
	for index := range response.RetrievalHits {
		hit := &response.RetrievalHits[index]
		hit.Evidence = remapEvidenceRef(hit.Evidence, nodeMap, rangeMap, pathMap, documentMap)
		for relatedIndex := range hit.RelatedEvidence {
			hit.RelatedEvidence[relatedIndex] = remapEvidenceRef(hit.RelatedEvidence[relatedIndex], nodeMap, rangeMap, pathMap, documentMap)
		}
		if hit.Seed != nil {
			hit.Seed.Evidence = remapEvidenceRef(hit.Seed.Evidence, nodeMap, rangeMap, pathMap, documentMap)
		}
	}
	return response
}

func remapEvidenceRef(ref investigation.EvidenceRef, nodes, ranges, paths, documents map[int]int) investigation.EvidenceRef {
	switch ref.Kind {
	case "node":
		ref.Index = nodes[ref.Index]
	case "source":
		ref.Index = ranges[ref.Index]
	case "path":
		ref.Index = paths[ref.Index]
	case "document":
		ref.Index = documents[ref.Index]
	}
	return ref
}

func appendUniqueString(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func dropLastInvestigationAnnotation(response *investigation.Response) bool {
	switch {
	case len(response.AcceptedBranches) > 0:
		response.AcceptedBranches = response.AcceptedBranches[:len(response.AcceptedBranches)-1]
	case len(response.RejectedBranches) > 0:
		response.RejectedBranches = response.RejectedBranches[:len(response.RejectedBranches)-1]
	case len(response.UnresolvedQuestions) > 0:
		response.UnresolvedQuestions = response.UnresolvedQuestions[:len(response.UnresolvedQuestions)-1]
	default:
		return false
	}
	return true
}
