package main

const (
	adapterVersion = "0.2.0"
	language       = "lotusscript"
)

type span struct {
	EndColumn   int    `json:"end_column"`
	EndLine     int    `json:"end_line"`
	Path        string `json:"path"`
	StartColumn int    `json:"start_column"`
	StartLine   int    `json:"start_line"`
}

type logicalLine struct {
	span span
	text string
}

type declaration struct {
	className     string
	external      bool
	id            string
	kind          string
	name          string
	ownerID       string
	ownerPath     string
	qualifiedName string
	span          *span
}

type useEvidence struct {
	dynamic    bool
	expression string
	importID   string
	keyword    string
	ownerPath  string
	span       *span
	target     string
}

type callEvidence struct {
	candidate  string
	className  string
	expression string
	ownerID    string
	ownerPath  string
	span       *span
}

type extendsEvidence struct {
	base      string
	classID   string
	ownerPath string
	span      *span
}

type parsedFile struct {
	content     []byte
	contentHash string
	invalid     bool
	moduleID    string
	path        string
}

type analysisState struct {
	basesByClass    map[string][]string
	callablesByName map[string][]declaration
	classesByName   map[string][]declaration
	extends         []extendsEvidence
	facts           *factSet
	fieldsByClass   map[string]map[string]string
	globalsByName   map[string][]declaration
	methodsByClass  map[string]map[string][]declaration
	modulesByName   map[string][]string
	moduleVariables map[string]map[string]string
	typesByCallable map[string]map[string]string
	uses            []useEvidence
	calls           []callEvidence
}
