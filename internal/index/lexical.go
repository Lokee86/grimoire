package index

import (
	"fmt"
	"strings"

	"github.com/Lokee86/grimoire/internal/lexical"
)

func buildLexicalIndex(files []FileRecord, previous *lexical.Index) *lexical.Index {
	inputs := make([]lexical.Input, 0)
	documentIndex := 0
	for _, file := range files {
		for _, chunk := range file.Chunks {
			inputs = append(inputs, lexical.Input{
				Key:  chunkLexicalKey(chunk, documentIndex),
				Path: chunk.Path,
				Text: chunk.Text,
			})
			documentIndex++
		}
	}
	return lexical.Rebuild(inputs, previous)
}

func buildFileLexicalIndex(files []FileRecord, previous *lexical.Index) *lexical.Index {
	inputs := make([]lexical.Input, 0, len(files))
	for _, file := range files {
		var text strings.Builder
		for index, chunk := range file.Chunks {
			if index > 0 {
				text.WriteByte('\n')
			}
			text.WriteString(chunk.Text)
		}
		inputs = append(inputs, lexical.Input{Key: file.Path, Path: file.Path, Text: text.String()})
	}
	return lexical.Rebuild(inputs, previous)
}

func (snapshot Snapshot) LexicalIndex() *lexical.Index {
	if snapshot.lexicalIndex != nil {
		return snapshot.lexicalIndex
	}
	return buildLexicalIndex(snapshot.Files, nil)
}

func (snapshot Snapshot) FileLexicalIndex() *lexical.Index {
	if snapshot.fileLexicalIndex != nil {
		return snapshot.fileLexicalIndex
	}
	return buildFileLexicalIndex(snapshot.Files, nil)
}

func (snapshot Snapshot) PrepareLexical() Snapshot {
	if snapshot.lexicalIndex == nil {
		snapshot.lexicalIndex = buildLexicalIndex(snapshot.Files, nil)
	}
	if snapshot.fileLexicalIndex == nil {
		snapshot.fileLexicalIndex = buildFileLexicalIndex(snapshot.Files, nil)
	}
	return snapshot
}

func validateLexicalIndex(snapshot Snapshot, lexicalIndex *lexical.Index) error {
	chunks := snapshot.AllChunks()
	if lexicalIndex == nil || lexicalIndex.DocumentCount() != len(chunks) {
		return fmt.Errorf("prepared lexical index does not match chunk count")
	}
	for documentIndex, chunk := range chunks {
		want := chunkLexicalKey(chunk, documentIndex)
		if lexicalIndex.Document(documentIndex).Key != want {
			return fmt.Errorf("prepared lexical index document %d does not match chunk %q", documentIndex, want)
		}
	}
	return nil
}

func chunkLexicalKey(chunk Chunk, documentIndex int) string {
	if chunk.ID != "" {
		return chunk.ID
	}
	return chunkID(chunk.Path, chunk.StartLine, chunk.EndLine, chunk.Text, documentIndex)
}
