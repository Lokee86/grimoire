package main

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"strings"
	"unicode/utf8"
)

type dxlNote struct {
	Codes []dxlCode `xml:"code"`
	Items []dxlItem `xml:"item"`
}

type dxlCode struct {
	Event  string `xml:"event,attr"`
	Source string `xml:"lotusscript"`
}

type dxlItem struct {
	Raw []dxlRaw `xml:"rawitemdata"`
}

type dxlRaw struct {
	Text string `xml:",chardata"`
	Type string `xml:"type,attr"`
}

func lotusScriptExtension(extension string) bool {
	switch extension {
	case ".ls", ".lsa", ".lsdb", ".lss":
		return true
	default:
		return false
	}
}

func lotusScriptContent(extension string, raw []byte) ([]byte, bool) {
	if extension == ".ls" || extension == ".lss" {
		return raw, true
	}
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return nil, false
	}
	if trimmed[0] != '<' {
		return raw, utf8.Valid(raw) && !bytes.Contains(raw, []byte{0})
	}
	content := extractDXLSource(raw)
	return content, len(content) != 0
}

func extractDXLSource(raw []byte) []byte {
	var note dxlNote
	if xml.Unmarshal(raw, &note) != nil {
		return nil
	}
	var sections []string
	for _, code := range note.Codes {
		if source := strings.TrimSpace(code.Source); source != "" {
			sections = append(sections, source)
		}
	}
	if len(sections) != 0 {
		return []byte(strings.Join(sections, "\n\n"))
	}
	var best []byte
	for _, item := range note.Items {
		for _, value := range item.Raw {
			if !strings.EqualFold(value.Type, "10") {
				continue
			}
			payload, err := base64.StdEncoding.DecodeString(strings.Join(strings.Fields(value.Text), ""))
			if err != nil {
				continue
			}
			candidate := lotusScriptPayload(payload)
			if len(candidate) > len(best) {
				best = candidate
			}
		}
	}
	return best
}

func lotusScriptPayload(payload []byte) []byte {
	markers := [][]byte{
		[]byte("'++LotusScript Development Environment"),
		[]byte("Option Public"),
		[]byte("Option Declare"),
	}
	start := -1
	for _, marker := range markers {
		if index := bytes.Index(payload, marker); index >= 0 && (start < 0 || index < start) {
			start = index
		}
	}
	if start < 0 {
		return nil
	}
	end := bytes.IndexByte(payload[start:], 0)
	if end < 0 {
		end = len(payload) - start
	}
	result := append([]byte(nil), payload[start:start+end]...)
	for index, value := range result {
		if value == '\t' || value == '\n' || value == '\r' || value >= 0x20 && value <= 0x7e {
			continue
		}
		result[index] = ' '
	}
	return bytes.TrimSpace(result)
}
