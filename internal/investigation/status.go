package investigation

func statusFromManifest(current manifest) Status {
	status := Status{
		Version: current.Version, SessionID: current.SessionID, Snapshot: current.Snapshot,
		CreatedAt: current.CreatedAt, UpdatedAt: current.UpdatedAt, ClosedAt: current.ClosedAt,
		Responses: len(current.Responses),
	}
	for _, meta := range current.Evidence {
		switch meta.Kind {
		case "node":
			status.UniqueNodes++
		case "source":
			status.UniqueSourceRanges++
		case "path":
			status.UniqueGraphPaths++
		case "document":
			status.UniqueDocuments++
		case "question":
			status.UniqueQuestions++
		case "rejected_branch":
			status.UniqueRejected++
		case "accepted_branch":
			status.UniqueAccepted++
		}
	}
	return status
}
