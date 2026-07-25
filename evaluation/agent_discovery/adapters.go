package agentdiscovery

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Adapter imports a tool's recording. RegisterAdapter permits external CBM
// exporters to join the benchmark without coupling this package to CBM.
type Adapter func(path string) ([]Transcript, error)

var adapters = map[string]Adapter{}

func init() {
	RegisterAdapter("progressive-jsonl", progressiveJSONL)
	RegisterAdapter("raw", rawTranscript)
	RegisterAdapter("grimoire-context", grimoireContext)
}

func RegisterAdapter(name string, adapter Adapter) {
	adapters[strings.ToLower(strings.TrimSpace(name))] = adapter
}

func AdapterFor(name string) (Adapter, bool) {
	adapter, ok := adapters[strings.ToLower(strings.TrimSpace(name))]
	return adapter, ok
}

func progressiveJSONL(path string) ([]Transcript, error) {
	objects, err := jsonObjects(path)
	if err != nil {
		return nil, err
	}
	byRun := map[string]*Transcript{}
	for _, object := range objects {
		var transcript Transcript
		if json.Unmarshal(object, &transcript) == nil && len(transcript.Events) > 0 {
			byRun[transcript.RunID+"\x00"+transcript.CaseID] = &transcript
			continue
		}
		var event Event
		if err := json.Unmarshal(object, &event); err != nil {
			return nil, fmt.Errorf("decode JSONL event: %w", err)
		}
		var header struct {
			Adapter string `json:"adapter"`
			RunID   string `json:"run_id"`
			CaseID  string `json:"case_id"`
		}
		if err := json.Unmarshal(object, &header); err != nil {
			return nil, err
		}
		key := header.RunID + "\x00" + header.CaseID
		if byRun[key] == nil {
			byRun[key] = &Transcript{Adapter: header.Adapter, RunID: header.RunID, CaseID: header.CaseID}
		}
		byRun[key].Events = append(byRun[key].Events, event)
	}
	return sortedTranscripts(byRun), nil
}

func rawTranscript(path string) ([]Transcript, error) {
	objects, err := jsonObjects(path)
	if err != nil {
		return nil, err
	}
	byRun := map[string]*Transcript{}
	for _, object := range objects {
		var record map[string]json.RawMessage
		if err := json.Unmarshal(object, &record); err != nil {
			return nil, err
		}
		event, header := rawEvent(record)
		key := header.RunID + "\x00" + header.CaseID
		if byRun[key] == nil {
			byRun[key] = &Transcript{Adapter: "raw", RunID: header.RunID, CaseID: header.CaseID}
		}
		byRun[key].Events = append(byRun[key].Events, event)
	}
	return sortedTranscripts(byRun), nil
}

func rawEvent(record map[string]json.RawMessage) (Event, struct{ RunID, CaseID string }) {
	var event Event
	data, _ := json.Marshal(record)
	_ = json.Unmarshal(data, &event)
	var header struct {
		RunID  string `json:"run_id"`
		CaseID string `json:"case_id"`
		Type   string `json:"type"`
		Tool   string `json:"tool"`
	}
	_ = json.Unmarshal(data, &header)
	if event.Kind == "" {
		event.Kind = header.Type
	}
	if header.Tool != "" && (strings.Contains(header.Tool, "open") || strings.Contains(header.Tool, "read")) {
		event.Kind = "source_open"
	}
	var args struct {
		Path      string `json:"path"`
		StartLine int    `json:"start_line"`
		EndLine   int    `json:"end_line"`
		Symbol    string `json:"symbol"`
	}
	if value := record["arguments"]; value != nil {
		_ = json.Unmarshal(value, &args)
	}
	if event.Path == "" {
		event.Path, event.StartLine, event.EndLine, event.Symbol = args.Path, args.StartLine, args.EndLine, args.Symbol
	}
	var usage struct {
		Input  int `json:"input_tokens"`
		Output int `json:"output_tokens"`
	}
	if value := record["usage"]; value != nil {
		_ = json.Unmarshal(value, &usage)
		event.InputTokens += usage.Input
		event.OutputTokens += usage.Output
	}
	return event, struct{ RunID, CaseID string }{header.RunID, header.CaseID}
}

func grimoireContext(path string) ([]Transcript, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var pkg struct {
		Query      string `json:"query"`
		TokenCount int    `json:"token_count"`
		Selections []struct {
			Path    string `json:"path"`
			Start   int    `json:"start_line"`
			End     int    `json:"end_line"`
			Tokens  int    `json:"token_count"`
			Content string `json:"content"`
		} `json:"selections"`
		Structural []struct {
			Path   string `json:"path"`
			Start  int    `json:"start_line"`
			End    int    `json:"end_line"`
			Symbol string `json:"symbol"`
		} `json:"structural_evidence"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("decode grimoire context: %w", err)
	}
	result := Transcript{Adapter: "grimoire-context", Events: []Event{{Kind: "query", InputTokens: pkg.TokenCount, InputText: pkg.Query}}}
	for _, selection := range pkg.Selections {
		result.Events = append(result.Events, Event{Kind: "source_open", Path: selection.Path, StartLine: selection.Start, EndLine: selection.End, InputTokens: selection.Tokens, Claim: selection.Content})
	}
	for _, evidence := range pkg.Structural {
		result.Events = append(result.Events, Event{Kind: "source_open", Path: evidence.Path, StartLine: evidence.Start, EndLine: evidence.End, Symbol: evidence.Symbol})
	}
	return []Transcript{result}, nil
}

func jsonObjects(path string) ([]json.RawMessage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	trimmed := bytes.TrimSpace(data)
	if bytes.HasPrefix(trimmed, []byte("[")) {
		var values []json.RawMessage
		err = json.Unmarshal(data, &values)
		return values, err
	}
	if bytes.HasPrefix(trimmed, []byte("{")) {
		var value json.RawMessage
		if err := json.Unmarshal(trimmed, &value); err == nil {
			return []json.RawMessage{value}, nil
		}
	}
	var values []json.RawMessage
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 1024), 4*1024*1024)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) > 0 {
			values = append(values, append(json.RawMessage(nil), line...))
		}
	}
	return values, scanner.Err()
}

func sortedTranscripts(byRun map[string]*Transcript) []Transcript {
	result := make([]Transcript, 0, len(byRun))
	for _, transcript := range byRun {
		if transcript.Adapter == "" {
			transcript.Adapter = "progressive-jsonl"
		}
		result = append(result, *transcript)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].RunID+result[i].CaseID < result[j].RunID+result[j].CaseID })
	return result
}
