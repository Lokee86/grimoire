package agentquery

func isDocumentationPath(path string) bool {
	kind, _ := classifyPath(path)
	return kind == "document"
}
