package index

import "errors"

const FormatVersion = 2

var ErrIncompatibleIndex = errors.New("incompatible prepared index")

type Snapshot struct {
	Version   int
	Tokenizer string
	Files     []FileRecord

	baseRoot    string
	baseShards  map[string]string
	dirtyShards map[string]bool
}

type FileRecord struct {
	Path   string
	Hash   string
	Size   int64
	Chunks []Chunk
}

type Chunk struct {
	ID         string
	Path       string
	StartLine  int
	EndLine    int
	TokenCount int
	Text       string
}

func (snapshot Snapshot) AllChunks() []Chunk {
	count := 0
	for _, file := range snapshot.Files {
		count += len(file.Chunks)
	}

	chunks := make([]Chunk, 0, count)
	for _, file := range snapshot.Files {
		chunks = append(chunks, file.Chunks...)
	}
	return chunks
}
