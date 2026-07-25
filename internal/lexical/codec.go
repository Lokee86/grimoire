package lexical

import (
	"encoding/json"
	"fmt"
)

const FormatVersion = 1

type encodedIndex struct {
	Version   int        `json:"version"`
	Documents []Document `json:"documents"`
}

func Encode(index *Index) ([]byte, error) {
	if index == nil {
		return nil, fmt.Errorf("lexical index is nil")
	}
	return json.Marshal(encodedIndex{Version: FormatVersion, Documents: index.documents})
}

func Decode(data []byte) (*Index, error) {
	var encoded encodedIndex
	if err := json.Unmarshal(data, &encoded); err != nil {
		return nil, fmt.Errorf("decode lexical index: %w", err)
	}
	if encoded.Version != FormatVersion {
		return nil, fmt.Errorf("unsupported lexical index version %d", encoded.Version)
	}
	index, err := New(encoded.Documents)
	if err != nil {
		return nil, fmt.Errorf("decode lexical index: %w", err)
	}
	return index, nil
}
