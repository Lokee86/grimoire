package interstack

import (
	"regexp"
	"strings"
	"unicode"
)

var constantPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^\s*const\s+([A-Za-z_][A-Za-z0-9_]*)[^=]*=\s*["']([^"']+)["']`),
	regexp.MustCompile(`^\s*(?:pub(?:\([^)]*\))?\s+)?const\s+([A-Za-z_][A-Za-z0-9_]*)[^=]*=\s*["']([^"']+)["']`),
	regexp.MustCompile(`^\s*(?:export\s+)?const\s+([A-Za-z_][A-Za-z0-9_]*)[^=]*=\s*["']([^"']+)["']`),
	regexp.MustCompile(`^\s*([A-Z][A-Z0-9_]*)\s*=\s*["']([^"']+)["']`),
}
var indexedTypeAssignmentPattern = regexp.MustCompile(`\[\s*(?:FIELD_TYPE|["']type["'])\s*\]\s*=\s*([A-Za-z_][A-Za-z0-9_\.:]*|["'][^"']+["'])`)
var structTypeAssignmentPattern = regexp.MustCompile(`\bType\s*:\s*([A-Za-z_][A-Za-z0-9_\.:]*|["'][^"']+["'])`)
var mapTypeAssignmentPattern = regexp.MustCompile(`["']type["']\s*:\s*([A-Za-z_][A-Za-z0-9_\.:]*|["'][^"']+["'])`)
var switchCasePattern = regexp.MustCompile(`^\s*case\s+(.+?)\s*:`)
var gdscriptMatchBranchPattern = regexp.MustCompile(`^\s*(TYPE_[A-Z0-9_]+)\s*:`)
var registrationPattern = regexp.MustCompile(`(?i)\b(?:register_(?:handler|message|packet)|subscribe_(?:message|packet)|on_(?:message|packet)|handle_(?:message|packet))\s*\(\s*([A-Za-z_][A-Za-z0-9_\.:]*|["'][^"']+["'])`)

func (r *resolver) collectConstants(file sourceFile) {
	for _, line := range file.Lines {
		for _, pattern := range constantPatterns {
			match := pattern.FindStringSubmatch(line)
			if len(match) != 3 {
				continue
			}
			name := match[1]
			value := match[2]
			values := r.constants[name]
			if values == nil {
				values = make(map[string]struct{})
				r.constants[name] = values
			}
			values[value] = struct{}{}
		}
	}
}

func (r *resolver) detectMessageProducers(file sourceFile) {
	for index, line := range file.Lines {
		owner, ok := r.index.ownerAt(file.Path, uint32(index+1))
		if !ok {
			continue
		}
		tokens := messageAssignmentTokens(line)
		for _, token := range tokens {
			value, ok := r.resolveMessageValue(token)
			if !ok || !looksLikeMessageValue(value) {
				continue
			}
			channel := r.addMessageChannel(value)
			r.addEdge(factEdge{
				Source: owner.ID, Target: channel.ID, Relation: "publishes",
				Span: lineSpan(file.Path, index+1, line),
				Attributes: map[string]any{
					"confidence": 1.0, "evidence": []string{strings.TrimSpace(line)}, "transport": "packet",
				},
			})
			r.result.Summary.MessageLinks++
		}
	}
}

func (r *resolver) detectMessageConsumers(file sourceFile) {
	packetDispatchUntil := -1
	for index, line := range file.Lines {
		trimmed := strings.ToLower(strings.TrimSpace(line))
		if (strings.Contains(trimmed, "switch ") || strings.Contains(trimmed, "match ")) &&
			(strings.Contains(trimmed, ".type") || strings.Contains(trimmed, `["type"]`) || strings.Contains(trimmed, "packet")) {
			packetDispatchUntil = index + 120
		}
		owner, ok := r.index.ownerAt(file.Path, uint32(index+1))
		if !ok {
			continue
		}
		tokens := make([]string, 0)
		if index <= packetDispatchUntil {
			if match := switchCasePattern.FindStringSubmatch(line); len(match) == 2 {
				tokens = append(tokens, strings.Split(match[1], ",")...)
			}
			if match := gdscriptMatchBranchPattern.FindStringSubmatch(line); len(match) == 2 {
				tokens = append(tokens, match[1])
			}
		}
		if match := registrationPattern.FindStringSubmatch(line); len(match) == 2 {
			tokens = append(tokens, match[1])
		}
		for _, token := range tokens {
			value, ok := r.resolveMessageValue(token)
			if !ok || !looksLikeMessageValue(value) {
				continue
			}
			channel := r.addMessageChannel(value)
			r.addEdge(factEdge{
				Source: channel.ID, Target: owner.ID, Relation: "consumes",
				Span: lineSpan(file.Path, index+1, line),
				Attributes: map[string]any{
					"confidence": 1.0, "evidence": []string{strings.TrimSpace(line)}, "transport": "packet",
				},
			})
			r.result.Summary.MessageLinks++
		}
	}
}

func messageAssignmentTokens(line string) []string {
	result := make([]string, 0, 3)
	for _, pattern := range []*regexp.Regexp{indexedTypeAssignmentPattern, structTypeAssignmentPattern, mapTypeAssignmentPattern} {
		if match := pattern.FindStringSubmatch(line); len(match) == 2 {
			result = append(result, match[1])
		}
	}
	return result
}

func (r *resolver) resolveMessageValue(token string) (string, bool) {
	token = strings.Trim(strings.TrimSpace(token), `"'`)
	if token == "" {
		return "", false
	}
	if strings.Contains(token, "_") && strings.ToLower(token) == token {
		return token, true
	}
	name := lastIdentifier(token)
	if name == "" {
		return "", false
	}
	value, ok := uniqueString(r.constants[name])
	return value, ok
}

func (r *resolver) addMessageChannel(value string) Node {
	identity := "packet\x00" + value
	id := stableNodeID("message-channel", identity)
	if existing, ok := r.nodes[id]; ok {
		return existing
	}
	node := Node{
		ID: id, Kind: "message-channel", Name: value,
		Path: syntheticPath("messages", identity), QualifiedName: "packet:" + value,
		Attributes: map[string]any{"message": value, "transport": "packet"},
	}
	r.addNode(node)
	r.result.Summary.MessageChannels++
	return node
}

func looksLikeMessageValue(value string) bool {
	if value == "" || len(value) > 160 || strings.ContainsAny(value, "/ \\:") {
		return false
	}
	for _, character := range value {
		if unicode.IsUpper(character) {
			return false
		}
	}
	value = strings.ToLower(value)
	for _, suffix := range []string{
		"_request", "-request", ".request", "_response", "-response", ".response",
		"_event", "-event", ".event", "_command", "-command", ".command",
		"_message", "-message", ".message", "_packet", "-packet", ".packet",
	} {
		if strings.HasSuffix(value, suffix) {
			return true
		}
	}
	return false
}
