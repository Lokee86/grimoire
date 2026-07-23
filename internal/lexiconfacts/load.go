package lexiconfacts

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func Load(directory string) (*Corpus, error) {
	if strings.TrimSpace(directory) == "" {
		return nil, nil
	}
	facts, err := loadDirectory(directory)
	if err != nil {
		return nil, err
	}
	return &Corpus{facts: facts}, nil
}

func loadDirectory(directory string) (library, error) {
	entries, err := filepath.Glob(filepath.Join(directory, "*.jsonl"))
	if err != nil {
		return library{}, fmt.Errorf("find Lexicon exports: %w", err)
	}
	if len(entries) == 0 {
		return library{}, fmt.Errorf("no Lexicon JSONL exports found in %s", directory)
	}
	sort.Strings(entries)
	result := library{nodes: make(map[string]Node)}
	for _, path := range entries {
		if err := loadFile(path, &result); err != nil {
			return library{}, err
		}
	}
	return result, nil
}

func loadFile(path string, result *library) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open Lexicon export %s: %w", path, err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	line := 0
	for scanner.Scan() {
		line++
		data := scanner.Bytes()
		var header recordHeader
		if err := json.Unmarshal(data, &header); err != nil {
			return fmt.Errorf("decode Lexicon export %s line %d: %w", path, line, err)
		}
		switch header.Record {
		case "node":
			var node Node
			if err := json.Unmarshal(data, &node); err != nil {
				return fmt.Errorf("decode Lexicon node %s line %d: %w", path, line, err)
			}
			if node.ID != "" {
				result.nodes[node.ID] = node
			}
		case "edge":
			var edge Edge
			if err := json.Unmarshal(data, &edge); err != nil {
				return fmt.Errorf("decode Lexicon edge %s line %d: %w", path, line, err)
			}
			if edge.Source != "" && edge.Target != "" {
				result.edges = append(result.edges, edge)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read Lexicon export %s: %w", path, err)
	}
	return nil
}
