package main

import "strings"

const (
	adapterVersion = "0.2.0"
	streamLanguage = "c-family"
)

type sourceSpan struct {
	Path        string
	StartLine   int
	StartColumn int
	EndLine     int
	EndColumn   int
}

func (span sourceSpan) record() map[string]any {
	return map[string]any{
		"end_column":   span.EndColumn,
		"end_line":     span.EndLine,
		"path":         span.Path,
		"start_column": span.StartColumn,
		"start_line":   span.StartLine,
	}
}

type declaration struct {
	ID                 string
	Kind               string
	Name               string
	QualifiedName      string
	Path               string
	ContainerID        string
	ContainerQualified string
	ParentTypeID       string
	Signature          string
	FileLanguage       string
	Span               sourceSpan
	Attributes         map[string]any
	Callable           bool
	Definition         bool
	FileLocal          bool
	MacroFunction      bool
}

type includeObservation struct {
	ID         string
	ModuleID   string
	Path       string
	Target     string
	Expression string
	System     bool
	Span       sourceSpan
}

type callObservation struct {
	SourceID    string
	SourceScope string
	Path        string
	Expression  string
	Candidate   string
	Member      bool
	Span        sourceSpan
}

type accessObservation struct {
	SourceID    string
	SourceScope string
	ParentType  string
	Path        string
	Expression  string
	Candidate   string
	Relation    string
	Member      bool
	Span        sourceSpan
}

type inheritanceObservation struct {
	SourceID    string
	SourceScope string
	Path        string
	Expression  string
	Candidate   string
	Span        sourceSpan
}

type sourceFile struct {
	Path           string
	Language       string
	ParserLanguage string
	Content        []byte
	FileID         string
	ModuleID       string
	ParseError     bool
	Declarations   []*declaration
	Includes       []includeObservation
	Calls          []callObservation
	Accesses       []accessObservation
	Inheritance    []inheritanceObservation
}

type repositoryModel struct {
	Repository   string
	Files        []*sourceFile
	Declarations []*declaration
}

type extractionContext struct {
	ContainerID        string
	ContainerQualified string
	TypeID             string
	TypeName           string
	CallableID         string
	CallableScope      string
	Template           bool
}

func qualify(scope, name string) string {
	raw := strings.TrimSpace(name)
	absolute := strings.HasPrefix(raw, "::")
	name = normalizeQualified(raw)
	if name == "" || scope == "" || absolute || name == scope || strings.HasPrefix(name, scope+"::") {
		return name
	}
	return scope + "::" + name
}

func normalizeQualified(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "::")
	value = strings.ReplaceAll(value, " ", "")
	return value
}
