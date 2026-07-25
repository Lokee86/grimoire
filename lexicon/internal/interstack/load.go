package interstack

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
)

type wireHeader struct {
	Record     string `json:"record"`
	Language   string `json:"language"`
	Repository string `json:"repository"`
}

type wireRecord struct {
	Record        string         `json:"record"`
	ID            string         `json:"id"`
	Kind          string         `json:"kind"`
	Name          string         `json:"name"`
	Path          string         `json:"path"`
	QualifiedName string         `json:"qualified_name"`
	Span          *Span          `json:"span"`
	Attributes    map[string]any `json:"attributes"`
}

func ParseLibrary(data []byte) (Library, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 64*1024), 32*1024*1024)
	if !scanner.Scan() {
		return Library{}, fmt.Errorf("empty Lexicon library")
	}
	var header wireHeader
	if err := json.Unmarshal(scanner.Bytes(), &header); err != nil {
		return Library{}, fmt.Errorf("decode Lexicon header: %w", err)
	}
	if header.Record != "lexicon" || header.Language == "" || header.Repository == "" {
		return Library{}, fmt.Errorf("invalid Lexicon header")
	}
	library := Library{Language: header.Language, Repository: header.Repository}
	for scanner.Scan() {
		var record wireRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return Library{}, fmt.Errorf("decode %s Lexicon record: %w", header.Language, err)
		}
		if record.Record != "node" || record.ID == "" {
			continue
		}
		library.Nodes = append(library.Nodes, Node{
			ID: record.ID, Kind: record.Kind, Name: record.Name, Path: record.Path,
			QualifiedName: record.QualifiedName, Span: record.Span, Attributes: record.Attributes,
		})
	}
	if err := scanner.Err(); err != nil {
		return Library{}, err
	}
	return library, nil
}
