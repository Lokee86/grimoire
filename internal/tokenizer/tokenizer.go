package tokenizer

import (
	"fmt"
	"sync"

	tiktoken "github.com/tiktoken-go/tokenizer"
)

const Name = "o200k_base"

var (
	codecOnce sync.Once
	codec     tiktoken.Codec
	codecErr  error
)

func Count(text string) (int, error) {
	codecOnce.Do(func() {
		codec, codecErr = tiktoken.Get(tiktoken.O200kBase)
		if codecErr != nil {
			codecErr = fmt.Errorf("initialize %s tokenizer: %w", Name, codecErr)
		}
	})
	if codecErr != nil {
		return 0, codecErr
	}

	count, err := codec.Count(text)
	if err != nil {
		return 0, fmt.Errorf("count %s tokens: %w", Name, err)
	}
	return count, nil
}
