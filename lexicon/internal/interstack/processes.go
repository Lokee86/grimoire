package interstack

import (
	"path/filepath"
	"regexp"
	"strings"
)

var execCommandPattern = regexp.MustCompile(`\bexec\.Command\(\s*([^,]+)(.*)\)`)
var execCommandContextPattern = regexp.MustCompile(`\bexec\.CommandContext\(\s*[^,]+,\s*([^,]+)(.*)\)`)
var quotedArgumentPattern = regexp.MustCompile(`["']([^"']+)["']`)
var goCLICommandPattern = regexp.MustCompile(`^\s*case\s+["']([a-z][a-z0-9-]*)["']\s*:`)
var rustCLICommandPattern = regexp.MustCompile(`cli::Command::([A-Z][A-Za-z0-9]*)`)

var boundaryCommands = map[string]map[string]struct{}{
	"lexicon": {
		"init": {}, "scan": {}, "demon": {}, "rebuild": {}, "export": {}, "gc": {},
		"languages": {}, "consumer": {}, "status": {}, "doctor": {}, "version": {},
	},
	"arcana": {
		"benchmark": {}, "import-facts": {}, "sync": {}, "update-facts": {}, "query": {},
		"vectorize": {}, "semantic-query": {}, "protocol": {}, "version": {},
	},
}

func (r *resolver) detectProcessContracts(file sourceFile) {
	r.detectProcessInvocations(file)
	r.detectCLICommandOwnership(file)
	r.detectArcanaProtocol(file)
}

func (r *resolver) detectProcessInvocations(file sourceFile) {
	for index, line := range file.Lines {
		executable, arguments, ok := parseExecCommand(line)
		if !ok {
			continue
		}
		processName, ok := r.resolveBoundaryProcess(file.Path, executable)
		if !ok {
			continue
		}
		owner, ok := r.index.ownerAt(file.Path, uint32(index+1))
		if !ok {
			continue
		}
		process := r.addProcess(processName)
		r.addEdge(factEdge{
			Source: owner.ID, Target: process.ID, Relation: "invokes-process",
			Span:       lineSpan(file.Path, index+1, line),
			Attributes: map[string]any{"confidence": 1.0, "evidence": []string{strings.TrimSpace(line)}},
		})
		r.result.Summary.ProcessLinks++

		if command := firstBoundaryCommand(processName, arguments); command != "" {
			r.linkCommandInvocation(owner, processName, command, file, index, line)
		}
	}

	processName := boundaryProcessForPath(file.Path)
	if processName == "" {
		return
	}
	for index, line := range file.Lines {
		for _, match := range quotedArgumentPattern.FindAllStringSubmatch(line, -1) {
			command := strings.ToLower(match[1])
			if _, ok := boundaryCommands[processName][command]; !ok {
				continue
			}
			if !strings.Contains(line, "Arguments") && !strings.Contains(line, "[]string") && !strings.HasPrefix(strings.TrimSpace(line), `"`+command+`"`) {
				continue
			}
			owner, ok := r.index.ownerAt(file.Path, uint32(index+1))
			if ok {
				r.linkCommandInvocation(owner, processName, command, file, index, line)
			}
		}
	}
}

func parseExecCommand(line string) (string, []string, bool) {
	match := execCommandContextPattern.FindStringSubmatch(line)
	if len(match) != 3 {
		match = execCommandPattern.FindStringSubmatch(line)
	}
	if len(match) != 3 {
		return "", nil, false
	}
	arguments := make([]string, 0)
	for _, value := range quotedArgumentPattern.FindAllStringSubmatch(match[2], -1) {
		arguments = append(arguments, value[1])
	}
	return strings.TrimSpace(match[1]), arguments, true
}

func (r *resolver) resolveBoundaryProcess(path, token string) (string, bool) {
	value := strings.Trim(strings.TrimSpace(token), `"'`)
	if constant, ok := uniqueString(r.constants[lastIdentifier(value)]); ok {
		value = constant
	}
	base := strings.ToLower(strings.TrimSuffix(filepath.Base(filepath.FromSlash(value)), ".exe"))
	if _, ok := boundaryCommands[base]; ok {
		return base, true
	}
	value = boundaryProcessForPath(path)
	return value, value != ""
}

func boundaryProcessForPath(path string) string {
	path = strings.ToLower(filepath.ToSlash(path))
	switch {
	case strings.Contains(path, "/arcanagraph/") || strings.HasPrefix(path, "internal/arcanagraph/"):
		return "arcana"
	case strings.Contains(path, "/lexiconfacts/") || strings.HasPrefix(path, "internal/lexiconfacts/"):
		return "lexicon"
	default:
		return ""
	}
}

func firstBoundaryCommand(process string, arguments []string) string {
	for _, argument := range arguments {
		argument = strings.ToLower(strings.TrimSpace(argument))
		if _, ok := boundaryCommands[process][argument]; ok {
			return argument
		}
	}
	return ""
}

func (r *resolver) linkCommandInvocation(owner Node, processName, command string, file sourceFile, index int, line string) {
	process := r.addProcess(processName)
	commandNode := r.addCLICommand(processName, command)
	r.addEdge(factEdge{Source: process.ID, Target: commandNode.ID, Relation: "contains"})
	r.addEdge(factEdge{
		Source: owner.ID, Target: commandNode.ID, Relation: "calls",
		Span:       lineSpan(file.Path, index+1, line),
		Attributes: map[string]any{"confidence": 1.0, "evidence": []string{strings.TrimSpace(line)}},
	})
	r.result.Summary.CommandLinks++
}

func (r *resolver) detectCLICommandOwnership(file sourceFile) {
	path := strings.ToLower(filepath.ToSlash(file.Path))
	processName := ""
	pattern := (*regexp.Regexp)(nil)
	switch {
	case strings.Contains(path, "lexicon/internal/cli/"):
		processName, pattern = "lexicon", goCLICommandPattern
	case path == "arcana/src/main.rs" || strings.HasSuffix(path, "/arcana/src/main.rs"):
		processName, pattern = "arcana", rustCLICommandPattern
	default:
		return
	}
	process := r.addProcess(processName)
	for index, line := range file.Lines {
		match := pattern.FindStringSubmatch(line)
		if len(match) != 2 {
			continue
		}
		command := strings.ToLower(match[1])
		if processName == "arcana" {
			command = camelToKebab(match[1])
		}
		if command == "help" {
			continue
		}
		if _, ok := boundaryCommands[processName][command]; !ok {
			continue
		}
		owner, ok := r.index.ownerAt(file.Path, uint32(index+1))
		if !ok {
			continue
		}
		commandNode := r.addCLICommand(processName, command)
		r.addEdge(factEdge{Source: process.ID, Target: commandNode.ID, Relation: "contains"})
		r.addEdge(factEdge{
			Source: owner.ID, Target: commandNode.ID, Relation: "defines",
			Span:       lineSpan(file.Path, index+1, line),
			Attributes: map[string]any{"confidence": 1.0, "evidence": []string{strings.TrimSpace(line)}},
		})
		r.result.Summary.CommandLinks++
	}
}
