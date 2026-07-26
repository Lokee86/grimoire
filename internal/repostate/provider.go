package repostate

func currentLexiconSnapshot(status Status) string {
	if status.Lexicon.Status != "current" {
		return ""
	}
	return status.Lexicon.Snapshot
}
