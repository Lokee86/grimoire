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
	current, err := getCodec()
	if err != nil {
		return 0, err
	}
	count, err := current.Count(text)
	if err != nil {
		return 0, fmt.Errorf("count %s tokens: %w", Name, err)
	}
	return count, nil
}

func Encode(text string) ([]uint, error) {
	current, err := getCodec()
	if err != nil {
		return nil, err
	}
	tokens, _, err := current.Encode(text)
	if err != nil {
		return nil, fmt.Errorf("encode %s tokens: %w", Name, err)
	}
	return tokens, nil
}

func Decode(tokens []uint) (string, error) {
	current, err := getCodec()
	if err != nil {
		return "", err
	}
	text, err := current.Decode(tokens)
	if err != nil {
		return "", fmt.Errorf("decode %s tokens: %w", Name, err)
	}
	return text, nil
}

func getCodec() (tiktoken.Codec, error) {
	codecOnce.Do(func() {
		codec, codecErr = tiktoken.Get(tiktoken.O200kBase)
		if codecErr != nil {
			codecErr = fmt.Errorf("initialize %s tokenizer: %w", Name, codecErr)
		}
	})
	return codec, codecErr
}
