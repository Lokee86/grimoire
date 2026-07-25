package interstack

import (
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"
)

type httpContract struct {
	Method      string
	Path        string
	Shape       string
	Service     string
	EndpointID  string
	HandlerID   string
	Span        *Span
	Framework   string
	HandlerName string
}

type httpProducer struct {
	SourceID string
	Method   string
	Path     string
	Shape    string
	Span     *Span
	Evidence string
}

var quotedHTTPValue = regexp.MustCompile(`["']([^"']*(?:https?://|/api/|/health\b|/ws\b)[^"']*)["']`)
var placeholderPattern = regexp.MustCompile(`\{[^}/]+\}|:[A-Za-z_][A-Za-z0-9_]*|%\{[^}]+\}|%s|#\{[^}]+\}`)
var configuredHTTPMethodPattern = regexp.MustCompile(`(?i)["']method["']\s*:\s*["'](GET|POST|PUT|PATCH|DELETE)["']`)
var httpClientMethodPattern = regexp.MustCompile(`(?i)(?:^|[^A-Za-z0-9_])(get|post|put|patch|delete)(?:_json)?\s*\(`)
var httpHelperCallPattern = regexp.MustCompile(`\b([A-Za-z_][A-Za-z0-9_]*)\s*\(`)

func (r *resolver) collectHTTPPathProviders(file sourceFile) {
	if file.Ext == ".rb" && strings.HasSuffix(file.Path, "config/routes.rb") {
		return
	}
	for index, line := range file.Lines {
		owner, ok := r.index.ownerAt(file.Path, uint32(index+1))
		if !ok || !isHTTPProviderName(owner.Name) || !strings.Contains(strings.ToLower(line), "return") {
			continue
		}
		for _, match := range quotedHTTPValue.FindAllStringSubmatch(line, -1) {
			normalized := normalizeHTTPPath(match[1])
			if normalized == "" {
				continue
			}
			producer := httpProducer{
				SourceID: owner.ID,
				Path:     normalized,
				Shape:    httpPathShape(normalized),
				Span:     lineSpan(file.Path, index+1, line),
				Evidence: strings.TrimSpace(line),
			}
			r.httpProviders[owner.Name] = append(r.httpProviders[owner.Name], producer)
			r.httpSources = append(r.httpSources, producer)
		}
	}
}

func (r *resolver) detectHTTPProducers(file sourceFile) {
	if file.Ext == ".rb" && strings.HasSuffix(file.Path, "config/routes.rb") {
		return
	}
	for index, line := range file.Lines {
		if !looksLikeHTTPClientCall(line) {
			continue
		}
		owner, ok := r.index.ownerAt(file.Path, uint32(index+1))
		if !ok {
			continue
		}
		block, end := httpCallBlock(file.Lines, index)
		method := httpMethodFromLine(block)
		span := blockSpan(file.Path, index+1, end+1, block)
		evidence := compactEvidence(block)
		for _, match := range quotedHTTPValue.FindAllStringSubmatch(block, -1) {
			r.appendHTTPProducer(owner.ID, method, match[1], span, evidence)
		}
		for _, match := range httpHelperCallPattern.FindAllStringSubmatch(block, -1) {
			helperName := match[1]
			if !isHTTPProviderName(helperName) {
				continue
			}
			provider, found := uniqueHTTPProvider(r.httpProviders[helperName])
			if !found {
				reason := "missing-target"
				if len(r.httpProviders[helperName]) > 1 {
					reason = "ambiguous-target"
				}
				r.addUnresolved(factUnresolved{
					Source: owner.ID, Relation: "calls-endpoint",
					Expression: helperName + "()", Reason: reason, Span: span,
					Attributes: map[string]any{
						"candidate_count": len(r.httpProviders[helperName]),
						"strategy":        "route-provider",
						"transport":       "http",
					},
				})
				continue
			}
			r.httpSources = append(r.httpSources, httpProducer{
				SourceID: owner.ID,
				Method:   method,
				Path:     provider.Path,
				Shape:    provider.Shape,
				Span:     span,
				Evidence: evidence,
			})
		}
	}
}

func (r *resolver) appendHTTPProducer(sourceID, method, rawPath string, span *Span, evidence string) {
	normalized := normalizeHTTPPath(rawPath)
	if normalized == "" {
		return
	}
	r.httpSources = append(r.httpSources, httpProducer{
		SourceID: sourceID,
		Method:   method,
		Path:     normalized,
		Shape:    httpPathShape(normalized),
		Span:     span,
		Evidence: evidence,
	})
}

func (r *resolver) resolveHTTPProducers() {
	seen := make(map[string]struct{})
	for _, producer := range r.httpSources {
		key := strings.Join([]string{
			producer.SourceID,
			producer.Method,
			producer.Shape,
			spanSortKey(producer.Span),
		}, "\x00")
		if _, duplicate := seen[key]; duplicate {
			continue
		}
		seen[key] = struct{}{}
		candidates := make([]httpContract, 0)
		for _, contract := range r.http {
			if contract.Shape != producer.Shape {
				continue
			}
			if producer.Method != "" && contract.Method != "*" && contract.Method != producer.Method {
				continue
			}
			candidates = append(candidates, contract)
		}
		sort.Slice(candidates, func(left, right int) bool {
			return candidates[left].EndpointID < candidates[right].EndpointID
		})
		if len(candidates) == 1 {
			confidence := 1.0
			strategy := "method-and-route"
			if producer.Method == "" || candidates[0].Method == "*" {
				confidence = 0.85
				strategy = "unique-route"
			}
			r.addEdge(factEdge{
				Source: producer.SourceID, Target: candidates[0].EndpointID,
				Relation: "calls-endpoint", Span: producer.Span,
				Attributes: map[string]any{
					"confidence": confidence, "evidence": []string{producer.Evidence},
					"strategy": strategy, "transport": "http",
				},
			})
			r.result.Summary.HTTPLinks++
			continue
		}
		reason := "missing-target"
		if len(candidates) > 1 {
			reason = "ambiguous-target"
		}
		r.addUnresolved(factUnresolved{
			Source: producer.SourceID, Relation: "calls-endpoint",
			Expression: displayHTTPExpression(producer.Method, producer.Path),
			Reason:     reason, Span: producer.Span,
			Attributes: map[string]any{"candidate_count": len(candidates), "transport": "http"},
		})
	}
}

func (r *resolver) addHTTPContract(contract httpContract) {
	identity := strings.Join([]string{contract.Service, contract.Method, contract.Path}, "\x00")
	contract.EndpointID = stableNodeID("http-endpoint", identity)
	name := displayHTTPExpression(contract.Method, contract.Path)
	r.addNode(Node{
		ID: contract.EndpointID, Kind: "http-endpoint", Name: name,
		Path: syntheticPath("http", identity), QualifiedName: "http:" + contract.Service + ":" + name,
		Attributes: map[string]any{
			"framework": contract.Framework, "method": contract.Method,
			"route": contract.Path, "service": contract.Service, "transport": "http",
		},
	})
	r.http = append(r.http, contract)
	r.result.Summary.HTTPContracts++
	if contract.HandlerID != "" {
		r.addEdge(factEdge{
			Source: contract.EndpointID, Target: contract.HandlerID,
			Relation: "handled-by", Span: contract.Span,
			Attributes: map[string]any{
				"confidence": 1.0, "framework": contract.Framework, "transport": "http",
			},
		})
		return
	}
	r.addUnresolved(factUnresolved{
		Source: contract.EndpointID, Relation: "handled-by",
		Expression: contract.HandlerName, Reason: "missing-target", Span: contract.Span,
		Attributes: map[string]any{"framework": contract.Framework, "transport": "http"},
	})
}

func looksLikeHTTPClientCall(line string) bool {
	lower := strings.ToLower(line)
	if strings.Contains(lower, "handlefunc(") || strings.Contains(lower, ".handler(") {
		return false
	}
	return httpClientMethodPattern.MatchString(line) ||
		strings.Contains(lower, "request(") || strings.Contains(lower, "fetch(") ||
		strings.Contains(lower, "axios.") || strings.Contains(lower, "requests.") ||
		strings.Contains(lower, "httpclient") || strings.Contains(lower, "net::http")
}

func isHTTPProviderName(name string) bool {
	name = strings.ToLower(name)
	return strings.HasSuffix(name, "_path") || strings.HasSuffix(name, "_url") || strings.HasSuffix(name, "_endpoint")
}

func uniqueHTTPProvider(providers []httpProducer) (httpProducer, bool) {
	unique := make(map[string]httpProducer)
	for _, provider := range providers {
		unique[provider.Shape] = provider
	}
	if len(unique) != 1 {
		return httpProducer{}, false
	}
	for _, provider := range unique {
		return provider, true
	}
	return httpProducer{}, false
}

func httpCallBlock(lines []string, start int) (string, int) {
	end := start
	balance := 0
	started := false
	parts := make([]string, 0, 4)
	for end < len(lines) && end < start+12 {
		line := lines[end]
		parts = append(parts, strings.TrimSpace(line))
		balance += strings.Count(line, "(")
		balance -= strings.Count(line, ")")
		started = started || strings.Contains(line, "(")
		if started && balance <= 0 {
			break
		}
		end++
	}
	return strings.Join(parts, "\n"), end
}

func blockSpan(path string, startLine, endLine int, text string) *Span {
	lines := strings.Split(text, "\n")
	endColumn := 2
	if len(lines) > 0 {
		endColumn = len([]rune(lines[len(lines)-1])) + 1
		if endColumn < 2 {
			endColumn = 2
		}
	}
	return &Span{
		Path: path, StartLine: uint32(startLine), StartColumn: 1,
		EndLine: uint32(endLine), EndColumn: uint32(endColumn),
	}
}

func compactEvidence(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func httpMethodFromLine(line string) string {
	if match := httpClientMethodPattern.FindStringSubmatch(line); len(match) == 2 {
		return strings.ToUpper(match[1])
	}
	upper := strings.ToUpper(line)
	for _, method := range []string{"DELETE", "PATCH", "POST", "PUT", "GET"} {
		if strings.Contains(upper, "METHOD_"+method) || strings.Contains(upper, "METHOD"+method) ||
			strings.Contains(strings.ToLower(line), "."+strings.ToLower(method)+"(") {
			return method
		}
	}
	if match := configuredHTTPMethodPattern.FindStringSubmatch(line); len(match) == 2 {
		return strings.ToUpper(match[1])
	}
	return ""
}

func normalizeHTTPPath(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, `\/`, "/"))
	if parsed, err := url.Parse(value); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		value = parsed.Path
	} else if index := strings.Index(value, "/api/"); index >= 0 {
		value = value[index:]
	} else if index := strings.Index(value, "/health"); index >= 0 {
		value = value[index:]
	} else if index := strings.Index(value, "/ws"); index >= 0 {
		value = value[index:]
	}
	if query := strings.IndexAny(value, "?#"); query >= 0 {
		value = value[:query]
	}
	value = placeholderPattern.ReplaceAllString(value, "{param}")
	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}
	value = path.Clean(strings.ReplaceAll(value, "//", "/"))
	if value == "." || value == "/" {
		return ""
	}
	return value
}

func httpPathShape(value string) string {
	return placeholderPattern.ReplaceAllString(value, "{}")
}

func displayHTTPExpression(method, route string) string {
	if method == "" {
		method = "*"
	}
	return method + " " + route
}
