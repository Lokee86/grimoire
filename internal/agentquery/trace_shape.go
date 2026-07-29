package agentquery

import (
	"sort"
	"strings"
	"unicode"

	"github.com/Lokee86/grimoire/internal/arcanagraph"
	"github.com/Lokee86/grimoire/internal/structure"
)

const maxTraceCandidates = 64

func traceCandidateLimit(limit int) int {
	if limit <= 0 {
		return 1
	}
	candidateLimit := limit * 2
	if candidateLimit < limit {
		candidateLimit = limit
	}
	if candidateLimit > maxTraceCandidates {
		candidateLimit = maxTraceCandidates
	}
	return candidateLimit
}

func finalizeTraceResponse(request Request, response *Response, limit int) {
	if len(response.Paths) == 0 {
		return
	}
	query := strings.TrimSpace(request.Query + " " + request.Anchor + " " + request.Target)
	for index := range response.Paths {
		response.Paths[index].Score = tracePathScore(response.Paths[index], query)
		response.Paths[index].Summary = tracePathSummary(response.Paths[index])
	}
	sort.SliceStable(response.Paths, func(left, right int) bool {
		if response.Paths[left].Score != response.Paths[right].Score {
			return response.Paths[left].Score > response.Paths[right].Score
		}
		if len(response.Paths[left].Steps) != len(response.Paths[right].Steps) {
			return len(response.Paths[left].Steps) < len(response.Paths[right].Steps)
		}
		return response.Paths[left].Summary < response.Paths[right].Summary
	})

	hasBehavior := false
	for _, path := range response.Paths {
		hasBehavior = hasBehavior || tracePathHasBehavior(path)
	}
	seen := make(map[string]bool, len(response.Paths))
	entryCounts := make(map[string]int)
	uniquePaths := make([]Path, 0, len(response.Paths))
	for _, path := range response.Paths {
		if directRuntimeIntrinsicPath(path) {
			response.Truncated = true
			continue
		}
		if hasBehavior && !tracePathHasBehavior(path) {
			response.Truncated = true
			continue
		}
		key := tracePathSignature(path)
		if seen[key] {
			response.Truncated = true
			continue
		}
		if request.Target == "" {
			entry := traceEntryKey(path)
			if entry != "" && entryCounts[entry] >= 2 {
				response.Truncated = true
				continue
			}
			entryCounts[entry]++
		}
		seen[key] = true
		uniquePaths = append(uniquePaths, path)
	}
	if limit > 0 && len(uniquePaths) > limit {
		uniquePaths = uniquePaths[:limit]
		response.Truncated = true
	}
	for index := range uniquePaths {
		uniquePaths[index].Rank = index + 1
		if request.Detail != "full" {
			compactTracePath(&uniquePaths[index])
		}
	}
	response.Paths = uniquePaths
}

func compactTracePath(path *Path) {
	start := traceBehaviorStartIndex(*path)
	path.ContinuationHandles = make([]string, 0, len(path.Nodes)-start)
	for _, node := range path.Nodes[start:] {
		if node.Handle.Value == "" {
			continue
		}
		if len(path.ContinuationHandles) == 0 || path.ContinuationHandles[len(path.ContinuationHandles)-1] != node.Handle.Value {
			path.ContinuationHandles = append(path.ContinuationHandles, node.Handle.Value)
		}
	}
	path.Relations = make([]string, 0, len(path.Steps)-start)
	path.Evidence = make([]TraceEvidence, 0, len(path.Steps)-start)
	for _, step := range path.Steps[start:] {
		path.Relations = append(path.Relations, step.Relation)
		if len(step.Spans) == 0 {
			continue
		}
		span := step.Spans[0]
		path.Evidence = append(path.Evidence, TraceEvidence{
			Relation: step.Relation, Path: span.Path,
			StartLine: span.StartLine, EndLine: span.EndLine,
			Handle: span.Handle.Value,
		})
	}
	path.Nodes = nil
	path.Steps = nil
}

func tracePathHasBehavior(path Path) bool {
	for _, step := range path.Steps {
		if !traceContextRelation(step.Relation) {
			return true
		}
	}
	return false
}

func rankTraceNeighbors(neighbors []arcanagraph.QueryNeighbor, query string) []arcanagraph.QueryNeighbor {
	ranked := append([]arcanagraph.QueryNeighbor(nil), neighbors...)
	terms := traceTerms(query)
	sort.SliceStable(ranked, func(left, right int) bool {
		leftScore := traceRelationScore(ranked[left].Relation) + traceNodeScore(ranked[left].Node, terms) + traceNodeLocationScore(ranked[left].Node)
		rightScore := traceRelationScore(ranked[right].Relation) + traceNodeScore(ranked[right].Node, terms) + traceNodeLocationScore(ranked[right].Node)
		if leftScore != rightScore {
			return leftScore > rightScore
		}
		if ranked[left].Relation != ranked[right].Relation {
			return ranked[left].Relation < ranked[right].Relation
		}
		return traceStructureNodeLabel(ranked[left].Node) < traceStructureNodeLabel(ranked[right].Node)
	})
	return ranked
}

func traceRelationsHaveBehavior(relations []string) bool {
	for _, relation := range relations {
		if !traceContextRelation(relation) {
			return true
		}
	}
	return false
}

func traceContextRelation(relation string) bool {
	relation = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(relation)), "possible-")
	switch relation {
	case "contains", "defines", "imports":
		return true
	default:
		return false
	}
}

func traceRelationScore(relation string) float64 {
	relation = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(relation)), "possible-")
	switch relation {
	case "calls-endpoint", "handled-by", "publishes", "consumes":
		return 120
	case "calls", "invokes", "dispatches", "routes-to", "sends", "receives":
		return 105
	case "creates", "deletes", "removes", "updates", "mutates", "writes", "reads":
		return 95
	case "overrides", "implements", "extends":
		return 80
	case "references", "reads-config":
		return 55
	case "imports":
		return 20
	case "contains", "defines":
		return 8
	default:
		return 45
	}
}

func tracePathScore(path Path, query string) float64 {
	terms := traceTerms(query)
	start := traceBehaviorStartIndex(path)
	score := 0.0
	behavioral := false
	if start < len(path.Nodes) {
		score += traceAgentNodeScore(path.Nodes[start], terms) * 5
	}
	for index, step := range path.Steps {
		if traceContextRelation(step.Relation) {
			continue
		}
		behavioral = true
		weight := 0.4
		if index == start {
			weight = 0.7
		}
		score += traceRelationScore(step.Relation) * weight
	}
	for index := start + 1; index < len(path.Nodes); index++ {
		score += traceAgentNodeScore(path.Nodes[index], terms) * 0.25
	}
	if tracePathEntersDiagnostics(path, start) {
		score -= 60
	}
	if tracePathEndsInLowValueRead(path) {
		score -= 70
	}
	if !behavioral {
		score -= 100
	}
	score -= float64(len(path.Steps)-start) * 2
	return score
}

func tracePathEntersDiagnostics(path Path, start int) bool {
	for index := start + 1; index < len(path.Nodes); index++ {
		value := strings.ToLower(path.Nodes[index].Path + " " + path.Nodes[index].QualifiedName)
		if strings.Contains(value, "/logging/") || strings.Contains(value, "/observability/") ||
			strings.Contains(value, "logger") || strings.Contains(value, "telemetry") || strings.Contains(value, "metrics") {
			return true
		}
	}
	return false
}

func tracePathEndsInLowValueRead(path Path) bool {
	if len(path.Steps) == 0 || len(path.Nodes) == 0 {
		return false
	}
	lastStep := path.Steps[len(path.Steps)-1]
	if strings.TrimPrefix(strings.ToLower(lastStep.Relation), "possible-") != "reads" {
		return false
	}
	kind := strings.ToLower(path.Nodes[len(path.Nodes)-1].Kind)
	switch kind {
	case "parameter", "local", "local-variable", "variable":
		return true
	default:
		return false
	}
}

func traceAgentNodeScore(node Node, terms []string) float64 {
	value := strings.ToLower(strings.Join([]string{node.Name, node.QualifiedName, node.Path, node.Kind}, " "))
	return traceTextScore(value, terms) + tracePathLocationScore(node.Path)
}

func directRuntimeIntrinsicPath(path Path) bool {
	if len(path.Steps) != 1 || len(path.Nodes) < 2 {
		return false
	}
	return tracePathLocationScore(path.Nodes[len(path.Nodes)-1].Path) <= -100
}

func tracePathLocationScore(path string) float64 {
	path = strings.ToLower(strings.TrimSpace(path))
	switch {
	case strings.HasPrefix(path, "@builtin/"), strings.HasPrefix(path, "@stdlib/"):
		return -140
	case strings.HasPrefix(path, "@external/"):
		return -50
	default:
		return 0
	}
}

func traceNodeScore(node structure.Node, terms []string) float64 {
	value := strings.ToLower(strings.Join([]string{node.Name, node.QualifiedName, node.Path, node.Kind}, " "))
	return traceTextScore(value, terms)
}

func traceNodeLocationScore(node structure.Node) float64 {
	return tracePathLocationScore(node.Path)
}

func traceTextScore(value string, terms []string) float64 {
	score := 0.0
	for _, term := range terms {
		if strings.Contains(value, term) {
			score += 18
		}
	}
	for _, mutation := range []string{
		"apply", "buffer", "clear", "create", "delete", "despawn", "insert", "mutate",
		"remove", "spawn", "sync", "tombstone", "update", "write",
	} {
		if strings.Contains(value, mutation) {
			score += 10
		}
	}
	for _, behavior := range []string{
		"consume", "dispatch", "handle", "packet", "process", "publish", "receive", "send",
	} {
		if strings.Contains(value, behavior) {
			score += 4
		}
	}
	return score
}

func tracePathSummary(path Path) string {
	if len(path.Nodes) == 0 {
		return ""
	}
	start := traceBehaviorStartIndex(path)
	if start >= len(path.Nodes) {
		start = 0
	}
	var builder strings.Builder
	if start > 0 {
		builder.WriteString(traceNodeLabel(path.Nodes[0]))
		builder.WriteString(" member ")
	}
	builder.WriteString(traceNodeLabel(path.Nodes[start]))
	for index := start; index < len(path.Steps); index++ {
		step := path.Steps[index]
		builder.WriteString(" --")
		builder.WriteString(step.Relation)
		builder.WriteString("→ ")
		if index+1 < len(path.Nodes) {
			builder.WriteString(traceNodeLabel(path.Nodes[index+1]))
		} else {
			builder.WriteString(step.To.Path)
		}
	}
	return builder.String()
}

func traceNodeLabel(node Node) string {
	if node.QualifiedName != "" {
		return node.QualifiedName
	}
	if node.Name != "" {
		return node.Name
	}
	if node.Path != "" {
		return node.Path
	}
	return node.Kind
}

func traceStructureNodeLabel(node structure.Node) string {
	if node.QualifiedName != "" {
		return node.QualifiedName
	}
	if node.Name != "" {
		return node.Name
	}
	return node.Path
}

func tracePathSignature(path Path) string {
	start := traceBehaviorStartIndex(path)
	var builder strings.Builder
	for index := start; index < len(path.Nodes); index++ {
		if index > start {
			builder.WriteByte('|')
		}
		builder.WriteString(strings.ToLower(traceNodeLabel(path.Nodes[index])))
		if index < len(path.Steps) {
			builder.WriteByte('>')
			builder.WriteString(strings.ToLower(path.Steps[index].Relation))
			builder.WriteByte('>')
		}
	}
	if builder.Len() == 0 {
		for _, step := range path.Steps {
			if traceContextRelation(step.Relation) {
				continue
			}
			builder.WriteString(step.From.Value)
			builder.WriteByte('>')
			builder.WriteString(step.Relation)
			builder.WriteByte('>')
			builder.WriteString(step.To.Value)
		}
	}
	return builder.String()
}

func traceBehaviorStartIndex(path Path) int {
	start := 0
	for start < len(path.Steps) && traceContextRelation(path.Steps[start].Relation) {
		start++
	}
	if start >= len(path.Nodes) {
		return max(0, len(path.Nodes)-1)
	}
	return start
}

func traceEntryKey(path Path) string {
	if len(path.Nodes) == 0 {
		return ""
	}
	start := traceBehaviorStartIndex(path)
	if start >= len(path.Nodes) {
		return ""
	}
	return strings.ToLower(traceNodeLabel(path.Nodes[start]))
}

func distinctBehaviorEntries(paths []Path) int {
	seen := make(map[string]bool)
	for _, path := range paths {
		if !tracePathHasBehavior(path) {
			continue
		}
		if key := traceEntryKey(path); key != "" {
			seen[key] = true
		}
	}
	return len(seen)
}

func traceTerms(value string) []string {
	fields := strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_'
	})
	seen := make(map[string]bool, len(fields))
	result := make([]string, 0, len(fields))
	for _, field := range fields {
		if len(field) < 3 || seen[field] {
			continue
		}
		seen[field] = true
		result = append(result, field)
	}
	return result
}
