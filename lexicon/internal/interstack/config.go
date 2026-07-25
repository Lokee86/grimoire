package interstack

import (
	"regexp"
	"strings"
)

type configReadPattern struct {
	Pattern   *regexp.Regexp
	DirectKey bool
}

var configReadPatterns = []configReadPattern{
	{Pattern: regexp.MustCompile(`\bos\.(?:Getenv|LookupEnv)\(\s*([A-Za-z_][A-Za-z0-9_]*|["'][A-Za-z_][A-Za-z0-9_]*["'])`)},
	{Pattern: regexp.MustCompile(`\bENV\s*\[\s*([A-Za-z_][A-Za-z0-9_]*|["'][A-Za-z_][A-Za-z0-9_]*["'])\s*\]`)},
	{Pattern: regexp.MustCompile(`\bENV\.fetch\(\s*([A-Za-z_][A-Za-z0-9_]*|["'][A-Za-z_][A-Za-z0-9_]*["'])`)},
	{Pattern: regexp.MustCompile(`\bprocess\.env\.([A-Za-z_][A-Za-z0-9_]*)`), DirectKey: true},
	{Pattern: regexp.MustCompile(`\bOS\.get_environment\(\s*([A-Za-z_][A-Za-z0-9_]*|["'][A-Za-z_][A-Za-z0-9_]*["'])`)},
	{Pattern: regexp.MustCompile(`\bos\.(?:getenv|environ\.get)\(\s*([A-Za-z_][A-Za-z0-9_]*|["'][A-Za-z_][A-Za-z0-9_]*["'])`)},
}

func (r *resolver) detectConfigReads(file sourceFile) {
	for index, line := range file.Lines {
		owner, ok := r.index.ownerAt(file.Path, uint32(index+1))
		if !ok {
			continue
		}
		for _, pattern := range configReadPatterns {
			for _, match := range pattern.Pattern.FindAllStringSubmatch(line, -1) {
				if len(match) != 2 {
					continue
				}
				key, resolved := r.resolveConfigKey(match[1], pattern.DirectKey)
				if !resolved {
					continue
				}
				node := r.addConfigKey(key)
				r.addEdge(factEdge{
					Source: owner.ID, Target: node.ID, Relation: "reads-config",
					Span: lineSpan(file.Path, index+1, line),
					Attributes: map[string]any{
						"confidence": 1.0, "evidence": []string{strings.TrimSpace(line)}, "transport": "configuration",
					},
				})
			}
		}
	}
}

func (r *resolver) resolveConfigKey(token string, direct bool) (string, bool) {
	token = strings.TrimSpace(token)
	if len(token) >= 2 && ((token[0] == '"' && token[len(token)-1] == '"') ||
		(token[0] == '\'' && token[len(token)-1] == '\'')) {
		return token[1 : len(token)-1], true
	}
	if value, ok := uniqueString(r.constants[token]); ok {
		return value, true
	}
	if direct {
		return token, token != ""
	}
	return "", false
}

func (r *resolver) addConfigKey(name string) Node {
	identity := "config\x00" + name
	id := stableNodeID("config-key", identity)
	if existing, ok := r.nodes[id]; ok {
		return existing
	}
	node := Node{
		ID: id, Kind: "config-key", Name: name,
		Path: syntheticPath("config", identity), QualifiedName: "config:" + name,
		Attributes: map[string]any{"key": name, "transport": "configuration"},
	}
	r.addNode(node)
	r.result.Summary.ConfigKeys++
	return node
}
