package compiler

// Config controls final exact-budget fitting behavior.
type Config struct {
	ProtectFacets        bool
	CompanionDepth       int
	ProtectRequiredLinks bool
}

// DefaultConfig protects required linked source spans, one source candidate
// per facet, and one additional same-file chunk that contributes new lexical
// evidence.
func DefaultConfig() Config {
	return Config{ProtectFacets: true, CompanionDepth: 1, ProtectRequiredLinks: true}
}

// LegacyConfig preserves rank-ordered final fitting.
func LegacyConfig() Config {
	return Config{}
}

func normalizedConfig(config Config) Config {
	if !config.ProtectFacets {
		config.CompanionDepth = 0
	}
	if config.CompanionDepth < 0 {
		config.CompanionDepth = 0
	}
	return config
}
