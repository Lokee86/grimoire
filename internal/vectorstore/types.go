package vectorstore

import "errors"

const ABIName = "grimoire_vector_ffi"

var (
	ErrUnavailable    = errors.New("Grimoire vector engine is unavailable")
	ErrBufferTooSmall = errors.New("vector engine output buffer is too small")
)

type Info struct {
	Model      string `json:"model"`
	Dimensions int    `json:"dimensions"`
	Count      int    `json:"count"`
}

type Hit struct {
	ID    string  `json:"id"`
	Score float32 `json:"score"`
	Index uint64  `json:"index"`
}
