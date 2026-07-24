package retrieve

// Config controls deterministic lexical ranking behavior.
type Config struct {
	DeclarationAliasBonus float64
}

// DefaultConfig returns the production lexical ranking configuration.
func DefaultConfig() Config {
	return Config{DeclarationAliasBonus: 1}
}

// LegacyConfig preserves exact-token lexical ranking without declaration aliases.
func LegacyConfig() Config {
	return Config{}
}

func normalizedConfig(config Config) Config {
	if config.DeclarationAliasBonus < 0 {
		config.DeclarationAliasBonus = 0
	}
	return config
}
