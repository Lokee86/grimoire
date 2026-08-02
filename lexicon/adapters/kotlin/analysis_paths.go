package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func qualify(owner, name string) string {
	if owner == "" || owner == "<default>" {
		return name
	}
	return owner + "." + name
}

func simpleQualifiedName(value string) string {
	if index := strings.LastIndexByte(value, '.'); index >= 0 {
		return value[index+1:]
	}
	return value
}

func sourceModuleQualifiedName(packageName, path string) string {
	return fmt.Sprintf("%s::source:%s", packageName, path)
}

func pathDirectory(path string) string {
	if index := strings.LastIndexByte(path, '/'); index >= 0 {
		return path[:index]
	}
	return "."
}

func wholeFileSpan(path string, content []byte) sourceSpan {
	line, column := 1, 1
	for offset := 0; offset < len(content); {
		if content[offset] == '\r' {
			if offset+1 < len(content) && content[offset+1] == '\n' {
				offset += 2
			} else {
				offset++
			}
			line++
			column = 1
			continue
		}
		value, size := utf8.DecodeRune(content[offset:])
		offset += size
		if value == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return sourceSpan{EndColumn: column, EndLine: line, Path: path, StartColumn: 1, StartLine: 1}
}
