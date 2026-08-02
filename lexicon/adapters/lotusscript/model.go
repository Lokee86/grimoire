package main

const (
	adapterVersion = "0.3.0"
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
	classID           string
	className         string
	external          bool
	id                string
	kind              string
	name              string
	ownerID           string
	ownerPath         string
	public            bool
	qualifiedName     string
	typeMembersPublic bool
	span              *span
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
	classID    string
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

type accessEvidence struct {
	classID   string
	ownerID   string
	ownerPath string
	span      *span
	text      string
}

type variableSymbol struct {
	dataType  string
	id        string
	ownerPath string
	public    bool
}

type parsedFile struct {
	content     []byte
	contentHash string
	invalid     bool
	moduleID    string
	path        string
}

type analysisState struct {
	accesses          []accessEvidence
	basesByClass      map[string][]string
	callablesByName   map[string][]declaration
	classesByName     map[string][]declaration
	extends           []extendsEvidence
	facts             *factSet
	fieldSymbols      map[string]map[string]variableSymbol
	fieldsByClass     map[string]map[string]string
	globalsByName     map[string][]declaration
	importsByPath     map[string][]string
	methodsByClass    map[string]map[string][]declaration
	moduleIDByPath    map[string]string
	modulePathsByName map[string][]string
	modulePublic      map[string]bool
	modulesByName     map[string][]string
	moduleSymbols     map[string]map[string]variableSymbol
	moduleVariables   map[string]map[string]string
	typesByCallable   map[string]map[string]string
	variableSymbols   map[string]map[string]variableSymbol
	uses              []useEvidence
	calls             []callEvidence
}
