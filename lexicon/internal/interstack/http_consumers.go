package interstack

import (
	"regexp"
	"strings"
)

type railsNamespace struct {
	Depth int
	Name  string
	Path  string
}

var railsNamespacePattern = regexp.MustCompile(`^\s*namespace\s+:([A-Za-z_][A-Za-z0-9_]*)(?:\s*,\s*path:\s*["']([^"']+)["'])?\s+do\b`)
var railsRoutePattern = regexp.MustCompile(`^\s*(get|post|put|patch|delete)\s+["']([^"']+)["']\s*(?:(?:,?\s*to:\s*|=>\s*)["']([^"']+)#([^"']+)["'])`)
var goHandlePattern = regexp.MustCompile(`(?:HandleFunc|Handle)\(\s*["']([^"']+)["']\s*,\s*([A-Za-z_][A-Za-z0-9_\.]*)`)
var goHandlerBindingPattern = regexp.MustCompile(`^\s*([A-Za-z_][A-Za-z0-9_]*)\s*(?::=|=)\s*([A-Za-z_][A-Za-z0-9_\.]*)\s*\(`)
var goRouterPattern = regexp.MustCompile(`\.\s*(GET|POST|PUT|PATCH|DELETE)\(\s*["']([^"']+)["']\s*,\s*([A-Za-z_][A-Za-z0-9_\.]*)`)

func (r *resolver) detectHTTPConsumers(file sourceFile) {
	if file.Ext == ".rb" && strings.HasSuffix(file.Path, "config/routes.rb") {
		r.detectRailsRoutes(file)
	}
	if file.Ext == ".go" {
		r.detectGoRoutes(file)
	}
}

func (r *resolver) detectRailsRoutes(file sourceFile) {
	stack := make([]railsNamespace, 0)
	depth := 0
	for index, line := range file.Lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "end" || strings.HasPrefix(trimmed, "end #") {
			if depth > 0 {
				depth--
			}
			for len(stack) > 0 && stack[len(stack)-1].Depth > depth {
				stack = stack[:len(stack)-1]
			}
			continue
		}
		if match := railsNamespacePattern.FindStringSubmatch(line); len(match) == 3 {
			pathPart := match[2]
			if pathPart == "" {
				pathPart = strings.ReplaceAll(match[1], "_", "-")
			}
			depth++
			stack = append(stack, railsNamespace{Depth: depth, Name: match[1], Path: pathPart})
			continue
		}
		match := railsRoutePattern.FindStringSubmatch(line)
		if len(match) == 5 {
			prefix := make([]string, 0, len(stack)+1)
			controllerParts := make([]string, 0, len(stack)+1)
			for _, namespace := range stack {
				prefix = append(prefix, namespace.Path)
				controllerParts = append(controllerParts, camelize(namespace.Name))
			}
			prefix = append(prefix, match[2])
			route := normalizeHTTPPath("/" + strings.Join(prefix, "/"))
			controller := strings.TrimPrefix(match[3], "/")
			controllerSegments := strings.Split(controller, "/")
			for _, segment := range controllerSegments {
				controllerParts = append(controllerParts, camelize(segment))
			}
			if len(controllerParts) > 0 {
				controllerParts[len(controllerParts)-1] += "Controller"
				handlerName := strings.Join(controllerParts, "::") + "#" + match[4]
				handler, found := r.index.exactQName(handlerName)
				handlerID := ""
				if found {
					handlerID = handler.ID
				}
				r.addHTTPContract(httpContract{
					Method: strings.ToUpper(match[1]), Path: route, Shape: httpPathShape(route),
					Service: componentForPath(file.Path), HandlerID: handlerID,
					Span: lineSpan(file.Path, index+1, line), Framework: "rails", HandlerName: handlerName,
				})
			}
		}
		if strings.HasSuffix(trimmed, " do") {
			depth++
		}
	}
}

func (r *resolver) detectGoRoutes(file sourceFile) {
	bindings := make(map[string]Node)
	for _, line := range file.Lines {
		match := goHandlerBindingPattern.FindStringSubmatch(line)
		if len(match) != 3 {
			continue
		}
		provider, found := r.index.callableByName(lastIdentifier(match[2]), file.Path)
		if found {
			bindings[match[1]] = provider
		}
	}

	for index, line := range file.Lines {
		if match := goHandlePattern.FindStringSubmatch(line); len(match) == 3 {
			method, route := splitGoRoute(match[1])
			handlerName := lastIdentifier(match[2])
			handler, found, confidence, strategy := r.resolveGoHTTPHandler(
				handlerName, file.Path, uint32(index+1), bindings,
			)
			handlerID := ""
			if found {
				handlerID = handler.ID
			}
			r.addHTTPContract(httpContract{
				Method: method, Path: route, Shape: httpPathShape(route),
				Service: componentForPath(file.Path), HandlerID: handlerID,
				HandlerConfidence: confidence, HandlerStrategy: strategy,
				Span: lineSpan(file.Path, index+1, line), Framework: "net/http", HandlerName: handlerName,
			})
			continue
		}
		if match := goRouterPattern.FindStringSubmatch(line); len(match) == 4 {
			route := normalizeHTTPPath(match[2])
			handlerName := lastIdentifier(match[3])
			handler, found, confidence, strategy := r.resolveGoHTTPHandler(
				handlerName, file.Path, uint32(index+1), bindings,
			)
			handlerID := ""
			if found {
				handlerID = handler.ID
			}
			r.addHTTPContract(httpContract{
				Method: strings.ToUpper(match[1]), Path: route, Shape: httpPathShape(route),
				Service: componentForPath(file.Path), HandlerID: handlerID,
				HandlerConfidence: confidence, HandlerStrategy: strategy,
				Span: lineSpan(file.Path, index+1, line), Framework: "go-router", HandlerName: handlerName,
			})
		}
	}
}

func (r *resolver) resolveGoHTTPHandler(name, path string, line uint32, bindings map[string]Node) (Node, bool, float64, string) {
	if provider, found := bindings[name]; found {
		return provider, true, 0.95, "handler-provider-binding"
	}
	if handler, found := r.index.callableByName(name, path); found {
		return handler, true, 1.0, "direct-handler"
	}
	if owner, found := r.index.ownerAt(path, line); found {
		return owner, true, 0.7, "registration-owner"
	}
	return Node{}, false, 0, ""
}

func splitGoRoute(value string) (string, string) {
	parts := strings.Fields(value)
	if len(parts) >= 2 {
		method := strings.ToUpper(parts[0])
		switch method {
		case "GET", "POST", "PUT", "PATCH", "DELETE":
			return method, normalizeHTTPPath(parts[1])
		}
	}
	return "*", normalizeHTTPPath(value)
}

func indentation(value string) int {
	count := 0
	for _, character := range value {
		switch character {
		case ' ':
			count++
		case '\t':
			count += 2
		default:
			return count
		}
	}
	return count
}
