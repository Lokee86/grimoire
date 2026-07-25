package extraction

import "github.com/Lokee86/grimoire/internal/index"

// Request contains the query material available when retrieval candidates are
// converted into final source spans. FacetQueries lets the extractor focus a
// candidate on the decomposed query facet that caused it to be retrieved.
type Request struct {
	Query        string
	FacetQueries map[string]string
}

// DiscoveryRequest is the provider-neutral contract for a span discoverer.
// Language-aware discoverers can be inserted ahead of the generic line-window
// discoverer without changing candidate assembly or compilation.
type DiscoveryRequest struct {
	Chunk index.Chunk
	Query string
	Terms []string
}

// Span is an inclusive, chunk-relative line range.
type Span struct {
	StartLine int
	EndLine   int
	Reason    string
}

// Discoverer finds useful source spans inside one prepared index chunk.
type Discoverer interface {
	Discover(DiscoveryRequest) ([]Span, error)
}

// Config controls only the extraction boundary. Retrieval ranking and package
// assembly remain separate concerns.
type Config struct {
	MaxSpans         int
	MinChunkLines    int
	MinChunkTokens   int
	MinTokenSavings  int
	MaxRetainedRatio float64
}

func DefaultConfig() Config {
	return Config{
		MaxSpans:         2,
		MinChunkLines:    24,
		MinChunkTokens:   320,
		MinTokenSavings:  96,
		MaxRetainedRatio: 0.80,
	}
}

// Extractor owns candidate-to-span conversion. It uses the first discoverer
// that returns a usable result and otherwise preserves the original chunk.
type Extractor struct {
	config      Config
	discoverers []Discoverer
}

func New(config Config, discoverers ...Discoverer) Extractor {
	return Extractor{config: normalizedConfig(config), discoverers: append([]Discoverer(nil), discoverers...)}
}

func normalizedConfig(config Config) Config {
	defaults := DefaultConfig()
	if config.MaxSpans <= 0 {
		config.MaxSpans = defaults.MaxSpans
	}
	if config.MinChunkLines <= 0 {
		config.MinChunkLines = defaults.MinChunkLines
	}
	if config.MinChunkTokens <= 0 {
		config.MinChunkTokens = defaults.MinChunkTokens
	}
	if config.MinTokenSavings < 0 {
		config.MinTokenSavings = defaults.MinTokenSavings
	}
	if config.MaxRetainedRatio <= 0 || config.MaxRetainedRatio > 1 {
		config.MaxRetainedRatio = defaults.MaxRetainedRatio
	}
	return config
}
