package compiler

// Config controls final exact-budget fitting behavior.
type Config struct {
	ProtectFacets        bool
	FacetFileDepth       int
	CompanionDepth       int
	ProtectRequiredLinks bool
}

// DefaultConfig protects required linked source spans, one source candidate
// per facet, and one additional same-file chunk that contributes new lexical
// evidence.
func DefaultConfig() Config {
	return Config{
		ProtectFacets: true, FacetFileDepth: 2,
		CompanionDepth: 1, ProtectRequiredLinks: true,
	}
}

// LegacyConfig preserves rank-ordered final fitting.
func LegacyConfig() Config {
	return Config{}
}

func normalizedConfig(config Config) Config {
	if !config.ProtectFacets {
		config.FacetFileDepth = 0
		config.CompanionDepth = 0
	} else if config.FacetFileDepth <= 0 {
		config.FacetFileDepth = 1
	}
	if config.CompanionDepth < 0 {
		config.CompanionDepth = 0
	}
	return config
}
