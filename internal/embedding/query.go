package embedding

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Lokee86/grimoire/internal/tokenizer"
)

type QueryMode string

const (
	QueryModeFast    QueryMode = "fast"
	QueryModeFull    QueryMode = "full"
	QueryModeQuality QueryMode = "quality"

	DefaultQueryWindowTokens     = 16
	DefaultQueryBatchTokens      = 64
	DefaultQueryBatchConcurrency = 2
	DefaultQueryMaxTokens        = 0
)

type QueryOptions struct {
	Mode             QueryMode
	WindowTokens     int
	BatchTokens      int
	BatchConcurrency int
	MaxTokens        int
}

type QueryInput struct {
	Text  string
	Label string
}

func (client *Client) EmbedQueries(ctx context.Context, queries []string) ([][]float32, error) {
	if len(queries) == 0 {
		return nil, errors.New("no embedding queries supplied")
	}
	formatted := make([]string, len(queries))
	for index, query := range queries {
		if strings.TrimSpace(query) == "" {
			return nil, errors.New("embedding query is empty")
		}
		formatted[index] = FormatQuery(query)
	}
	return client.embed(ctx, formatted)
}

func DefaultQueryOptions() QueryOptions {
	return QueryOptions{
		Mode: QueryModeFast, WindowTokens: DefaultQueryWindowTokens,
		BatchTokens: DefaultQueryBatchTokens, BatchConcurrency: DefaultQueryBatchConcurrency,
		MaxTokens: DefaultQueryMaxTokens,
	}
}

func ParseQueryMode(value string) (QueryMode, error) {
	mode := QueryMode(strings.ToLower(strings.TrimSpace(value)))
	switch mode {
	case QueryModeFast, QueryModeFull, QueryModeQuality:
		return mode, nil
	default:
		return "", fmt.Errorf("unknown query embedding mode %q", value)
	}
}

func PlanQuery(query string, options QueryOptions) ([]QueryInput, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("embedding query is empty")
	}
	if err := options.Validate(); err != nil {
		return nil, err
	}
	tokens, err := tokenizer.Encode(query)
	if err != nil {
		return nil, err
	}
	if options.MaxTokens > 0 && len(tokens) > options.MaxTokens {
		tokens = tokens[:options.MaxTokens]
	}
	full, err := tokenizer.Decode(tokens)
	if err != nil {
		return nil, err
	}
	windows, err := queryWindows(tokens, options.WindowTokens)
	if err != nil {
		return nil, err
	}

	switch options.Mode {
	case QueryModeFast:
		return windows, nil
	case QueryModeFull:
		return []QueryInput{{Text: full, Label: "full query"}}, nil
	case QueryModeQuality:
		if len(windows) == 1 && windows[0].Text == full {
			return []QueryInput{{Text: full, Label: "full query"}}, nil
		}
		inputs := make([]QueryInput, 0, len(windows)+1)
		inputs = append(inputs, QueryInput{Text: full, Label: "full query"})
		return append(inputs, windows...), nil
	default:
		panic("validated query mode became invalid")
	}
}

func (options QueryOptions) Validate() error {
	if _, err := ParseQueryMode(string(options.Mode)); err != nil {
		return err
	}
	if options.WindowTokens <= 0 || options.BatchTokens <= 0 || options.BatchConcurrency <= 0 {
		return errors.New("positive query window, batch, and concurrency values are required")
	}
	if options.MaxTokens < 0 {
		return errors.New("query maximum tokens cannot be negative")
	}
	if options.WindowTokens > options.BatchTokens {
		return errors.New("query window tokens cannot exceed query batch tokens")
	}
	return nil
}

func queryWindows(tokens []uint, windowSize int) ([]QueryInput, error) {
	count := (len(tokens) + windowSize - 1) / windowSize
	inputs := make([]QueryInput, 0, count)
	for start := 0; start < len(tokens); start += windowSize {
		end := min(start+windowSize, len(tokens))
		text, err := tokenizer.Decode(tokens[start:end])
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, QueryInput{
			Text: text, Label: fmt.Sprintf("split window %d/%d", len(inputs)+1, count),
		})
	}
	return inputs, nil
}
