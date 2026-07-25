package interstack

import (
	"regexp"
	"strings"
)

type railsNamespace struct {
	Indent int
	Name   string
	Path   string
}

var railsNamespacePattern = regexp.MustCompile(`^\s*namespace\s+:([A-Za-z_][A-Za-z0-9_]*)(?:\s*,\s*path:\s*["']([^"']+)["'])?\s+do\b`)
var railsRoutePattern = regexp.MustCompile(`^\s*(get|post|put|patch|delete)\s+["']([^"']+)["']\s*(?:(?:,?\s*to:\s*|=>\s*)["']([^"']+)#([^"']+)["'])`)
var goHandleFuncPattern = regexp.MustCompile(`HandleFunc\(\s*["']([^"']+)["']\s*,\s*([A-Za-z_][A-Za-z0-9_\.]*)`)
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
	for index, line := range file.Lines {
		trimmed := strings.TrimSpace(line)
		indent := indentation(line)
		for len(stack) > 0 && indent <= stack[len(stack)-1].Indent && !strings.HasPrefix(trimmed, "namespace ") {
			stack = stack[:len(stack)-1]
		}
		if match := railsNamespacePattern.FindStringSubmatch(line); len(match) == 3 {
			pathPart := match[2]
			if pathPart == "" {
				pathPart = strings.ReplaceAll(match[1], "_", "-")
			}
			stack = append(stack, railsNamespace{Indent: indent, Name: match[1], Path: pathPart})
			continue
		}
		match := railsRoutePattern.FindStringSubmatch(line)
		if len(match) != 5 {
			continue
		}
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
		if len(controllerParts) == 0 {
			continue
		}
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

func (r *resolver) detectGoRoutes(file sourceFile) {
	for index, line := range file.Lines {
		if match := goHandleFuncPattern.FindStringSubmatch(line); len(match) == 3 {
			method, route := splitGoRoute(match[1])
			handlerName := lastIdentifier(match[2])
			handler, found := r.index.callableByName(handlerName, file.Path)
			handlerID := ""
			if found {
				handlerID = handler.ID
			}
			r.addHTTPContract(httpContract{
				Method: method, Path: route, Shape: httpPathShape(route),
				Service: componentForPath(file.Path), HandlerID: handlerID,
				Span: lineSpan(file.Path, index+1, line), Framework: "net/http", HandlerName: handlerName,
			})
			continue
		}
		if match := goRouterPattern.FindStringSubmatch(line); len(match) == 4 {
			route := normalizeHTTPPath(match[2])
			handlerName := lastIdentifier(match[3])
			handler, found := r.index.callableByName(handlerName, file.Path)
			handlerID := ""
			if found {
				handlerID = handler.ID
			}
			r.addHTTPContract(httpContract{
				Method: strings.ToUpper(match[1]), Path: route, Shape: httpPathShape(route),
				Service: componentForPath(file.Path), HandlerID: handlerID,
				Span: lineSpan(file.Path, index+1, line), Framework: "go-router", HandlerName: handlerName,
			})
		}
	}
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
