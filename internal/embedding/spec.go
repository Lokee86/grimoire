package embedding

import "strings"

const (
	ModelName        = "Qwen3-Embedding-0.6B"
	ModelRepository  = "Qwen/Qwen3-Embedding-0.6B-GGUF"
	ModelVariant     = "Q8_0"
	ModelReference   = ModelRepository + ":" + ModelVariant
	NativeDimensions = 1024
	Dimensions       = 512
	DefaultEndpoint  = "http://127.0.0.1:8080/v1"

	QueryInstruction = "Given a software development query, retrieve relevant source code and documentation from a repository"
)

func Identity() string {
	return "qwen3-embedding-0.6b-q8_0-512d"
}

func FormatQuery(query string) string {
	return "Instruct: " + QueryInstruction + "\nQuery:" + strings.TrimSpace(query)
}
