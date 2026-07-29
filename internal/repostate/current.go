package repostate

func repositoryStateCurrent(status Status) bool {
	return status.Lexicon.Status == "current" &&
		status.Arcana.Status == "current" &&
		status.Grimoire.Status == "current" &&
		status.Knowledge.Status == "current"
}
