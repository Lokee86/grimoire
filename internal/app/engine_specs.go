package app

var lexiconEngine = engineSpec{
	name:        "lexicon",
	displayName: "Lexicon",
	commandEnv:  "GRIMOIRE_LEXICON_COMMAND",
	versionArgs: []string{"version"},
	help: `Usage: grimoire lexicon <command> [options]

Forward Lexicon language-analysis, adapter, snapshot, and diagnostic commands
through the Grimoire product entry point.

Commands:
  check       Resolve Lexicon and report its command and version
  init        Initialize language analysis for a repository
  scan        Incrementally analyze repository changes
  demon       Watch and continuously update analysis state
  rebuild     Rebuild analysis state
  export      Export an immutable snapshot
  gc          Collect unreachable analysis objects
  languages   Inspect or configure language adapters
  consumer    Manage snapshot consumers
  status      Inspect repository analysis state
  doctor      Run adapter and state diagnostics
  version     Print the Lexicon version
  help        Show this help

Except for check, help, and version aliases, arguments after "lexicon" are
passed directly to Lexicon. Set GRIMOIRE_LEXICON_COMMAND to override discovery.
`,
}

var arcanaEngine = engineSpec{
	name:        "arcana",
	displayName: "Arcana",
	commandEnv:  "GRIMOIRE_ARCANA_COMMAND",
	versionArgs: []string{"--version"},
	help: `Usage: grimoire arcana <command> [options]

Forward Arcana graph-state, structural-query, synchronization, and diagnostic
commands through the Grimoire product entry point.

Commands:
  check           Resolve Arcana and report its command and version
  sync            Synchronize graph state from a Lexicon snapshot
  query           Query exact nodes and relationships
  semantic-query  Search the optional semantic graph index
  vectorize       Build the optional semantic graph index
  protocol        Serve machine-readable graph queries over JSONL
  import-facts    Compile facts into a repository snapshot
  update-facts    Apply changed-file facts as a graph overlay
  benchmark       Compare overlay and packed-snapshot behavior
  version         Print the Arcana version
  help            Show this help

Except for check, help, and version aliases, arguments after "arcana" are
passed directly to Arcana. Set GRIMOIRE_ARCANA_COMMAND to override discovery.
`,
}
