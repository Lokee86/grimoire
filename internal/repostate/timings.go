package repostate

// PreparationTimings separates provider work from repository-state overhead.
type PreparationTimings struct {
	InitialInspectionMS int64 `json:"initial_inspection_ms"`
	LockWaitMS          int64 `json:"lock_wait_ms"`
	ReinspectionMS      int64 `json:"reinspection_ms"`
	LexiconMS           int64 `json:"lexicon_ms"`
	ArcanaMS            int64 `json:"arcana_ms"`
	SourceIndexMS       int64 `json:"source_index_ms"`
	DocumentationMS     int64 `json:"documentation_ms"`
	MarkerWriteMS       int64 `json:"marker_write_ms"`
	FinalVerificationMS int64 `json:"final_verification_ms"`
	TotalMS             int64 `json:"total_ms"`
}

func (timings *PreparationTimings) addAction(name string, elapsed int64) {
	switch name {
	case "refresh-lexicon":
		timings.LexiconMS += elapsed
	case "synchronize-arcana":
		timings.ArcanaMS += elapsed
	case "prepare-grimoire":
		timings.SourceIndexMS += elapsed
	case "prepare-knowledge":
		timings.DocumentationMS += elapsed
	}
}
