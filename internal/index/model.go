package index

const FormatVersion = 1

type Snapshot struct {
	Version int          `json:"version"`
	Files   []FileRecord `json:"files"`
}

type FileRecord struct {
	Path   string  `json:"path"`
	Hash   string  `json:"hash"`
	Size   int64   `json:"size"`
	Chunks []Chunk `json:"chunks"`
}

type Chunk struct {
	ID              string `json:"id"`
	Path            string `json:"path"`
	StartLine       int    `json:"start_line"`
	EndLine         int    `json:"end_line"`
	EstimatedTokens int    `json:"estimated_tokens"`
	Text            string `json:"text"`
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
